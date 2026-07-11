package database

import (
	"context"
	"database/sql"
	"time"
)

// ConnectionPoolConfig 数据库连接池配置
type ConnectionPoolConfig struct {
	MaxOpenConnections int
	MaxIdleConnections int
	ConnMaxLifetime    time.Duration
	ConnMaxIdleTime    time.Duration
}

// DefaultConnectionPoolConfig 默认连接池配置
func DefaultConnectionPoolConfig() *ConnectionPoolConfig {
	return &ConnectionPoolConfig{
		MaxOpenConnections: 25,
		MaxIdleConnections: 5,
		ConnMaxLifetime:    5 * time.Minute,
		ConnMaxIdleTime:    1 * time.Minute,
	}
}

// SetConnectionPool 设置数据库连接池
func (s *SQLiteDB) SetConnectionPool(config *ConnectionPoolConfig) error {
	if config == nil {
		config = DefaultConnectionPoolConfig()
	}

	s.db.SetMaxOpenConns(config.MaxOpenConnections)
	s.db.SetMaxIdleConns(config.MaxIdleConnections)
	s.db.SetConnMaxLifetime(config.ConnMaxLifetime)
	s.db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	return nil
}

// GetConnectionStats 获取连接池状态
func (s *SQLiteDB) GetConnectionStats() sql.DBStats {
	stats := s.db.Stats()
	return stats
}

// Ping 检查数据库连接
func (s *SQLiteDB) Ping() error {
	return s.db.Ping()
}

// Close 关闭数据库连接
func (s *SQLiteDB) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Begin 开始一个事务
func (s *SQLiteDB) Begin() (*sql.Tx, error) {
	return s.db.Begin()
}

// BeginTx 开始一个带配置的事务
func (s *SQLiteDB) BeginTx(opts *sql.TxOptions) (*sql.Tx, error) {
	return s.db.BeginTx(context.Background(), opts)
}