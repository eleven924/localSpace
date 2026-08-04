package models

import "encoding/json"

const (
	JobTypeBatchImport  = "batch_import"
	JobTypeSingleImport = "single_import"
	JobTypeCleanup      = "job_cleanup"
)

const (
	JobStatusPending        = "pending"
	JobStatusRunning        = "running"
	JobStatusPaused         = "paused"
	JobStatusAwaitingResume = "awaiting_resume"
	JobStatusCompleted      = "completed"
	JobStatusFailed         = "failed"
	JobStatusCancelled      = "cancelled"
	JobStatusTimedOut       = "timed_out"
	JobStatusRecovering     = "recovering"
	JobStatusCleanupFailed  = "cleanup_failed"
)

const (
	BatchImportItemStatusPending    = "pending"
	BatchImportItemStatusProcessing = "processing"
	BatchImportItemStatusCompleted  = "completed"
	BatchImportItemStatusFailed     = "failed"
	BatchImportItemStatusRecovering = "recovering"
	BatchImportTempSuffix           = ".localspace-importing"
)

type Job struct {
	ID                uint            `json:"id"`
	JobType           string          `json:"jobType"`
	Status            string          `json:"status"`
	Title             string          `json:"title"`
	Payload           json.RawMessage `json:"payload"`
	Result            json.RawMessage `json:"result"`
	ProgressTotal     int             `json:"progressTotal"`
	ProgressCompleted int             `json:"progressCompleted"`
	ProgressMessage   string          `json:"progressMessage"`
	ExclusiveKey      string          `json:"exclusiveKey"`
	CanResume         bool            `json:"canResume"`
	StartedAt         string          `json:"startedAt"`
	HeartbeatAt       string          `json:"heartbeatAt"`
	FinishedAt        string          `json:"finishedAt"`
	TimeoutAt         string          `json:"timeoutAt"`
	ErrorMessage      string          `json:"errorMessage"`
	CreatedAt         string          `json:"createdAt"`
	UpdatedAt         string          `json:"updatedAt"`
}

type JobListResponse struct {
	Items    []*Job `json:"items"`
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Total    int    `json:"total"`
}

type ExitGuardSnapshot struct {
	HasProtectedJobs bool                    `json:"hasProtectedJobs"`
	Total            int                     `json:"total"`
	StatusCounts     map[string]int          `json:"statusCounts"`
	Jobs             []*ExitGuardJobSnapshot `json:"jobs"`
}

type ExitGuardJobSnapshot struct {
	ID                uint   `json:"id"`
	JobType           string `json:"jobType"`
	Status            string `json:"status"`
	Title             string `json:"title"`
	ProgressTotal     int    `json:"progressTotal"`
	ProgressCompleted int    `json:"progressCompleted"`
	ProgressMessage   string `json:"progressMessage"`
	CanResume         bool   `json:"canResume"`
}

type JobRetentionConfig struct {
	ID        uint   `json:"id"`
	Enabled   bool   `json:"enabled"`
	MaxCount  int    `json:"maxCount"`
	MaxDays   int    `json:"maxDays"`
	UpdatedAt string `json:"updatedAt"`
}

type BatchImportJobRequest struct {
	Files                        []BatchImportFileInput `json:"files"`
	SharedTags                   []string               `json:"sharedTags"`
	SharedDescription            string                 `json:"sharedDescription"`
	CollectionID                 *uint                  `json:"collectionId"`
	EnableAIGeneratedTags        bool                   `json:"enableAIGeneratedTags"`
	EnableAIGeneratedDescription bool                   `json:"enableAIGeneratedDescription"`
}

type BatchImportFileInput struct {
	SourcePath  string `json:"sourcePath"`
	DisplayName string `json:"displayName"`
}

type BatchImportResult struct {
	SuccessCount int                     `json:"successCount"`
	FailedCount  int                     `json:"failedCount"`
	FailedItems  []BatchImportFailedItem `json:"failedItems"`
}

type BatchImportFailedItem struct {
	SourcePath  string `json:"sourcePath"`
	DisplayName string `json:"displayName"`
	Error       string `json:"error"`
}

type BatchImportItem struct {
	ID               uint   `json:"id"`
	JobID            uint   `json:"jobId"`
	SourcePath       string `json:"sourcePath"`
	DisplayName      string `json:"displayName"`
	DetectedFileType string `json:"detectedFileType"`
	Status           string `json:"status"`
	ItemIndex        int    `json:"itemIndex"`
	TempPath         string `json:"tempPath"`
	FinalPath        string `json:"finalPath"`
	ExpectedSize     int64  `json:"expectedSize"`
	BytesCopied      int64  `json:"bytesCopied"`
	Checksum         string `json:"checksum"`
	ErrorMessage     string `json:"errorMessage"`
	StartedAt        string `json:"startedAt"`
	FinishedAt       string `json:"finishedAt"`
	UpdatedAt        string `json:"updatedAt"`
}

type SelectedFile struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Size int64  `json:"size"`
}

func (j *Job) ProgressPercent() int {
	if j == nil || j.ProgressTotal <= 0 {
		return 0
	}
	if j.ProgressCompleted <= 0 {
		return 0
	}
	percent := (j.ProgressCompleted * 100) / j.ProgressTotal
	if percent > 100 {
		return 100
	}
	return percent
}
