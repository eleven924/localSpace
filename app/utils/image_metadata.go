package utils

import (
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ImageMetadata 包含图片的元数据信息
type ImageMetadata struct {
	Width      int       `json:"width"`
	Height     int       `json:"height"`
	Format     string    `json:"format"`
	ColorSpace string    `json:"colorSpace,omitempty"`
	HasAlpha   bool      `json:"hasAlpha"`
	FileSize   int64     `json:"fileSize"`
	ModTime    time.Time `json:"modTime"`
}

// ExtractImageMetadata 提取图片文件的元数据
func ExtractImageMetadata(filePath string) (*ImageMetadata, error) {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// 解码图片配置
	config, format, err := image.DecodeConfig(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// 判断是否有 Alpha 通道
	hasAlpha := false
	switch format {
	case "png", "gif":
		hasAlpha = true
	case "jpeg", "jpg":
		hasAlpha = false
	}

	// 判断色彩空间
	colorSpace := detectColorSpace(config)

	metadata := &ImageMetadata{
		Width:      config.Width,
		Height:     config.Height,
		Format:     format,
		ColorSpace: colorSpace,
		HasAlpha:   hasAlpha,
		FileSize:   fileInfo.Size(),
		ModTime:    fileInfo.ModTime(),
	}

	return metadata, nil
}

// detectColorSpace 检测图片的色彩空间
func detectColorSpace(config image.Config) string {
	// 根据色彩模型判断
	if config.ColorModel == nil {
		return "Unknown"
	}

	// 尝试将一个颜色转换到该模型
	// 这可以帮助我们判断模型的类型
	testColor := color.NRGBA{0, 0, 0, 255}
	converted := config.ColorModel.Convert(testColor)

	// 检查转换后的颜色
	if _, _, _, a := converted.RGBA(); a != 0 {
		// 模型支持 Alpha 通道
		if config.Width > 0 && config.Height > 0 {
			// 对于有内容的图片，检查是否真的使用了 alpha
			return "RGBA"
		}
	}

	// 检查是否是灰度图
	// 如果模型是 YCbCr 或类似的，可能是 JPEG 的原生格式
	// 这里做一个简单的判断
	if strings.Contains(fmt.Sprintf("%T", config.ColorModel), "YCbCr") {
		return "YCbCr (JPEG)"
	}

	// 默认情况下返回 RGB
	return "RGB"
}

// GetImageAspectRatio 计算图片的宽高比
func GetImageAspectRatio(width, height int) float64 {
	if height == 0 {
		return 0
	}
	return float64(width) / float64(height)
}

// FormatImageDimensions 格式化图片尺寸
func FormatImageDimensions(width, height int) string {
	return fmt.Sprintf("%dx%d", width, height)
}

// IsImageSupported 检查图片格式是否支持
func IsImageSupported(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	supportedFormats := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"}

	for _, format := range supportedFormats {
		if ext == format {
			return true
		}
	}

	return false
}

// GetImageInfoString 获取图片信息的字符串表示
func GetImageInfoString(metadata *ImageMetadata) string {
	if metadata == nil {
		return "No metadata available"
	}

	var info strings.Builder

	info.WriteString(fmt.Sprintf("Format: %s\n", metadata.Format))
	info.WriteString(fmt.Sprintf("Dimensions: %dx%d\n", metadata.Width, metadata.Height))
	info.WriteString(fmt.Sprintf("Aspect Ratio: %.2f\n", GetImageAspectRatio(metadata.Width, metadata.Height)))

	if metadata.ColorSpace != "" {
		info.WriteString(fmt.Sprintf("Color Space: %s\n", metadata.ColorSpace))
	}

	info.WriteString(fmt.Sprintf("Has Alpha: %t\n", metadata.HasAlpha))
	info.WriteString(fmt.Sprintf("File Size: %s\n", FormatBytes(metadata.FileSize)))
	info.WriteString(fmt.Sprintf("Modified: %s", metadata.ModTime.Format("2006-01-02 15:04:05")))

	return info.String()
}

// FormatBytes 格式化字节数为可读字符串
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// ExtractBasicImageInfo 提取基本的图片信息（不包含元数据）
func ExtractBasicImageInfo(filePath string) (width, height int, format string, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, 0, "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	config, fmtStr, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, "", fmt.Errorf("failed to decode image: %w", err)
	}

	return config.Width, config.Height, fmtStr, nil
}