package services

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

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
	FilePath    string   // source path
	FileName    string   // target display name
	Description string   // user-provided description
	Tags        []string // user-provided tags
	Keywords    string   // user-provided keywords
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

	fileInfo, err := os.Stat(req.FilePath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	extension := filepath.Ext(req.FilePath)
	fileType, err := parseFileType(extension)
	if err != nil {
		return fmt.Errorf("failed to parse file type: %w", err)
	}

	masters, err := s.storageService.GetMasterDirectories()
	if err != nil {
		return fmt.Errorf("failed to get master directories: %w", err)
	}
	if len(masters) == 0 {
		return fmt.Errorf("no master directory configured. Please add a master directory in Settings > Storage Directories before importing files.")
	}

	var defaultMaster *models.StorageDir
	for _, m := range masters {
		if m.IsDefault {
			defaultMaster = &m
			break
		}
	}
	if defaultMaster == nil {
		defaultMaster = &masters[0]
	}

	hasSpace, err := s.storageService.CheckMasterStorageSpace(defaultMaster.ID, fileInfo.Size())
	if err != nil {
		return fmt.Errorf("failed to check storage space: %w", err)
	}
	if !hasSpace {
		return fmt.Errorf("not enough storage space for this file")
	}

	destPath, err := s.storageService.GetStoragePathForFileWithMaster(defaultMaster.ID, fileType, req.FileName)
	if err != nil {
		return fmt.Errorf("failed to get storage path: %w", err)
	}
	if err := s.storageService.EnsureStorageDirExists(filepath.Dir(destPath)); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	checksum, err := utils.CalculateFileChecksum(req.FilePath)
	if err != nil {
		return fmt.Errorf("failed to calculate file checksum: %w", err)
	}

	isDuplicate, duplicateID, err := s.fileRepo.CheckDuplicateByChecksum(checksum)
	if err != nil {
		return fmt.Errorf("failed to check for duplicate files: %w", err)
	}
	if isDuplicate {
		duplicateFile, err := s.fileRepo.FindByID(duplicateID)
		if err == nil {
			return fmt.Errorf("duplicate file detected. A file with identical content already exists: %s (ID: %d)", duplicateFile.FileName, duplicateID)
		}
		return fmt.Errorf("duplicate file detected. A file with identical content already exists (ID: %d)", duplicateID)
	}

	exists, err := s.fileRepo.ExistsByPath(destPath)
	if err != nil {
		return fmt.Errorf("failed to check if file exists: %w", err)
	}
	if exists {
		return fmt.Errorf("file already exists at destination: %s", destPath)
	}

	if err := os.Rename(req.FilePath, destPath); err != nil {
		if err := copyFile(req.FilePath, destPath); err != nil {
			return fmt.Errorf("failed to move/copy file: %w", err)
		}
	}

	metadata, err := s.ExtractMetadata(destPath, fileType)
	if err != nil {
		fmt.Printf("Warning: Failed to extract metadata: %v\n", err)
	}

	formMetadata := s.resolveMetadataForImport(&req)
	tags := formMetadata.Tags
	description := formMetadata.Description

	originalName := req.FileName
	if req.FilePath != "" {
		originalName = filepath.Base(req.FilePath)
	}

	file := &models.File{
		FileName:     req.FileName,
		OriginalName: originalName,
		FilePath:     destPath,
		FileType:     fileType,
		FileSubType:  strings.TrimPrefix(extension, "."),
		FileSize:     fileInfo.Size(),
		Tags:         tags,
		Description:  description,
		Metadata:     metadata,
		Thumbnail:    "",
		Checksum:     checksum,
		IsDeleted:    false,
		DeletedAt:    "",
	}

	if err := s.fileRepo.Create(file); err != nil {
		_ = os.Remove(destPath)
		return fmt.Errorf("failed to create file record: %w", err)
	}

	if supportsGeneratedThumbnail(fileType) {
		// Generate during import so the first file listing can use the cached thumbnail.
		s.generateThumbnailForFile(file.ID, destPath, fileType)
	}

	if err := s.storageService.UpdateMasterDirectorySize(defaultMaster.ID, fileType, fileInfo.Size()); err != nil {
		fmt.Printf("Failed to update storage size: %v\n", err)
	}

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

	if !supportsGeneratedThumbnail(file.FileType) {
		return "", fmt.Errorf("thumbnails are only supported for image and video files")
	}

	if file.Thumbnail == "" || !s.thumbnailIsCurrent(file) {
		thumbnailPath, err := s.thumbnailService.GetThumbnail(file.FilePath, file.FileType, file.ID)
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

	if err := os.Remove(file.FilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	if file.Thumbnail != "" {
		_ = s.thumbnailService.RemoveThumbnail(id, file.FileType)
	}

	if err := s.storageService.UpdateStorageSize(file.FileType, -file.FileSize); err != nil {
		fmt.Printf("Failed to update storage size: %v\n", err)
	}

	return s.fileRepo.Delete(id)
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

	if err := os.Rename(file.FilePath, newPath); err != nil {
		return fmt.Errorf("failed to rename file: %w", err)
	}

	originalPath := file.FilePath
	file.FileName = newName
	file.FilePath = newPath
	file.ModifiedAt = time.Now().Format(time.RFC3339)

	if err := s.fileRepo.Update(file); err != nil {
		_ = os.Rename(newPath, originalPath)
		file.FilePath = originalPath
		return fmt.Errorf("failed to update file record: %w", err)
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
				_ = s.thumbnailService.RemoveThumbnail(file.ID, file.FileType)
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

		".exe": "game",
		".app": "game",
		".iso": "game",
		".zip": "game",
		".rar": "game",
		".7z":  "game",

		".msi": "installer",
		".pkg": "installer",
		".deb": "installer",
		".rpm": "installer",
		".apk": "installer",
		".dmg": "installer",

		".img":  "image",
		".vmdk": "image",

		".jpg":  "image",
		".jpeg": "image",
		".png":  "image",
		".gif":  "image",
		".bmp":  "image",
		".webp": "image",
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
				return file, nil
			}

			newPath, err := s.findRenamedFile(file.FilePath, file.Checksum)
			if err != nil {
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

	return file, nil
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
