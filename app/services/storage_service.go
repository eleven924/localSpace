package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"LocalSpace/app/models"
	"LocalSpace/app/repositories"
)

// StorageService handles storage directory operations
type StorageService struct {
	configRepo    *repositories.ConfigRepository
	storageBasePath string
}

// NewStorageService creates a new StorageService
func NewStorageService(configRepo *repositories.ConfigRepository) *StorageService {
	return &StorageService{configRepo: configRepo}
}

// SetStorageBasePath sets the base storage path
func (s *StorageService) SetStorageBasePath(path string) {
	s.storageBasePath = path
}

func sanitizePathSegment(segment string) string {
	replacer := strings.NewReplacer(
		`<`, "_",
		`>`, "_",
		`:`, "_",
		`"`, "_",
		`/`, "_",
		`\`, "_",
		`|`, "_",
		`?`, "_",
		`*`, "_",
	)

	cleaned := strings.TrimSpace(replacer.Replace(segment))
	cleaned = strings.TrimRight(cleaned, ". ")
	if cleaned == "" {
		return "_"
	}
	return cleaned
}

// GetStorageBasePath returns the base storage path
func (s *StorageService) GetStorageBasePath() (string, error) {
	if s.storageBasePath != "" {
		return s.storageBasePath, nil
	}

	// Try to get from config
	path, err := s.configRepo.Get("file_storage_path")
	if err == nil && path != "" {
		s.storageBasePath = path
		return s.storageBasePath, nil
	}

	// Not configured - use default path in data directory
	// Get the application directory
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}
	execDir := filepath.Dir(execPath)
	defaultPath := filepath.Join(execDir, "data", "files")

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(defaultPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create default storage directory: %w", err)
	}

	// Save to config
	if err := s.configRepo.Set("file_storage_path", defaultPath); err != nil {
		fmt.Printf("Warning: failed to save default storage path: %v\n", err)
	}

	s.storageBasePath = defaultPath
	return s.storageBasePath, nil
}

// GetStorageDirectories returns all storage directories
func (s *StorageService) GetStorageDirectories() ([]models.StorageDir, error) {
	return s.configRepo.GetStorageDirectories()
}

// GetStorageLayoutConfig returns the current storage layout configuration.
func (s *StorageService) GetStorageLayoutConfig() (*models.StorageLayoutConfig, error) {
	return s.configRepo.GetStorageLayoutConfig()
}

// AddStorageDirectory adds a new storage directory
func (s *StorageService) AddStorageDirectory(path, fileType string) error {
	// Validate path
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Check if directory exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", path)
	}

	// Create storage directory
	dir := &models.StorageDir{
		Path:        path,
		FileType:    fileType,
		CurrentSize: 0,
		MaxSize:     0, // 0 means unlimited
		IsActive:    true,
	}

	_, err := s.configRepo.AddStorageDirectory(dir)
	if err != nil {
		return err
	}

	return nil
}

// RemoveStorageDirectory removes a storage directory and clears its contents
func (s *StorageService) RemoveStorageDirectory(id uint) error {
	// Get the directory info before removing
	dirs, err := s.GetStorageDirectories()
	if err != nil {
		return err
	}

	// Find the directory to remove
	var dirToRemove *models.StorageDir
	for _, dir := range dirs {
		if dir.ID == id {
			dirToRemove = &dir
			break
		}
	}

	if dirToRemove == nil {
		return fmt.Errorf("storage directory not found")
	}

	// Clear directory contents if it exists
	if dirToRemove.Path != "" {
		if err := s.clearDirectoryContents(dirToRemove.Path); err != nil {
			// Log error but don't fail the removal
			fmt.Printf("Warning: failed to clear directory contents: %v\n", err)
		}
	}

	// Remove from database
	return s.configRepo.RemoveStorageDirectory(id)
}

// clearDirectoryContents clears all contents of a directory
func (s *StorageService) clearDirectoryContents(path string) error {
	// Check if directory exists
	dirInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Directory doesn't exist, nothing to clear
			return nil
		}
		return fmt.Errorf("failed to stat directory: %w", err)
	}

	// Make sure it's a directory
	if !dirInfo.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	// Read all entries
	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// Remove all entries
	for _, entry := range entries {
		entryPath := filepath.Join(path, entry.Name())

		if err := os.RemoveAll(entryPath); err != nil {
			// Log error but continue removing other files
			fmt.Printf("Warning: failed to remove %s: %v\n", entryPath, err)
		}
	}

	return nil
}

// ToggleStorageDirectory toggles the active status of a storage directory
func (s *StorageService) ToggleStorageDirectory(id uint, isActive bool) error {
	return s.configRepo.ToggleStorageDirectory(id, isActive)
}

// CheckStorageSpace checks if there's enough storage space for a file
func (s *StorageService) CheckStorageSpace(fileSize int64) (bool, error) {
	dirs, err := s.GetStorageDirectories()
	if err != nil {
		return false, fmt.Errorf("failed to get storage directories: %w", err)
	}

	// Check if any directory has enough space
	for _, dir := range dirs {
		if !dir.IsActive {
			continue
		}

		// If max_size is 0, it means unlimited
		if dir.MaxSize == 0 {
			return true, nil
		}

		// Check if current size + file size <= max size
		if dir.CurrentSize+fileSize <= dir.MaxSize {
			return true, nil
		}
	}

	// If no directories exist, create default storage directory with unlimited space
	if len(dirs) == 0 {
		basePath, err := s.GetStorageBasePath()
		if err != nil {
			return false, fmt.Errorf("failed to get storage base path: %w", err)
		}

		defaultPath := filepath.Join(basePath, "files")
		if err := os.MkdirAll(defaultPath, 0755); err != nil {
			return false, fmt.Errorf("failed to create default storage directory: %w", err)
		}

		// Add this directory to the database with unlimited space (max_size = 0)
		if err := s.AddStorageDirectory(defaultPath, "all"); err != nil {
			// Log error but continue - the directory exists
			fmt.Printf("Warning: failed to add default storage directory: %v\n", err)
		}

		return true, nil
	}

	return false, fmt.Errorf("not enough storage space")
}

// GetStoragePath returns the best storage path for a given file type
func (s *StorageService) GetStoragePath(fileType string) (string, error) {
	dirs, err := s.GetStorageDirectories()
	if err != nil {
		return "", fmt.Errorf("failed to get storage directories: %w", err)
	}

	// Find a directory that matches the file type
	for _, dir := range dirs {
		if dir.FileType == fileType && dir.IsActive {
			return dir.Path, nil
		}
	}

	// If no matching directory found, create a default one under base path
	basePath, err := s.GetStorageBasePath()
	if err != nil {
		return "", fmt.Errorf("file storage path not configured. Please configure file storage in settings before importing files.")
	}

	defaultPath := filepath.Join(basePath, fileType)
	if err := os.MkdirAll(defaultPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create default storage directory: %w", err)
	}

	// Add this directory to the database
	if err := s.AddStorageDirectory(defaultPath, fileType); err != nil {
		// Log error but continue
		fmt.Printf("Warning: failed to add default storage directory: %v\n", err)
	}

	return defaultPath, nil
}

// GetStoragePathForFile returns the full storage path for a file
func (s *StorageService) GetStoragePathForFile(fileType, fileName string) (string, error) {
	basePath, err := s.GetStoragePath(fileType)
	if err != nil {
		return "", err
	}

	// Use the base path directly (no subdirectories for now)
	return filepath.Join(basePath, fileName), nil
}

// UpdateStorageSize updates the storage size after file operations
func (s *StorageService) UpdateStorageSize(fileType string, sizeChange int64) error {
	fmt.Printf("UpdateStorageSize called: fileType=%s, sizeChange=%d\n", fileType, sizeChange)

	dirs, err := s.GetStorageDirectories()
	if err != nil {
		return fmt.Errorf("failed to get storage directories: %w", err)
	}

	fmt.Printf("Found %d storage directories\n", len(dirs))
	for _, dir := range dirs {
		fmt.Printf("  - ID: %d, Path: %s, FileType: %s, CurrentSize: %d\n", dir.ID, dir.Path, dir.FileType, dir.CurrentSize)
	}

	// Find the directory for this file type and update its size
	for _, dir := range dirs {
		if dir.FileType == fileType {
			fmt.Printf("Updating size for directory ID %d (type: %s): %d bytes\n", dir.ID, fileType, sizeChange)
			return s.configRepo.UpdateStorageDirSize(dir.ID, sizeChange)
		}
	}

	return fmt.Errorf("no storage directory found for file type: %s", fileType)
}

// EnsureStorageDirExists ensures that the storage directory exists
func (s *StorageService) EnsureStorageDirExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create storage directory: %w", err)
		}
	}
	return nil
}

// UpdateMasterDirectorySize updates the size of a subdirectory under a master directory
func (s *StorageService) UpdateMasterDirectorySize(masterID uint, fileType string, sizeChange int64) error {
	// Get subdirectories for this master directory
	subs, err := s.configRepo.GetSubDirectories(masterID)
	if err != nil {
		return fmt.Errorf("failed to get subdirectories: %w", err)
	}

	// Find the subdirectory with the matching file type
	for _, sub := range subs {
		if sub.FileType == fileType {
			fmt.Printf("Updating master directory subdirectory ID %d (type: %s): %d bytes\n", sub.ID, fileType, sizeChange)
			return s.configRepo.UpdateStorageDirSize(sub.ID, sizeChange)
		}
	}

	return fmt.Errorf("no subdirectory found for file type %s under master directory %d", fileType, masterID)
}

// GetMasterDirectories 获取所有主目录及其子目录
func (s *StorageService) GetMasterDirectories() ([]models.StorageDir, error) {
	masters, err := s.configRepo.GetMasterDirectories()
	if err != nil {
		return nil, err
	}

	// 为每个主目录加载子目录
	for i := range masters {
		subs, err := s.configRepo.GetSubDirectories(masters[i].ID)
		if err != nil {
			return nil, err
		}
		masters[i].SubDirs = subs

		fmt.Printf("Master directory ID %d (%s) has %d subdirectories:\n", masters[i].ID, masters[i].Path, len(subs))
		for _, sub := range subs {
			fmt.Printf("  - Subdirectory ID %d, Type: %s, Path: %s, CurrentSize: %d\n", sub.ID, sub.FileType, sub.Path, sub.CurrentSize)
		}

		// 计算总大小
		totalSize, err := s.configRepo.CalculateMasterDirectorySize(masters[i].ID)
		if err != nil {
			return nil, err
		}
		masters[i].TotalSize = totalSize
		fmt.Printf("Calculated total size for master directory ID %d: %d bytes\n", masters[i].ID, totalSize)
	}

	return masters, nil
}

// AddMasterDirectory 添加主目录
func (s *StorageService) AddMasterDirectory(path string, maxSizeGB float64) error {
	// 验证路径
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	normalizedPath := filepath.Clean(path)

	// 检查路径冲突
	hasConflict, err := s.configRepo.CheckPathConflict(normalizedPath, 0)
	if err != nil {
		return err
	}
	if hasConflict {
		return fmt.Errorf("path conflicts with existing directory")
	}

	// 检查目录是否存在
	if _, err := os.Stat(normalizedPath); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", normalizedPath)
	}

	// 如果是第一个主目录，自动设为默认
	masters, err := s.configRepo.GetMasterDirectories()
	if err != nil {
		return err
	}
	fmt.Printf("Current master directories count: %d\n", len(masters))
	for _, m := range masters {
		fmt.Printf("  - ID: %d, Path: %s, FileType: %s, IsDefault: %v\n", m.ID, m.Path, m.FileType, m.IsDefault)
	}
	isDefault := len(masters) == 0
	fmt.Printf("Setting IsDefault to: %v for new master directory\n", isDefault)

	// 创建主目录
	dir := &models.StorageDir{
		Path:      normalizedPath,
		FileType:  "master", // 主目录类型
		MaxSize:   int64(maxSizeGB * 1024 * 1024 * 1024), // 转换为字节
		IsActive:  true,
		IsDefault: isDefault,
		ParentID:  nil,
	}

	_, err = s.configRepo.AddStorageDirectory(dir)
	return err
}

// SetDefaultMasterDirectory 设置默认主目录
func (s *StorageService) SetDefaultMasterDirectory(id uint) error {
	return s.configRepo.SetDefaultMasterDirectory(id)
}

// EnsureSubDirectory 确保子目录存在，不存在则创建
func (s *StorageService) EnsureSubDirectory(masterID uint, fileType string) (uint, string, error) {
	// 获取主目录信息
	masters, err := s.configRepo.GetMasterDirectories()
	if err != nil {
		return 0, "", err
	}

	var master *models.StorageDir
	for _, m := range masters {
		if m.ID == masterID {
			master = &m
			break
		}
	}

	if master == nil {
		return 0, "", fmt.Errorf("master directory not found")
	}

	// 检查子目录是否已存在
	subs, err := s.configRepo.GetSubDirectories(masterID)
	if err != nil {
		return 0, "", err
	}

	for _, sub := range subs {
		if sub.FileType == fileType {
			return sub.ID, sub.Path, nil // 已存在
		}
	}

	// 创建新的子目录
	subPath := filepath.Join(master.Path, fileType)
	if err := os.MkdirAll(subPath, 0755); err != nil {
		return 0, "", fmt.Errorf("failed to create sub directory: %w", err)
	}

	// 检查是否已经存在相同路径的记录（可能是旧数据）
	// 如果存在，先删除它
	dirs, err := s.GetStorageDirectories()
	if err != nil {
		return 0, "", fmt.Errorf("failed to get storage directories: %w", err)
	}

	for _, dir := range dirs {
		if dir.Path == subPath {
			// 删除冲突的记录
			if err := s.configRepo.RemoveStorageDirectory(dir.ID); err != nil {
				// 忽略删除错误，继续尝试创建
				fmt.Printf("Warning: failed to remove conflicting directory: %v\n", err)
			}
			break
		}
	}

	subDir := &models.StorageDir{
		Path:     subPath,
		FileType: fileType,
		MaxSize:  0, // 子目录无限制
		IsActive: true,
		ParentID: &masterID,
	}

	subDirID, err := s.configRepo.AddStorageDirectory(subDir)
	if err != nil {
		return 0, "", err
	}

	return subDirID, subPath, nil
}

// GetStoragePathForFileWithMaster 使用指定主目录获取文件存储路径
func (s *StorageService) GetStoragePathForFileWithMaster(masterID uint, fileType, collectionName, fileName string) (string, error) {
	// 获取指定的主目录
	masters, err := s.configRepo.GetMasterDirectories()
	if err != nil {
		return "", fmt.Errorf("failed to get master directories: %w", err)
	}

	var master *models.StorageDir
	for _, m := range masters {
		if m.ID == masterID {
			master = &m
			break
		}
	}

	if master == nil {
		return "", fmt.Errorf("master directory not found")
	}

	// 确保子目录存在
	_, subPath, err := s.EnsureSubDirectory(masterID, fileType)
	if err != nil {
		return "", err
	}

	layoutConfig, err := s.GetStorageLayoutConfig()
	if err != nil {
		return "", fmt.Errorf("failed to get storage layout config: %w", err)
	}

	if layoutConfig.Strategy == "type_collection" {
		segment := strings.TrimSpace(collectionName)
		if segment == "" {
			segment = layoutConfig.UnsortedFolderName
		}
		if layoutConfig.SanitizeFolderName {
			segment = sanitizePathSegment(segment)
		}
		subPath = filepath.Join(subPath, segment)
	}

	return filepath.Join(subPath, fileName), nil
}

// CheckMasterStorageSpace 检查主目录存储空间
func (s *StorageService) CheckMasterStorageSpace(masterID uint, fileSize int64) (bool, error) {
	masters, err := s.configRepo.GetMasterDirectories()
	if err != nil {
		return false, fmt.Errorf("failed to get master directories: %w", err)
	}

	var master *models.StorageDir
	for _, m := range masters {
		if m.ID == masterID {
			master = &m
			break
		}
	}

	if master == nil {
		return false, fmt.Errorf("master directory not found")
	}

	// 如果max_size为0，表示无限制
	if master.MaxSize == 0 {
		return true, nil
	}

	// 检查当前总大小 + 文件大小是否超过最大容量
	return (master.TotalSize + fileSize) <= master.MaxSize, nil
}
