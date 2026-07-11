package database

import (
	"context"
	"database/sql"
	"fmt"
)

// Transaction 事务执行器
type Transaction struct {
	tx *sql.Tx
}

// NewTransaction 创建新的事务
func NewTransaction(db *sql.DB) (*Transaction, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return &Transaction{tx: tx}, nil
}

// NewTransactionWithContext 创建带上下文的事务
func NewTransactionWithContext(ctx context.Context, db *sql.DB, opts *sql.TxOptions) (*Transaction, error) {
	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return &Transaction{tx: tx}, nil
}

// Commit 提交事务
func (t *Transaction) Commit() error {
	if err := t.tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// Rollback 回滚事务
func (t *Transaction) Rollback() error {
	if err := t.tx.Rollback(); err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}
	return nil
}

// Exec 执行 SQL 语句
func (t *Transaction) Exec(query string, args ...interface{}) (sql.Result, error) {
	return t.tx.Exec(query, args...)
}

// Query 执行查询
func (t *Transaction) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return t.tx.Query(query, args...)
}

// QueryRow 执行单行查询
func (t *Transaction) QueryRow(query string, args ...interface{}) *sql.Row {
	return t.tx.QueryRow(query, args...)
}

// TransactionCallback 事务回调函数类型
type TransactionCallback func(*Transaction) error

// ExecuteInTransaction 在事务中执行回调函数
func ExecuteInTransaction(db *sql.DB, callback TransactionCallback) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	transaction := &Transaction{tx: tx}

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r) // 重新抛出 panic
		}
	}()

	err = callback(transaction)
	if err != nil {
		// 如果回调出错，回滚事务
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("callback error: %w, rollback error: %v", err, rbErr)
		}
		return fmt.Errorf("transaction failed and was rolled back: %w", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ExecuteInTransactionWithContext 在带上下文的事务中执行回调函数
func ExecuteInTransactionWithContext(ctx context.Context, db *sql.DB, opts *sql.TxOptions, callback TransactionCallback) error {
	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	transaction := &Transaction{tx: tx}

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r) // 重新抛出 panic
		}
	}()

	err = callback(transaction)
	if err != nil {
		// 如果回调出错，回滚事务
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("callback error: %w, rollback error: %v", err, rbErr)
		}
		return fmt.Errorf("transaction failed and was rolled back: %w", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}