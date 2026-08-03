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
	Page           int
	PageSize       int
	FileType       string
	CollectionID   *uint
	CollectionName string
	UnsortedOnly   bool
	SortBy         string
	SortOrder      string
}

var allowedSortColumns = map[string]bool{
	"id":          true,
	"created_at":  true,
	"modified_at": true,
	"file_name":   true,
}

func normalizeSortBy(sortBy string) string {
	if allowedSortColumns[sortBy] {
		return sortBy
	}
	return "created_at"
}

func normalizeSortOrder(sortOrder string) string {
	order := strings.ToUpper(sortOrder)
	if order == "ASC" || order == "DESC" {
		return order
	}
	return "DESC"
}

func qualifiedFileSortColumn(sortBy string) string {
	column := normalizeSortBy(sortBy)
	return "f." + column
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
			INSERT INTO files (file_name, original_name, collection_name, collection_id, file_path, file_type, file_sub_type, file_size, tags, description, metadata, thumbnail, checksum, is_deleted, deleted_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query,
		file.FileName,
		file.OriginalName,
		file.CollectionName,
		file.CollectionID,
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
			SELECT f.id, f.file_name, f.original_name, COALESCE(c.name, f.collection_name), f.collection_id, f.file_path, f.file_type, f.file_sub_type, f.file_size,
			       f.tags, f.description, f.metadata, f.thumbnail, f.checksum, f.is_deleted, f.deleted_at, f.created_at, f.modified_at
			FROM files f
			LEFT JOIN collections c ON c.id = f.collection_id
			WHERE f.id = ?`

	row := r.db.QueryRow(query, id)

	var file models.File
	var tagsJSON, metadataJSON string
	var collectionID sql.NullInt64

	err := row.Scan(
		&file.ID,
		&file.FileName,
		&file.OriginalName,
		&file.CollectionName,
		&collectionID,
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

	if collectionID.Valid {
		cid := uint(collectionID.Int64)
		file.CollectionID = &cid
	}

	return &file, nil
}

// List returns a list of files with filters
func (r *FileRepository) List(filter FileFilter) ([]*models.File, error) {
	query := `
			SELECT f.id, f.file_name, f.original_name, COALESCE(c.name, f.collection_name), f.collection_id, f.file_path, f.file_type, f.file_sub_type, f.file_size,
			       f.tags, f.description, f.metadata, f.thumbnail, f.checksum, f.is_deleted, f.deleted_at, f.created_at, f.modified_at
			FROM files f
			LEFT JOIN collections c ON c.id = f.collection_id
			WHERE 1=1`

	args := []interface{}{}
	argIndex := 1

	// Add file type filter
	if filter.FileType != "" && filter.FileType != "all" {
		query += fmt.Sprintf(" AND f.file_type = $%d", argIndex)
		args = append(args, filter.FileType)
		argIndex++
	}

	// Add collection filter
	if filter.UnsortedOnly {
		// 未分配筛选要排除只有旧 collection_name、但名称已经匹配现有合集的历史记录。
		query += " AND f.collection_id IS NULL AND NOT EXISTS (SELECT 1 FROM collections existing WHERE existing.name = TRIM(f.collection_name))"
	} else if filter.CollectionID != nil {
		trimmedCollectionName := strings.TrimSpace(filter.CollectionName)
		if trimmedCollectionName != "" {
			// 兼容旧数据：部分文件只有 collection_name 没有 collection_id，按合集筛选时一并命中。
			query += fmt.Sprintf(" AND (f.collection_id = $%d OR (f.collection_id IS NULL AND TRIM(f.collection_name) = $%d))", argIndex, argIndex+1)
			args = append(args, *filter.CollectionID, trimmedCollectionName)
			argIndex += 2
		} else {
			query += fmt.Sprintf(" AND f.collection_id = $%d", argIndex)
			args = append(args, *filter.CollectionID)
			argIndex++
		}
	}

	// Add sorting
	sortBy := qualifiedFileSortColumn(filter.SortBy)
	sortOrder := normalizeSortOrder(filter.SortOrder)

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
		var collectionID sql.NullInt64

		err := rows.Scan(
			&file.ID,
			&file.FileName,
			&file.OriginalName,
			&file.CollectionName,
			&collectionID,
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

		if collectionID.Valid {
			cid := uint(collectionID.Int64)
			file.CollectionID = &cid
		}

		files = append(files, &file)
	}

	return files, nil
}

// Count returns the total number of files matching the filter
func (r *FileRepository) Count(filter FileFilter) (int, error) {
	whereParts := []string{"1=1"}
	args := []interface{}{}
	argIndex := 1

	if filter.FileType != "" && filter.FileType != "all" {
		whereParts = append(whereParts, fmt.Sprintf("file_type = $%d", argIndex))
		args = append(args, filter.FileType)
		argIndex++
	}
	if filter.UnsortedOnly {
		// 未分配计数和列表保持同一规则，避免旧合集名记录重复出现在“未分配”和目标合集下。
		whereParts = append(whereParts, "collection_id IS NULL AND NOT EXISTS (SELECT 1 FROM collections c WHERE c.name = TRIM(files.collection_name))")
	} else if filter.CollectionID != nil {
		trimmedCollectionName := strings.TrimSpace(filter.CollectionName)
		if trimmedCollectionName != "" {
			// 计数和列表使用同一套兼容条件，避免点击合集后筛选数字归零。
			whereParts = append(whereParts, fmt.Sprintf("(collection_id = $%d OR (collection_id IS NULL AND TRIM(collection_name) = $%d))", argIndex, argIndex+1))
			args = append(args, *filter.CollectionID, trimmedCollectionName)
			argIndex += 2
		} else {
			whereParts = append(whereParts, fmt.Sprintf("collection_id = $%d", argIndex))
			args = append(args, *filter.CollectionID)
			argIndex++
		}
	}

	query := fmt.Sprintf(`SELECT COUNT(*) FROM files WHERE %s`, strings.Join(whereParts, " AND "))
	var count int
	if err := r.db.QueryRow(query, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count files: %w", err)
	}
	return count, nil
}

func (r *FileRepository) CountCollectionFilters(fileType string) (*models.CollectionFilterCounts, error) {
	result := &models.CollectionFilterCounts{
		Collections: map[uint]int{},
	}

	fileTypeWhere := ""
	totalArgs := []interface{}{}
	if fileType != "" && fileType != "all" {
		fileTypeWhere = " AND file_type = ?"
		totalArgs = append(totalArgs, fileType)
	}

	if err := r.db.QueryRow(`SELECT COUNT(*) FROM files WHERE 1=1`+fileTypeWhere, totalArgs...).Scan(&result.Total); err != nil {
		return nil, fmt.Errorf("failed to count all collection filter files: %w", err)
	}

	unsortedQuery := `
		SELECT COUNT(*)
		FROM files f
		WHERE f.collection_id IS NULL
		  AND NOT EXISTS (
		      SELECT 1 FROM collections c WHERE c.name = TRIM(f.collection_name)
		  )`
	unsortedArgs := []interface{}{}
	if fileType != "" && fileType != "all" {
		unsortedQuery += " AND f.file_type = ?"
		unsortedArgs = append(unsortedArgs, fileType)
	}
	if err := r.db.QueryRow(unsortedQuery, unsortedArgs...).Scan(&result.Unsorted); err != nil {
		return nil, fmt.Errorf("failed to count unsorted collection filter files: %w", err)
	}

	collectionQuery := `
		SELECT c.id, COUNT(f.id)
		FROM collections c
		LEFT JOIN files f
		  ON (
		      f.collection_id = c.id
		      OR (f.collection_id IS NULL AND c.name = TRIM(f.collection_name))
		  )`
	collectionArgs := []interface{}{}
	if fileType != "" && fileType != "all" {
		collectionQuery += " AND f.file_type = ?"
		collectionArgs = append(collectionArgs, fileType)
	}
	collectionQuery += " GROUP BY c.id"

	rows, err := r.db.Query(collectionQuery, collectionArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to count collection filter files: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var collectionID uint
		var count int
		if err := rows.Scan(&collectionID, &count); err != nil {
			return nil, fmt.Errorf("failed to scan collection filter count: %w", err)
		}
		result.Collections[collectionID] = count
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate collection filter counts: %w", err)
	}

	return result, nil
}

// Search searches for files by query
func (r *FileRepository) Search(query string) ([]*models.File, error) {
	searchQuery := "%" + strings.ToLower(query) + "%"

	sqlQuery := `
			SELECT f.id, f.file_name, f.original_name, COALESCE(c.name, f.collection_name), f.collection_id, f.file_path, f.file_type, f.file_sub_type, f.file_size,
			       f.tags, f.description, f.metadata, f.thumbnail, f.checksum, f.is_deleted, f.deleted_at, f.created_at, f.modified_at
			FROM files f
			LEFT JOIN collections c ON c.id = f.collection_id
			WHERE LOWER(f.file_name) LIKE ?
			   OR LOWER(COALESCE(c.name, f.collection_name)) LIKE ?
			   OR LOWER(f.tags) LIKE ?
			   OR LOWER(f.description) LIKE ?
			ORDER BY f.created_at DESC
			LIMIT 100`

	rows, err := r.db.Query(sqlQuery, searchQuery, searchQuery, searchQuery, searchQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to search files: %w", err)
	}
	defer rows.Close()

	files := []*models.File{}
	for rows.Next() {
		var file models.File
		var tagsJSON, metadataJSON string
		var collectionID sql.NullInt64

		err := rows.Scan(
			&file.ID,
			&file.FileName,
			&file.OriginalName,
			&file.CollectionName,
			&collectionID,
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

		if collectionID.Valid {
			cid := uint(collectionID.Int64)
			file.CollectionID = &cid
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

// UpdateMetadataWithCollection updates tags, description, collection_id, and modified time for a file
func (r *FileRepository) UpdateMetadataWithCollection(id uint, tags []string, description string, collectionID *uint) error {
	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}
	query := `UPDATE files SET tags = ?, description = ?, collection_id = ?, modified_at = ? WHERE id = ?`
	modifiedAt := time.Now().Format(time.RFC3339)
	result, err := r.db.Exec(query, string(tagsJSON), description, collectionID, modifiedAt, id)
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

// UpdateCollectionMove updates collection display fields and path after a collection move.
func (r *FileRepository) UpdateCollectionMove(id uint, collectionID *uint, collectionName string, filePath string) error {
	// 移动合集需要同步 collection_name，避免卡片、分组和搜索继续显示旧合集。
	query := `UPDATE files SET collection_id = ?, collection_name = ?, file_path = ?, modified_at = CURRENT_TIMESTAMP WHERE id = ?`
	result, err := r.db.Exec(query, collectionID, collectionName, filePath, id)
	if err != nil {
		return fmt.Errorf("failed to update file collection move: %w", err)
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
			SET file_name = ?, collection_id = ?, file_path = ?, file_type = ?, file_sub_type = ?,
			    file_size = ?, tags = ?, description = ?, metadata = ?, thumbnail = ?,
			    modified_at = CURRENT_TIMESTAMP
			WHERE id = ?`

	result, err := r.db.Exec(query,
		file.FileName,
		file.CollectionID,
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
		SELECT f.id, f.file_name, f.original_name, COALESCE(c.name, f.collection_name), f.collection_id, f.file_path, f.file_type, f.file_sub_type, f.file_size,
		       f.tags, f.description, f.metadata, f.thumbnail, f.checksum, f.is_deleted, f.deleted_at, f.created_at, f.modified_at
		FROM files f
		LEFT JOIN collections c ON c.id = f.collection_id
		WHERE f.checksum = ? AND f.is_deleted = FALSE`

	row := r.db.QueryRow(query, checksum)

	var file models.File
	var tagsJSON, metadataJSON string
	var collectionID sql.NullInt64

	err := row.Scan(
		&file.ID,
		&file.FileName,
		&file.OriginalName,
		&file.CollectionName,
		&collectionID,
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

	if collectionID.Valid {
		cid := uint(collectionID.Int64)
		file.CollectionID = &cid
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
