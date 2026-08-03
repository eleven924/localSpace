package models

import "strings"

type Collection struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type FileListResponse struct {
	Items    []*File `json:"items"`
	Page     int     `json:"page"`
	PageSize int     `json:"pageSize"`
	Total    int     `json:"total"`
}

type BatchMoveResult struct {
	SuccessCount int                   `json:"successCount"`
	FailedCount  int                   `json:"failedCount"`
	FailedItems  []BatchMoveFailedItem `json:"failedItems"`
}

type BatchMoveFailedItem struct {
	FileID   uint   `json:"fileId"`
	FileName string `json:"fileName"`
	Error    string `json:"error"`
}

type BatchDeleteResult struct {
	SuccessCount int                     `json:"successCount"`
	FailedCount  int                     `json:"failedCount"`
	FailedItems  []BatchDeleteFailedItem `json:"failedItems"`
}

type BatchDeleteFailedItem struct {
	FileID   uint   `json:"fileId"`
	FileName string `json:"fileName"`
	Error    string `json:"error"`
}

func NormalizeCollectionName(name string) string {
	return strings.TrimSpace(name)
}
