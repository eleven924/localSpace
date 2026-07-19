package models

import "encoding/json"

// File represents a file in LocalSpace
type File struct {
	ID          uint      `json:"id"`
	FileName    string    `json:"fileName"`
	OriginalName string   `json:"originalName"`
	CollectionName string `json:"collectionName"`
	FilePath    string    `json:"filePath"`
	FileType    string    `json:"fileType"`
	FileSubType string    `json:"fileSubType"`
	FileSize    int64     `json:"fileSize"`
	Tags        []string  `json:"tags" gorm:"serializer:json"`
	Description string    `json:"description"`
	Metadata    Metadata  `json:"metadata" gorm:"serializer:json"`
	Thumbnail   string    `json:"thumbnail"`
	Checksum    string    `json:"checksum"`
	IsDeleted   bool      `json:"isDeleted"`
	DeletedAt   string    `json:"deletedAt"`
	CreatedAt   string    `json:"createdAt"`
	ModifiedAt  string    `json:"modifiedAt"`
}

// Metadata represents file metadata
type Metadata struct {
	Size      int64  `json:"size,omitempty"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
	Duration  int    `json:"duration,omitempty"`
	PageCount int    `json:"pageCount,omitempty"`
	Author    string `json:"author,omitempty"`
	Title     string `json:"title,omitempty"`
}

// AIAnalysis represents AI analysis results
type AIAnalysis struct {
	Tags        []string `json:"tags"`
	Description string   `json:"description"`
}

// FileType represents a file type
type FileType struct {
	ID          uint        `json:"id"`
	Name        string      `json:"name"`
	DisplayName string      `json:"displayName"`
	Extensions  string      `json:"extensions"`
	SubTypes    []string    `json:"subTypes"`
	CreatedAt   string      `json:"createdAt"`
}

// GetExtensionsSlice returns extensions as a slice
func (ft *FileType) GetExtensionsSlice() []string {
	var extensions []string
	json.Unmarshal([]byte(ft.Extensions), &extensions)
	return extensions
}
