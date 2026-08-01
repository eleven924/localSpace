package models

type NotificationType string

const (
	NotificationTypeJobCompleted NotificationType = "job_completed"
	NotificationTypeJobFailed    NotificationType = "job_failed"
	NotificationTypeJobCancelled NotificationType = "job_cancelled"
)

type NotificationEvent struct {
	ID        string                 `json:"id"`
	Type      NotificationType       `json:"type"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Payload   map[string]interface{} `json:"payload"`
	CreatedAt string                 `json:"createdAt"`
	Read      bool                   `json:"read"`
}
