package services

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"LocalSpace/app/utils"
)

// ThumbnailService 缩略图缓存服务
type ThumbnailService struct {
	cacheDir      string
	cache         map[string]*ThumbnailCacheEntry
	cacheMutex    sync.RWMutex
	maxCacheSize  int64
	currentSize   int64
	defaultWidth  int
	defaultHeight int
	maxAge        time.Duration
}

// ThumbnailCacheEntry 缓存条目
type ThumbnailCacheEntry struct {
	Path         string
	LastAccessed time.Time
	CreatedAt    time.Time
	Size         int64
}

// NewThumbnailService 创建新的缩略图服务
func NewThumbnailService(cacheDir string) *ThumbnailService {
	// 确保缓存目录存在
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		panic(fmt.Sprintf("failed to create thumbnail cache directory: %v", err))
	}

	service := &ThumbnailService{
		cacheDir:      cacheDir,
		cache:         make(map[string]*ThumbnailCacheEntry),
		maxCacheSize:  100 * 1024 * 1024, // 100MB 默认缓存大小
		defaultWidth:  300,
		defaultHeight: 300,
		maxAge:        7 * 24 * time.Hour, // 7天最大缓存时间
	}

	// 启动缓存清理协程
	go service.cleanupExpiredEntries()

	// 加载现有缓存
	service.loadExistingCache()

	return service
}

// GetThumbnail 获取或生成缩略图
func (s *ThumbnailService) GetThumbnail(filePath, fileType string, fileID uint) (string, error) {
	// 生成缓存键
	cacheKey := s.generateCacheKey(fileID, fileType)
	cachePath := filepath.Join(s.cacheDir, cacheKey)

	// 检查缓存
	s.cacheMutex.RLock()
	entry, exists := s.cache[cacheKey]
	if exists {
		entry.LastAccessed = time.Now()
		s.cacheMutex.RUnlock()
		return cachePath, nil
	}
	s.cacheMutex.RUnlock()

	// 缓存不存在，生成缩略图
	if err := s.generateThumbnail(filePath, fileType, cachePath); err != nil {
		return "", fmt.Errorf("failed to generate thumbnail: %w", err)
	}

	// 获取文件信息
	fileInfo, err := os.Stat(cachePath)
	if err != nil {
		return cachePath, nil // 文件信息获取失败不影响返回
	}

	// 添加到缓存
	s.cacheMutex.Lock()
	s.cache[cacheKey] = &ThumbnailCacheEntry{
		Path:         cachePath,
		LastAccessed: time.Now(),
		CreatedAt:    time.Now(),
		Size:         fileInfo.Size(),
	}
	s.currentSize += fileInfo.Size()
	s.cacheMutex.Unlock()

	// 检查缓存大小，必要时清理
	s.checkCacheSize()

	return cachePath, nil
}

// GetThumbnailAsync 异步获取或生成缩略图
func (s *ThumbnailService) GetThumbnailAsync(filePath, fileType string, fileID uint) <-chan string {
	resultChan := make(chan string, 1)

	go func() {
		thumbnailPath, err := s.GetThumbnail(filePath, fileType, fileID)
		if err != nil {
			// 生成失败，返回空字符串
			resultChan <- ""
		} else {
			resultChan <- thumbnailPath
		}
		close(resultChan)
	}()

	return resultChan
}

// RemoveThumbnail 移除指定文件的缩略图
func (s *ThumbnailService) RemoveThumbnail(fileID uint, fileType string) error {
	cacheKey := s.generateCacheKey(fileID, fileType)

	// 从缓存中移除
	s.cacheMutex.Lock()
	if entry, exists := s.cache[cacheKey]; exists {
		s.currentSize -= entry.Size
		delete(s.cache, cacheKey)
	}
	legacyCacheKeys := s.generateLegacyCacheKeys(fileID, fileType)
	for _, legacyCacheKey := range legacyCacheKeys {
		if entry, exists := s.cache[legacyCacheKey]; exists {
			s.currentSize -= entry.Size
			delete(s.cache, legacyCacheKey)
		}
	}
	s.cacheMutex.Unlock()

	// 删除文件
	for _, key := range append([]string{cacheKey}, legacyCacheKeys...) {
		cachePath := filepath.Join(s.cacheDir, key)
		if err := os.Remove(cachePath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove thumbnail file: %w", err)
		}
	}

	return nil
}

// ClearCache 清空所有缓存
func (s *ThumbnailService) ClearCache() error {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	// 删除所有缓存文件
	for _, entry := range s.cache {
		if err := os.Remove(entry.Path); err != nil && !os.IsNotExist(err) {
			// 记录错误但继续
			fmt.Printf("Warning: Failed to remove thumbnail file %s: %v\n", entry.Path, err)
		}
	}

	// 清空缓存
	s.cache = make(map[string]*ThumbnailCacheEntry)
	s.currentSize = 0

	return nil
}

// GetCacheInfo 获取缓存信息
func (s *ThumbnailService) GetCacheInfo() (count int, size int64, maxSize int64) {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()

	return len(s.cache), s.currentSize, s.maxCacheSize
}

// SetMaxCacheSize 设置最大缓存大小
func (s *ThumbnailService) SetMaxCacheSize(maxSize int64) {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()
	s.maxCacheSize = maxSize
	s.checkCacheSize()
}

// SetMaxAge 设置最大缓存时间
func (s *ThumbnailService) SetMaxAge(maxAge time.Duration) {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()
	s.maxAge = maxAge
	go s.cleanupExpiredEntries()
}

// generateThumbnail 生成缩略图
func (s *ThumbnailService) generateThumbnail(filePath, fileType, outputPath string) error {
	switch fileType {
	case "image":
		return utils.GenerateThumbnail(filePath, outputPath, s.defaultWidth, s.defaultHeight)
	case "video":
		return utils.GenerateVideoThumbnail(filePath, outputPath, s.defaultWidth, s.defaultHeight)
	case "document":
		return utils.GenerateDocumentThumbnail(filePath, outputPath, s.defaultWidth, s.defaultHeight)
	case "music":
		return utils.GenerateAudioThumbnail(filePath, outputPath, s.defaultWidth, s.defaultHeight)
	default:
		return utils.GeneratePlaceholderThumbnail(outputPath, s.defaultWidth, s.defaultHeight)
	}
}

// generateCacheKey 生成缓存键
func (s *ThumbnailService) generateCacheKey(fileID uint, fileType string) string {
	return fmt.Sprintf("%d_%s_wide.png", fileID, fileType)
}

func (s *ThumbnailService) generateLegacyCacheKeys(fileID uint, fileType string) []string {
	return []string{
		fmt.Sprintf("%d_%s_square.png", fileID, fileType),
		fmt.Sprintf("%d_%s.png", fileID, fileType),
		fmt.Sprintf("%d_%s", fileID, fileType),
	}
}

// checkCacheSize 检查缓存大小，必要时清理
func (s *ThumbnailService) checkCacheSize() {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	// 如果缓存大小超过限制，清理最少使用的条目
	if s.currentSize > s.maxCacheSize {
		s.evictLeastRecentlyUsed()
	}
}

// evictLeastRecentlyUsed 清除最少使用的缓存条目
func (s *ThumbnailService) evictLeastRecentlyUsed() {
	if len(s.cache) == 0 {
		return
	}

	// 找到最少使用的条目
	var lruKey string
	var lruTime time.Time
	first := true

	for key, entry := range s.cache {
		if first || entry.LastAccessed.Before(lruTime) {
			lruKey = key
			lruTime = entry.LastAccessed
			first = false
		}
	}

	// 删除 LRU 条目
	if lruKey != "" {
		if entry, exists := s.cache[lruKey]; exists {
			if err := os.Remove(entry.Path); err != nil && !os.IsNotExist(err) {
				fmt.Printf("Warning: Failed to remove LRU thumbnail: %v\n", err)
			}
			s.currentSize -= entry.Size
		}
		delete(s.cache, lruKey)
	}
}

// cleanupExpiredEntries 清理过期条目
func (s *ThumbnailService) cleanupExpiredEntries() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		s.removeExpiredEntries()
	}
}

// removeExpiredEntries 移除过期条目
func (s *ThumbnailService) removeExpiredEntries() {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	expiredKeys := make([]string, 0)
	now := time.Now()

	for key, entry := range s.cache {
		if now.Sub(entry.CreatedAt) > s.maxAge {
			expiredKeys = append(expiredKeys, key)
		}
	}

	// 删除过期条目
	for _, key := range expiredKeys {
		if entry, exists := s.cache[key]; exists {
			if err := os.Remove(entry.Path); err != nil && !os.IsNotExist(err) {
				fmt.Printf("Warning: Failed to remove expired thumbnail: %v\n", err)
			}
			s.currentSize -= entry.Size
		}
		delete(s.cache, key)
	}
}

// loadExistingCache 加载现有缓存
func (s *ThumbnailService) loadExistingCache() {
	entries, err := os.ReadDir(s.cacheDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(s.cacheDir, entry.Name())
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			continue
		}

		// 检查是否过期
		if time.Since(fileInfo.ModTime()) > s.maxAge {
			os.Remove(filePath)
			continue
		}

		// 添加到缓存
		s.cache[entry.Name()] = &ThumbnailCacheEntry{
			Path:         filePath,
			LastAccessed: fileInfo.ModTime(),
			CreatedAt:    fileInfo.ModTime(),
			Size:         fileInfo.Size(),
		}
		s.currentSize += fileInfo.Size()
	}
}

// GetCacheStats 获取缓存统计信息
func (s *ThumbnailService) GetCacheStats() map[string]interface{} {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()

	return map[string]interface{}{
		"count":         len(s.cache),
		"currentSize":   s.currentSize,
		"maxSize":       s.maxCacheSize,
		"usagePercent":  float64(s.currentSize) / float64(s.maxCacheSize) * 100,
		"maxAge":        s.maxAge.String(),
		"defaultWidth":  s.defaultWidth,
		"defaultHeight": s.defaultHeight,
	}
}

// WarmupCache 预热缓存（为指定文件生成缩略图）
func (s *ThumbnailService) WarmupCache(files []struct {
	FilePath string
	FileType string
	FileID   uint
}) (success, failed int) {
	for _, file := range files {
		_, err := s.GetThumbnail(file.FilePath, file.FileType, file.FileID)
		if err != nil {
			failed++
		} else {
			success++
		}
	}
	return success, failed
}
