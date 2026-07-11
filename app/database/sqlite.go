package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/glebarez/sqlite"
)

// SQLiteDB wraps the sql.DB with additional functionality
type SQLiteDB struct {
	db *sql.DB
}


// NewSQLiteDB creates a new SQLite database connection
func NewSQLiteDB(dbPath string) (*sql.DB, error) {
	// Ensure the directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open the database
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Run migrations
	if err := RunMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Configure connection pool
	config := DefaultConnectionPoolConfig()
	db.SetMaxOpenConns(config.MaxOpenConnections)
	db.SetMaxIdleConns(config.MaxIdleConnections)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// Insert initial data
	if err := insertInitialData(db); err != nil {
		return nil, fmt.Errorf("failed to insert initial data: %w", err)
	}

	return db, nil
}

// NewSQLiteDBWrapper creates a wrapper for the SQLiteDB
func NewSQLiteDBWrapper(db *sql.DB) *SQLiteDB {
	return &SQLiteDB{db: db}
}

// GetDB returns the underlying sql.DB
func (s *SQLiteDB) GetDB() *sql.DB {
	return s.db
}

// insertInitialData inserts initial data into the database
func insertInitialData(db *sql.DB) error {
	// Insert file types
	fileTypes := []string{
		`INSERT OR IGNORE INTO file_types (name, display_name, extensions, sub_types) VALUES ('video', '视频', '.mp4,.avi,.mkv,.mov,.wmv,.flv,.webm', '["mp4","avi","mkv","mov","wmv","flv","webm"]')`,
		`INSERT OR IGNORE INTO file_types (name, display_name, extensions, sub_types) VALUES ('document', '文档', '.pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.txt,.md', '["pdf","doc","docx","xls","xlsx","ppt","pptx","txt","md"]')`,
		`INSERT OR IGNORE INTO file_types (name, display_name, extensions, sub_types) VALUES ('music', '音乐', '.mp3,.wav,.flac,.aac,.ogg,.m4a', '["mp3","wav","flac","aac","ogg","m4a"]')`,
		`INSERT OR IGNORE INTO file_types (name, display_name, extensions, sub_types) VALUES ('game', '游戏', '.exe,.app,.dmg,.iso,.zip,.rar,.7z', '["exe","app","dmg","iso","zip","rar","7z"]')`,
		`INSERT OR IGNORE INTO file_types (name, display_name, extensions, sub_types) VALUES ('installer', '安装包', '.msi,.pkg,.deb,.rpm,.apk', '["msi","pkg","deb","rpm","apk"]')`,
		`INSERT OR IGNORE INTO file_types (name, display_name, extensions, sub_types) VALUES ('image', '镜像', '.iso,.img,.dmg,.vdi,.vmdk', '["iso","img","dmg","vdi","vmdk"]')`,
	}

	for _, sql := range fileTypes {
		if _, err := db.Exec(sql); err != nil {
			return fmt.Errorf("failed to insert file types: %w", err)
		}
	}

	// Insert default configs
	configs := []string{
		`INSERT OR IGNORE INTO configs (key, value, description) VALUES ('file_storage_path', '', 'File storage directory path')`,
		`INSERT OR IGNORE INTO configs (key, value, description) VALUES ('storage_initialized', 'false', 'Storage initialization status')`,
		`INSERT OR IGNORE INTO configs (key, value, description) VALUES ('ai_enabled', 'false', 'AI feature enabled')`,
		`INSERT OR IGNORE INTO configs (key, value, description) VALUES ('version', '1.0.0', 'Application version')`,
	}

	for _, sql := range configs {
		if _, err := db.Exec(sql); err != nil {
			return fmt.Errorf("failed to insert configs: %w", err)
		}
	}

	// Insert default theme config
	themeConfig := `INSERT OR IGNORE INTO theme_configs (theme_mode, primary_color, background_image) VALUES ('light', '#2196F3', '')`
	if _, err := db.Exec(themeConfig); err != nil {
		return fmt.Errorf("failed to insert theme config: %w", err)
	}

	return nil
}