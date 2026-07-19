package services

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"LocalSpace/app/models"
)

func TestSupportsGeneratedThumbnailOnlyAllowsImageAndVideo(t *testing.T) {
	for _, fileType := range []string{"image", "video"} {
		if !supportsGeneratedThumbnail(fileType) {
			t.Fatalf("expected %q to support generated thumbnails", fileType)
		}
	}

	for _, fileType := range []string{"document", "music", "game", "other"} {
		if supportsGeneratedThumbnail(fileType) {
			t.Fatalf("expected %q to skip generated thumbnails", fileType)
		}
	}
}

func TestGetThumbnailRejectsUnsupportedFileTypes(t *testing.T) {
	service := &ThumbnailService{
		cacheDir: t.TempDir(),
		cache:    make(map[string]*ThumbnailCacheEntry),
	}

	_, err := service.GetThumbnail("missing.txt", "document", 1)
	if err == nil {
		t.Fatal("expected unsupported file type error")
	}
	if !strings.Contains(err.Error(), "only supported for image and video") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestThumbnailCacheKeyUsesPNGExtension(t *testing.T) {
	service := &ThumbnailService{}

	got := service.generateCacheKey(2, "video")
	if got != "2_video_wide.png" {
		t.Fatalf("expected cache key to include png extension, got %q", got)
	}
}

func TestRemoveThumbnailDeletesCurrentAndLegacyCacheFiles(t *testing.T) {
	cacheDir := t.TempDir()
	service := &ThumbnailService{
		cacheDir: cacheDir,
		cache:    make(map[string]*ThumbnailCacheEntry),
	}

	currentKey := service.generateCacheKey(2, "video")
	legacyKeys := service.generateLegacyCacheKeys(2, "video")
	currentPath := filepath.Join(cacheDir, currentKey)

	if err := os.WriteFile(currentPath, []byte("current"), 0644); err != nil {
		t.Fatalf("failed to write current thumbnail: %v", err)
	}
	for _, legacyKey := range legacyKeys {
		legacyPath := filepath.Join(cacheDir, legacyKey)
		if err := os.WriteFile(legacyPath, []byte("legacy"), 0644); err != nil {
			t.Fatalf("failed to write legacy thumbnail: %v", err)
		}
		service.cache[legacyKey] = &ThumbnailCacheEntry{Path: legacyPath, Size: 6, CreatedAt: time.Now()}
		service.currentSize += 6
	}

	service.cache[currentKey] = &ThumbnailCacheEntry{Path: currentPath, Size: 7, CreatedAt: time.Now()}
	service.currentSize += 7

	if err := service.RemoveThumbnail(2, "video"); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if _, err := os.Stat(currentPath); !os.IsNotExist(err) {
		t.Fatalf("expected current thumbnail to be removed, stat err: %v", err)
	}
	for _, legacyKey := range legacyKeys {
		legacyPath := filepath.Join(cacheDir, legacyKey)
		if _, err := os.Stat(legacyPath); !os.IsNotExist(err) {
			t.Fatalf("expected legacy thumbnail to be removed, stat err: %v", err)
		}
	}
	if service.currentSize != 0 {
		t.Fatalf("expected current size to be 0, got %d", service.currentSize)
	}
}

func TestThumbnailExistsRejectsLegacyPathWithoutExtension(t *testing.T) {
	cacheDir := t.TempDir()
	legacyPath := filepath.Join(cacheDir, "2_video")
	if err := os.WriteFile(legacyPath, []byte("legacy"), 0644); err != nil {
		t.Fatalf("failed to write legacy thumbnail: %v", err)
	}

	service := &FileService{}
	if service.thumbnailExists(legacyPath) {
		t.Fatal("expected legacy thumbnail path without extension to be treated as missing")
	}
}

func TestThumbnailIsCurrentRejectsPreviousPNGCacheKey(t *testing.T) {
	cacheDir := t.TempDir()
	previousPath := filepath.Join(cacheDir, "2_video.png")
	if err := os.WriteFile(previousPath, []byte("previous"), 0644); err != nil {
		t.Fatalf("failed to write previous thumbnail: %v", err)
	}

	thumbnailService := &ThumbnailService{}
	service := &FileService{thumbnailService: thumbnailService}
	file := &models.File{
		ID:        2,
		FileType:  "video",
		Thumbnail: previousPath,
	}

	if service.thumbnailIsCurrent(file) {
		t.Fatal("expected previous png cache key to be treated as stale")
	}
}

func TestThumbnailDataURIEncodesThumbnailFile(t *testing.T) {
	cacheDir := t.TempDir()
	thumbnailPath := filepath.Join(cacheDir, "2_video.png")
	content := []byte("png data")
	if err := os.WriteFile(thumbnailPath, content, 0644); err != nil {
		t.Fatalf("failed to write thumbnail: %v", err)
	}

	service := &FileService{}
	got, err := service.thumbnailDataURI(thumbnailPath)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	wantPrefix := "data:image/png;base64,"
	if !strings.HasPrefix(got, wantPrefix) {
		t.Fatalf("expected data URI prefix %q, got %q", wantPrefix, got)
	}

	encoded := strings.TrimPrefix(got, wantPrefix)
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("failed to decode data URI: %v", err)
	}
	if string(decoded) != string(content) {
		t.Fatalf("expected decoded content %q, got %q", content, decoded)
	}
}
