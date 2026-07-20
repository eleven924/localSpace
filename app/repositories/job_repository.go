package repositories

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"LocalSpace/app/models"
)

type JobRepository struct {
	db *sql.DB
}

func NewJobRepository(dbWrapper *SQLiteDBWrapper) *JobRepository {
	return &JobRepository{db: dbWrapper.GetDB()}
}

func (r *JobRepository) Create(job *models.Job) error {
	now := time.Now().Format(time.RFC3339)
	result, err := r.db.Exec(`
		INSERT INTO jobs (
			job_type, status, title, payload, result, progress_total, progress_completed, progress_message,
			exclusive_key, can_resume, started_at, heartbeat_at, finished_at, timeout_at, error_message,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		job.JobType,
		job.Status,
		job.Title,
		nullableJSON(job.Payload),
		nullableJSON(job.Result),
		job.ProgressTotal,
		job.ProgressCompleted,
		job.ProgressMessage,
		job.ExclusiveKey,
		job.CanResume,
		nullIfEmpty(job.StartedAt),
		nullIfEmpty(job.HeartbeatAt),
		nullIfEmpty(job.FinishedAt),
		nullIfEmpty(job.TimeoutAt),
		job.ErrorMessage,
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get job id: %w", err)
	}
	job.ID = uint(id)
	job.CreatedAt = now
	job.UpdatedAt = now
	return nil
}

func (r *JobRepository) Update(job *models.Job) error {
	if job == nil {
		return fmt.Errorf("job is nil")
	}

	job.UpdatedAt = time.Now().Format(time.RFC3339)
	_, err := r.db.Exec(`
		UPDATE jobs
		SET status = ?, title = ?, payload = ?, result = ?, progress_total = ?, progress_completed = ?,
		    progress_message = ?, exclusive_key = ?, can_resume = ?, started_at = ?, heartbeat_at = ?,
		    finished_at = ?, timeout_at = ?, error_message = ?, updated_at = ?
		WHERE id = ?`,
		job.Status,
		job.Title,
		nullableJSON(job.Payload),
		nullableJSON(job.Result),
		job.ProgressTotal,
		job.ProgressCompleted,
		job.ProgressMessage,
		job.ExclusiveKey,
		job.CanResume,
		nullIfEmpty(job.StartedAt),
		nullIfEmpty(job.HeartbeatAt),
		nullIfEmpty(job.FinishedAt),
		nullIfEmpty(job.TimeoutAt),
		job.ErrorMessage,
		job.UpdatedAt,
		job.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}
	return nil
}

func (r *JobRepository) FindByID(id uint) (*models.Job, error) {
	row := r.db.QueryRow(`
		SELECT id, job_type, status, title, payload, result, progress_total, progress_completed, progress_message,
		       exclusive_key, can_resume, started_at, heartbeat_at, finished_at, timeout_at, error_message,
		       created_at, updated_at
		FROM jobs
		WHERE id = ?`, id)
	return scanJob(row)
}

func (r *JobRepository) ListActive() ([]*models.Job, error) {
	return r.listByStatuses(
		models.JobStatusPending,
		models.JobStatusRunning,
		models.JobStatusRecovering,
		models.JobStatusAwaitingResume,
	)
}

func (r *JobRepository) ListResumable() ([]*models.Job, error) {
	return r.listByStatuses(models.JobStatusAwaitingResume)
}

func (r *JobRepository) ListUnfinishedForStartup() ([]*models.Job, error) {
	return r.listByStatuses(
		models.JobStatusRunning,
		models.JobStatusRecovering,
		models.JobStatusAwaitingResume,
	)
}

func (r *JobRepository) List(page, pageSize int, jobType string) ([]*models.Job, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	whereParts := []string{"1=1"}
	args := make([]interface{}, 0, 4)
	if strings.TrimSpace(jobType) != "" && jobType != "all" {
		whereParts = append(whereParts, "job_type = ?")
		args = append(args, jobType)
	}

	whereClause := strings.Join(whereParts, " AND ")

	var total int
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM jobs WHERE %s`, whereClause)
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count jobs: %w", err)
	}

	queryArgs := append([]interface{}{}, args...)
	offset := (page - 1) * pageSize
	queryArgs = append(queryArgs, pageSize, offset)

	query := fmt.Sprintf(`
		SELECT id, job_type, status, title, payload, result, progress_total, progress_completed, progress_message,
		       exclusive_key, can_resume, started_at, heartbeat_at, finished_at, timeout_at, error_message,
		       created_at, updated_at
		FROM jobs
		WHERE %s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`, whereClause)

	rows, err := r.db.Query(query, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list jobs: %w", err)
	}
	defer rows.Close()

	items := make([]*models.Job, 0, pageSize)
	for rows.Next() {
		job, err := scanJob(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, job)
	}

	return items, total, nil
}

func (r *JobRepository) CountActiveByExclusiveKey(exclusiveKey string, excludeJobID uint) (int, error) {
	if strings.TrimSpace(exclusiveKey) == "" {
		return 0, nil
	}

	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*)
		FROM jobs
		WHERE exclusive_key = ?
		  AND status IN (?, ?, ?, ?)
		  AND id != ?`,
		exclusiveKey,
		models.JobStatusPending,
		models.JobStatusRunning,
		models.JobStatusRecovering,
		models.JobStatusAwaitingResume,
		excludeJobID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count active jobs: %w", err)
	}
	return count, nil
}

func (r *JobRepository) CreateBatchImportItems(items []*models.BatchImportItem) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin batch item transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now().Format(time.RFC3339)
	for _, item := range items {
		result, err := tx.Exec(`
			INSERT INTO batch_import_items (
				job_id, source_path, display_name, detected_file_type, status, item_index,
				temp_path, final_path, expected_size, bytes_copied, checksum, error_message,
				started_at, finished_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			item.JobID,
			item.SourcePath,
			item.DisplayName,
			item.DetectedFileType,
			item.Status,
			item.ItemIndex,
			item.TempPath,
			item.FinalPath,
			item.ExpectedSize,
			item.BytesCopied,
			item.Checksum,
			item.ErrorMessage,
			nullIfEmpty(item.StartedAt),
			nullIfEmpty(item.FinishedAt),
			now,
		)
		if err != nil {
			return fmt.Errorf("failed to create batch import item: %w", err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get batch import item id: %w", err)
		}
		item.ID = uint(id)
		item.UpdatedAt = now
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit batch import item transaction: %w", err)
	}

	return nil
}

func (r *JobRepository) ListBatchImportItems(jobID uint) ([]*models.BatchImportItem, error) {
	rows, err := r.db.Query(`
		SELECT id, job_id, source_path, display_name, detected_file_type, status, item_index,
		       temp_path, final_path, expected_size, bytes_copied, checksum, error_message,
		       started_at, finished_at, updated_at
		FROM batch_import_items
		WHERE job_id = ?
		ORDER BY item_index ASC`, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to list batch import items: %w", err)
	}
	defer rows.Close()

	items := make([]*models.BatchImportItem, 0)
	for rows.Next() {
		item, err := scanBatchImportItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *JobRepository) UpdateBatchImportItem(item *models.BatchImportItem) error {
	if item == nil {
		return fmt.Errorf("batch import item is nil")
	}

	item.UpdatedAt = time.Now().Format(time.RFC3339)
	_, err := r.db.Exec(`
		UPDATE batch_import_items
		SET source_path = ?, display_name = ?, detected_file_type = ?, status = ?, item_index = ?,
		    temp_path = ?, final_path = ?, expected_size = ?, bytes_copied = ?, checksum = ?,
		    error_message = ?, started_at = ?, finished_at = ?, updated_at = ?
		WHERE id = ?`,
		item.SourcePath,
		item.DisplayName,
		item.DetectedFileType,
		item.Status,
		item.ItemIndex,
		item.TempPath,
		item.FinalPath,
		item.ExpectedSize,
		item.BytesCopied,
		item.Checksum,
		item.ErrorMessage,
		nullIfEmpty(item.StartedAt),
		nullIfEmpty(item.FinishedAt),
		item.UpdatedAt,
		item.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update batch import item: %w", err)
	}
	return nil
}

func (r *JobRepository) listByStatuses(statuses ...string) ([]*models.Job, error) {
	if len(statuses) == 0 {
		return []*models.Job{}, nil
	}

	placeholders := make([]string, 0, len(statuses))
	args := make([]interface{}, 0, len(statuses))
	for _, status := range statuses {
		placeholders = append(placeholders, "?")
		args = append(args, status)
	}

	query := fmt.Sprintf(`
		SELECT id, job_type, status, title, payload, result, progress_total, progress_completed, progress_message,
		       exclusive_key, can_resume, started_at, heartbeat_at, finished_at, timeout_at, error_message,
		       created_at, updated_at
		FROM jobs
		WHERE status IN (%s)
		ORDER BY created_at DESC`, strings.Join(placeholders, ", "))

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}
	defer rows.Close()

	jobs := make([]*models.Job, 0)
	for rows.Next() {
		job, err := scanJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func scanJob(scanner interface {
	Scan(dest ...interface{}) error
}) (*models.Job, error) {
	var job models.Job
	var payload sql.NullString
	var result sql.NullString
	var startedAt sql.NullString
	var heartbeatAt sql.NullString
	var finishedAt sql.NullString
	var timeoutAt sql.NullString

	err := scanner.Scan(
		&job.ID,
		&job.JobType,
		&job.Status,
		&job.Title,
		&payload,
		&result,
		&job.ProgressTotal,
		&job.ProgressCompleted,
		&job.ProgressMessage,
		&job.ExclusiveKey,
		&job.CanResume,
		&startedAt,
		&heartbeatAt,
		&finishedAt,
		&timeoutAt,
		&job.ErrorMessage,
		&job.CreatedAt,
		&job.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("job not found")
		}
		return nil, fmt.Errorf("failed to scan job: %w", err)
	}

	job.Payload = rawJSON(payload)
	job.Result = rawJSON(result)
	job.StartedAt = startedAt.String
	job.HeartbeatAt = heartbeatAt.String
	job.FinishedAt = finishedAt.String
	job.TimeoutAt = timeoutAt.String

	return &job, nil
}

func scanBatchImportItem(scanner interface {
	Scan(dest ...interface{}) error
}) (*models.BatchImportItem, error) {
	var item models.BatchImportItem
	var startedAt sql.NullString
	var finishedAt sql.NullString

	err := scanner.Scan(
		&item.ID,
		&item.JobID,
		&item.SourcePath,
		&item.DisplayName,
		&item.DetectedFileType,
		&item.Status,
		&item.ItemIndex,
		&item.TempPath,
		&item.FinalPath,
		&item.ExpectedSize,
		&item.BytesCopied,
		&item.Checksum,
		&item.ErrorMessage,
		&startedAt,
		&finishedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan batch import item: %w", err)
	}

	item.StartedAt = startedAt.String
	item.FinishedAt = finishedAt.String
	return &item, nil
}

func nullableJSON(data json.RawMessage) interface{} {
	if len(data) == 0 {
		return nil
	}
	return string(data)
}

func rawJSON(value sql.NullString) json.RawMessage {
	if !value.Valid || strings.TrimSpace(value.String) == "" {
		return nil
	}
	return json.RawMessage(value.String)
}

func nullIfEmpty(value string) interface{} {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}
