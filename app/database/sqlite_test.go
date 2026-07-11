package database

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"database/sql"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/glebarez/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	// 创建临时数据库文件
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// 删除临时数据库（如果存在）
	os.Remove(dbPath)

	// 创建数据库
	db, err := NewSQLiteDB(dbPath)
	require.NoError(t, err)
	require.NotNil(t, db)

	return db
}

func setupBenchmarkDB(b *testing.B) *sql.DB {
	// 创建临时数据库文件
	tmpDir := b.TempDir()
	dbPath := filepath.Join(tmpDir, "bench.db")

	// 删除临时数据库（如果存在）
	os.Remove(dbPath)

	// 创建数据库
	db, err := NewSQLiteDB(dbPath)
	if err != nil {
		b.Fatal(err)
	}

	return db
}

func TestNewSQLiteDB(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// 测试数据库是否正常工作
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM file_types").Scan(&count)
	require.NoError(t, err)
	assert.Greater(t, count, 0, "Should have some file types inserted")
}

func TestConnectionPool(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// 测试连接池配置
	wrapper := NewSQLiteDBWrapper(db)
	config := DefaultConnectionPoolConfig()

	err := wrapper.SetConnectionPool(config)
	require.NoError(t, err)

	// 获取连接池状态
	stats := wrapper.GetConnectionStats()
	assert.NotNil(t, stats)
	assert.Equal(t, config.MaxOpenConnections, stats.MaxOpenConnections)
}

func TestTransaction(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// 测试事务执行
	callback := func(tx *Transaction) error {
		// 在事务中插入测试数据
		_, err := tx.Exec("INSERT INTO configs (key, value) VALUES (?, ?)", "test_key", "test_value")
		return err
	}

	err := ExecuteInTransaction(db, callback)
	require.NoError(t, err)

	// 验证数据已插入
	var value string
	err = db.QueryRow("SELECT value FROM configs WHERE key = ?", "test_key").Scan(&value)
	require.NoError(t, err)
	assert.Equal(t, "test_value", value)
}

func TestTransactionRollback(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// 测试事务回滚
	callback := func(tx *Transaction) error {
		_, err := tx.Exec("INSERT INTO configs (key, value) VALUES (?, ?)", "rollback_test", "value")
		if err != nil {
			return err
		}
		// 故意返回错误以触发回滚
		return assert.AnError
	}

	err := ExecuteInTransaction(db, callback)
	require.Error(t, err)

	// 验证数据未插入
	var value string
	err = db.QueryRow("SELECT value FROM configs WHERE key = ?", "rollback_test").Scan(&value)
	require.Error(t, err)
}

func TestValidator(t *testing.T) {
	validator := NewValidator()

	// 测试字符串验证
	validator.ValidateString("name", "test", 1, 10, true)
	assert.False(t, validator.HasErrors(), "Should have no errors for valid string")

	// 测试空字符串（required）
	validator = NewValidator()
	validator.ValidateString("name", "", 1, 10, true)
	assert.True(t, validator.HasErrors(), "Should have error for empty required string")

	// 测试字符串太短
	validator = NewValidator()
	validator.ValidateString("name", "x", 3, 10, true)
	assert.True(t, validator.HasErrors(), "Should have error for string too short")

	// 测试字符串太长
	validator = NewValidator()
	validator.ValidateString("name", "very_long_name_exceeds_max", 1, 10, true)
	assert.True(t, validator.HasErrors(), "Should have error for string too long")

	// 测试标签验证
	validator = NewValidator()
	validator.ValidateTags("tags", []string{"work", "important"}, 10)
	assert.False(t, validator.HasErrors(), "Should have no errors for valid tags")

	// 测试标签太多
	validator = NewValidator()
	validator.ValidateTags("tags", make([]string, 15), 10)
	assert.True(t, validator.HasErrors(), "Should have error for too many tags")

	// 测试空标签
	validator = NewValidator()
	validator.ValidateTags("tags", []string{"work", ""}, 10)
	assert.True(t, validator.HasErrors(), "Should have error for empty tag")

	// 测试文件路径验证
	validator = NewValidator()
	validator.ValidateFilePath("path", "/valid/path/file.txt", true)
	assert.False(t, validator.HasErrors(), "Should have no errors for valid path")

	// 测试文件路径包含非法字符
	validator = NewValidator()
	validator.ValidateFilePath("path", "/invalid/path/file<>.txt", true)
	assert.True(t, validator.HasErrors(), "Should have error for path with invalid characters")
}

func TestMigrations(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_migrations.db")

	// 删除临时数据库（如果存在）
	os.Remove(dbPath)

	// 创建数据库（会自动运行迁移）
	db, err := NewSQLiteDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	// 检查迁移历史表是否存在
	var tableExists int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='schema_migrations'").Scan(&tableExists)
	require.NoError(t, err)
	assert.Equal(t, 1, tableExists, "Migration history table should exist")

	// 检查迁移版本
	var version int
	err = db.QueryRow("SELECT MAX(version) FROM schema_migrations").Scan(&version)
	require.NoError(t, err)
	assert.Greater(t, version, 0, "Should have at least one migration applied")

	// 检查所有表是否已创建
	tables := []string{"file_types", "configs", "ai_configs", "theme_configs", "storage_dirs", "tags", "files"}
	for _, table := range tables {
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
		require.NoError(t, err, "Table %s should exist", table)
		assert.Equal(t, 1, count, "Table %s should exist", table)
	}
}

func TestDatabaseIntegrity(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// 测试外键约束
	// SQLite 默认不启用外键约束，需要手动启用
	_, err := db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	// 测试唯一约束
	_, err = db.Exec("INSERT INTO configs (key, value) VALUES ('unique_test', 'value1')")
	require.NoError(t, err)

	_, err = db.Exec("INSERT INTO configs (key, value) VALUES ('unique_test', 'value2')")
	require.Error(t, err, "Should fail on unique constraint violation")

	// 测试 NOT NULL 约束
	_, err = db.Exec("INSERT INTO file_types (name, display_name, extensions) VALUES (NULL, 'test', '.txt')")
	require.Error(t, err, "Should fail on NOT NULL constraint")
}

func TestPerformance(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// 测试批量插入性能
	stmt, err := db.Prepare("INSERT INTO configs (key, value) VALUES (?, ?)")
	require.NoError(t, err)
	defer stmt.Close()

	const batchSize = 1000
	for i := 0; i < batchSize; i++ {
		_, err := stmt.Exec(fmt.Sprintf("key_%d", i), fmt.Sprintf("value_%d", i))
		require.NoError(t, err)
	}

	// 测试查询性能
	start := time.Now()
	rows, err := db.Query("SELECT key, value FROM configs WHERE key LIKE ?", "key_5%")
	require.NoError(t, err)
	defer rows.Close()

	count := 0
	for rows.Next() {
		var key, value string
		err = rows.Scan(&key, &value)
		require.NoError(t, err)
		count++
	}
	duration := time.Since(start)

	assert.Greater(t, count, 0, "Should have some results")
	assert.Less(t, duration.Milliseconds(), int64(100), "Query should be fast (<100ms)")
}

func TestDatabaseBackup(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// 插入测试数据
	_, err := db.Exec("INSERT INTO configs (key, value) VALUES ('backup_test', 'backup_value')")
	require.NoError(t, err)

	// 创建备份
	tmpDir := t.TempDir()
	backupPath := filepath.Join(tmpDir, "backup.db")

	_, err = db.Exec(fmt.Sprintf("VACUUM INTO '%s'", backupPath))
	if err != nil {
		t.Skip("VACUUM INTO not supported, skipping backup test")
		return
	}

	// 验证备份文件存在
	_, err = os.Stat(backupPath)
	require.NoError(t, err)

	// 打开备份并验证数据
	backupDB, err := sql.Open("sqlite", backupPath)
	require.NoError(t, err)
	defer backupDB.Close()

	var value string
	err = backupDB.QueryRow("SELECT value FROM configs WHERE key = ?", "backup_test").Scan(&value)
	require.NoError(t, err)
	assert.Equal(t, "backup_value", value)
}

// 基准测试
func BenchmarkInsert(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO configs (key, value) VALUES (?, ?)")
	if err != nil {
		b.Fatal(err)
	}
	defer stmt.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stmt.Exec(fmt.Sprintf("key_%d", i), fmt.Sprintf("value_%d", i))
	}
}

func BenchmarkQuery(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer db.Close()

	// 预填充数据
	const recordCount = 1000
	stmt, _ := db.Prepare("INSERT INTO configs (key, value) VALUES (?, ?)")
	defer stmt.Close()
	for i := 0; i < recordCount; i++ {
		stmt.Exec(fmt.Sprintf("key_%d", i), fmt.Sprintf("value_%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows, _ := db.Query("SELECT key, value FROM configs")
		rows.Close()
	}
}

func BenchmarkTransaction(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ExecuteInTransaction(db, func(tx *Transaction) error {
			_, err := tx.Exec("INSERT INTO configs (key, value) VALUES (?, ?)", "bench_key", "bench_value")
			return err
		})
	}
}
