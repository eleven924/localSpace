package services

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"LocalSpace/app/models"
	"LocalSpace/app/repositories"
)

// ConfigService handles configuration operations
type ConfigService struct {
	configRepo *repositories.ConfigRepository
}

// NewConfigService creates a new ConfigService
func NewConfigService(configRepo *repositories.ConfigRepository) *ConfigService {
	return &ConfigService{configRepo: configRepo}
}

// InitConfig initializes the application configuration
func (s *ConfigService) InitConfig(fileStoragePath string) error {
	// Validate file storage path
	if fileStoragePath == "" {
		return fmt.Errorf("file storage path cannot be empty")
	}

	// Ensure directory exists
	if err := os.MkdirAll(fileStoragePath, 0755); err != nil {
		return fmt.Errorf("failed to create file storage directory: %w", err)
	}

	// Set file storage path
	if err := s.configRepo.Set("file_storage_path", fileStoragePath); err != nil {
		return fmt.Errorf("failed to set file storage path: %w", err)
	}

	// Set storage initialized flag
	if err := s.configRepo.Set("storage_initialized", "true"); err != nil {
		return fmt.Errorf("failed to set storage initialized: %w", err)
	}

	return nil
}

// GetConfig gets a configuration value
func (s *ConfigService) GetConfig(key string) (string, error) {
	return s.configRepo.Get(key)
}

// UpdateConfig updates a configuration value
func (s *ConfigService) UpdateConfig(key, value string) error {
	return s.configRepo.Set(key, value)
}

// GetFileTypes returns all file types
func (s *ConfigService) GetFileTypes() ([]models.FileType, error) {
	return s.configRepo.GetFileTypes()
}

// ParseFileType parses file type from extension
func (s *ConfigService) ParseFileType(extension string) (models.FileType, error) {
	return s.configRepo.ParseFileType(extension)
}

// GetAIConfig returns the AI configuration
func (s *ConfigService) GetAIConfig() (*models.AIConfig, error) {
	return s.configRepo.GetAIConfig()
}

// UpdateAIConfig updates the AI configuration
func (s *ConfigService) UpdateAIConfig(config models.AIConfig) error {
	return s.configRepo.SetAIConfig(&config)
}

// GetThemeConfig returns the theme configuration
func (s *ConfigService) GetThemeConfig() (*models.ThemeConfig, error) {
	return s.configRepo.GetThemeConfig()
}

// UpdateThemeConfig updates the theme configuration
func (s *ConfigService) UpdateThemeConfig(config models.ThemeConfig) error {
	return s.configRepo.SetThemeConfig(&config)
}

// GetOpenWithConfig returns the preferred open configuration.
func (s *ConfigService) GetOpenWithConfig() (*models.OpenWithConfig, error) {
	return s.configRepo.GetOpenWithConfig()
}

// UpdateOpenWithConfig updates the preferred open configuration.
func (s *ConfigService) UpdateOpenWithConfig(config models.OpenWithConfig) error {
	return s.configRepo.SetOpenWithConfig(&config)
}

// GetStorageLayoutConfig returns the storage layout configuration.
func (s *ConfigService) GetStorageLayoutConfig() (*models.StorageLayoutConfig, error) {
	return s.configRepo.GetStorageLayoutConfig()
}

// UpdateStorageLayoutConfig updates the storage layout configuration.
func (s *ConfigService) UpdateStorageLayoutConfig(config models.StorageLayoutConfig) error {
	return s.configRepo.SetStorageLayoutConfig(&config)
}

// IsStorageInitialized checks if storage is initialized
func (s *ConfigService) IsStorageInitialized() (bool, error) {
	value, err := s.configRepo.Get("storage_initialized")
	if err != nil {
		return false, err
	}
	return value == "true", nil
}

// GetStoragePath returns the storage path
func (s *ConfigService) GetStoragePath() (string, error) {
	value, err := s.configRepo.Get("storage_path")
	if err != nil {
		return "", err
	}
	return value, nil
}

// CheckPathConflict checks if a path conflicts with existing directories
func (s *ConfigService) CheckPathConflict(path string, excludeID uint) (bool, error) {
	return s.configRepo.CheckPathConflict(path, excludeID)
}

const jobRetentionConfigKey = "job_retention_config"

func defaultJobRetentionConfig() *models.JobRetentionConfig {
	return &models.JobRetentionConfig{
		ID:       1,
		Enabled:  false,
		MaxCount: 0,
		MaxDays:  0,
	}
}

// GetJobRetentionConfig returns the job retention configuration.
func (s *ConfigService) GetJobRetentionConfig() (*models.JobRetentionConfig, error) {
	value, err := s.configRepo.Get(jobRetentionConfigKey)
	if err != nil {
		if strings.Contains(err.Error(), "config key not found") {
			return defaultJobRetentionConfig(), nil
		}
		return nil, err
	}
	if strings.TrimSpace(value) == "" {
		return defaultJobRetentionConfig(), nil
	}
	var config models.JobRetentionConfig
	if err := json.Unmarshal([]byte(value), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job retention config: %w", err)
	}
	if config.ID == 0 {
		config.ID = 1
	}
	return &config, nil
}

// UpdateJobRetentionConfig updates the job retention configuration.
func (s *ConfigService) UpdateJobRetentionConfig(config models.JobRetentionConfig) error {
	if config.MaxCount < 0 {
		config.MaxCount = 0
	}
	if config.MaxDays < 0 {
		config.MaxDays = 0
	}
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal job retention config: %w", err)
	}
	return s.configRepo.Set(jobRetentionConfigKey, string(data))
}
