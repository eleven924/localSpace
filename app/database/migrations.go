package database

import (
	"database/sql"
	"fmt"
)

// Migration 数据库迁移
type Migration struct {
	Version int
	Name    string
	Up      func(*sql.DB) error
	Down    func(*sql.DB) error
}

// migrations 所有的数据库迁移
var migrations = []Migration{
	{
		Version: 1,
		Name:    "create_initial_tables",
		Up:      migration001_Up,
		Down:    migration001_Down,
	},
	{
		Version: 2,
		Name:    "add_indexes",
		Up:      migration002_Up,
		Down:    migration002_Down,
	},
	{
		Version: 3,
		Name:    "add_file_metadata_columns",
		Up:      migration003_Up,
		Down:    migration003_Down,
	},
	{
		Version: 4,
		Name:    "add_master_directory_support",
		Up:      migration004_Up,
		Down:    migration004_Down,
	},
	{
		Version: 5,
		Name:    "add_ai_config_runtime_fields",
		Up:      migration005_Up,
		Down:    migration005_Down,
	},
	{
		Version: 6,
		Name:    "add_ai_config_web_search_provider_fields",
		Up:      migration006_Up,
		Down:    migration006_Down,
	},
	{
		Version: 7,
		Name:    "add_file_collection_name",
		Up:      migration007_Up,
		Down:    migration007_Down,
	},
	{
		Version: 8,
		Name:    "add_job_system_tables",
		Up:      migration008_Up,
		Down:    migration008_Down,
	},
	{
		Version: 9,
		Name:    "add_collections_and_file_collection_id",
		Up:      migration009_Up,
		Down:    migration009_Down,
	},
}

// RunMigrations 运行数据库迁移
func RunMigrations(db *sql.DB) error {
	// 创建迁移历史表
	if err := createMigrationHistoryTable(db); err != nil {
		return fmt.Errorf("failed to create migration history table: %w", err)
	}

	// 执行每个迁移
	for _, migration := range migrations {
		if err := runMigration(db, migration); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", migration.Name, err)
		}
	}

	return nil
}

// createMigrationHistoryTable 创建迁移历史表
func createMigrationHistoryTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`

	_, err := db.Exec(query)
	return err
}

// runMigration 运行单个迁移
func runMigration(db *sql.DB, migration Migration) error {
	// 检查迁移是否已经执行
	var exists bool
	err := db.QueryRow("SELECT COUNT(*) > 0 FROM schema_migrations WHERE version = ?", migration.Version).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}

	if exists {
		// 迁移已经执行，跳过
		return nil
	}

	// 执行迁移
	if err := migration.Up(db); err != nil {
		return fmt.Errorf("failed to apply migration: %w", err)
	}

	// 记录迁移历史
	query := `INSERT INTO schema_migrations (version, name, applied_at) VALUES (?, ?, CURRENT_TIMESTAMP)`
	_, err = db.Exec(query, migration.Version, migration.Name)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return nil
}

// migration001_Up: 创建初始表
func migration001_Up(db *sql.DB) error {
	// 创建文件类型表
	query := `
		CREATE TABLE IF NOT EXISTS file_types (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			display_name TEXT NOT NULL,
			extensions TEXT NOT NULL,
			sub_types TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`

	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf("failed to create file_types table: %w", err)
	}

	// 创建配置表
	query = `
		CREATE TABLE IF NOT EXISTS configs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			key TEXT NOT NULL UNIQUE,
			value TEXT NOT NULL,
			description TEXT,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`

	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf("failed to create configs table: %w", err)
	}

	// 创建 AI 配置表
	query = `
		CREATE TABLE IF NOT EXISTS ai_configs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			api_key TEXT NOT NULL,
			model TEXT NOT NULL,
			base_url TEXT NOT NULL,
			enabled BOOLEAN DEFAULT 0,
			enable_agent BOOLEAN DEFAULT 0,
			enable_web_search BOOLEAN DEFAULT 0,
			max_tokens INTEGER DEFAULT 500,
			timeout INTEGER DEFAULT 30,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`

	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf("failed to create ai_configs table: %w", err)
	}

	// 创建主题配置表
	query = `
		CREATE TABLE IF NOT EXISTS theme_configs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			theme_mode TEXT NOT NULL,
			primary_color TEXT NOT NULL,
			background_image TEXT,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`

	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf("failed to create theme_configs table: %w", err)
	}

	// 创建存储目录表
	query = `
		CREATE TABLE IF NOT EXISTS storage_dirs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			path TEXT NOT NULL UNIQUE,
			file_type TEXT NOT NULL,
			current_size INTEGER DEFAULT 0,
			max_size INTEGER,
			is_active BOOLEAN DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`

	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf("failed to create storage_dirs table: %w", err)
	}

	// 创建标签表
	query = `
		CREATE TABLE IF NOT EXISTS tags (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			color TEXT DEFAULT '#2196F3',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`

	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf("failed to create tags table: %w", err)
	}

	// 创建文件表
	query = `
		CREATE TABLE IF NOT EXISTS files (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_name TEXT NOT NULL,
			file_path TEXT NOT NULL UNIQUE,
			file_type TEXT NOT NULL,
			file_sub_type TEXT,
			file_size INTEGER NOT NULL,
			tags TEXT NOT NULL DEFAULT '[]',
			description TEXT,
			metadata TEXT DEFAULT '{}',
			thumbnail TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			modified_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`

	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf("failed to create files table: %w", err)
	}

	return nil
}

// migration001_Down: 删除初始表
func migration001_Down(db *sql.DB) error {
	tables := []string{"files", "tags", "storage_dirs", "theme_configs", "ai_configs", "configs", "file_types"}

	for _, table := range tables {
		query := fmt.Sprintf("DROP TABLE IF EXISTS %s", table)
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}

	return nil
}

// migration002_Up: 添加索引
func migration002_Up(db *sql.DB) error {
	// 创建文件类型索引
	_, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_files_file_type ON files(file_type)`)
	if err != nil {
		return fmt.Errorf("failed to create file_type index: %w", err)
	}

	// 创建创建时间索引
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_files_created_at ON files(created_at)`)
	if err != nil {
		return fmt.Errorf("failed to create created_at index: %w", err)
	}

	// 创建存储目录文件类型索引
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_storage_dirs_file_type ON storage_dirs(file_type)`)
	if err != nil {
		return fmt.Errorf("failed to create storage_dir file_type index: %w", err)
	}

	// 创建存储目录激活状态索引
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_storage_dirs_is_active ON storage_dirs(is_active)`)
	if err != nil {
		return fmt.Errorf("failed to create storage_dir is_active index: %w", err)
	}

	return nil
}

// migration002_Down: 删除索引
func migration002_Down(db *sql.DB) error {
	indexes := []string{
		"idx_files_file_type",
		"idx_files_created_at",
		"idx_storage_dirs_file_type",
		"idx_storage_dirs_is_active",
	}

	for _, index := range indexes {
		_, err := db.Exec(fmt.Sprintf("DROP INDEX IF EXISTS %s", index))
		if err != nil {
			return fmt.Errorf("failed to drop index %s: %w", index, err)
		}
	}

	return nil
}

// migration003_Up: 添加文件元数据列
func migration003_Up(db *sql.DB) error {
	// 检查列是否已存在
	var columnExists bool
	err := db.QueryRow(`
		SELECT COUNT(*) > 0
		FROM pragma_table_info('files')
		WHERE name = 'original_name'
	`).Scan(&columnExists)

	if err != nil {
		return fmt.Errorf("failed to check column existence: %w", err)
	}

	if !columnExists {
		// 添加原始文件名列
		_, err := db.Exec(`ALTER TABLE files ADD COLUMN original_name TEXT`)
		if err != nil {
			return fmt.Errorf("failed to add original_name column: %w", err)
		}
	}

	// 检查 checksum 列
	err = db.QueryRow(`
		SELECT COUNT(*) > 0
		FROM pragma_table_info('files')
		WHERE name = 'checksum'
	`).Scan(&columnExists)

	if err != nil {
		return fmt.Errorf("failed to check column existence: %w", err)
	}

	if !columnExists {
		// 添加校验和列
		_, err = db.Exec(`ALTER TABLE files ADD COLUMN checksum TEXT`)
		if err != nil {
			return fmt.Errorf("failed to add checksum column: %w", err)
		}
	}

	// 检查 is_deleted 列
	err = db.QueryRow(`
		SELECT COUNT(*) > 0
		FROM pragma_table_info('files')
		WHERE name = 'is_deleted'
	`).Scan(&columnExists)

	if err != nil {
		return fmt.Errorf("failed to check column existence: %w", err)
	}

	if !columnExists {
		// 添加软删除列
		_, err = db.Exec(`ALTER TABLE files ADD COLUMN is_deleted BOOLEAN DEFAULT 0`)
		if err != nil {
			return fmt.Errorf("failed to add is_deleted column: %w", err)
		}
	}

	// 检查 deleted_at 列
	err = db.QueryRow(`
		SELECT COUNT(*) > 0
		FROM pragma_table_info('files')
		WHERE name = 'deleted_at'
	`).Scan(&columnExists)

	if err != nil {
		return fmt.Errorf("failed to check column existence: %w", err)
	}

	if !columnExists {
		// 添加删除时间列
		_, err = db.Exec(`ALTER TABLE files ADD COLUMN deleted_at DATETIME`)
		if err != nil {
			return fmt.Errorf("failed to add deleted_at column: %w", err)
		}
	}

	return nil
}

// migration003_Down: 删除文件元数据列
func migration003_Down(db *sql.DB) error {
	columns := []string{
		"original_name",
		"checksum",
		"is_deleted",
		"deleted_at",
	}

	for _, column := range columns {
		_, err := db.Exec(fmt.Sprintf("ALTER TABLE files DROP COLUMN IF EXISTS %s", column))
		if err != nil {
			return fmt.Errorf("failed to drop column %s: %w", column, err)
		}
	}

	return nil
}

// GetCurrentMigrationVersion 获取当前迁移版本
func GetCurrentMigrationVersion(db *sql.DB) (int, error) {
	var version int
	err := db.QueryRow("SELECT MAX(version) FROM schema_migrations").Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}

// GetPendingMigrations 获取待执行的迁移
func GetPendingMigrations() []Migration {
	return migrations
}

// migration004_Up: 添加主目录支持
func migration004_Up(db *sql.DB) error {
	// 添加is_default字段（确保只有一个默认主目录）
	_, err := db.Exec(`ALTER TABLE storage_dirs ADD COLUMN is_default BOOLEAN DEFAULT 0`)
	if err != nil {
		return fmt.Errorf("failed to add is_default column: %w", err)
	}

	// 添加parent_id字段（支持子目录）
	_, err = db.Exec(`ALTER TABLE storage_dirs ADD COLUMN parent_id INTEGER DEFAULT NULL`)
	if err != nil {
		return fmt.Errorf("failed to add parent_id column: %w", err)
	}

	// 添加唯一约束防止路径冲突
	_, err = db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_storage_dirs_path ON storage_dirs(path)`)
	if err != nil {
		return fmt.Errorf("failed to create path unique index: %w", err)
	}

	// 添加复合索引优化查询
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_storage_dirs_parent_default ON storage_dirs(parent_id, is_default)`)
	if err != nil {
		return fmt.Errorf("failed to create parent_default index: %w", err)
	}

	return nil
}

// migration004_Down: 回滚主目录支持
func migration004_Down(db *sql.DB) error {
	// 删除索引
	_, err := db.Exec(`DROP INDEX IF EXISTS idx_storage_dirs_parent_default`)
	if err != nil {
		return fmt.Errorf("failed to drop parent_default index: %w", err)
	}

	_, err = db.Exec(`DROP INDEX IF EXISTS idx_storage_dirs_path`)
	if err != nil {
		return fmt.Errorf("failed to drop path index: %w", err)
	}

	// SQLite不支持DROP COLUMN，需要重建表（略）
	return fmt.Errorf("SQLite rollback not supported for column additions")
}

// migration005_Up: add AI runtime config fields
func migration005_Up(db *sql.DB) error {
	columns := []struct {
		name       string
		definition string
	}{
		{name: "enable_agent", definition: "BOOLEAN DEFAULT 0"},
		{name: "enable_web_search", definition: "BOOLEAN DEFAULT 0"},
		{name: "max_tokens", definition: "INTEGER DEFAULT 500"},
		{name: "timeout", definition: "INTEGER DEFAULT 30"},
	}

	for _, column := range columns {
		var exists bool
		err := db.QueryRow(`
			SELECT COUNT(*) > 0
			FROM pragma_table_info('ai_configs')
			WHERE name = ?
		`, column.name).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check %s column existence: %w", column.name, err)
		}
		if exists {
			continue
		}
		if _, err := db.Exec(fmt.Sprintf("ALTER TABLE ai_configs ADD COLUMN %s %s", column.name, column.definition)); err != nil {
			return fmt.Errorf("failed to add %s column: %w", column.name, err)
		}
	}

	return nil
}

func migration005_Down(db *sql.DB) error {
	return fmt.Errorf("SQLite rollback not supported for column additions")
}

// migration006_Up: add AI web search config fields
func migration006_Up(db *sql.DB) error {
	columns := []struct {
		name       string
		definition string
	}{
		{name: "web_search_provider", definition: "TEXT DEFAULT ''"},
		{name: "web_search_base_url", definition: "TEXT DEFAULT ''"},
		{name: "web_search_api_key", definition: "TEXT DEFAULT ''"},
		{name: "web_search_timeout", definition: "INTEGER DEFAULT 10"},
		{name: "web_search_max_results", definition: "INTEGER DEFAULT 3"},
	}

	for _, column := range columns {
		var exists bool
		err := db.QueryRow(`
			SELECT COUNT(*) > 0
			FROM pragma_table_info('ai_configs')
			WHERE name = ?
		`, column.name).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check %s column existence: %w", column.name, err)
		}
		if exists {
			continue
		}
		if _, err := db.Exec(fmt.Sprintf("ALTER TABLE ai_configs ADD COLUMN %s %s", column.name, column.definition)); err != nil {
			return fmt.Errorf("failed to add %s column: %w", column.name, err)
		}
	}

	return nil
}

func migration006_Down(db *sql.DB) error {
	return fmt.Errorf("SQLite rollback not supported for column additions")
}

// migration007_Up: add file collection name column
func migration007_Up(db *sql.DB) error {
	var exists bool
	err := db.QueryRow(`
		SELECT COUNT(*) > 0
		FROM pragma_table_info('files')
		WHERE name = 'collection_name'
	`).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check collection_name column existence: %w", err)
	}
	if exists {
		return nil
	}

	if _, err := db.Exec(`ALTER TABLE files ADD COLUMN collection_name TEXT DEFAULT ''`); err != nil {
		return fmt.Errorf("failed to add collection_name column: %w", err)
	}

	return nil
}

func migration007_Down(db *sql.DB) error {
	return fmt.Errorf("SQLite rollback not supported for column additions")
}

func migration008_Up(db *sql.DB) error {
	jobsTable := `
		CREATE TABLE IF NOT EXISTS jobs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			job_type TEXT NOT NULL,
			status TEXT NOT NULL,
			title TEXT NOT NULL,
			payload TEXT,
			result TEXT,
			progress_total INTEGER NOT NULL DEFAULT 0,
			progress_completed INTEGER NOT NULL DEFAULT 0,
			progress_message TEXT NOT NULL DEFAULT '',
			exclusive_key TEXT NOT NULL DEFAULT '',
			can_resume BOOLEAN NOT NULL DEFAULT 0,
			started_at TEXT,
			heartbeat_at TEXT,
			finished_at TEXT,
			timeout_at TEXT,
			error_message TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`
	if _, err := db.Exec(jobsTable); err != nil {
		return fmt.Errorf("failed to create jobs table: %w", err)
	}

	itemsTable := `
		CREATE TABLE IF NOT EXISTS batch_import_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			job_id INTEGER NOT NULL,
			source_path TEXT NOT NULL,
			display_name TEXT NOT NULL,
			detected_file_type TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL,
			item_index INTEGER NOT NULL,
			temp_path TEXT NOT NULL DEFAULT '',
			final_path TEXT NOT NULL DEFAULT '',
			expected_size INTEGER NOT NULL DEFAULT 0,
			bytes_copied INTEGER NOT NULL DEFAULT 0,
			checksum TEXT NOT NULL DEFAULT '',
			error_message TEXT NOT NULL DEFAULT '',
			started_at TEXT,
			finished_at TEXT,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(job_id) REFERENCES jobs(id) ON DELETE CASCADE
		);`
	if _, err := db.Exec(itemsTable); err != nil {
		return fmt.Errorf("failed to create batch_import_items table: %w", err)
	}

	if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status)`); err != nil {
		return fmt.Errorf("failed to create jobs status index: %w", err)
	}
	if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_jobs_type_status ON jobs(job_type, status)`); err != nil {
		return fmt.Errorf("failed to create jobs type/status index: %w", err)
	}
	if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_jobs_exclusive_status ON jobs(exclusive_key, status)`); err != nil {
		return fmt.Errorf("failed to create jobs exclusive/status index: %w", err)
	}
	if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_batch_import_items_job_index ON batch_import_items(job_id, item_index)`); err != nil {
		return fmt.Errorf("failed to create batch_import_items job/index index: %w", err)
	}

	return nil
}

func migration008_Down(db *sql.DB) error {
	_, _ = db.Exec(`DROP INDEX IF EXISTS idx_batch_import_items_job_index`)
	_, _ = db.Exec(`DROP INDEX IF EXISTS idx_jobs_exclusive_status`)
	_, _ = db.Exec(`DROP INDEX IF EXISTS idx_jobs_type_status`)
	_, _ = db.Exec(`DROP INDEX IF EXISTS idx_jobs_status`)
	_, _ = db.Exec(`DROP TABLE IF EXISTS batch_import_items`)
	_, _ = db.Exec(`DROP TABLE IF EXISTS jobs`)
	return nil
}

func migration009_Up(db *sql.DB) error {
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS collections (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		return fmt.Errorf("failed to create collections table: %w", err)
	}

	var exists bool
	if err := db.QueryRow(`
		SELECT COUNT(*) > 0
		FROM pragma_table_info('files')
		WHERE name = 'collection_id'
	`).Scan(&exists); err != nil {
		return fmt.Errorf("failed to check collection_id column: %w", err)
	}
	if !exists {
		if _, err := db.Exec(`ALTER TABLE files ADD COLUMN collection_id INTEGER DEFAULT NULL`); err != nil {
			return fmt.Errorf("failed to add collection_id column: %w", err)
		}
	}

	if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_files_collection_id ON files(collection_id)`); err != nil {
		return fmt.Errorf("failed to create collection_id index: %w", err)
	}
	return nil
}

func migration009_Down(db *sql.DB) error {
	return fmt.Errorf("SQLite rollback not supported for collection_id addition")
}
