package repositories

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"LocalSpace/app/models"
)

// FileRepository handles file data operations
type FileRepository struct {
	db *sql.DB
}

// NewFileRepository creates a new FileRepository
func NewFileRepository(dbWrapper *SQLiteDBWrapper) *FileRepository {
	return &FileRepository{db: dbWrapper.GetDB()}
}

// FileFilter represents filters for file queries
type FileFilter struct {
	Page      int
	PageSize  int
	FileType  string
	SortBy    string
	SortOrder string
}

// Create creates a new file record
func (r *FileRepository) Create(file *models.File) error {
	tagsJSON, err := json.Marshal(file.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	metadataJSON, err := json.Marshal(file.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
			INSERT INTO files (file_name, original_name, file_path, file_type, file_sub_type, file_size, tags, description, metadata, thumbnail, checksum, is_deleted, deleted_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query,
		file.FileName,
		file.OriginalName,
		file.FilePath,
		file.FileType,
		file.FileSubType,
		file.FileSize,
		string(tagsJSON),
		file.Description,
		string(metadataJSON),
		file.Thumbnail,
		file.Checksum,
		file.IsDeleted,
		file.DeletedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	file.ID = uint(id)
	return nil
}

// FindByID finds a file by ID
func (r *FileRepository) FindByID(id uint) (*models.File, error) {
	query := `
			SELECT id, file_name, original_name, file_path, file_type, file_sub_type, file_size,
			       tags, description, metadata, thumbnail, checksum, is_deleted, deleted_at, created_at, modified_at
			FROM files WHERE id = ?`

	row := r.db.QueryRow(query, id)

	var file models.File
	var tagsJSON, metadataJSON string

	err := row.Scan(
		&file.ID,
		&file.FileName,
		&file.OriginalName,
		&file.FilePath,
		&file.FileType,
		&file.FileSubType,
		&file.FileSize,
		&tagsJSON,
		&file.Description,
		&metadataJSON,
		&file.Thumbnail,
		&file.Checksum,
		&file.IsDeleted,
		&file.DeletedAt,
		&file.CreatedAt,
		&file.ModifiedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("file not found")
		}
		return nil, fmt.Errorf("failed to find file: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal([]byte(tagsJSON), &file.Tags); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
	}

	if err := json.Unmarshal([]byte(metadataJSON), &file.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &file, nil
}

// List returns a list of files with filters
func (r *FileRepository) List(filter FileFilter) ([]*models.File, error) {
	query := `
			SELECT id, file_name, original_name, file_path, file_type, file_sub_type, file_size,
			       tags, description, metadata, thumbnail, checksum, is_deleted, deleted_at, created_at, modified_at
			FROM files WHERE 1=1`

	args := []interface{}{}
	argIndex := 1

	// Add file type filter
	if filter.FileType != "" && filter.FileType != "all" {
		query += fmt.Sprintf(" AND file_type = $%d", argIndex)
		args = append(args, filter.FileType)
		argIndex++
	}

	// Add sorting
	sortBy := "created_at"
	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}

	sortOrder := "DESC"
	if filter.SortOrder != "" {
		sortOrder = filter.SortOrder
	}

	query += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

	// Add pagination
	offset := (filter.Page - 1) * filter.PageSize
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, filter.PageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}
	defer rows.Close()

	files := []*models.File{}
	for rows.Next() {
		var file models.File
		var tagsJSON, metadataJSON string

		err := rows.Scan(
			&file.ID,
			&file.FileName,
			&file.OriginalName,
			&file.FilePath,
			&file.FileType,
			&file.FileSubType,
			&file.FileSize,
			&tagsJSON,
			&file.Description,
			&metadataJSON,
			&file.Thumbnail,
			&file.Checksum,
			&file.IsDeleted,
			&file.DeletedAt,
			&file.CreatedAt,
			&file.ModifiedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan file: %w", err)
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal([]byte(tagsJSON), &file.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}

		if err := json.Unmarshal([]byte(metadataJSON), &file.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		files = append(files, &file)
	}

	return files, nil
}

// Search searches for files by query
func (r *FileRepository) Search(query string) ([]*models.File, error) {
	searchQuery := "%" + strings.ToLower(query) + "%"

	sqlQuery := `
			SELECT id, file_name, original_name, file_path, file_type, file_sub_type, file_size,
			       tags, description, metadata, thumbnail, checksum, is_deleted, deleted_at, created_at, modified_at
			FROM files
			WHERE LOWER(file_name) LIKE ?
			   OR LOWER(tags) LIKE ?
			   OR LOWER(description) LIKE ?
			ORDER BY created_at DESC
			LIMIT 100`

	rows, err := r.db.Query(sqlQuery, searchQuery, searchQuery, searchQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to search files: %w", err)
	}
	defer rows.Close()

	files := []*models.File{}
	for rows.Next() {
		var file models.File
		var tagsJSON, metadataJSON string

		err := rows.Scan(
			&file.ID,
			&file.FileName,
			&file.OriginalName,
			&file.FilePath,
			&file.FileType,
			&file.FileSubType,
			&file.FileSize,
			&tagsJSON,
			&file.Description,
			&metadataJSON,
			&file.Thumbnail,
			&file.Checksum,
			&file.IsDeleted,
			&file.DeletedAt,
			&file.CreatedAt,
			&file.ModifiedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan file: %w", err)
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal([]byte(tagsJSON), &file.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}

		if err := json.Unmarshal([]byte(metadataJSON), &file.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		files = append(files, &file)
	}

	return files, nil
}

// Delete deletes a file by ID
func (r *FileRepository) Delete(id uint) error {
	query := `DELETE FROM files WHERE id = ?`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("file not found")
	}

	return nil
}

// DeleteByPath deletes all files with a specific path prefix
func (r *FileRepository) DeleteByPath(pathPrefix string) (int64, error) {
	query := `DELETE FROM files WHERE file_path LIKE ? || '%'`

	result, err := r.db.Exec(query, pathPrefix)
	if err != nil {
		return 0, fmt.Errorf("failed to delete files by path: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// UpdateMetadata updates only tags, description, and modified time for a file
func (r *FileRepository) UpdateMetadata(id uint, tags []string, description string) error {
	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	query := `UPDATE files SET tags = ?, description = ?, modified_at = ? WHERE id = ?`
	modifiedAt := time.Now().Format(time.RFC3339)

	result, err := r.db.Exec(query, string(tagsJSON), description, modifiedAt, id)
	if err != nil {
		return fmt.Errorf("failed to update file metadata: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("file not found")
	}

	return nil
}

// Update updates a file
func (r *FileRepository) Update(file *models.File) error {
	tagsJSON, err := json.Marshal(file.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	metadataJSON, err := json.Marshal(file.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
			UPDATE files
			SET file_name = ?, file_path = ?, file_type = ?, file_sub_type = ?,
			    file_size = ?, tags = ?, description = ?, metadata = ?, thumbnail = ?,
			    modified_at = CURRENT_TIMESTAMP
			WHERE id = ?`

	result, err := r.db.Exec(query,
		file.FileName,
		file.FilePath,
		file.FileType,
		file.FileSubType,
		file.FileSize,
		string(tagsJSON),
		file.Description,
		string(metadataJSON),
		file.Thumbnail,
		file.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update file: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("file not found")
	}

	return nil
}

// UpdateThumbnail updates only the thumbnail path for a file
func (r *FileRepository) UpdateThumbnail(id uint, thumbnailPath string) error {
	query := `UPDATE files SET thumbnail = ?, modified_at = CURRENT_TIMESTAMP WHERE id = ?`

	result, err := r.db.Exec(query, thumbnailPath, id)
	if err != nil {
		return fmt.Errorf("failed to update thumbnail: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("file not found")
	}

	return nil
}

// ExistsByPath checks if a file exists by path
func (r *FileRepository) ExistsByPath(path string) (bool, error) {
	query := `SELECT COUNT(*) FROM files WHERE file_path = ?`

	var count int
	err := r.db.QueryRow(query, path).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}

	return count > 0, nil
}

// FindByChecksum finds a file by its checksum
func (r *FileRepository) FindByChecksum(checksum string) (*models.File, error) {
	query := `
		SELECT id, file_name, original_name, file_path, file_type, file_sub_type, file_size,
		       tags, description, metadata, thumbnail, checksum, is_deleted, deleted_at, created_at, modified_at
		FROM files WHERE checksum = ? AND is_deleted = FALSE`

	row := r.db.QueryRow(query, checksum)

	var file models.File
	var tagsJSON, metadataJSON string

	err := row.Scan(
		&file.ID,
		&file.FileName,
		&file.OriginalName,
		&file.FilePath,
		&file.FileType,
		&file.FileSubType,
		&file.FileSize,
		&tagsJSON,
		&file.Description,
		&metadataJSON,
		&file.Thumbnail,
		&file.Checksum,
		&file.IsDeleted,
		&file.DeletedAt,
		&file.CreatedAt,
		&file.ModifiedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No duplicate found
		}
		return nil, fmt.Errorf("failed to find file by checksum: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal([]byte(tagsJSON), &file.Tags); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
	}

	if err := json.Unmarshal([]byte(metadataJSON), &file.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &file, nil
}

// CheckDuplicateByChecksum checks if a file with the given checksum exists
func (r *FileRepository) CheckDuplicateByChecksum(checksum string) (bool, uint, error) {
	query := `SELECT id FROM files WHERE checksum = ? AND is_deleted = FALSE LIMIT 1`

	var id uint
	err := r.db.QueryRow(query, checksum).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, 0, nil // No duplicate found
		}
		return false, 0, fmt.Errorf("failed to check duplicate: %w", err)
	}

	return true, id, nil
}
