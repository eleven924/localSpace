package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DocumentMetadata 包含文档的元数据信息
type DocumentMetadata struct {
	PageCount int       `json:"pageCount"`
	Title     string    `json:"title,omitempty"`
	Author    string    `json:"author,omitempty"`
	Subject   string    `json:"subject,omitempty"`
	Keywords  string    `json:"keywords,omitempty"`
	Creator   string    `json:"creator,omitempty"`
	Producer  string    `json:"producer,omitempty"`
	Created   time.Time `json:"created,omitempty"`
	Modified  time.Time `json:"modified,omitempty"`
	FileSize  int64     `json:"fileSize"`
	ModTime   time.Time `json:"modTime"`
}

// ExtractDocumentMetadata 提取文档文件的元数据
func ExtractDocumentMetadata(filePath, fileType string) (*DocumentMetadata, error) {
	// 获取文件信息
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	metadata := &DocumentMetadata{
		FileSize: fileInfo.Size(),
		ModTime:  fileInfo.ModTime(),
	}

	// 根据文件类型提取元数据
	switch fileType {
	case "pdf":
		return extractPDFMetadata(filePath, metadata)
	case "docx", "doc":
		return extractWordMetadata(filePath, metadata)
	case "xlsx", "xls":
		return extractExcelMetadata(filePath, metadata)
	case "pptx", "ppt":
		return extractPowerPointMetadata(filePath, metadata)
	case "txt":
		return extractTextMetadata(filePath, metadata)
	case "md":
		return extractMarkdownMetadata(filePath, metadata)
	default:
		return metadata, nil
	}
}

// extractPDFMetadata 提取 PDF 文件元数据
func extractPDFMetadata(filePath string, metadata *DocumentMetadata) (*DocumentMetadata, error) {
	// 注意：这里需要集成 PDF 解析库
	// 由于 Go 生态中 PDF 库的限制，这里提供基本实现
	// 实际项目中可以考虑使用：
	// - github.com/unidoc/unipdf (需要商业许可)
	// - 调用外部工具如 pdftk 或 pdfinfo

	// 使用文件名作为标题
	fileName := filepath.Base(filePath)
	metadata.Title = strings.TrimSuffix(fileName, filepath.Ext(fileName))

	// 尝试估算页数（基于文件大小）
	// 这是一个粗略的估算，实际需要 PDF 库支持
	if metadata.FileSize > 0 {
		// 假设平均每页 PDF 约 100KB
		estimatedPages := int(metadata.FileSize / (100 * 1024))
		if estimatedPages < 1 {
			estimatedPages = 1
		}
		metadata.PageCount = estimatedPages
	}

	return metadata, nil
}

// extractWordMetadata 提取 Word 文件元数据
func extractWordMetadata(filePath string, metadata *DocumentMetadata) (*DocumentMetadata, error) {
	// Word 文件实际上是一个 ZIP 压缩包
	// 元数据存储在 docProps 目录下
	// 可以通过以下方式提取：
	// 1. 使用 github.com/unidoc/unioffice
	// 2. 解压 ZIP 并解析 XML

	// 使用文件名作为标题
	fileName := filepath.Base(filePath)
	metadata.Title = strings.TrimSuffix(fileName, filepath.Ext(filePath))

	return metadata, nil
}

// extractExcelMetadata 提取 Excel 文件元数据
func extractExcelMetadata(filePath string, metadata *DocumentMetadata) (*DocumentMetadata, error) {
	// Excel 文件元数据提取
	// 可以使用 github.com/360EntSecGroup-Skylar/excelize 或类似库

	// 使用文件名作为标题
	fileName := filepath.Base(filePath)
	metadata.Title = strings.TrimSuffix(fileName, filepath.Ext(filePath))

	return metadata, nil
}

// extractPowerPointMetadata 提取 PowerPoint 文件元数据
func extractPowerPointMetadata(filePath string, metadata *DocumentMetadata) (*DocumentMetadata, error) {
	// PowerPoint 文件元数据提取
	// 可以使用 github.com/unidoc/unioffice 或类似库

	// 使用文件名作为标题
	fileName := filepath.Base(filePath)
	metadata.Title = strings.TrimSuffix(fileName, filepath.Ext(filePath))

	return metadata, nil
}

// extractTextMetadata 提取文本文件元数据
func extractTextMetadata(filePath string, metadata *DocumentMetadata) (*DocumentMetadata, error) {
	// 读取文本文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read text file: %w", err)
	}

	// 估算页数（基于字符数）
	// 假设每页约 3000 字符
	estimatedPages := len(content) / 3000
	if estimatedPages < 1 {
		estimatedPages = 1
	}
	metadata.PageCount = estimatedPages

	// 使用文件名作为标题
	fileName := filepath.Base(filePath)
	metadata.Title = strings.TrimSuffix(fileName, filepath.Ext(fileName))

	return metadata, nil
}

// extractMarkdownMetadata 提取 Markdown 文件元数据
func extractMarkdownMetadata(filePath string, metadata *DocumentMetadata) (*DocumentMetadata, error) {
	// 读取 Markdown 文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read markdown file: %w", err)
	}

	// 解析 Markdown 元数据
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		// 检查标题（# 开头）
		if strings.HasPrefix(line, "#") {
			if metadata.Title == "" {
				metadata.Title = strings.TrimSpace(strings.TrimPrefix(line, "#"))
				metadata.Title = strings.TrimPrefix(metadata.Title, "#")
				metadata.Title = strings.TrimSpace(metadata.Title)
			}
		}
	}

	// 如果没有找到标题，使用文件名
	if metadata.Title == "" {
		fileName := filepath.Base(filePath)
		metadata.Title = strings.TrimSuffix(fileName, filepath.Ext(filePath))
	}

	// 估算页数（基于字符数）
	estimatedPages := len(content) / 3000
	if estimatedPages < 1 {
		estimatedPages = 1
	}
	metadata.PageCount = estimatedPages

	return metadata, nil
}

// FormatPageCount 格式化页数
func FormatPageCount(pageCount int) string {
	if pageCount == 0 {
		return "Unknown"
	}
	return fmt.Sprintf("%d page(s)", pageCount)
}

// IsDocumentSupported 检查文档格式是否支持
func IsDocumentSupported(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	supportedFormats := []string{
		".pdf", ".doc", ".docx", ".xls", ".xlsx",
		".ppt", ".pptx", ".txt", ".md", ".rtf",
		".odt", ".ods", ".odp",
	}

	for _, format := range supportedFormats {
		if ext == format {
			return true
		}
	}

	return false
}

// GetDocumentInfoString 获取文档信息的字符串表示
func GetDocumentInfoString(metadata *DocumentMetadata) string {
	if metadata == nil {
		return "No metadata available"
	}

	var info strings.Builder

	if metadata.Title != "" {
		info.WriteString(fmt.Sprintf("Title: %s\n", metadata.Title))
	}

	if metadata.Author != "" {
		info.WriteString(fmt.Sprintf("Author: %s\n", metadata.Author))
	}

	if metadata.Subject != "" {
		info.WriteString(fmt.Sprintf("Subject: %s\n", metadata.Subject))
	}

	if metadata.PageCount > 0 {
		info.WriteString(fmt.Sprintf("Pages: %d\n", metadata.PageCount))
	}

	if metadata.Creator != "" {
		info.WriteString(fmt.Sprintf("Creator: %s\n", metadata.Creator))
	}

	if metadata.Producer != "" {
		info.WriteString(fmt.Sprintf("Producer: %s\n", metadata.Producer))
	}

	if !metadata.Created.IsZero() {
		info.WriteString(fmt.Sprintf("Created: %s\n", metadata.Created.Format("2006-01-02 15:04:05")))
	}

	if !metadata.Modified.IsZero() {
		info.WriteString(fmt.Sprintf("Modified: %s\n", metadata.Modified.Format("2006-01-02 15:04:05")))
	}

	info.WriteString(fmt.Sprintf("File Size: %s\n", FormatBytes(metadata.FileSize)))
	info.WriteString(fmt.Sprintf("File Modified: %s", metadata.ModTime.Format("2006-01-02 15:04:05")))

	return info.String()
}

// GetDocumentType 获取文档类型
func GetDocumentType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))

	documentTypes := map[string]string{
		".pdf":   "PDF Document",
		".doc":   "Word Document (Legacy)",
		".docx":  "Word Document",
		".xls":   "Excel Spreadsheet (Legacy)",
		".xlsx":  "Excel Spreadsheet",
		".ppt":   "PowerPoint Presentation (Legacy)",
		".pptx":  "PowerPoint Presentation",
		".txt":   "Plain Text",
		".md":    "Markdown",
		".rtf":   "Rich Text Format",
		".odt":   "OpenDocument Text",
		".ods":   "OpenDocument Spreadsheet",
		".odp":   "OpenDocument Presentation",
	}

	if docType, exists := documentTypes[ext]; exists {
		return docType
	}

	return "Unknown Document"
}

// ExtractBasicDocumentInfo 提取基本的文档信息
func ExtractBasicDocumentInfo(filePath string) (title string, pageCount int, err error) {
	// 获取文件类型
	fileType := getDocumentFileType(filePath)

	// 提取元数据
	metadata, err := ExtractDocumentMetadata(filePath, fileType)
	if err != nil {
		return "", 0, err
	}

	return metadata.Title, metadata.PageCount, nil
}

// getDocumentFileType 根据文件扩展名获取文件类型
func getDocumentFileType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))

	fileTypes := map[string]string{
		".pdf":   "pdf",
		".doc":   "docx",
		".docx":  "docx",
		".xls":   "xlsx",
		".xlsx":  "xlsx",
		".ppt":   "pptx",
		".pptx":  "pptx",
		".txt":   "txt",
		".md":    "md",
		".rtf":   "document",
		".odt":   "document",
		".ods":   "document",
		".odp":   "document",
	}

	if fileType, exists := fileTypes[ext]; exists {
		return fileType
	}

	return "document"
}