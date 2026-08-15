package repositories

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"LocalSpace/app/models"
)

// ConfigRepository handles configuration data operations
type ConfigRepository struct {
	db *sql.DB
}

// NewConfigRepository creates a new ConfigRepository
func NewConfigRepository(dbWrapper *SQLiteDBWrapper) *ConfigRepository {
	return &ConfigRepository{db: dbWrapper.GetDB()}
}

// Get gets a configuration value by key
func (r *ConfigRepository) Get(key string) (string, error) {
	query := `SELECT value FROM configs WHERE key = ?`

	var value string
	err := r.db.QueryRow(query, key).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("config key not found: %s", key)
		}
		return "", fmt.Errorf("failed to get config: %w", err)
	}

	return value, nil
}

// Set sets a configuration value
func (r *ConfigRepository) Set(key, value string) error {
	query := `
		INSERT INTO configs (key, value, updated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(key) DO UPDATE SET
			value = excluded.value,
			updated_at = CURRENT_TIMESTAMP`

	_, err := r.db.Exec(query, key, value)
	if err != nil {
		return fmt.Errorf("failed to set config: %w", err)
	}

	return nil
}

func defaultOpenWithConfig() *models.OpenWithConfig {
	return &models.OpenWithConfig{
		ByFileType:  map[string]string{},
		ByExtension: map[string]string{},
	}
}

func defaultStorageLayoutConfig() *models.StorageLayoutConfig {
	return &models.StorageLayoutConfig{
		Strategy:           "type_collection",
		UnsortedFolderName: models.BuiltinUnsortedFolderName,
		SanitizeFolderName: true,
	}
}

// normalizeLegacyUnsortedFiles 将旧配置中的未分配目录名只归一化数据库语义，不搬动物理文件。
// 这样旧文件仍留在原目录，但不会因为设置改名而继续伪装成一个可编辑合集。
func (r *ConfigRepository) normalizeLegacyUnsortedFiles(legacyName string) error {
	legacyName = strings.TrimSpace(legacyName)
	if legacyName == "" || legacyName == models.BuiltinUnsortedFolderName {
		return nil
	}

	_, err := r.db.Exec(`
		UPDATE files
		SET collection_name = ?, modified_at = CURRENT_TIMESTAMP
		WHERE collection_id IS NULL AND TRIM(collection_name) = ?`,
		models.BuiltinUnsortedFolderName, legacyName)
	if err != nil {
		return fmt.Errorf("failed to normalize legacy unsorted files: %w", err)
	}
	return nil
}

// GetAll returns all configurations
func (r *ConfigRepository) GetAll() (map[string]string, error) {
	query := `SELECT key, value FROM configs`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all configs: %w", err)
	}
	defer rows.Close()

	configs := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("failed to scan config: %w", err)
		}
		configs[key] = value
	}

	return configs, nil
}

// GetOpenWithConfig returns the preferred app configuration.
func (r *ConfigRepository) GetOpenWithConfig() (*models.OpenWithConfig, error) {
	value, err := r.Get("open_with_config")
	if err != nil {
		if strings.Contains(err.Error(), "config key not found") {
			return defaultOpenWithConfig(), nil
		}
		return nil, err
	}

	config := defaultOpenWithConfig()
	if strings.TrimSpace(value) == "" {
		return config, nil
	}
	if err := json.Unmarshal([]byte(value), config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal open-with config: %w", err)
	}
	if config.ByFileType == nil {
		config.ByFileType = map[string]string{}
	}
	if config.ByExtension == nil {
		config.ByExtension = map[string]string{}
	}

	return config, nil
}

// SetOpenWithConfig saves the preferred app configuration.
func (r *ConfigRepository) SetOpenWithConfig(config *models.OpenWithConfig) error {
	if config == nil {
		config = defaultOpenWithConfig()
	}
	if config.ByFileType == nil {
		config.ByFileType = map[string]string{}
	}
	if config.ByExtension == nil {
		config.ByExtension = map[string]string{}
	}

	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal open-with config: %w", err)
	}

	return r.Set("open_with_config", string(data))
}

// GetStorageLayoutConfig returns the storage layout configuration.
func (r *ConfigRepository) GetStorageLayoutConfig() (*models.StorageLayoutConfig, error) {
	value, err := r.Get("storage_layout_config")
	if err != nil {
		if strings.Contains(err.Error(), "config key not found") {
			return defaultStorageLayoutConfig(), nil
		}
		return nil, err
	}

	config := defaultStorageLayoutConfig()
	if strings.TrimSpace(value) == "" {
		return config, nil
	}
	if err := json.Unmarshal([]byte(value), config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal storage layout config: %w", err)
	}
	if err := r.normalizeLegacyUnsortedFiles(config.UnsortedFolderName); err != nil {
		return nil, err
	}
	if config.Strategy == "" {
		config.Strategy = "type_collection"
	}
	// 兼容旧配置字段，但不再允许它改变内置未分配目录。
	config.UnsortedFolderName = models.BuiltinUnsortedFolderName

	return config, nil
}

// SetStorageLayoutConfig saves the storage layout configuration.
func (r *ConfigRepository) SetStorageLayoutConfig(config *models.StorageLayoutConfig) error {
	if config == nil {
		config = defaultStorageLayoutConfig()
	}
	if err := r.normalizeLegacyUnsortedFiles(config.UnsortedFolderName); err != nil {
		return err
	}
	if strings.TrimSpace(config.Strategy) == "" {
		config.Strategy = "type_collection"
	}
	config.UnsortedFolderName = models.BuiltinUnsortedFolderName

	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal storage layout config: %w", err)
	}

	return r.Set("storage_layout_config", string(data))
}

// GetFileTypes returns all file types
func (r *ConfigRepository) GetFileTypes() ([]models.FileType, error) {
	query := `SELECT id, name, display_name, extensions, sub_types, created_at FROM file_types ORDER BY name`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get file types: %w", err)
	}
	defer rows.Close()

	fileTypes := []models.FileType{}
	for rows.Next() {
		var ft models.FileType
		var subTypesJSON string

		err := rows.Scan(
			&ft.ID,
			&ft.Name,
			&ft.DisplayName,
			&ft.Extensions,
			&subTypesJSON,
			&ft.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan file type: %w", err)
		}

		// Unmarshal sub types
		if err := json.Unmarshal([]byte(subTypesJSON), &ft.SubTypes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal sub types: %w", err)
		}

		fileTypes = append(fileTypes, ft)
	}

	return fileTypes, nil
}

// ParseFileType parses file type from extension
func (r *ConfigRepository) ParseFileType(extension string) (models.FileType, error) {
	// Normalize extension (ensure it starts with a dot)
	if !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}

	query := `
		SELECT id, name, display_name, extensions, sub_types, created_at
		FROM file_types
		WHERE extensions LIKE ?`

	extSearch := "%" + extension + "%"

	var ft models.FileType
	var subTypesJSON string

	err := r.db.QueryRow(query, extSearch).Scan(
		&ft.ID,
		&ft.Name,
		&ft.DisplayName,
		&ft.Extensions,
		&subTypesJSON,
		&ft.CreatedAt,
	)

	if err != nil {
		return models.FileType{}, fmt.Errorf("file type not found for extension %s: %w", extension, err)
	}

	// Unmarshal sub types
	if err := json.Unmarshal([]byte(subTypesJSON), &ft.SubTypes); err != nil {
		return models.FileType{}, fmt.Errorf("failed to unmarshal sub types: %w", err)
	}

	return ft, nil
}

// GetAIConfig returns the AI configuration
func (r *ConfigRepository) GetAIConfig() (*models.AIConfig, error) {
	query := `SELECT id, api_key, model, base_url, enabled FROM ai_configs WHERE id = 1`

	var config models.AIConfig
	err := r.db.QueryRow(query).Scan(
		&config.ID,
		&config.APIKey,
		&config.Model,
		&config.BaseURL,
		&config.Enabled,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return defaultAIConfig(), nil
		}
		return nil, fmt.Errorf("failed to get AI config: %w", err)
	}

	populateAIConfigDefaults(&config)
	if err := r.loadOptionalAIConfigFields(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (r *ConfigRepository) loadOptionalAIConfigFields(config *models.AIConfig) error {
	var exists bool
	err := r.db.QueryRow(`SELECT COUNT(*) > 0 FROM pragma_table_info('ai_configs') WHERE name = 'enable_agent'`).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to inspect AI config schema: %w", err)
	}
	if !exists {
		return nil
	}

	query := `SELECT enable_agent, enable_web_search, max_tokens, timeout, web_search_provider, web_search_base_url, web_search_api_key, web_search_timeout, web_search_max_results FROM ai_configs WHERE id = 1`
	if err := r.db.QueryRow(query).Scan(
		&config.EnableAgent,
		&config.EnableWebSearch,
		&config.MaxTokens,
		&config.Timeout,
		&config.WebSearchProvider,
		&config.WebSearchBaseURL,
		&config.WebSearchAPIKey,
		&config.WebSearchTimeout,
		&config.WebSearchMaxResults,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("failed to get AI config optional fields: %w", err)
	}

	populateAIConfigDefaults(config)
	return nil
}

func defaultAIConfig() *models.AIConfig {
	return &models.AIConfig{
		ID:                  0,
		APIKey:              "",
		Model:               "gpt-3.5-turbo",
		BaseURL:             "https://api.openai.com/v1",
		Enabled:             false,
		EnableAgent:         false,
		EnableWebSearch:     false,
		MaxTokens:           500,
		Timeout:             30,
		WebSearchProvider:   "",
		WebSearchBaseURL:    "",
		WebSearchAPIKey:     "",
		WebSearchTimeout:    10,
		WebSearchMaxResults: 3,
	}
}

func populateAIConfigDefaults(config *models.AIConfig) {
	if config.Model == "" {
		config.Model = "gpt-3.5-turbo"
	}
	if config.BaseURL == "" {
		config.BaseURL = "https://api.openai.com/v1"
	}
	if config.MaxTokens == 0 {
		config.MaxTokens = 500
	}
	if config.Timeout == 0 {
		config.Timeout = 30
	}
	if config.WebSearchTimeout == 0 {
		config.WebSearchTimeout = 10
	}
	if config.WebSearchMaxResults == 0 {
		config.WebSearchMaxResults = 3
	}
}

// SetAIConfig sets the AI configuration
func (r *ConfigRepository) SetAIConfig(config *models.AIConfig) error {
	enableAgent := config.EnableAgent
	enableWebSearch := config.EnableWebSearch
	maxTokens := config.MaxTokens
	if maxTokens == 0 {
		maxTokens = 500
	}
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30
	}
	webSearchTimeout := config.WebSearchTimeout
	if webSearchTimeout == 0 {
		webSearchTimeout = 10
	}
	webSearchMaxResults := config.WebSearchMaxResults
	if webSearchMaxResults == 0 {
		webSearchMaxResults = 3
	}

	query := `
		INSERT INTO ai_configs (
			id, api_key, model, base_url, enabled,
			enable_agent, enable_web_search, max_tokens, timeout,
			web_search_provider, web_search_base_url, web_search_api_key, web_search_timeout, web_search_max_results,
			updated_at
		)
		VALUES (1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(id) DO UPDATE SET
			api_key = excluded.api_key,
			model = excluded.model,
			base_url = excluded.base_url,
			enabled = excluded.enabled,
			enable_agent = excluded.enable_agent,
			enable_web_search = excluded.enable_web_search,
			max_tokens = excluded.max_tokens,
			timeout = excluded.timeout,
			web_search_provider = excluded.web_search_provider,
			web_search_base_url = excluded.web_search_base_url,
			web_search_api_key = excluded.web_search_api_key,
			web_search_timeout = excluded.web_search_timeout,
			web_search_max_results = excluded.web_search_max_results,
			updated_at = CURRENT_TIMESTAMP`

	_, err := r.db.Exec(
		query,
		config.APIKey,
		config.Model,
		config.BaseURL,
		config.Enabled,
		enableAgent,
		enableWebSearch,
		maxTokens,
		timeout,
		config.WebSearchProvider,
		config.WebSearchBaseURL,
		config.WebSearchAPIKey,
		webSearchTimeout,
		webSearchMaxResults,
	)
	if err != nil {
		return fmt.Errorf("failed to set AI config: %w", err)
	}

	return nil
}

// GetThemeConfig returns the theme configuration
func (r *ConfigRepository) GetThemeConfig() (*models.ThemeConfig, error) {
	query := `SELECT id, theme_mode, primary_color, background_image FROM theme_configs WHERE id = 1`

	var config models.ThemeConfig
	err := r.db.QueryRow(query).Scan(
		&config.ID,
		&config.ThemeMode,
		&config.PrimaryColor,
		&config.BackgroundImage,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Return default config if not exists
			return &models.ThemeConfig{
				ID:              0,
				ThemeMode:       "light",
				PrimaryColor:    "#2196F3",
				BackgroundImage: "",
			}, nil
		}
		return nil, fmt.Errorf("failed to get theme config: %w", err)
	}

	return &config, nil
}

// SetThemeConfig sets the theme configuration
func (r *ConfigRepository) SetThemeConfig(config *models.ThemeConfig) error {
	query := `
		INSERT INTO theme_configs (id, theme_mode, primary_color, background_image, updated_at)
		VALUES (1, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(id) DO UPDATE SET
			theme_mode = excluded.theme_mode,
			primary_color = excluded.primary_color,
			background_image = excluded.background_image,
			updated_at = CURRENT_TIMESTAMP`

	_, err := r.db.Exec(query, config.ThemeMode, config.PrimaryColor, config.BackgroundImage)
	if err != nil {
		return fmt.Errorf("failed to set theme config: %w", err)
	}

	return nil
}

// GetStorageDirectories returns all storage directories
func (r *ConfigRepository) GetStorageDirectories() ([]models.StorageDir, error) {
	query := `
		SELECT id, path, file_type, current_size, max_size, is_active, created_at
		FROM storage_dirs
		ORDER BY file_type`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage directories: %w", err)
	}
	defer rows.Close()

	dirs := []models.StorageDir{}
	for rows.Next() {
		var dir models.StorageDir
		err := rows.Scan(
			&dir.ID,
			&dir.Path,
			&dir.FileType,
			&dir.CurrentSize,
			&dir.MaxSize,
			&dir.IsActive,
			&dir.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan storage directory: %w", err)
		}

		dirs = append(dirs, dir)
	}

	return dirs, nil
}

// AddStorageDirectory adds a new storage directory
func (r *ConfigRepository) AddStorageDirectory(dir *models.StorageDir) (uint, error) {
	var parentID interface{} = nil
	if dir.ParentID != nil {
		parentID = *dir.ParentID
	}

	query := `
		INSERT INTO storage_dirs (path, file_type, current_size, max_size, is_active, is_default, parent_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query, dir.Path, dir.FileType, dir.CurrentSize, dir.MaxSize, dir.IsActive, dir.IsDefault, parentID)
	if err != nil {
		return 0, fmt.Errorf("failed to add storage directory: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return uint(id), nil
}

// RemoveStorageDirectory removes a storage directory
func (r *ConfigRepository) RemoveStorageDirectory(id uint) error {
	fmt.Printf("Removing storage directory ID: %d\n", id)

	// 开启事务以确保删除操作的原子性
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 先删除所有子目录（parent_id = id）
	result, err := tx.Exec(`DELETE FROM storage_dirs WHERE parent_id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to remove subdirectories: %w", err)
	}
	subDirRows, _ := result.RowsAffected()
	fmt.Printf("Deleted %d subdirectories for parent ID: %d\n", subDirRows, id)

	// 然后删除主目录
	query := `DELETE FROM storage_dirs WHERE id = ?`
	result, err = tx.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to remove storage directory: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("storage directory not found")
	}

	fmt.Printf("Deleted master directory ID: %d\n", id)

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	fmt.Printf("Successfully removed storage directory ID: %d\n", id)
	return nil
}

// ToggleStorageDirectory toggles the active status of a storage directory
func (r *ConfigRepository) ToggleStorageDirectory(id uint, isActive bool) error {
	query := `UPDATE storage_dirs SET is_active = ? WHERE id = ?`

	result, err := r.db.Exec(query, isActive, id)
	if err != nil {
		return fmt.Errorf("failed to toggle storage directory: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("storage directory not found")
	}

	return nil
}

// UpdateStorageDirSize updates the size of a storage directory
func (r *ConfigRepository) UpdateStorageDirSize(id uint, sizeChange int64) error {
	query := `UPDATE storage_dirs SET current_size = current_size + ? WHERE id = ?`

	result, err := r.db.Exec(query, sizeChange, id)
	if err != nil {
		return fmt.Errorf("failed to update storage directory size: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("storage directory not found")
	}

	return nil
}

// GetMasterDirectories 获取所有主目录（parent_id IS NULL AND file_type = 'master'）
func (r *ConfigRepository) GetMasterDirectories() ([]models.StorageDir, error) {
	query := `
		SELECT id, path, file_type, current_size, max_size, is_active, is_default, parent_id, created_at
		FROM storage_dirs
		WHERE parent_id IS NULL AND file_type = 'master'
		ORDER BY is_default DESC, created_at ASC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get master directories: %w", err)
	}
	defer rows.Close()

	dirs := []models.StorageDir{}
	for rows.Next() {
		var dir models.StorageDir
		var parentID sql.NullInt64
		err := rows.Scan(
			&dir.ID, &dir.Path, &dir.FileType, &dir.CurrentSize,
			&dir.MaxSize, &dir.IsActive, &dir.IsDefault, &parentID,
			&dir.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan master directory: %w", err)
		}
		if parentID.Valid {
			pid := uint(parentID.Int64)
			dir.ParentID = &pid
		}
		dirs = append(dirs, dir)
	}
	return dirs, nil
}

// GetDefaultMasterDirectory 获取默认主目录
func (r *ConfigRepository) GetDefaultMasterDirectory() (*models.StorageDir, error) {
	query := `
		SELECT id, path, file_type, current_size, max_size, is_active, is_default, parent_id, created_at
		FROM storage_dirs
		WHERE parent_id IS NULL AND is_default = 1 AND (file_type = 'master' OR file_type = '')
		LIMIT 1`

	var dir models.StorageDir
	var parentID sql.NullInt64
	err := r.db.QueryRow(query).Scan(
		&dir.ID, &dir.Path, &dir.FileType, &dir.CurrentSize,
		&dir.MaxSize, &dir.IsActive, &dir.IsDefault, &parentID,
		&dir.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 没有默认目录
		}
		return nil, fmt.Errorf("failed to get default master directory: %w", err)
	}

	if parentID.Valid {
		pid := uint(parentID.Int64)
		dir.ParentID = &pid
	}
	return &dir, nil
}

// SetDefaultMasterDirectory 设置默认主目录（确保只有一个）
func (r *ConfigRepository) SetDefaultMasterDirectory(id uint) error {
	// 开启事务
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 取消所有默认标记
	_, err = tx.Exec(`UPDATE storage_dirs SET is_default = 0 WHERE parent_id IS NULL AND (file_type = 'master' OR file_type = '')`)
	if err != nil {
		return fmt.Errorf("failed to clear default flags: %w", err)
	}

	// 设置新的默认目录
	_, err = tx.Exec(`UPDATE storage_dirs SET is_default = 1 WHERE id = ? AND parent_id IS NULL AND (file_type = 'master' OR file_type = '')`, id)
	if err != nil {
		return fmt.Errorf("failed to set default directory: %w", err)
	}

	return tx.Commit()
}

// CheckPathConflict 检查路径冲突（包括父目录冲突）
func (r *ConfigRepository) CheckPathConflict(path string, excludeID uint) (bool, error) {
	normalizedPath := filepath.Clean(path)
	separator := string(filepath.Separator)

	// 只检查真正的主目录（parent_id IS NULL 且 file_type = 'master' 或 file_type = ''）
	// 忽略其他所有记录
	query := `SELECT id, path, file_type, parent_id FROM storage_dirs WHERE id != ?`
	rows, err := r.db.Query(query, excludeID)
	if err != nil {
		return false, fmt.Errorf("failed to check path conflict: %w", err)
	}
	defer rows.Close()

	fmt.Printf("Checking path conflict for: %s (excludeID: %d)\n", path, excludeID)

	for rows.Next() {
		var existingID uint
		var existingPath string
		var fileType string
		var parentID sql.NullInt64
		if err := rows.Scan(&existingID, &existingPath, &fileType, &parentID); err != nil {
			return false, fmt.Errorf("failed to scan path: %w", err)
		}

		// 只检查真正的主目录
		isMasterDir := parentID.Valid == false && (fileType == "master" || fileType == "")
		if !isMasterDir {
			continue // 跳过子目录和其他类型的目录
		}

		fmt.Printf("Checking against master directory ID %d, path: %s, file_type: %s\n", existingID, existingPath, fileType)

		normalizedExisting := filepath.Clean(existingPath)

		// 精确匹配检查
		if normalizedPath == normalizedExisting {
			fmt.Printf("Path conflict detected: exact match with existing path (ID: %d): %s\n", existingID, existingPath)
			return true, nil
		}

		// 检查新路径是否是现有路径的子目录
		// 例如：H:/test/sub 是 H:/test 的子目录
		if strings.HasPrefix(normalizedPath+separator, normalizedExisting+separator) {
			fmt.Printf("Path conflict detected: new path is subdirectory of existing (ID: %d): %s\n", existingID, existingPath)
			return true, nil
		}

		// 检查现有路径是否是新路径的子目录
		// 例如：H:/test 是 H:/test/sub 的父目录
		if strings.HasPrefix(normalizedExisting+separator, normalizedPath+separator) {
			fmt.Printf("Path conflict detected: existing path is subdirectory of new (ID: %d): %s\n", existingID, existingPath)
			return true, nil
		}

		// 兄弟目录（如 H:/test 和 H:/test2）不会匹配，因为它们不以对方+separator开头
	}

	fmt.Printf("No path conflict detected for: %s (excludeID: %d)\n", path, excludeID)
	return false, nil
}

// GetSubDirectories 获取主目录的所有子目录
func (r *ConfigRepository) GetSubDirectories(parentID uint) ([]models.StorageDir, error) {
	query := `
		SELECT id, path, file_type, current_size, max_size, is_active, is_default, parent_id, created_at
		FROM storage_dirs
		WHERE parent_id = ?
		ORDER BY file_type`

	rows, err := r.db.Query(query, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sub directories: %w", err)
	}
	defer rows.Close()

	dirs := []models.StorageDir{}
	for rows.Next() {
		var dir models.StorageDir
		var parentID sql.NullInt64
		err := rows.Scan(
			&dir.ID, &dir.Path, &dir.FileType, &dir.CurrentSize,
			&dir.MaxSize, &dir.IsActive, &dir.IsDefault, &parentID,
			&dir.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sub directory: %w", err)
		}
		if parentID.Valid {
			pid := uint(parentID.Int64)
			dir.ParentID = &pid
		}
		dirs = append(dirs, dir)
	}
	return dirs, nil
}

// CalculateMasterDirectorySize 计算主目录总大小
func (r *ConfigRepository) CalculateMasterDirectorySize(masterID uint) (int64, error) {
	query := `
		SELECT COALESCE(SUM(current_size), 0)
		FROM storage_dirs
		WHERE parent_id = ? OR id = ?`

	var totalSize int64
	err := r.db.QueryRow(query, masterID, masterID).Scan(&totalSize)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate master directory size: %w", err)
	}
	return totalSize, nil
}
