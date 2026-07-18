package models

// Config represents a configuration key-value pair
type Config struct {
	ID          uint   `json:"id"`
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description"`
	UpdatedAt   string `json:"updatedAt"`
}

// AIConfig represents AI configuration
type AIConfig struct {
	ID      uint   `json:"id"`
	APIKey  string `json:"apiKey"`
	Model   string `json:"model"`
	BaseURL string `json:"baseURL"`
	Enabled bool   `json:"enabled"`

	WebSearchProvider   string `json:"webSearchProvider"`
	WebSearchBaseURL    string `json:"webSearchBaseURL"`
	WebSearchAPIKey     string `json:"webSearchAPIKey"`
	WebSearchTimeout    int    `json:"webSearchTimeout"`
	WebSearchMaxResults int    `json:"webSearchMaxResults"`
}

// ThemeConfig represents theme configuration
type ThemeConfig struct {
	ID              uint   `json:"id"`
	ThemeMode       string `json:"themeMode"`
	PrimaryColor    string `json:"primaryColor"`
	BackgroundColor string `json:"backgroundColor"`
	BackgroundImage string `json:"backgroundImage"`
}

// StorageDir represents a storage directory
type StorageDir struct {
	ID          uint   `json:"id"`
	Path        string `json:"path"`
	FileType    string `json:"fileType"`
	CurrentSize int64  `json:"currentSize"`
	MaxSize     int64  `json:"maxSize"`
	IsActive    bool   `json:"isActive"`
	IsDefault   bool   `json:"isDefault"`      // 新增：是否为默认主目录
	ParentID    *uint  `json:"parentId"`       // 新增：父目录ID
	CreatedAt   string `json:"createdAt"`

	// 计算字段（不存储在数据库）
	TotalSize   int64        `json:"totalSize,omitempty"`   // 主目录总大小
	SubDirs     []StorageDir `json:"subDirs,omitempty"` // 子目录列表
}

// Tag represents a file tag
type Tag struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	CreatedAt string `json:"createdAt"`
}