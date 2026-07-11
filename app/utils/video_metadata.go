package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// VideoMetadata 包含视频的元数据信息
type VideoMetadata struct {
	Duration   float64   `json:"duration"`    // 秒
	Width      int       `json:"width"`
	Height     int       `json:"height"`
	BitRate    int64     `json:"bitRate"`    // bps
	FrameRate  float64   `json:"frameRate"`  // fps
	Codec      string    `json:"codec"`
	AudioCodec string    `json:"audioCodec"`
	Format     string    `json:"format"`
	FileSize   int64     `json:"fileSize"`
	ModTime    time.Time `json:"modTime"`
	HasAudio   bool      `json:"hasAudio"`
}

// ffprobeOutput ffprobe JSON 输出结构
type ffprobeOutput struct {
	Streams []struct {
		CodecType string `json:"codec_type"`
		CodecName string `json:"codec_name"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
		BitRate   string `json:"bit_rate"`
		RFrameRate string `json:"r_frame_rate"`
	} `json:"streams"`
	Format struct {
		Duration   string `json:"duration"`
		BitRate   string `json:"bit_rate"`
		Size      string `json:"size"`
		FormatName string `json:"format_name"`
	} `json:"format"`
}

// ExtractVideoMetadata 提取视频文件的元数据
func ExtractVideoMetadata(filePath string) (*VideoMetadata, error) {
	// 检查 ffprobe 是否可用
	if !isFFProbeAvailable() {
		return nil, fmt.Errorf("ffprobe is not available on the system")
	}

	// 获取文件信息
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// 使用 ffprobe 提取元数据
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		filePath)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute ffprobe: %w", err)
	}

	// 解析 JSON 输出
	var ffprobeData ffprobeOutput
	if err := json.Unmarshal(output, &ffprobeData); err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	// 提取视频流信息
	metadata := &VideoMetadata{
		FileSize: fileInfo.Size(),
		ModTime:  fileInfo.ModTime(),
	}

	// 解析格式信息
	if ffprobeData.Format.FormatName != "" {
		metadata.Format = ffprobeData.Format.FormatName
	}

	// 解析时长
	if ffprobeData.Format.Duration != "" {
		duration, err := strconv.ParseFloat(ffprobeData.Format.Duration, 64)
		if err == nil {
			metadata.Duration = duration
		}
	}

	// 解析比特率
	if ffprobeData.Format.BitRate != "" {
		bitRate, err := strconv.ParseInt(ffprobeData.Format.BitRate, 10, 64)
		if err == nil {
			metadata.BitRate = bitRate
		}
	}

	// 解析流信息
	for _, stream := range ffprobeData.Streams {
		if stream.CodecType == "video" {
			metadata.Width = stream.Width
			metadata.Height = stream.Height
			metadata.Codec = stream.CodecName

			// 解析帧率
			if stream.RFrameRate != "" {
				frameRate, err := parseFrameRate(stream.RFrameRate)
				if err == nil {
					metadata.FrameRate = frameRate
				}
			}
		} else if stream.CodecType == "audio" {
			metadata.HasAudio = true
			metadata.AudioCodec = stream.CodecName
		}
	}

	return metadata, nil
}

// isFFProbeAvailable 检查 ffprobe 是否可用
func isFFProbeAvailable() bool {
	cmd := exec.Command("ffprobe", "-version")
	return cmd.Run() == nil
}

// parseFrameRate 解析帧率字符串
func parseFrameRate(frameRateStr string) (float64, error) {
	// 移除可能的 fps 后缀
	frameRateStr = strings.TrimSuffix(strings.TrimSpace(frameRateStr), "fps")

	// 解析分数形式 (如 30/1)
	parts := strings.Split(frameRateStr, "/")
	if len(parts) == 2 {
		numerator, err1 := strconv.ParseFloat(parts[0], 64)
		denominator, err2 := strconv.ParseFloat(parts[1], 64)
		if err1 == nil && err2 == nil && denominator != 0 {
			return numerator / denominator, nil
		}
	}

	// 解析小数形式
	return strconv.ParseFloat(frameRateStr, 64)
}

// FormatDuration 格式化时长为可读字符串
func FormatDuration(duration float64) string {
	hours := int(duration / 3600)
	minutes := int((duration - float64(hours*3600)) / 60)
	seconds := int(duration) % 60

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

// FormatBitRate 格式化比特率为可读字符串
func FormatBitRate(bitRate int64) string {
	if bitRate == 0 {
		return "Unknown"
	}

	const unit = 1000
	if bitRate < unit {
		return fmt.Sprintf("%d bps", bitRate)
	}

	div, exp := int64(unit), 0
	for n := bitRate / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.2f %cbps", float64(bitRate)/float64(div), "kMGT"[exp])
}

// GetVideoAspectRatio 计算视频的宽高比
func GetVideoAspectRatio(width, height int) float64 {
	if height == 0 {
		return 0
	}
	return float64(width) / float64(height)
}

// IsVideoSupported 检查视频格式是否支持
func IsVideoSupported(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	supportedFormats := []string{
		".mp4", ".avi", ".mkv", ".mov", ".wmv", ".flv",
		".webm", ".m4v", ".3gp", ".ts", ".m2ts",
	}

	for _, format := range supportedFormats {
		if ext == format {
			return true
		}
	}

	return false
}

// GetVideoInfoString 获取视频信息的字符串表示
func GetVideoInfoString(metadata *VideoMetadata) string {
	if metadata == nil {
		return "No metadata available"
	}

	var info strings.Builder

	info.WriteString(fmt.Sprintf("Format: %s\n", metadata.Format))
	info.WriteString(fmt.Sprintf("Dimensions: %dx%d\n", metadata.Width, metadata.Height))
	info.WriteString(fmt.Sprintf("Aspect Ratio: %.2f\n", GetVideoAspectRatio(metadata.Width, metadata.Height)))
	info.WriteString(fmt.Sprintf("Duration: %s\n", FormatDuration(metadata.Duration)))

	if metadata.FrameRate > 0 {
		info.WriteString(fmt.Sprintf("Frame Rate: %.2f fps\n", metadata.FrameRate))
	}

	if metadata.Codec != "" {
		info.WriteString(fmt.Sprintf("Video Codec: %s\n", metadata.Codec))
	}

	if metadata.HasAudio && metadata.AudioCodec != "" {
		info.WriteString(fmt.Sprintf("Audio Codec: %s\n", metadata.AudioCodec))
	}

	info.WriteString(fmt.Sprintf("Bit Rate: %s\n", FormatBitRate(metadata.BitRate)))
	info.WriteString(fmt.Sprintf("File Size: %s\n", FormatBytes(metadata.FileSize)))
	info.WriteString(fmt.Sprintf("Modified: %s", metadata.ModTime.Format("2006-01-02 15:04:05")))

	return info.String()
}

// ExtractBasicVideoInfo 提取基本的视频信息（不包含元数据）
func ExtractBasicVideoInfo(filePath string) (width, height int, duration float64, err error) {
	if !isFFProbeAvailable() {
		return 0, 0, 0, fmt.Errorf("ffprobe is not available on the system")
	}

	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		filePath)

	output, err := cmd.Output()
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to execute ffprobe: %w", err)
	}

	var ffprobeData ffprobeOutput
	if err := json.Unmarshal(output, &ffprobeData); err != nil {
		return 0, 0, 0, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	// 解析时长
	if ffprobeData.Format.Duration != "" {
		duration, err = strconv.ParseFloat(ffprobeData.Format.Duration, 64)
		if err != nil {
			duration = 0
		}
	}

	// 解析尺寸
	for _, stream := range ffprobeData.Streams {
		if stream.CodecType == "video" {
			width = stream.Width
			height = stream.Height
			break
		}
	}

	return width, height, duration, nil
}

// GetFFProbeVersion 获取 ffprobe 版本信息
func GetFFProbeVersion() string {
	cmd := exec.Command("ffprobe", "-version")
	output, err := cmd.Output()
	if err != nil {
		return "Not available"
	}

	// 提取版本号
	re := regexp.MustCompile(`ffprobe version (\d+\.\d+(\.\d+)?)`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) > 1 {
		return matches[1]
	}

	return "Unknown"
}