package app

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	std_runtime "runtime"
	"strings"
	"time"

	"LocalSpace/app/agents"
	"LocalSpace/app/database"
	"LocalSpace/app/models"
	"LocalSpace/app/repositories"
	"LocalSpace/app/services"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx              context.Context
	db               *sql.DB
	fileService      *services.FileService
	storageService   *services.StorageService
	aiService        *services.AIService
	agentService     *services.AgentService
	configService    *services.ConfigService
	thumbnailService *services.ThumbnailService
	jobService       *services.JobService
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		// Initialize with empty ready state
	}
}

// isInitialized checks if the app services are initialized
func (a *App) isInitialized() bool {
	// Check if both config and file services are ready
	return a.configService != nil && a.fileService != nil
}

// waitForInitialization gives startup a short window to finish before
// file-related calls fall back to an empty result.
func (a *App) waitForInitialization(timeout time.Duration) bool {
	if a.isInitialized() {
		return true
	}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		time.Sleep(50 * time.Millisecond)
		if a.isInitialized() {
			return true
		}
	}

	return a.isInitialized()
}

// Startup is called when the app starts
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	a.initializeApp()
}

// Shutdown is called when the app is closing
func (a *App) Shutdown(ctx context.Context) {
	if a.jobService != nil {
		if err := a.jobService.PrepareForShutdown(); err != nil {
			fmt.Printf("Failed to prepare jobs for shutdown: %v\n", err)
		}
	}
	if a.db != nil {
		a.db.Close()
	}
}

// BeforeClose intercepts app shutdown when a recoverable job is still active.
func (a *App) BeforeClose(ctx context.Context) (prevent bool) {
	if a.jobService == nil {
		return false
	}

	hasJobs, count, err := a.jobService.HasBackgroundJobsForCloseProtection()
	if err != nil || !hasJobs {
		return false
	}

	response, dialogErr := runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
		Type:          runtime.QuestionDialog,
		Title:         "批量导入仍在运行",
		Message:       fmt.Sprintf("当前有 %d 个后台任务仍在运行，关闭程序可能中断导入。是否仍然退出？", count),
		Buttons:       []string{"继续运行", "仍然退出"},
		DefaultButton: "继续运行",
		CancelButton:  "继续运行",
	})
	if dialogErr != nil {
		fmt.Printf("Failed to show before-close dialog: %v\n", dialogErr)
		return false
	}

	if response == "继续运行" {
		return true
	}

	if err := a.jobService.PrepareForShutdown(); err != nil {
		fmt.Printf("Failed to update jobs before close: %v\n", err)
	}
	return false
}

// initializeApp initializes the application
func (a *App) initializeApp() {
	// Get executable directory for database storage
	execPath, err := os.Executable()
	if err != nil {
		fmt.Printf("Failed to get executable path: %v\n", err)
		return
	}
	execDir := filepath.Dir(execPath)

	// Ensure data directory exists in program directory
	dataDir := filepath.Join(execDir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		fmt.Printf("Failed to create data directory: %v\n", err)
		return
	}

	// Initialize database in program directory
	dbPath := filepath.Join(dataDir, "localspace.db")
	db, err := database.NewSQLiteDB(dbPath)
	if err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		return
	}
	a.db = db

	// Initialize repositories
	fileRepo := repositories.NewFileRepository(repositories.NewSQLiteDBWrapper(db))
	configRepo := repositories.NewConfigRepository(repositories.NewSQLiteDBWrapper(db))

	// Initialize services
	a.storageService = services.NewStorageService(configRepo)
	a.aiService = services.NewAIService(configRepo)

	// Initialize agent service
	a.agentService = services.NewAgentService(configRepo)

	// Initialize thumbnail service
	thumbnailDir := filepath.Join(dataDir, "thumbnails")
	a.thumbnailService = services.NewThumbnailService(thumbnailDir)

	a.fileService = services.NewFileService(fileRepo, a.storageService, a.aiService, a.thumbnailService)
	a.fileService.SetAgentService(a.agentService)
	a.configService = services.NewConfigService(configRepo)
	a.jobService = services.NewJobService(
		repositories.NewJobRepository(repositories.NewSQLiteDBWrapper(db)),
		a.fileService,
		a.aiService,
		a.agentService,
	)
	a.jobService.SetEventEmitter(func(eventName string, data interface{}) {
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, eventName, data)
		}
	})
	if err := a.jobService.NormalizeUnfinishedJobs(); err != nil {
		fmt.Printf("Failed to normalize unfinished jobs: %v\n", err)
	}

	if err := a.fileService.CleanupOrphanedTempFiles(24 * time.Hour); err != nil {
		fmt.Printf("Failed to cleanup orphaned temp files: %v\n", err)
	}

	fmt.Printf("LocalSpace initialized\n")
	fmt.Printf("Database: %s\n", dbPath)
	fmt.Printf("File storage: %s\n", getFileStoragePath(configRepo))
	fmt.Printf("Thumbnail cache: %s\n", thumbnailDir)
}

// getFileStoragePath gets the configured file storage path
func getFileStoragePath(configRepo *repositories.ConfigRepository) string {
	path, err := configRepo.Get("file_storage_path")
	if err == nil && path != "" {
		return path
	}
	// Return default message if not configured
	return "(not configured - requires user setup)"
}

// Configuration Methods

// InitConfig initializes the application configuration
func (a *App) InitConfig(fileStoragePath string) error {
	// Set the file storage path (user specified)
	if fileStoragePath != "" {
		// Validate path exists
		if _, err := os.Stat(fileStoragePath); os.IsNotExist(err) {
			// Try to create it
			if err := os.MkdirAll(fileStoragePath, 0755); err != nil {
				return fmt.Errorf("failed to create file storage directory: %w", err)
			}
		}

		// Save to config
		if err := a.configService.UpdateConfig("file_storage_path", fileStoragePath); err != nil {
			return fmt.Errorf("failed to save file storage path: %w", err)
		}

		// Update storage service
		a.storageService.SetStorageBasePath(fileStoragePath)
	}

	return nil
}

// GetConfig gets a configuration value
func (a *App) GetConfig(key string) (string, error) {
	return a.configService.GetConfig(key)
}

// UpdateConfig updates a configuration value
func (a *App) UpdateConfig(key, value string) error {
	return a.configService.UpdateConfig(key, value)
}

// File Type Methods

// GetFileTypes returns all supported file types
func (a *App) GetFileTypes() ([]models.FileType, error) {
	return a.configService.GetFileTypes()
}

// ParseFileType parses file type from extension
func (a *App) ParseFileType(extension string) (models.FileType, error) {
	return a.configService.ParseFileType(extension)
}

// File Management Methods

// ImportFile imports a file into LocalSpace (backward compatible)
func (a *App) ImportFile(filePath, fileName, description string, tags []string) error {
	return a.fileService.ImportFile(services.ImportFileRequest{
		FilePath:    filePath,
		FileName:    fileName,
		Description: description,
		Tags:        tags,
		Keywords:    "", // Default empty keywords for backward compatibility
	})
}

// ImportFileWithKeywords imports a file into LocalSpace with keywords
func (a *App) ImportFileWithKeywords(filePath, fileName, description string, tags []string, keywords string) error {
	return a.fileService.ImportFile(services.ImportFileRequest{
		FilePath:    filePath,
		FileName:    fileName,
		Description: description,
		Tags:        tags,
		Keywords:    keywords,
	})
}

// ImportFileWithMetadata imports a file with keywords and collection metadata.
func (a *App) ImportFileWithMetadata(filePath, fileName, description string, tags []string, keywords, collectionName string) error {
	return a.fileService.ImportFile(services.ImportFileRequest{
		FilePath:       filePath,
		FileName:       fileName,
		Description:    description,
		Tags:           tags,
		Keywords:       keywords,
		CollectionName: collectionName,
	})
}

// GetFiles returns a list of files
func (a *App) GetFiles(page, pageSize int, fileType string) ([]*models.File, error) {
	if !a.waitForInitialization(5 * time.Second) {
		return []*models.File{}, fmt.Errorf("app not initialized")
	}
	return a.fileService.ListFiles(services.FileFilter{
		Page:     page,
		PageSize: pageSize,
		FileType: fileType,
	})
}

// SearchFiles searches for files
func (a *App) SearchFiles(query string) ([]*models.File, error) {
	return a.fileService.SearchFiles(query)
}

// GetFile returns a single file by ID
func (a *App) GetFile(id uint) (*models.File, error) {
	return a.fileService.GetFile(id)
}

func (a *App) RefreshFile(id uint) (*models.File, error) {
	if !a.isInitialized() {
		return nil, fmt.Errorf("app not initialized")
	}
	return a.fileService.RefreshFile(id)
}

// DeleteFile deletes a file
func (a *App) DeleteFile(id uint) error {
	return a.fileService.DeleteFile(id)
}

// RenameFile renames a file
func (a *App) RenameFile(id uint, newName string) error {
	if !a.isInitialized() {
		return fmt.Errorf("app not initialized")
	}
	return a.fileService.RenameFile(id, newName)
}

// UpdateFileMetadata updates file tags and description.
func (a *App) UpdateFileMetadata(id uint, tags []string, description string) error {
	if !a.isInitialized() {
		return fmt.Errorf("app not initialized")
	}
	return a.fileService.UpdateFileMetadata(id, tags, description)
}

// OpenFile opens a file with a preferred app when configured, otherwise system default.
func (a *App) OpenFile(id uint) error {
	return a.OpenFileWithPreferredApp(id)
}

// OpenFileWithPreferredApp opens a file with a configured preferred application when available.
func (a *App) OpenFileWithPreferredApp(id uint) error {
	file, err := a.fileService.GetFile(id)
	if err != nil {
		return err
	}

	openWithConfig, err := a.configService.GetOpenWithConfig()
	if err != nil {
		return fmt.Errorf("failed to load preferred open config: %w", err)
	}

	appPath := resolvePreferredApp(file.FileName, file.FileType, openWithConfig)
	if appPath == "" {
		return a.openFileWithDefaultApp(file.FilePath)
	}
	if _, err := os.Stat(appPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("preferred app not found: %s", appPath)
		}
		return fmt.Errorf("failed to access preferred app: %w", err)
	}

	return exec.Command(appPath, file.FilePath).Start()
}

// OpenFileWithSystemDefault opens a file with the system default application.
func (a *App) OpenFileWithSystemDefault(id uint) error {
	file, err := a.fileService.GetFile(id)
	if err != nil {
		return err
	}

	return a.openFileWithDefaultApp(file.FilePath)
}

// OpenFileLocation opens the file's location in the file manager
func (a *App) OpenFileLocation(id uint) error {
	file, err := a.fileService.GetFile(id)
	if err != nil {
		return err
	}

	// Debug: print file info
	fmt.Printf("Opening location for file ID %d: %s\n", id, file.FileName)
	fmt.Printf("File path: %s\n", file.FilePath)
	fmt.Printf("File type: %s\n", file.FileType)

	return a.openFileLocation(file.FilePath)
}

// openFileWithDefaultApp opens a file with the default system application
func (a *App) openFileWithDefaultApp(filePath string) error {
	var cmd *exec.Cmd

	switch std_runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", filePath)
	case "darwin":
		cmd = exec.Command("open", filePath)
	case "linux":
		cmd = exec.Command("xdg-open", filePath)
	default:
		return fmt.Errorf("unsupported platform: %s", std_runtime.GOOS)
	}

	return cmd.Start()
}

// openFileLocation opens the file's location in the file manager
func (a *App) openFileLocation(filePath string) error {
	// Debug: print the path we're trying to open
	fmt.Printf("Opening location for path: %s\n", filePath)

	switch std_runtime.GOOS {
	case "windows":
		// Windows: use explorer with /select to select the file
		// Convert to absolute path first
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %w", err)
		}

		fmt.Printf("Opening location for path: %s\n", absPath)
		return exec.Command("explorer", "/select,", absPath).Start()
	case "darwin":
		// macOS: use open -R to reveal the file in Finder
		cmd := exec.Command("open", "-R", filePath)
		return cmd.Start()
	case "linux":
		// Linux: try to open the parent directory with the default file manager
		dirPath := filepath.Dir(filePath)
		cmd := exec.Command("xdg-open", dirPath)
		return cmd.Start()
	default:
		return fmt.Errorf("unsupported platform: %s", std_runtime.GOOS)
	}
}

func resolvePreferredApp(fileName, fileType string, config *models.OpenWithConfig) string {
	if config == nil {
		return ""
	}

	extension := strings.ToLower(filepath.Ext(fileName))
	if extension != "" && config.ByExtension != nil {
		if appPath := strings.TrimSpace(config.ByExtension[extension]); appPath != "" {
			return appPath
		}
	}
	if config.ByFileType != nil {
		if appPath := strings.TrimSpace(config.ByFileType[fileType]); appPath != "" {
			return appPath
		}
	}

	return ""
}

// AI Methods

// GetAIAnalysis gets AI analysis for a file
func (a *App) GetAIAnalysis(fileName, fileType, userKeywords string, userTags []string, userDescription string) (*models.AIAnalysis, error) {
	// Try to use agent service first if available
	if a.agentService != nil {
		analysis, err := a.agentService.AnalyzeMetadata(context.Background(), &agents.MetadataGenerationInput{
			FileName:        fileName,
			FileType:        fileType,
			UserKeywords:    userKeywords,
			UserTags:        userTags,
			UserDescription: userDescription,
		})
		if err != nil || analysis == nil {
			return &models.AIAnalysis{Tags: []string{}, Description: ""}, nil
		}

		return &models.AIAnalysis{
			Tags:        analysis.Tags,
			Description: analysis.Description,
		}, nil
	}

	// Fallback to old AI service if agent service is not available
	tags, err := a.aiService.GenerateTags(fileName, fileType)
	if err != nil {
		// AI failure is not critical, continue with empty tags
		tags = []string{}
	}

	description, err := a.aiService.GenerateDescription(fileName, fileType)
	if err != nil {
		// AI failure is not critical, continue with empty description
		description = ""
	}

	return &models.AIAnalysis{
		Tags:        tags,
		Description: description,
	}, nil
}

// Storage Methods

// GetStorageDirectories returns all storage directories
func (a *App) GetStorageDirectories() ([]models.StorageDir, error) {
	return a.storageService.GetStorageDirectories()
}

// AddStorageDirectory adds a new storage directory
func (a *App) AddStorageDirectory(path, fileType string) error {
	return a.storageService.AddStorageDirectory(path, fileType)
}

// RemoveStorageDirectory removes a storage directory
func (a *App) RemoveStorageDirectory(id uint) error {
	// Get the directory info before removing
	dirs, err := a.storageService.GetStorageDirectories()
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

	// Delete file records from database if directory path exists
	if dirToRemove.Path != "" {
		deletedCount, err := a.fileService.DeleteFilesByPath(dirToRemove.Path)
		if err != nil {
			// Log error but don't fail the removal
			fmt.Printf("Warning: failed to delete file records: %v\n", err)
		} else {
			fmt.Printf("Deleted %d file records from database\n", deletedCount)
		}
	}

	// Remove storage directory (which will clear physical directory contents)
	return a.storageService.RemoveStorageDirectory(id)
}

// ToggleStorageDirectory toggles the active status of a storage directory
func (a *App) ToggleStorageDirectory(id uint, isActive bool) error {
	return a.storageService.ToggleStorageDirectory(id, isActive)
}

// CheckStorageSpace checks if there's enough storage space for a file
func (a *App) CheckStorageSpace(fileSize int64) (bool, error) {
	return a.storageService.CheckStorageSpace(fileSize)
}

// GetMasterDirectories returns all master directories
func (a *App) GetMasterDirectories() ([]models.StorageDir, error) {
	return a.storageService.GetMasterDirectories()
}

// AddMasterDirectory adds a new master directory
func (a *App) AddMasterDirectory(path string, maxSizeGB float64) error {
	return a.storageService.AddMasterDirectory(path, maxSizeGB)
}

// SetDefaultMasterDirectory sets a master directory as default
func (a *App) SetDefaultMasterDirectory(id uint) error {
	return a.storageService.SetDefaultMasterDirectory(id)
}

// CheckPathConflict checks if a path conflicts with existing directories
func (a *App) CheckPathConflict(path string, excludeID uint) (bool, error) {
	if a.configService == nil {
		return false, fmt.Errorf("config service not initialized")
	}
	return a.configService.CheckPathConflict(path, excludeID)
}

// GetMasterStoragePathForFile returns the storage path for a file using master directory structure
func (a *App) GetMasterStoragePathForFile(masterID uint, fileType, collectionName, fileName string) (string, error) {
	return a.storageService.GetStoragePathForFileWithMaster(masterID, fileType, collectionName, fileName)
}

// CheckMasterStorageSpace checks storage space for a master directory
func (a *App) CheckMasterStorageSpace(masterID uint, fileSize int64) (bool, error) {
	return a.storageService.CheckMasterStorageSpace(masterID, fileSize)
}

// Theme Methods

// GetThemeConfig returns the current theme configuration
func (a *App) GetThemeConfig() (*models.ThemeConfig, error) {
	if !a.isInitialized() {
		// Return default theme config if not yet initialized
		return &models.ThemeConfig{
			ThemeMode:       "light",
			PrimaryColor:    "#2196F3",
			BackgroundColor: "#1B2636",
		}, nil
	}
	return a.configService.GetThemeConfig()
}

// UpdateThemeConfig updates the theme configuration
func (a *App) UpdateThemeConfig(config models.ThemeConfig) error {
	return a.configService.UpdateThemeConfig(config)
}

// GetOpenWithConfig returns the preferred open configuration.
func (a *App) GetOpenWithConfig() (*models.OpenWithConfig, error) {
	return a.configService.GetOpenWithConfig()
}

// UpdateOpenWithConfig updates the preferred open configuration.
func (a *App) UpdateOpenWithConfig(config models.OpenWithConfig) error {
	return a.configService.UpdateOpenWithConfig(config)
}

// GetStorageLayoutConfig returns the storage layout configuration.
func (a *App) GetStorageLayoutConfig() (*models.StorageLayoutConfig, error) {
	return a.configService.GetStorageLayoutConfig()
}

// UpdateStorageLayoutConfig updates the storage layout configuration.
func (a *App) UpdateStorageLayoutConfig(config models.StorageLayoutConfig) error {
	return a.configService.UpdateStorageLayoutConfig(config)
}

// AI Config Methods

// GetAIConfig returns the AI configuration
func (a *App) GetAIConfig() (*models.AIConfig, error) {
	return a.configService.GetAIConfig()
}

// UpdateAIConfig updates the AI configuration
func (a *App) UpdateAIConfig(config models.AIConfig) error {
	return a.configService.UpdateAIConfig(config)
}

// File Selection Method

// SelectFile opens a file selection dialog
func (a *App) SelectFile() (string, error) {
	dialog, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择文件",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "所有文件",
				Pattern:     "*.*",
			},
		},
	})

	if err != nil {
		return "", fmt.Errorf("file selection failed: %w", err)
	}

	return dialog, nil
}

// SelectFiles opens a multi-file selection dialog and returns file info.
func (a *App) SelectFiles() ([]models.SelectedFile, error) {
	paths, err := runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择要导入的文件",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "所有文件",
				Pattern:     "*.*",
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("file selection failed: %w", err)
	}

	files := make([]models.SelectedFile, 0, len(paths))
	for _, path := range paths {
		info, statErr := os.Stat(path)
		if statErr != nil {
			continue
		}
		files = append(files, models.SelectedFile{
			Name: filepath.Base(path),
			Path: path,
			Size: info.Size(),
		})
	}

	return files, nil
}

// SelectExecutable opens an executable selection dialog.
func (a *App) SelectExecutable() (string, error) {
	dialog, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择打开软件",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "可执行文件",
				Pattern:     "*.exe",
			},
			{
				DisplayName: "所有文件",
				Pattern:     "*.*",
			},
		},
	})

	if err != nil {
		return "", fmt.Errorf("executable selection failed: %w", err)
	}

	return dialog, nil
}

// SelectDirectory opens a directory selection dialog
func (a *App) SelectDirectory() (string, error) {
	dialog, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择目录",
	})

	if err != nil {
		return "", fmt.Errorf("directory selection failed: %w", err)
	}

	return dialog, nil
}

// GetFileMetadata extracts metadata from a file
func (a *App) GetFileMetadata(filePath, fileType string) (models.Metadata, error) {
	return a.fileService.ExtractMetadata(filePath, fileType)
}

// GenerateThumbnail generates a thumbnail for a file
func (a *App) GenerateThumbnail(fileID uint) (string, error) {
	return a.fileService.GetThumbnail(fileID)
}

// GetThumbnailCacheInfo returns information about the thumbnail cache
func (a *App) GetThumbnailCacheInfo() (count int, size int64, maxSize int64) {
	return a.thumbnailService.GetCacheInfo()
}

// ClearThumbnailCache clears all cached thumbnails
func (a *App) ClearThumbnailCache() error {
	return a.thumbnailService.ClearCache()
}

// GetThumbnailCacheStats returns detailed cache statistics
func (a *App) GetThumbnailCacheStats() map[string]interface{} {
	return a.thumbnailService.GetCacheStats()
}

// Duplicate Detection Methods

// CheckDuplicate checks if a file is a duplicate
func (a *App) CheckDuplicate(filePath string) (map[string]interface{}, error) {
	isDuplicate, duplicateID, duplicateName, err := a.fileService.CheckDuplicate(filePath)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"is_duplicate":   isDuplicate,
		"duplicate_id":   duplicateID,
		"duplicate_name": duplicateName,
		"checksum":       "", // Could be returned if needed
	}, nil
}

// GetDuplicateFiles finds all duplicate files
func (a *App) GetDuplicateFiles() (map[string][]map[string]interface{}, error) {
	duplicates, err := a.fileService.GetDuplicateFiles()
	if err != nil {
		return nil, err
	}

	// Convert to a format suitable for JSON serialization
	result := make(map[string][]map[string]interface{})
	for checksum, files := range duplicates {
		fileMaps := make([]map[string]interface{}, len(files))
		for i, file := range files {
			fileMaps[i] = map[string]interface{}{
				"id":         file.ID,
				"file_name":  file.FileName,
				"file_path":  file.FilePath,
				"file_type":  file.FileType,
				"file_size":  file.FileSize,
				"checksum":   file.Checksum,
				"created_at": file.CreatedAt,
			}
		}
		result[checksum] = fileMaps
	}

	return result, nil
}

// Job Methods

func (a *App) SubmitBatchImportJob(req models.BatchImportJobRequest) (*models.Job, error) {
	if a.jobService == nil {
		return nil, fmt.Errorf("job service not initialized")
	}
	return a.jobService.SubmitBatchImportJob(req)
}

func (a *App) GetActiveJobs() ([]*models.Job, error) {
	if a.jobService == nil {
		return []*models.Job{}, nil
	}
	return a.jobService.GetActiveJobs()
}

func (a *App) GetResumableJobs() ([]*models.Job, error) {
	if a.jobService == nil {
		return []*models.Job{}, nil
	}
	return a.jobService.GetResumableJobs()
}

func (a *App) GetJob(jobID uint) (*models.Job, error) {
	if a.jobService == nil {
		return nil, fmt.Errorf("job service not initialized")
	}
	return a.jobService.GetJob(jobID)
}

func (a *App) ListJobs(page, pageSize int, jobType string) (*models.JobListResponse, error) {
	if a.jobService == nil {
		return &models.JobListResponse{
			Items:    []*models.Job{},
			Page:     page,
			PageSize: pageSize,
			Total:    0,
		}, nil
	}
	return a.jobService.ListJobs(page, pageSize, jobType)
}

func (a *App) ResumeJob(jobID uint) error {
	if a.jobService == nil {
		return fmt.Errorf("job service not initialized")
	}
	return a.jobService.ResumeJob(jobID)
}

func (a *App) CancelJob(jobID uint) error {
	if a.jobService == nil {
		return fmt.Errorf("job service not initialized")
	}
	return a.jobService.CancelJob(jobID)
}

// Helper method to get current timestamp
func (a *App) getCurrentTimestamp() time.Time {
	return time.Now()
}
