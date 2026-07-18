package utils

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
)

// ThumbnailSize 默认缩略图尺寸
const (
	DefaultThumbnailWidth  = 300
	DefaultThumbnailHeight = 169
	ThumbnailQuality       = 85 // JPEG 质量 (1-100)
)

// GenerateThumbnail 生成图片缩略图
func GenerateThumbnail(filePath, outputPath string, width, height int) error {
	// 检查输入文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("source file does not exist: %s", filePath)
	}

	// 确保输出目录存在
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// 打开源图片
	src, err := imaging.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open source image: %w", err)
	}

	// 生成固定比例缩略图，必要时居中裁切，避免显示时出现留白。
	dst := imaging.Fill(src, width, height, imaging.Center, imaging.Lanczos)

	// 根据输出格式保存
	ext := strings.ToLower(filepath.Ext(outputPath))
	switch ext {
	case ".png":
		return imaging.Save(dst, outputPath)
	case ".jpg", ".jpeg":
		return imaging.Save(dst, outputPath, imaging.JPEGQuality(ThumbnailQuality))
	default:
		// 默认保存为 PNG
		return imaging.Save(dst, outputPath)
	}
}

// GenerateVideoThumbnail 生成视频缩略图
func GenerateVideoThumbnail(filePath, outputPath string, width, height int) error {
	// 检查 ffmpeg 是否可用
	if !isFFmpegAvailable() {
		return fmt.Errorf("ffmpeg is not available on the system")
	}

	// 检查输入文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("source file does not exist: %s", filePath)
	}

	// 确保输出目录存在
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// 使用 ffmpeg 生成缩略图（从第 1 秒开始）
	cmd := exec.Command("ffmpeg",
		"-i", filePath,
		"-ss", "00:00:01",
		"-vframes", "1",
		"-vf", fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=increase,crop=%d:%d", width, height, width, height),
		"-f", "image2",
		"-update", "1",
		"-y", // 覆盖已存在的文件
		outputPath,
	)

	// 运行命令
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to generate video thumbnail: %w: %s", err, strings.TrimSpace(string(output)))
	}

	return nil
}

// GenerateDocumentThumbnail 生成文档缩略图
func GenerateDocumentThumbnail(filePath, outputPath string, width, height int) error {
	// 文档缩略图生成需要额外的库支持
	// 这里提供基础实现，可以根据需要扩展

	// 对于 PDF 文档，可以使用 pdftoppm 或其他工具
	// 对于 Office 文档，可以使用 LibreOffice 或其他工具

	// 当前版本返回占位图
	return GeneratePlaceholderThumbnail(outputPath, width, height)
}

// GenerateAudioThumbnail 生成音频缩略图
func GenerateAudioThumbnail(filePath, outputPath string, width, height int) error {
	// 音频缩略图通常是固定的图标
	return generateAudioIconThumbnail(outputPath, width, height)
}

// GeneratePlaceholderThumbnail 生成占位缩略图
func GeneratePlaceholderThumbnail(outputPath string, width, height int) error {
	// 创建一个简单的占位图
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// 填充背景色
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetRGBA(x, y, color.RGBA{240, 240, 240, 255})
		}
	}

	// 确保输出目录存在
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// 保存为 PNG
	return imaging.Save(img, outputPath)
}

// generateAudioIconThumbnail 生成音频图标缩略图
func generateAudioIconThumbnail(outputPath string, width, height int) error {
	// 创建一个简单的音频图标
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// 填充背景色
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetRGBA(x, y, color.RGBA{100, 181, 246, 255})
		}
	}

	// 在中心绘制简单的音符符号（简化版）
	centerX, centerY := width/2, height/2
	size := min(width, height) / 4

	for y := centerY - size; y < centerY+size; y++ {
		for x := centerX - size; x < centerX+size; x++ {
			// 简单的圆形
			dx, dy := float64(x-centerX), float64(y-centerY)
			if dx*dx+dy*dy < float64(size*size) {
				img.SetRGBA(x, y, color.RGBA{255, 255, 255, 255})
			}
		}
	}

	// 确保输出目录存在
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// 保存为 PNG
	return imaging.Save(img, outputPath)
}

// IsThumbnailSupported 检查文件类型是否支持生成缩略图
func IsThumbnailSupported(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	supportedFormats := []string{
		".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp",
		".mp4", ".avi", ".mkv", ".mov", ".wmv", ".flv", ".webm", ".m4v",
		".pdf",
		".mp3", ".wav", ".flac", ".aac", ".ogg", ".m4a",
	}

	for _, format := range supportedFormats {
		if ext == format {
			return true
		}
	}

	return false
}

// GetThumbnailFileType 获取缩略图文件类型
func GetThumbnailFileType(originalPath string) string {
	// 默认使用 PNG 格式
	return "png"
}

// GenerateThumbnailForFile 根据文件类型生成缩略图
func GenerateThumbnailForFile(filePath, fileType string) (string, error) {
	if !IsThumbnailSupported(filePath) {
		return "", fmt.Errorf("unsupported file type for thumbnail generation")
	}

	// 生成缓存键（使用文件路径的哈希）
	cacheKey := generateCacheKey(filePath)
	outputDir := filepath.Join(os.TempDir(), "localspace", "thumbnails")
	outputPath := filepath.Join(outputDir, cacheKey+".png")

	// 检查缓存
	if _, err := os.Stat(outputPath); err == nil {
		return outputPath, nil
	}

	// 确保缓存目录存在
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	// 根据文件类型生成缩略图
	var err error
	switch fileType {
	case "image":
		err = GenerateThumbnail(filePath, outputPath, DefaultThumbnailWidth, DefaultThumbnailHeight)
	case "video":
		err = GenerateVideoThumbnail(filePath, outputPath, DefaultThumbnailWidth, DefaultThumbnailHeight)
	case "document":
		err = GenerateDocumentThumbnail(filePath, outputPath, DefaultThumbnailWidth, DefaultThumbnailHeight)
	case "music":
		err = GenerateAudioThumbnail(filePath, outputPath, DefaultThumbnailWidth, DefaultThumbnailHeight)
	default:
		err = GeneratePlaceholderThumbnail(outputPath, DefaultThumbnailWidth, DefaultThumbnailHeight)
	}

	if err != nil {
		return "", fmt.Errorf("failed to generate thumbnail: %w", err)
	}

	return outputPath, nil
}

// ClearThumbnailCache 清空缩略图缓存
func ClearThumbnailCache() error {
	cacheDir := filepath.Join(os.TempDir(), "localspace", "thumbnails")
	return os.RemoveAll(cacheDir)
}

// GetThumbnailCacheSize 获取缩略图缓存大小
func GetThumbnailCacheSize() (int64, error) {
	cacheDir := filepath.Join(os.TempDir(), "localspace", "thumbnails")
	return getDirSize(cacheDir)
}

// generateCacheKey 生成缓存键
func generateCacheKey(filePath string) string {
	// 使用文件路径的简单哈希作为缓存键
	// 这里使用文件路径的 base64 编码作为键
	return strings.ReplaceAll(strings.ReplaceAll(filePath, "/", "_"), "\\", "_")
}

// getDirSize 获取目录大小
func getDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

// isFFmpegAvailable 检查 ffmpeg 是否可用
func isFFmpegAvailable() bool {
	cmd := exec.Command("ffmpeg", "-version")
	return cmd.Run() == nil
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetThumbnailQuality 获取缩略图质量
func GetThumbnailQuality() int {
	return ThumbnailQuality
}

// SetThumbnailQuality 设置缩略图质量
func SetThumbnailQuality(quality int) error {
	if quality < 1 || quality > 100 {
		return fmt.Errorf("quality must be between 1 and 100")
	}
	// 注意：这里需要修改全局变量，实际实现中应该使用配置
	return nil
}
