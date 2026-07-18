package services

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"LocalSpace/app/models"
	"LocalSpace/app/repositories"
	"LocalSpace/app/utils"
)

// FileService handles file operations
type FileService struct {
	fileRepo         *repositories.FileRepository
	storageService   *StorageService
	aiService       *AIService
	agentService    *AgentService
	thumbnailService *ThumbnailService
}

// ImportFileRequest represents a file import request
type ImportFileRequest struct {
	FilePath    string   // 文件路径
	FileName    string   // 文件名
	Description string   // 用户提供的描述
	Tags        []string // 用户提供的标签
	Keywords    string   // 用户输入的关键词
}

// FileFilter represents filters for file queries
type FileFilter struct {
	Page      int
	PageSize  int
	FileType  string
	SortBy    string
	SortOrder string
}

// NewFileService creates a new FileService
func NewFileService(
	fileRepo *repositories.FileRepository,
	storageService *StorageService,
	aiService *AIService,
	thumbnailService *ThumbnailService,
) *FileService {
	return &FileService{
		fileRepo:         fileRepo,
		storageService:   storageService,
		aiService:       aiService,
		thumbnailService: thumbnailService,
	}
}

// SetAgentService sets the agent service (called after initialization)
func (s *FileService) SetAgentService(agentService *AgentService) {
	s.agentService = agentService
}

// ImportFile imports a file into LocalSpace
func (s *FileService) ImportFile(req ImportFileRequest) error {
	// Validate input
	if req.FilePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	// Check if file exists
	if _, err := os.Stat(req.FilePath); os.IsNotExist(err) {
		return fmt.Errorf("source file does not exist: %s", req.FilePath)
	}

	// Get file info
	fileInfo, err := os.Stat(req.FilePath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Parse file type
	extension := filepath.Ext(req.FilePath)
	fileType, err := parseFileType(extension)
	if err != nil {
		return fmt.Errorf("failed to parse file type: %w", err)
	}

	// Get default master directory for storage
	masters, err := s.storageService.GetMasterDirectories()
	if err != nil {
		return fmt.Errorf("failed to get master directories: %w", err)
	}

	// Check if any master directory exists
	if len(masters) == 0 {
		return fmt.Errorf("no master directory configured. Please add a master directory in Settings > Storage Directories before importing files.")
	}

	// Find the default master directory
	var defaultMaster *models.StorageDir
	for _, m := range masters {
		if m.IsDefault {
			defaultMaster = &m
			break
		}
	}

	// If no default master, use the first one
	if defaultMaster == nil {
		defaultMaster = &masters[0]
	}

	// Check storage space
	hasSpace, err := s.storageService.CheckMasterStorageSpace(defaultMaster.ID, fileInfo.Size())
	if err != nil {
		return fmt.Errorf("failed to check storage space: %w", err)
	}

	if !hasSpace {
		return fmt.Errorf("not enough storage space for this file")
	}

	// Determine destination path using master directory
	destPath, err := s.storageService.GetStoragePathForFileWithMaster(defaultMaster.ID, fileType, req.FileName)
	if err != nil {
		return fmt.Errorf("failed to get storage path: %w", err)
	}

	// Ensure destination directory exists
	if err := s.storageService.EnsureStorageDirExists(filepath.Dir(destPath)); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Calculate checksum for duplicate detection
	checksum, err := utils.CalculateFileChecksum(req.FilePath)
	if err != nil {
		return fmt.Errorf("failed to calculate file checksum: %w", err)
	}

	// Check for duplicate files by checksum
	isDuplicate, duplicateID, err := s.fileRepo.CheckDuplicateByChecksum(checksum)
	if err != nil {
		return fmt.Errorf("failed to check for duplicate files: %w", err)
	}

	if isDuplicate {
		// Get the duplicate file details for better error message
		duplicateFile, err := s.fileRepo.FindByID(duplicateID)
		if err == nil {
			return fmt.Errorf("duplicate file detected. A file with identical content already exists: %s (ID: %d)", duplicateFile.FileName, duplicateID)
		}
		return fmt.Errorf("duplicate file detected. A file with identical content already exists (ID: %d)", duplicateID)
	}

	// Check if file already exists at destination path
	exists, err := s.fileRepo.ExistsByPath(destPath)
	if err != nil {
		return fmt.Errorf("failed to check if file exists: %w", err)
	}

	if exists {
		return fmt.Errorf("file already exists at destination: %s", destPath)
	}

	// Move file
	if err := os.Rename(req.FilePath, destPath); err != nil {
		// If rename fails (cross-device), try copy
		if err := copyFile(req.FilePath, destPath); err != nil {
			return fmt.Errorf("failed to move/copy file: %w", err)
		}
	}

	// Extract metadata from the file
	metadata, err := s.ExtractMetadata(destPath, fileType)
	if err != nil {
		fmt.Printf("Warning: Failed to extract metadata: %v\n", err)
	}

	// AI generation using AgentService
	var tags []string
	var description string

	// Try AgentService first
	if s.agentService != nil {
		ctx := context.Background()

		// Generate tags
		if agentTags, err := s.agentService.GenerateTags(
			ctx,
			req.FileName,
			fileType,
			req.Keywords,
			req.Tags,
			req.Description,
		); err == nil {
			tags = agentTags
		}

		// Generate description
		if agentDesc, err := s.agentService.GenerateDescription(
			ctx,
			req.FileName,
			fileType,
			req.Keywords,
			req.Tags,
			req.Description,
		); err == nil {
			description = agentDesc
		}
	}

	// Fallback to user input if AI generation failed
	if len(tags) == 0 && len(req.Tags) > 0 {
		tags = req.Tags
	}
	if description == "" && req.Description != "" {
		description = req.Description
	}

	// Create file record
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
		Thumbnail:    "", // Will be set after file creation
		Checksum:     checksum,
		IsDeleted:    false,
		DeletedAt:    "",
	}

	// Save to database first to get file ID
	if err := s.fileRepo.Create(file); err != nil {
		// Rollback: remove moved file
		os.Remove(destPath)
		return fmt.Errorf("failed to create file record: %w", err)
	}

	// Generate thumbnail asynchronously
	go s.generateThumbnailForFile(file.ID, destPath, fileType)

	// Update storage size
	if err := s.storageService.UpdateMasterDirectorySize(defaultMaster.ID, fileType, fileInfo.Size()); err != nil {
		// This is not critical, log error
		fmt.Printf("Failed to update storage size: %v\n", err)
	}

	return nil
}

// ListFiles returns a list of files
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

	// 为每个文件生成或获取缩略图
	for _, file := range files {
		if file.Thumbnail == "" || !s.thumbnailExists(file.Thumbnail) {
			// 异步生成缩略图
			go s.generateThumbnailForFile(file.ID, file.FilePath, file.FileType)
		}
	}

	return files, nil
}

// SearchFiles searches for files
func (s *FileService) SearchFiles(query string) ([]*models.File, error) {
	return s.fileRepo.Search(query)
}

// GetFile returns a single file by ID
func (s *FileService) GetFile(id uint) (*models.File, error) {
	file, err := s.fileRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// 生成或获取缩略图
	if file.Thumbnail == "" || !s.thumbnailExists(file.Thumbnail) {
		go s.generateThumbnailForFile(file.ID, file.FilePath, file.FileType)
	}

	return file, nil
}

// GetThumbnail returns the thumbnail path for a file
func (s *FileService) GetThumbnail(id uint) (string, error) {
	file, err := s.fileRepo.FindByID(id)
	if err != nil {
		return "", fmt.Errorf("failed to get file: %w", err)
	}

	// 如果缩略图路径为空或文件不存在，生成新的缩略图
	if file.Thumbnail == "" || !s.thumbnailExists(file.Thumbnail) {
		thumbnailPath, err := s.thumbnailService.GetThumbnail(file.FilePath, file.FileType, file.ID)
		if err != nil {
			return "", fmt.Errorf("failed to get thumbnail: %w", err)
		}

		// 更新文件记录
		if err := s.fileRepo.UpdateThumbnail(id, thumbnailPath); err != nil {
			// 记录错误但不影响返回
			fmt.Printf("Warning: Failed to update thumbnail path: %v\n", err)
		}

		return thumbnailPath, nil
	}

	return file.Thumbnail, nil
}

// DeleteFile deletes a file
func (s *FileService) DeleteFile(id uint) error {
	// Get file info
	file, err := s.fileRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("failed to get file: %w", err)
	}

	// Delete physical file
	if err := os.Remove(file.FilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	// Remove thumbnail
	if file.Thumbnail != "" {
		s.thumbnailService.RemoveThumbnail(id, file.FileType)
	}

	// Update storage size
	if err := s.storageService.UpdateStorageSize(file.FileType, -file.FileSize); err != nil {
		// This is not critical, log error
		fmt.Printf("Failed to update storage size: %v\n", err)
	}

	// Delete from database
	return s.fileRepo.Delete(id)
}

// RenameFile renames a file
func (s *FileService) RenameFile(id uint, newName string) error {
	// Get current file info
	file, err := s.fileRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("failed to get file: %w", err)
	}

	// Validate new name
	if newName == "" {
		return fmt.Errorf("new file name cannot be empty")
	}

	// Check if new name is the same as current name
	if newName == file.FileName {
		return nil
	}

	// Get file extension
	ext := filepath.Ext(file.FilePath)
	if ext != "" && !strings.HasSuffix(newName, ext) {
		newName = newName + ext
	}

	// Construct new file path
	dir := filepath.Dir(file.FilePath)
	newPath := filepath.Join(dir, newName)

	// Check if new path already exists
	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("a file with this name already exists: %s", newName)
	}

	// Rename the physical file
	if err := os.Rename(file.FilePath, newPath); err != nil {
		return fmt.Errorf("failed to rename file: %w", err)
	}

	// Update file record
	file.FileName = newName
	file.FilePath = newPath
	file.ModifiedAt = time.Now().Format(time.RFC3339)

	// Save to database
	if err := s.fileRepo.Update(file); err != nil {
		// Rollback: rename back to original
		os.Rename(newPath, file.FilePath)
		return fmt.Errorf("failed to update file record: %w", err)
	}

	return nil
}

// DeleteFilesByPath deletes all files with a specific path prefix
func (s *FileService) DeleteFilesByPath(pathPrefix string) (int64, error) {
	// Get all files with this path prefix first (to clean up thumbnails)
	filter := repositories.FileFilter{
		Page:     1,
		PageSize: 10000, // Get all files
	}

	files, err := s.fileRepo.List(filter)
	if err != nil {
		return 0, fmt.Errorf("failed to list files: %w", err)
	}

	// Remove thumbnails for files that match the path prefix
	deletedCount := int64(0)
	for _, file := range files {
		if strings.HasPrefix(file.FilePath, pathPrefix) {
			// Remove thumbnail
			if file.Thumbnail != "" {
				s.thumbnailService.RemoveThumbnail(file.ID, file.FileType)
			}
			// Update storage size
			if err := s.storageService.UpdateStorageSize(file.FileType, -file.FileSize); err != nil {
				// This is not critical, log error
				fmt.Printf("Failed to update storage size: %v\n", err)
			}
			deletedCount++
		}
	}

	// Delete all files with this path prefix from database
	rowsAffected, err := s.fileRepo.DeleteByPath(pathPrefix)
	if err != nil {
		return deletedCount, fmt.Errorf("failed to delete files by path: %w", err)
	}

	return rowsAffected, nil
}

// ExtractMetadata extracts metadata from a file
func (s *FileService) ExtractMetadata(filePath, fileType string) (models.Metadata, error) {
	var metadata models.Metadata

	// Check file exists and get file info
	fileInfo, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return metadata, fmt.Errorf("file does not exist: %s", filePath)
	}
	if err != nil {
		return metadata, fmt.Errorf("failed to get file info: %w", err)
	}

	// Always include file size
	metadata.Size = fileInfo.Size()

	// Extract metadata based on file type
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
		// For unsupported types, return metadata with just file size
		return metadata, nil
	}
}

// extractImageMetadata extracts image metadata
func (s *FileService) extractImageMetadata(filePath string) (models.Metadata, error) {
	imgMeta, err := utils.ExtractImageMetadata(filePath)
	if err != nil {
		return models.Metadata{}, fmt.Errorf("failed to extract image metadata: %w", err)
	}

	metadata := models.Metadata{
		Width:  imgMeta.Width,
		Height: imgMeta.Height,
	}

	return metadata, nil
}

// extractVideoMetadata extracts video metadata
func (s *FileService) extractVideoMetadata(filePath string) (models.Metadata, error) {
	vidMeta, err := utils.ExtractVideoMetadata(filePath)
	if err != nil {
		return models.Metadata{}, fmt.Errorf("failed to extract video metadata: %w", err)
	}

	metadata := models.Metadata{
		Width:    vidMeta.Width,
		Height:   vidMeta.Height,
		Duration: int(vidMeta.Duration),
	}

	return metadata, nil
}

// extractDocumentMetadata extracts document metadata
func (s *FileService) extractDocumentMetadata(filePath, fileType string) (models.Metadata, error) {
	docMeta, err := utils.ExtractDocumentMetadata(filePath, fileType)
	if err != nil {
		return models.Metadata{}, fmt.Errorf("failed to extract document metadata: %w", err)
	}

	metadata := models.Metadata{
		PageCount: docMeta.PageCount,
		Author:    docMeta.Author,
		Title:     docMeta.Title,
	}

	return metadata, nil
}

// generateThumbnailForFile 为文件生成缩略图
func (s *FileService) generateThumbnailForFile(fileID uint, filePath, fileType string) {
	thumbnailPath, err := s.thumbnailService.GetThumbnail(filePath, fileType, fileID)
	if err != nil {
		fmt.Printf("Warning: Failed to generate thumbnail for file %d: %v\n", fileID, err)
		return
	}

	// 更新文件记录
	if err := s.fileRepo.UpdateThumbnail(fileID, thumbnailPath); err != nil {
		fmt.Printf("Warning: Failed to update thumbnail path for file %d: %v\n", fileID, err)
	}
}

// thumbnailExists 检查缩略图文件是否存在
func (s *FileService) thumbnailExists(thumbnailPath string) bool {
	if thumbnailPath == "" {
		return false
	}
	_, err := os.Stat(thumbnailPath)
	return err == nil
}

// copyFile copies a file from src to dst
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

// parseFileType parses file type from extension
func parseFileType(extension string) (string, error) {
	ext := strings.ToLower(extension)

	// Map extensions to file types
	typeMap := map[string]string{
		// Video files
		".mp4":  "video",
		".avi":  "video",
		".mkv":  "video",
		".mov":  "video",
		".wmv":  "video",
		".flv":  "video",
		".webm": "video",
		".m4v":  "video",

		// Document files
		".pdf":   "document",
		".doc":   "document",
		".docx":  "document",
		".xls":   "document",
		".xlsx":  "document",
		".ppt":   "document",
		".pptx":  "document",
		".txt":   "document",
		".md":    "document",
		".rtf":   "document",

		// Music files
		".mp3":  "music",
		".wav":  "music",
		".flac": "music",
		".aac":  "music",
		".ogg":  "music",
		".m4a":  "music",
		".wma":  "music",

		// Game files
		".exe":  "game",
		".app":  "game",
		".iso":  "game",
		".zip":  "game",
		".rar":  "game",
		".7z":   "game",

		// Installer files
		".msi":  "installer",
		".pkg":  "installer",
		".deb":  "installer",
		".rpm":  "installer",
		".apk":  "installer",
		".dmg":  "installer",

		// Image/Disk files (disk images, not pictures)
		".img":  "image",
		".vmdk": "image",

		// Picture files
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

// CheckDuplicate checks if a file is a duplicate by calculating its checksum
func (s *FileService) CheckDuplicate(filePath string) (bool, uint, string, error) {
	// Validate file path
	if filePath == "" {
		return false, 0, "", fmt.Errorf("file path cannot be empty")
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false, 0, "", fmt.Errorf("source file does not exist: %s", filePath)
	}

	// Calculate checksum
	checksum, err := utils.CalculateFileChecksum(filePath)
	if err != nil {
		return false, 0, "", fmt.Errorf("failed to calculate file checksum: %w", err)
	}

	// Check for duplicate
	isDuplicate, duplicateID, err := s.fileRepo.CheckDuplicateByChecksum(checksum)
	if err != nil {
		return false, 0, "", fmt.Errorf("failed to check for duplicate: %w", err)
	}

	// If duplicate found, get the duplicate file name
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

// GetDuplicateFiles finds all duplicate files based on checksums
func (s *FileService) GetDuplicateFiles() (map[string][]*models.File, error) {
	// Get all files with checksums
	allFiles, err := s.fileRepo.List(repositories.FileFilter{
		Page:     1,
		PageSize: 10000, // Get all files
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get all files: %w", err)
	}

	// Group files by checksum
	checksumMap := make(map[string][]*models.File)
	for _, file := range allFiles {
		if file.Checksum != "" && !file.IsDeleted {
			checksumMap[file.Checksum] = append(checksumMap[file.Checksum], file)
		}
	}

	// Filter only checksums with more than one file (duplicates)
	duplicates := make(map[string][]*models.File)
	for checksum, files := range checksumMap {
		if len(files) > 1 {
			duplicates[checksum] = files
		}
	}

	return duplicates, nil
}

// RefreshFile checks and updates file information if the file has been modified
func (s *FileService) RefreshFile(id uint) (*models.File, error) {
	// Get current file record
	file, err := s.fileRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	// Check if physical file exists
	fileInfo, err := os.Stat(file.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File might have been renamed or moved, try to find it by checksum
			if file.Checksum == "" {
				// Return the current file record without error - file may have been renamed
				return file, nil
			}

			// Try to find the file in the same directory or related directories
			newPath, err := s.findRenamedFile(file.FilePath, file.Checksum)
			if err != nil {
				// File not found - return current record without error
				// This allows frontend to handle it gracefully
				return file, nil
			}

			// Update the file path in the database
			file.FilePath = newPath
			file.FileName = filepath.Base(newPath)

			// Update the file record with new path
			if err := s.fileRepo.Update(file); err != nil {
				// Log error but continue with current file info
				fmt.Printf("Warning: Failed to update file record with new path: %v\n", err)
			}

			// Get the updated file info from new location
			fileInfo, err = os.Stat(newPath)
			if err != nil {
				// Return current file info
				return file, nil
			}

			// Update file size
			file.FileSize = fileInfo.Size()

			// Update modification time
			file.ModifiedAt = fileInfo.ModTime().Format(time.RFC3339)

			// Save the final updates
			if err := s.fileRepo.Update(file); err != nil {
				fmt.Printf("Warning: Failed to update file record: %v\n", err)
			}

			return file, nil
		}
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Check if file has been modified (compare modification time)
	currentModTime := fileInfo.ModTime()
	storedModTime, err := time.Parse(time.RFC3339, file.ModifiedAt)
	if err != nil {
		// If we can't parse the stored time, assume it's modified
		storedModTime = time.Time{}
	}

	updated := false
	originalSize := file.FileSize
	newName := file.FileName
	newSize := file.FileSize

	// Check if file size changed
	if fileInfo.Size() != file.FileSize {
		newSize = fileInfo.Size()
		updated = true
	}

	// Check if file name changed (get from file path)
	currentFileName := filepath.Base(file.FilePath)
	if currentFileName != file.FileName {
		newName = currentFileName
		updated = true
	}

	// Check if modification time changed
	if currentModTime.After(storedModTime) {
		updated = true
	}

	// If file was modified, update the record
	if updated {
		now := time.Now().Format(time.RFC3339)

		// Update file record
		file.FileName = newName
		file.FileSize = newSize
		file.ModifiedAt = now

		// Recalculate checksum if file content changed
		newChecksum, err := utils.CalculateFileChecksum(file.FilePath)
		if err != nil {
			fmt.Printf("Warning: Failed to recalculate checksum for file %d: %v\n", id, err)
		} else if newChecksum != file.Checksum {
			// Check if new checksum conflicts with existing files
			isDuplicate, _, err := s.fileRepo.CheckDuplicateByChecksum(newChecksum)
			if err == nil && !isDuplicate {
				file.Checksum = newChecksum
			}
		}

		// Save updates
		if err := s.fileRepo.Update(file); err != nil {
			return nil, fmt.Errorf("failed to update file record: %w", err)
		}

		// Update storage size if file size changed
		sizeDiff := int64(newSize) - int64(originalSize)
		if sizeDiff != 0 {
			// Get the master directory for this file type
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

// findRenamedFile tries to find a file that was renamed or moved by matching checksums
func (s *FileService) findRenamedFile(oldPath, checksum string) (string, error) {
	// Get the directory of the old file
	oldDir := filepath.Dir(oldPath)

	// First, search in the same directory
	if path, err := s.findFileByChecksumInDir(oldDir, checksum); err == nil {
		return path, nil
	}

	// If not found, search in parent directories and sibling directories
	parentDir := filepath.Dir(oldDir)
	entries, err := os.ReadDir(parentDir)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				subDir := filepath.Join(parentDir, entry.Name())
				// Skip the original directory (already checked)
				if subDir != oldDir {
					if path, err := s.findFileByChecksumInDir(subDir, checksum); err == nil {
						return path, nil
					}
				}
			}
		}
	}

	// As a fallback, search in all configured master directories
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

// findFileByChecksumInDir searches for a file with the given checksum in a specific directory
func (s *FileService) findFileByChecksumInDir(dir, checksum string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip directories, we only want files
		}

		filePath := filepath.Join(dir, entry.Name())
		fileChecksum, err := utils.CalculateFileChecksum(filePath)
		if err != nil {
			continue // Skip files that can't be read
		}

		if fileChecksum == checksum {
			return filePath, nil
		}
	}

	return "", fmt.Errorf("file not found in directory")
}