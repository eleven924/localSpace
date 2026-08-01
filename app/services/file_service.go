package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"LocalSpace/app/agents"
	"LocalSpace/app/models"
	"LocalSpace/app/repositories"
	"LocalSpace/app/utils"
)

type fileMetadataUpdater interface {
	FindByID(id uint) (*models.File, error)
	UpdateMetadata(id uint, tags []string, description string) error
}

// FileService handles file operations.
type FileService struct {
	fileRepo         *repositories.FileRepository
	metadataRepo     fileMetadataUpdater
	storageService   *StorageService
	aiService        *AIService
	agentService     *AgentService
	thumbnailService *ThumbnailService
	thumbnailMu      sync.Mutex
	thumbnailPending map[uint]struct{}
}

// ImportFileRequest represents a file import request.
type ImportFileRequest struct {
	FilePath       string   // source path
	FileName       string   // target display name
	Description    string   // user-provided description
	Tags           []string // user-provided tags
	Keywords       string   // user-provided keywords
	CollectionName string   // user-provided collection / series / project name
}

type BatchImportPlan struct {
	SourcePath     string
	FileName       string
	OriginalName   string
	CollectionName string
	FileType       string
	FileSubType    string
	FileSize       int64
	Checksum       string
	MasterID       uint
	FinalPath      string
	TempPath       string
}

type BatchImportMetadataRequest struct {
	FileName                     string
	FileType                     string
	SharedTags                   []string
	SharedDescription            string
	EnableAIGeneratedTags        bool
	EnableAIGeneratedDescription bool
}

// FileFilter represents filters for file queries.
type FileFilter struct {
	Page      int
	PageSize  int
	FileType  string
	SortBy    string
	SortOrder string
}

// NewFileService creates a new FileService.
func NewFileService(
	fileRepo *repositories.FileRepository,
	storageService *StorageService,
	aiService *AIService,
	thumbnailService *ThumbnailService,
) *FileService {
	return &FileService{
		fileRepo:         fileRepo,
		metadataRepo:     fileRepo,
		storageService:   storageService,
		aiService:        aiService,
		thumbnailService: thumbnailService,
		thumbnailPending: make(map[uint]struct{}),
	}
}

type stagedImport struct {
	SourcePath     string
	TempPath       string
	FinalPath      string
	Checksum       string
	FileSize       int64
	FileType       string
	FileSubType    string
	MasterID       uint
	CollectionName string
	FileName       string
	OriginalName   string
	Metadata       models.Metadata
}

const batchImportTempDir = ".localspace-temp"
const batchImportTempSuffix = ".localspace-importing"

func generateRandomSuffix() string {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		// fallback to timestamp if crypto/rand fails
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

func generateTempPath(masterPath, fileName, jobID string, itemIndex int) string {
	timestamp := time.Now().UnixNano()
	base := fmt.Sprintf("%s-%d-%s", fileName, timestamp, generateRandomSuffix())
	if jobID != "" {
		return filepath.Join(masterPath, batchImportTempDir, "batch-"+jobID, fmt.Sprintf("%d", itemIndex), base+batchImportTempSuffix)
	}
	return filepath.Join(masterPath, batchImportTempDir, base+batchImportTempSuffix)
}

func (s *FileService) prepareStagedImport(sourcePath, fileName, collectionName string, jobID string, itemIndex int, onProgress func(copied int64) error) (*stagedImport, error) {
	if sourcePath == "" {
		return nil, fmt.Errorf("source path cannot be empty")
	}
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("source file does not exist: %s", sourcePath)
	}

	fileInfo, err := os.Stat(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	extension := filepath.Ext(sourcePath)
	fileType, err := parseFileType(extension)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file type: %w", err)
	}

	masters, err := s.storageService.GetMasterDirectories()
	if err != nil {
		return nil, fmt.Errorf("failed to get master directories: %w", err)
	}
	if len(masters) == 0 {
		return nil, fmt.Errorf("no master directory configured. Please add a master directory in Settings > Storage Directories before importing files.")
	}

	defaultMaster := selectDefaultMasterDirectory(masters)
	if defaultMaster == nil {
		defaultMaster = &masters[0]
	}

	hasSpace, err := s.storageService.CheckMasterStorageSpace(defaultMaster.ID, fileInfo.Size())
	if err != nil {
		return nil, fmt.Errorf("failed to check storage space: %w", err)
	}
	if !hasSpace {
		return nil, fmt.Errorf("not enough storage space for this file")
	}

	layoutConfig, err := s.storageService.GetStorageLayoutConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get storage layout config: %w", err)
	}

	trimmedCollectionName := strings.TrimSpace(collectionName)
	effectiveCollectionName := trimmedCollectionName
	if layoutConfig.Strategy == "type_collection" && effectiveCollectionName == "" {
		effectiveCollectionName = layoutConfig.UnsortedFolderName
	}

	targetFileName := strings.TrimSpace(fileName)
	if targetFileName == "" {
		targetFileName = filepath.Base(sourcePath)
	}

	finalPath, err := s.storageService.GetStoragePathForFileWithMaster(defaultMaster.ID, fileType, effectiveCollectionName, targetFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage path: %w", err)
	}
	if err := s.storageService.EnsureStorageDirExists(filepath.Dir(finalPath)); err != nil {
		return nil, fmt.Errorf("failed to create destination directory: %w", err)
	}

	checksum, err := utils.CalculateFileChecksum(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate file checksum: %w", err)
	}

	isDuplicate, duplicateID, err := s.fileRepo.CheckDuplicateByChecksum(checksum)
	if err != nil {
		return nil, fmt.Errorf("failed to check for duplicate files: %w", err)
	}
	if isDuplicate {
		duplicateFile, findErr := s.fileRepo.FindByID(duplicateID)
		if findErr == nil {
			return nil, fmt.Errorf("duplicate file detected. A file with identical content already exists: %s (ID: %d)", duplicateFile.FileName, duplicateID)
		}
		return nil, fmt.Errorf("duplicate file detected. A file with identical content already exists (ID: %d)", duplicateID)
	}

	exists, err := s.fileRepo.ExistsByPath(finalPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check if file exists: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("file already exists at destination: %s", finalPath)
	}

	tempPath := generateTempPath(defaultMaster.Path, targetFileName, jobID, itemIndex)
	if err := s.storageService.EnsureStorageDirExists(filepath.Dir(tempPath)); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	if err := copyFileWithProgress(sourcePath, tempPath, onProgress); err != nil {
		_ = os.Remove(tempPath)
		return nil, fmt.Errorf("failed to copy file to staging: %w", err)
	}

	return &stagedImport{
		SourcePath:     sourcePath,
		TempPath:       tempPath,
		FinalPath:      finalPath,
		Checksum:       checksum,
		FileSize:       fileInfo.Size(),
		FileType:       fileType,
		FileSubType:    strings.TrimPrefix(extension, "."),
		MasterID:       defaultMaster.ID,
		CollectionName: effectiveCollectionName,
		FileName:       targetFileName,
		OriginalName:   filepath.Base(sourcePath),
	}, nil
}

func (s *FileService) commitStagedImport(staged *stagedImport, tags []string, description string, metadata models.Metadata) (*models.File, error) {
	if staged == nil {
		return nil, fmt.Errorf("staged import is nil")
	}

	file := &models.File{
		FileName:       staged.FileName,
		OriginalName:   staged.OriginalName,
		CollectionName: staged.CollectionName,
		FilePath:       staged.FinalPath,
		FileType:       staged.FileType,
		FileSubType:    staged.FileSubType,
		FileSize:       staged.FileSize,
		Tags:           s.normalizeMetadataTags(tags),
		Description:    s.normalizeMetadataDescription(description),
		Metadata:       metadata,
		Thumbnail:      "",
		Checksum:       staged.Checksum,
		IsDeleted:      false,
		DeletedAt:      "",
	}

	if err := s.fileRepo.Create(file); err != nil {
		_ = os.Remove(staged.TempPath)
		return nil, fmt.Errorf("failed to create file record: %w", err)
	}

	if err := os.Rename(staged.TempPath, staged.FinalPath); err != nil {
		_ = s.fileRepo.Delete(file.ID)
		_ = os.Remove(staged.TempPath)
		return nil, fmt.Errorf("failed to commit imported file: %w", err)
	}

	if err := os.Remove(staged.SourcePath); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Warning: imported file committed but failed to remove source %s: %v\n", staged.SourcePath, err)
	}

	if supportsGeneratedThumbnail(staged.FileType) {
		s.generateThumbnailForFile(file.ID, staged.FinalPath, staged.FileType)
	}

	if err := s.storageService.UpdateMasterDirectorySize(staged.MasterID, staged.FileType, staged.FileSize); err != nil {
		fmt.Printf("Failed to update storage size: %v\n", err)
	}

	return file, nil
}

func (s *FileService) cleanupStagedImport(staged *stagedImport) error {
	if staged == nil {
		return nil
	}
	if err := os.Remove(staged.TempPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove staged temp file %s: %w", staged.TempPath, err)
	}
	return nil
}

// SetAgentService sets the agent service after initialization.
func (s *FileService) SetAgentService(agentService *AgentService) {
	s.agentService = agentService
}

// ImportFile imports a file into LocalSpace.
func (s *FileService) ImportFile(req ImportFileRequest) error {
	if req.FilePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}
	if _, err := os.Stat(req.FilePath); os.IsNotExist(err) {
		return fmt.Errorf("source file does not exist: %s", req.FilePath)
	}

	fileName := strings.TrimSpace(req.FileName)
	if fileName == "" {
		fileName = filepath.Base(req.FilePath)
	}

	staged, err := s.prepareStagedImport(req.FilePath, fileName, req.CollectionName, "", 0, nil)
	if err != nil {
		return err
	}
	defer func() {
		if staged != nil {
			_ = s.cleanupStagedImport(staged)
		}
	}()

	metadata, err := s.ExtractMetadata(staged.TempPath, staged.FileType)
	if err != nil {
		fmt.Printf("Warning: Failed to extract metadata: %v\n", err)
	}

	formMetadata := s.resolveMetadataForImport(&req)

	_, err = s.commitStagedImport(staged, formMetadata.Tags, formMetadata.Description, metadata)
	if err != nil {
		return err
	}

	staged = nil
	return nil
}

type importMetadata struct {
	Tags        []string
	Description string
}

func (s *FileService) resolveMetadataForImport(req *ImportFileRequest) importMetadata {
	result := importMetadata{Tags: []string{}}
	if req != nil {
		result.Tags = append(result.Tags, req.Tags...)
		result.Description = strings.TrimSpace(req.Description)
	}
	return result
}

func (s *FileService) PrepareBatchImport(sourcePath, displayName, collectionName string, jobID uint, itemIndex int) (*BatchImportPlan, error) {
	fileName := strings.TrimSpace(displayName)
	if fileName == "" {
		fileName = filepath.Base(sourcePath)
	}

	staged, err := s.prepareStagedImport(sourcePath, fileName, collectionName, fmt.Sprintf("%d", jobID), itemIndex, nil)
	if err != nil {
		return nil, err
	}

	return &BatchImportPlan{
		SourcePath:     staged.SourcePath,
		FileName:       staged.FileName,
		OriginalName:   staged.OriginalName,
		CollectionName: staged.CollectionName,
		FileType:       staged.FileType,
		FileSubType:    staged.FileSubType,
		FileSize:       staged.FileSize,
		Checksum:       staged.Checksum,
		MasterID:       staged.MasterID,
		FinalPath:      staged.FinalPath,
		TempPath:       staged.TempPath,
	}, nil
}

func (s *FileService) CopyFileToTemp(plan *BatchImportPlan, onProgress func(copied int64) error) error {
	if plan == nil {
		return fmt.Errorf("batch import plan is nil")
	}

	if err := copyFileWithProgress(plan.SourcePath, plan.TempPath, onProgress); err != nil {
		_ = os.Remove(plan.TempPath)
		return fmt.Errorf("failed to copy file to staging: %w", err)
	}

	info, err := os.Stat(plan.TempPath)
	if err != nil {
		_ = os.Remove(plan.TempPath)
		return fmt.Errorf("failed to stat staged file: %w", err)
	}
	if info.Size() != plan.FileSize {
		_ = os.Remove(plan.TempPath)
		return fmt.Errorf("copied file size mismatch: expected %d bytes, got %d bytes", plan.FileSize, info.Size())
	}

	return nil
}

func (s *FileService) FinalizeBatchImport(plan *BatchImportPlan, tags []string, description string, metadata models.Metadata) (*models.File, error) {
	if plan == nil {
		return nil, fmt.Errorf("batch import plan is nil")
	}

	staged := &stagedImport{
		SourcePath:     plan.SourcePath,
		TempPath:       plan.TempPath,
		FinalPath:      plan.FinalPath,
		Checksum:       plan.Checksum,
		FileSize:       plan.FileSize,
		FileType:       plan.FileType,
		FileSubType:    plan.FileSubType,
		MasterID:       plan.MasterID,
		CollectionName: plan.CollectionName,
		FileName:       plan.FileName,
		OriginalName:   plan.OriginalName,
	}

	file, err := s.commitStagedImport(staged, tags, description, metadata)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (s *FileService) CleanupBatchImportArtifacts(tempPath, finalPath string) error {
	// Only delete the explicitly known temp path. finalPath is intentionally
	// ignored to avoid deleting real files that happen to match a suffix.
	path := strings.TrimSpace(tempPath)
	if path == "" {
		return nil
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove temp artifact %s: %w", path, err)
	}
	return nil
}

func (s *FileService) ResolveBatchImportMetadata(req BatchImportMetadataRequest) ([]string, string, error) {
	tags := append([]string{}, req.SharedTags...)
	description := strings.TrimSpace(req.SharedDescription)

	needsAI := req.EnableAIGeneratedTags || req.EnableAIGeneratedDescription
	if needsAI {
		analysisTags, analysisDescription, err := s.generateImportMetadata(req.FileName, req.FileType)
		if err != nil {
			return nil, "", err
		}
		if req.EnableAIGeneratedTags {
			tags = append(tags, analysisTags...)
		}
		if req.EnableAIGeneratedDescription && strings.TrimSpace(analysisDescription) != "" {
			description = analysisDescription
		}
	}

	return s.normalizeMetadataTags(tags), s.normalizeMetadataDescription(description), nil
}

func (s *FileService) generateImportMetadata(fileName, fileType string) ([]string, string, error) {
	if s.agentService != nil {
		analysis, err := s.agentService.AnalyzeMetadata(context.Background(), &agents.MetadataGenerationInput{
			FileName: fileName,
			FileType: fileType,
		})
		if err == nil && analysis != nil {
			return analysis.Tags, analysis.Description, nil
		}
	}

	var tags []string
	var description string
	if s.aiService != nil {
		generatedTags, err := s.aiService.GenerateTags(fileName, fileType)
		if err != nil {
			return nil, "", err
		}
		generatedDescription, err := s.aiService.GenerateDescription(fileName, fileType)
		if err != nil {
			return nil, "", err
		}
		tags = generatedTags
		description = generatedDescription
	}

	return tags, description, nil
}

func selectDefaultMasterDirectory(masters []models.StorageDir) *models.StorageDir {
	if len(masters) == 0 {
		return nil
	}
	for _, master := range masters {
		if master.IsDefault {
			m := master
			return &m
		}
	}
	master := masters[0]
	return &master
}

func (s *FileService) normalizeMetadataTags(tags []string) []string {
	seen := make(map[string]struct{}, len(tags))
	normalized := make([]string, 0, len(tags))

	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		normalized = append(normalized, trimmed)
	}

	return normalized
}

func (s *FileService) normalizeMetadataDescription(description string) string {
	return strings.TrimSpace(description)
}

func (s *FileService) UpdateFileMetadata(id uint, tags []string, description string) error {
	if id == 0 {
		return fmt.Errorf("file id cannot be empty")
	}
	if s.metadataRepo == nil {
		return fmt.Errorf("file repository not initialized")
	}

	if _, err := s.metadataRepo.FindByID(id); err != nil {
		return fmt.Errorf("failed to get file: %w", err)
	}

	normalizedTags := s.normalizeMetadataTags(tags)
	normalizedDescription := s.normalizeMetadataDescription(description)

	if err := s.metadataRepo.UpdateMetadata(id, normalizedTags, normalizedDescription); err != nil {
		return fmt.Errorf("failed to update file metadata: %w", err)
	}

	return nil
}

// ListFiles returns a list of files.
func (s *FileService) ListFiles(filter FileFilter) ([]*models.File, error) {
	files, err := s.fileRepo.List(repositories.FileFilter{
		Page:      filter.Page,
		PageSize:  filter.PageSize,
		FileType:  filter.FileType,
		SortBy:    filter.SortBy,
		SortOrder: filter.SortOrder,
	})
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		file.FileType = resolveDisplayFileType(file.FilePath, file.FileType)

		if !supportsGeneratedThumbnail(file.FileType) {
			file.Thumbnail = ""
			continue
		}

		if !s.thumbnailIsCurrent(file) {
			file.Thumbnail = ""
			s.scheduleThumbnailGeneration(file.ID, file.FilePath, file.FileType)
			continue
		}

		if thumbnailDataURI, err := s.thumbnailDataURI(file.Thumbnail); err == nil {
			file.Thumbnail = thumbnailDataURI
		}
	}

	return files, nil
}

// SearchFiles searches for files.
func (s *FileService) SearchFiles(query string) ([]*models.File, error) {
	files, err := s.fileRepo.Search(query)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		file.FileType = resolveDisplayFileType(file.FilePath, file.FileType)

		if !supportsGeneratedThumbnail(file.FileType) {
			file.Thumbnail = ""
			continue
		}

		if !s.thumbnailIsCurrent(file) {
			file.Thumbnail = ""
			s.scheduleThumbnailGeneration(file.ID, file.FilePath, file.FileType)
			continue
		}

		if thumbnailDataURI, err := s.thumbnailDataURI(file.Thumbnail); err == nil {
			file.Thumbnail = thumbnailDataURI
		}
	}

	return files, nil
}

// GetFile returns a single file by ID.
func (s *FileService) GetFile(id uint) (*models.File, error) {
	file, err := s.fileRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	file.FileType = resolveDisplayFileType(file.FilePath, file.FileType)

	if !supportsGeneratedThumbnail(file.FileType) {
		file.Thumbnail = ""
		return file, nil
	}

	if !s.thumbnailIsCurrent(file) {
		file.Thumbnail = ""
		return file, nil
	}

	if thumbnailDataURI, err := s.thumbnailDataURI(file.Thumbnail); err == nil {
		file.Thumbnail = thumbnailDataURI
	}

	return file, nil
}

// GetThumbnail returns the thumbnail path for a file.
func (s *FileService) GetThumbnail(id uint) (string, error) {
	file, err := s.fileRepo.FindByID(id)
	if err != nil {
		return "", fmt.Errorf("failed to get file: %w", err)
	}

	resolvedType := resolveDisplayFileType(file.FilePath, file.FileType)
	if !supportsGeneratedThumbnail(resolvedType) {
		return "", fmt.Errorf("thumbnails are only supported for image and video files")
	}

	if file.Thumbnail == "" || !s.thumbnailIsCurrent(file) {
		thumbnailPath, err := s.thumbnailService.GetThumbnail(file.FilePath, resolvedType, file.ID)
		if err != nil {
			return "", fmt.Errorf("failed to get thumbnail: %w", err)
		}

		if err := s.fileRepo.UpdateThumbnail(id, thumbnailPath); err != nil {
			fmt.Printf("Warning: Failed to update thumbnail path: %v\n", err)
		}

		thumbnailDataURI, err := s.thumbnailDataURI(thumbnailPath)
		if err != nil {
			return "", fmt.Errorf("failed to read thumbnail: %w", err)
		}
		return thumbnailDataURI, nil
	}

	thumbnailDataURI, err := s.thumbnailDataURI(file.Thumbnail)
	if err != nil {
		return "", fmt.Errorf("failed to read thumbnail: %w", err)
	}
	return thumbnailDataURI, nil
}

// DeleteFile deletes a file.
func (s *FileService) DeleteFile(id uint) error {
	file, err := s.fileRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("failed to get file: %w", err)
	}

	resolvedType := resolveDisplayFileType(file.FilePath, file.FileType)

	if err := s.fileRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete file record: %w", err)
	}

	if err := os.Remove(file.FilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	if file.Thumbnail != "" {
		_ = s.thumbnailService.RemoveThumbnail(id, resolvedType)
	}

	if err := s.storageService.UpdateStorageSize(file.FileType, -file.FileSize); err != nil {
		fmt.Printf("Failed to update storage size: %v\n", err)
	}

	return nil
}

// RenameFile renames a file.
func (s *FileService) RenameFile(id uint, newName string) error {
	file, err := s.fileRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("failed to get file: %w", err)
	}

	if newName == "" {
		return fmt.Errorf("new file name cannot be empty")
	}
	if newName == file.FileName {
		return nil
	}

	ext := filepath.Ext(file.FilePath)
	if ext != "" && !strings.HasSuffix(newName, ext) {
		newName = newName + ext
	}

	dir := filepath.Dir(file.FilePath)
	newPath := filepath.Join(dir, newName)
	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("a file with this name already exists: %s", newName)
	}

	originalPath := file.FilePath
	originalName := file.FileName
	file.FileName = newName
	file.FilePath = newPath
	file.ModifiedAt = time.Now().Format(time.RFC3339)

	if err := s.fileRepo.Update(file); err != nil {
		file.FileName = originalName
		file.FilePath = originalPath
		return fmt.Errorf("failed to update file record: %w", err)
	}

	if err := os.Rename(originalPath, newPath); err != nil {
		file.FileName = originalName
		file.FilePath = originalPath
		file.ModifiedAt = time.Now().Format(time.RFC3339)
		if rollbackErr := s.fileRepo.Update(file); rollbackErr != nil {
			fmt.Printf("Critical: failed to rollback file record after rename failure: %v\n", rollbackErr)
		}
		return fmt.Errorf("failed to rename file: %w", err)
	}

	return nil
}

// DeleteFilesByPath deletes all files with a specific path prefix.
func (s *FileService) DeleteFilesByPath(pathPrefix string) (int64, error) {
	filter := repositories.FileFilter{
		Page:     1,
		PageSize: 10000,
	}

	files, err := s.fileRepo.List(filter)
	if err != nil {
		return 0, fmt.Errorf("failed to list files: %w", err)
	}

	deletedCount := int64(0)
	for _, file := range files {
		if strings.HasPrefix(file.FilePath, pathPrefix) {
			if file.Thumbnail != "" {
				_ = s.thumbnailService.RemoveThumbnail(file.ID, resolveDisplayFileType(file.FilePath, file.FileType))
			}
			if err := s.storageService.UpdateStorageSize(file.FileType, -file.FileSize); err != nil {
				fmt.Printf("Failed to update storage size: %v\n", err)
			}
			deletedCount++
		}
	}

	rowsAffected, err := s.fileRepo.DeleteByPath(pathPrefix)
	if err != nil {
		return deletedCount, fmt.Errorf("failed to delete files by path: %w", err)
	}

	return rowsAffected, nil
}

// ExtractMetadata extracts metadata from a file.
func (s *FileService) ExtractMetadata(filePath, fileType string) (models.Metadata, error) {
	var metadata models.Metadata

	fileInfo, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return metadata, fmt.Errorf("file does not exist: %s", filePath)
	}
	if err != nil {
		return metadata, fmt.Errorf("failed to get file info: %w", err)
	}

	metadata.Size = fileInfo.Size()

	switch fileType {
	case "image", "photo":
		imageMetadata, err := s.extractImageMetadata(filePath)
		if err == nil {
			imageMetadata.Size = metadata.Size
			return imageMetadata, nil
		}
		return metadata, nil
	case "video":
		videoMetadata, err := s.extractVideoMetadata(filePath)
		if err == nil {
			videoMetadata.Size = metadata.Size
			return videoMetadata, nil
		}
		return metadata, nil
	case "document", "pdf", "docx", "doc", "xlsx", "xls", "pptx", "ppt", "txt", "md":
		documentMetadata, err := s.extractDocumentMetadata(filePath, fileType)
		if err == nil {
			documentMetadata.Size = metadata.Size
			return documentMetadata, nil
		}
		return metadata, nil
	default:
		return metadata, nil
	}
}

func (s *FileService) extractImageMetadata(filePath string) (models.Metadata, error) {
	imgMeta, err := utils.ExtractImageMetadata(filePath)
	if err != nil {
		return models.Metadata{}, fmt.Errorf("failed to extract image metadata: %w", err)
	}

	return models.Metadata{
		Width:  imgMeta.Width,
		Height: imgMeta.Height,
	}, nil
}

func (s *FileService) extractVideoMetadata(filePath string) (models.Metadata, error) {
	vidMeta, err := utils.ExtractVideoMetadata(filePath)
	if err != nil {
		return models.Metadata{}, fmt.Errorf("failed to extract video metadata: %w", err)
	}

	return models.Metadata{
		Width:    vidMeta.Width,
		Height:   vidMeta.Height,
		Duration: int(vidMeta.Duration),
	}, nil
}

func (s *FileService) extractDocumentMetadata(filePath, fileType string) (models.Metadata, error) {
	docMeta, err := utils.ExtractDocumentMetadata(filePath, fileType)
	if err != nil {
		return models.Metadata{}, fmt.Errorf("failed to extract document metadata: %w", err)
	}

	return models.Metadata{
		PageCount: docMeta.PageCount,
		Author:    docMeta.Author,
		Title:     docMeta.Title,
	}, nil
}

func (s *FileService) generateThumbnailForFile(fileID uint, filePath, fileType string) {
	if !supportsGeneratedThumbnail(fileType) {
		return
	}

	thumbnailPath, err := s.thumbnailService.GetThumbnail(filePath, fileType, fileID)
	if err != nil {
		fmt.Printf("Warning: Failed to generate thumbnail for file %d: %v\n", fileID, err)
		return
	}

	if err := s.fileRepo.UpdateThumbnail(fileID, thumbnailPath); err != nil {
		fmt.Printf("Warning: Failed to update thumbnail path for file %d: %v\n", fileID, err)
	}
}

func (s *FileService) scheduleThumbnailGeneration(fileID uint, filePath, fileType string) {
	if !supportsGeneratedThumbnail(fileType) {
		return
	}

	s.thumbnailMu.Lock()
	if s.thumbnailPending == nil {
		s.thumbnailPending = make(map[uint]struct{})
	}
	if _, pending := s.thumbnailPending[fileID]; pending {
		s.thumbnailMu.Unlock()
		return
	}
	s.thumbnailPending[fileID] = struct{}{}
	s.thumbnailMu.Unlock()

	go func() {
		defer func() {
			s.thumbnailMu.Lock()
			delete(s.thumbnailPending, fileID)
			s.thumbnailMu.Unlock()
		}()

		s.generateThumbnailForFile(fileID, filePath, fileType)
	}()
}

func (s *FileService) thumbnailExists(thumbnailPath string) bool {
	if thumbnailPath == "" {
		return false
	}
	if filepath.Ext(thumbnailPath) == "" {
		return false
	}
	_, err := os.Stat(thumbnailPath)
	return err == nil
}

func (s *FileService) thumbnailIsCurrent(file *models.File) bool {
	if file == nil || !supportsGeneratedThumbnail(file.FileType) || !s.thumbnailExists(file.Thumbnail) {
		return false
	}
	return filepath.Base(file.Thumbnail) == s.thumbnailService.generateCacheKey(file.ID, file.FileType)
}

func (s *FileService) thumbnailDataURI(thumbnailPath string) (string, error) {
	data, err := os.ReadFile(thumbnailPath)
	if err != nil {
		return "", err
	}

	mimeType := "image/png"
	switch strings.ToLower(filepath.Ext(thumbnailPath)) {
	case ".jpg", ".jpeg":
		mimeType = "image/jpeg"
	case ".gif":
		mimeType = "image/gif"
	case ".webp":
		mimeType = "image/webp"
	}

	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(data)), nil
}

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

func copyFileWithProgress(src, dst string, onProgress func(copied int64) error) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	buffer := make([]byte, 1024*1024)
	var copied int64
	for {
		readBytes, readErr := source.Read(buffer)
		if readBytes > 0 {
			written, writeErr := destination.Write(buffer[:readBytes])
			if writeErr != nil {
				return writeErr
			}
			copied += int64(written)
			if onProgress != nil {
				if err := onProgress(copied); err != nil {
					return err
				}
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}
	return nil
}

func parseFileType(extension string) (string, error) {
	ext := strings.ToLower(extension)

	typeMap := map[string]string{
		".mp4":  "video",
		".avi":  "video",
		".mkv":  "video",
		".mov":  "video",
		".wmv":  "video",
		".flv":  "video",
		".webm": "video",
		".m4v":  "video",

		".pdf":  "document",
		".doc":  "document",
		".docx": "document",
		".xls":  "document",
		".xlsx": "document",
		".ppt":  "document",
		".pptx": "document",
		".txt":  "document",
		".md":   "document",
		".rtf":  "document",

		".mp3":  "music",
		".wav":  "music",
		".flac": "music",
		".aac":  "music",
		".ogg":  "music",
		".m4a":  "music",
		".wma":  "music",

		".zip": "archive",
		".rar": "archive",
		".7z":  "archive",
		".tar": "archive",
		".gz":  "archive",

		".exe":  "installer",
		".app":  "installer",
		".msi":  "installer",
		".ipa":  "installer",
		".pkg":  "installer",
		".deb":  "installer",
		".rpm":  "installer",
		".apk":  "installer",
		".dmg":  "installer",
		".iso":  "installer",
		".img":  "installer",
		".vdi":  "installer",
		".vmdk": "installer",

		".jpg":  "image",
		".jpeg": "image",
		".png":  "image",
		".gif":  "image",
		".bmp":  "image",
		".webp": "image",
		".svg":  "image",
		".ico":  "image",
	}

	if fileType, ok := typeMap[ext]; ok {
		return fileType, nil
	}

	return "other", nil
}

// CheckDuplicate checks if a file is a duplicate by calculating its checksum.
func (s *FileService) CheckDuplicate(filePath string) (bool, uint, string, error) {
	if filePath == "" {
		return false, 0, "", fmt.Errorf("file path cannot be empty")
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false, 0, "", fmt.Errorf("source file does not exist: %s", filePath)
	}

	checksum, err := utils.CalculateFileChecksum(filePath)
	if err != nil {
		return false, 0, "", fmt.Errorf("failed to calculate file checksum: %w", err)
	}

	isDuplicate, duplicateID, err := s.fileRepo.CheckDuplicateByChecksum(checksum)
	if err != nil {
		return false, 0, "", fmt.Errorf("failed to check for duplicate: %w", err)
	}

	var duplicateName string
	if isDuplicate {
		duplicateFile, err := s.fileRepo.FindByID(duplicateID)
		if err == nil {
			duplicateName = duplicateFile.FileName
		} else {
			duplicateName = "Unknown"
		}
	}

	return isDuplicate, duplicateID, duplicateName, nil
}

// GetDuplicateFiles finds all duplicate files based on checksums.
func (s *FileService) GetDuplicateFiles() (map[string][]*models.File, error) {
	allFiles, err := s.fileRepo.List(repositories.FileFilter{
		Page:     1,
		PageSize: 10000,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get all files: %w", err)
	}

	checksumMap := make(map[string][]*models.File)
	for _, file := range allFiles {
		if file.Checksum != "" && !file.IsDeleted {
			checksumMap[file.Checksum] = append(checksumMap[file.Checksum], file)
		}
	}

	duplicates := make(map[string][]*models.File)
	for checksum, files := range checksumMap {
		if len(files) > 1 {
			duplicates[checksum] = files
		}
	}

	return duplicates, nil
}

// RefreshFile checks and updates file information if the file has been modified.
func (s *FileService) RefreshFile(id uint) (*models.File, error) {
	file, err := s.fileRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	fileInfo, err := os.Stat(file.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			if file.Checksum == "" {
				file.FileType = resolveDisplayFileType(file.FilePath, file.FileType)
				return file, nil
			}

			newPath, err := s.findRenamedFile(file.FilePath, file.Checksum)
			if err != nil {
				file.FileType = resolveDisplayFileType(file.FilePath, file.FileType)
				return file, nil
			}

			file.FilePath = newPath
			file.FileName = filepath.Base(newPath)

			if err := s.fileRepo.Update(file); err != nil {
				fmt.Printf("Warning: Failed to update file record with new path: %v\n", err)
			}

			fileInfo, err = os.Stat(newPath)
			if err != nil {
				return file, nil
			}

			file.FileSize = fileInfo.Size()
			file.ModifiedAt = fileInfo.ModTime().Format(time.RFC3339)

			if err := s.fileRepo.Update(file); err != nil {
				fmt.Printf("Warning: Failed to update file record: %v\n", err)
			}

			file.FileType = resolveDisplayFileType(file.FilePath, file.FileType)
			return file, nil
		}
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	currentModTime := fileInfo.ModTime()
	storedModTime, err := time.Parse(time.RFC3339, file.ModifiedAt)
	if err != nil {
		storedModTime = time.Time{}
	}

	updated := false
	originalSize := file.FileSize
	newName := file.FileName
	newSize := file.FileSize

	if fileInfo.Size() != file.FileSize {
		newSize = fileInfo.Size()
		updated = true
	}

	currentFileName := filepath.Base(file.FilePath)
	if currentFileName != file.FileName {
		newName = currentFileName
		updated = true
	}

	if currentModTime.After(storedModTime) {
		updated = true
	}

	if updated {
		now := time.Now().Format(time.RFC3339)

		file.FileName = newName
		file.FileSize = newSize
		file.ModifiedAt = now

		newChecksum, err := utils.CalculateFileChecksum(file.FilePath)
		if err != nil {
			fmt.Printf("Warning: Failed to recalculate checksum for file %d: %v\n", id, err)
		} else if newChecksum != file.Checksum {
			isDuplicate, _, err := s.fileRepo.CheckDuplicateByChecksum(newChecksum)
			if err == nil && !isDuplicate {
				file.Checksum = newChecksum
			}
		}

		if err := s.fileRepo.Update(file); err != nil {
			return nil, fmt.Errorf("failed to update file record: %w", err)
		}

		sizeDiff := int64(newSize) - int64(originalSize)
		if sizeDiff != 0 {
			masters, err := s.storageService.GetMasterDirectories()
			if err == nil {
				for _, m := range masters {
					if strings.Contains(file.FilePath, m.Path) {
						if err := s.storageService.UpdateMasterDirectorySize(m.ID, file.FileType, sizeDiff); err != nil {
							fmt.Printf("Warning: Failed to update storage size: %v\n", err)
						}
						break
					}
				}
			}
		}
	}

	file.FileType = resolveDisplayFileType(file.FilePath, file.FileType)
	return file, nil
}

func resolveDisplayFileType(filePath, storedType string) string {
	resolvedType, err := parseFileType(filepath.Ext(filePath))
	if err != nil || resolvedType == "other" {
		return storedType
	}

	return resolvedType
}

func (s *FileService) findRenamedFile(oldPath, checksum string) (string, error) {
	oldDir := filepath.Dir(oldPath)

	if path, err := s.findFileByChecksumInDir(oldDir, checksum); err == nil {
		return path, nil
	}

	parentDir := filepath.Dir(oldDir)
	entries, err := os.ReadDir(parentDir)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				subDir := filepath.Join(parentDir, entry.Name())
				if subDir != oldDir {
					if path, err := s.findFileByChecksumInDir(subDir, checksum); err == nil {
						return path, nil
					}
				}
			}
		}
	}

	masters, err := s.storageService.GetMasterDirectories()
	if err == nil {
		for _, m := range masters {
			if path, err := s.findFileByChecksumInDir(m.Path, checksum); err == nil {
				return path, nil
			}
		}
	}

	return "", fmt.Errorf("file with checksum %s not found", checksum)
}

func (s *FileService) findFileByChecksumInDir(dir, checksum string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(dir, entry.Name())
		fileChecksum, err := utils.CalculateFileChecksum(filePath)
		if err != nil {
			continue
		}

		if fileChecksum == checksum {
			return filePath, nil
		}
	}

	return "", fmt.Errorf("file not found in directory")
}
