package repositories

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"LocalSpace/app/models"
)

type CollectionRepository struct {
	db *sql.DB
}

func NewCollectionRepository(dbWrapper *SQLiteDBWrapper) *CollectionRepository {
	return &CollectionRepository{db: dbWrapper.GetDB()}
}

func (r *CollectionRepository) GetAll() ([]models.Collection, error) {
	rows, err := r.db.Query(`
		SELECT id, name, created_at, updated_at
		FROM collections
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}
	defer rows.Close()

	var result []models.Collection
	for rows.Next() {
		var c models.Collection
		if err := rows.Scan(&c.ID, &c.Name, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan collection: %w", err)
		}
		result = append(result, c)
	}
	return result, nil
}

func (r *CollectionRepository) FindByID(id uint) (*models.Collection, error) {
	row := r.db.QueryRow(`
		SELECT id, name, created_at, updated_at
		FROM collections WHERE id = ?
	`, id)
	var c models.Collection
	if err := row.Scan(&c.ID, &c.Name, &c.CreatedAt, &c.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("collection not found")
		}
		return nil, fmt.Errorf("failed to find collection: %w", err)
	}
	return &c, nil
}

func (r *CollectionRepository) Add(name string) (uint, error) {
	trimmed := models.NormalizeCollectionName(name)
	if trimmed == "" {
		return 0, fmt.Errorf("collection name cannot be empty")
	}
	// _unsorted 是系统内置目录名，不能创建同名合集，否则路由和物理目录会产生歧义。
	if strings.EqualFold(trimmed, models.BuiltinUnsortedFolderName) {
		return 0, fmt.Errorf("collection name %q is reserved", models.BuiltinUnsortedFolderName)
	}
	now := time.Now().Format(time.RFC3339)
	result, err := r.db.Exec(`
		INSERT INTO collections (name, created_at, updated_at) VALUES (?, ?, ?)
	`, trimmed, now, now)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return 0, fmt.Errorf("collection name already exists")
		}
		return 0, fmt.Errorf("failed to add collection: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get collection id: %w", err)
	}
	return uint(id), nil
}

func (r *CollectionRepository) Remove(id uint) error {
	result, err := r.db.Exec(`DELETE FROM collections WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to remove collection: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("collection not found")
	}
	return nil
}

func (r *CollectionRepository) CountFilesByCollectionID(id uint) (int, error) {
	var count int
	if err := r.db.QueryRow(`
		SELECT COUNT(*) FROM files WHERE collection_id = ? AND is_deleted = 0
	`, id).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count files in collection: %w", err)
	}
	return count, nil
}
