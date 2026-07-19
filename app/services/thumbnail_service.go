package services

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"LocalSpace/app/utils"
)

// ThumbnailService manages cached thumbnails for supported file types.
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

// ThumbnailCacheEntry represents one cached thumbnail file.
type ThumbnailCacheEntry struct {
	Path         string
	LastAccessed time.Time
	CreatedAt    time.Time
	Size         int64
}

func supportsGeneratedThumbnail(fileType string) bool {
	switch fileType {
	case "image", "video":
		return true
	default:
		return false
	}
}

// NewThumbnailService creates a new thumbnail service.
func NewThumbnailService(cacheDir string) *ThumbnailService {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		panic(fmt.Sprintf("failed to create thumbnail cache directory: %v", err))
	}

	service := &ThumbnailService{
		cacheDir:      cacheDir,
		cache:         make(map[string]*ThumbnailCacheEntry),
		maxCacheSize:  100 * 1024 * 1024,
		defaultWidth:  300,
		defaultHeight: 300,
		maxAge:        7 * 24 * time.Hour,
	}

	go service.cleanupExpiredEntries()
	service.loadExistingCache()

	return service
}

// GetThumbnail returns a cached thumbnail path or generates one on demand.
func (s *ThumbnailService) GetThumbnail(filePath, fileType string, fileID uint) (string, error) {
	if !supportsGeneratedThumbnail(fileType) {
		return "", fmt.Errorf("thumbnail generation is only supported for image and video files")
	}

	cacheKey := s.generateCacheKey(fileID, fileType)
	cachePath := filepath.Join(s.cacheDir, cacheKey)

	s.cacheMutex.RLock()
	entry, exists := s.cache[cacheKey]
	if exists {
		entry.LastAccessed = time.Now()
		s.cacheMutex.RUnlock()
		return cachePath, nil
	}
	s.cacheMutex.RUnlock()

	if err := s.generateThumbnail(filePath, fileType, cachePath); err != nil {
		return "", fmt.Errorf("failed to generate thumbnail: %w", err)
	}

	fileInfo, err := os.Stat(cachePath)
	if err != nil {
		return cachePath, nil
	}

	s.cacheMutex.Lock()
	s.cache[cacheKey] = &ThumbnailCacheEntry{
		Path:         cachePath,
		LastAccessed: time.Now(),
		CreatedAt:    time.Now(),
		Size:         fileInfo.Size(),
	}
	s.currentSize += fileInfo.Size()
	s.cacheMutex.Unlock()

	s.checkCacheSize()

	return cachePath, nil
}

// GetThumbnailAsync returns the thumbnail path through a channel.
func (s *ThumbnailService) GetThumbnailAsync(filePath, fileType string, fileID uint) <-chan string {
	resultChan := make(chan string, 1)

	go func() {
		thumbnailPath, err := s.GetThumbnail(filePath, fileType, fileID)
		if err != nil {
			resultChan <- ""
		} else {
			resultChan <- thumbnailPath
		}
		close(resultChan)
	}()

	return resultChan
}

// RemoveThumbnail removes current and legacy cache files for a file.
func (s *ThumbnailService) RemoveThumbnail(fileID uint, fileType string) error {
	cacheKey := s.generateCacheKey(fileID, fileType)

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

	for _, key := range append([]string{cacheKey}, legacyCacheKeys...) {
		cachePath := filepath.Join(s.cacheDir, key)
		if err := os.Remove(cachePath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove thumbnail file: %w", err)
		}
	}

	return nil
}

// ClearCache removes all cached thumbnails.
func (s *ThumbnailService) ClearCache() error {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	for _, entry := range s.cache {
		if err := os.Remove(entry.Path); err != nil && !os.IsNotExist(err) {
			fmt.Printf("Warning: Failed to remove thumbnail file %s: %v\n", entry.Path, err)
		}
	}

	s.cache = make(map[string]*ThumbnailCacheEntry)
	s.currentSize = 0

	return nil
}

// GetCacheInfo returns cache count, size, and limit.
func (s *ThumbnailService) GetCacheInfo() (count int, size int64, maxSize int64) {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()

	return len(s.cache), s.currentSize, s.maxCacheSize
}

// SetMaxCacheSize updates the max cache size.
func (s *ThumbnailService) SetMaxCacheSize(maxSize int64) {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()
	s.maxCacheSize = maxSize
	s.checkCacheSize()
}

// SetMaxAge updates the max cache age.
func (s *ThumbnailService) SetMaxAge(maxAge time.Duration) {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()
	s.maxAge = maxAge
	go s.cleanupExpiredEntries()
}

func (s *ThumbnailService) generateThumbnail(filePath, fileType, outputPath string) error {
	switch fileType {
	case "image":
		return utils.GenerateThumbnail(filePath, outputPath, s.defaultWidth, s.defaultHeight)
	case "video":
		return utils.GenerateVideoThumbnail(filePath, outputPath, s.defaultWidth, s.defaultHeight)
	default:
		return fmt.Errorf("thumbnail generation is only supported for image and video files")
	}
}

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

func (s *ThumbnailService) checkCacheSize() {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	if s.currentSize > s.maxCacheSize {
		s.evictLeastRecentlyUsed()
	}
}

func (s *ThumbnailService) evictLeastRecentlyUsed() {
	if len(s.cache) == 0 {
		return
	}

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

func (s *ThumbnailService) cleanupExpiredEntries() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		s.removeExpiredEntries()
	}
}

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

		if time.Since(fileInfo.ModTime()) > s.maxAge {
			_ = os.Remove(filePath)
			continue
		}

		s.cache[entry.Name()] = &ThumbnailCacheEntry{
			Path:         filePath,
			LastAccessed: fileInfo.ModTime(),
			CreatedAt:    fileInfo.ModTime(),
			Size:         fileInfo.Size(),
		}
		s.currentSize += fileInfo.Size()
	}
}

// GetCacheStats returns detailed cache stats.
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

// WarmupCache pre-generates thumbnails for the provided files.
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
