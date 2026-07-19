package utils

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateThumbnailOutputsRequestedCanvasSize(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "source.png")
	outputPath := filepath.Join(dir, "thumbnail.png")

	source := image.NewRGBA(image.Rect(0, 0, 640, 360))
	for y := 0; y < source.Bounds().Dy(); y++ {
		for x := 0; x < source.Bounds().Dx(); x++ {
			source.SetRGBA(x, y, color.RGBA{10, 120, 200, 255})
		}
	}

	sourceFile, err := os.Create(sourcePath)
	if err != nil {
		t.Fatalf("failed to create source image: %v", err)
	}
	if err := png.Encode(sourceFile, source); err != nil {
		sourceFile.Close()
		t.Fatalf("failed to encode source image: %v", err)
	}
	if err := sourceFile.Close(); err != nil {
		t.Fatalf("failed to close source image: %v", err)
	}

	if err := GenerateThumbnail(sourcePath, outputPath, DefaultThumbnailWidth, DefaultThumbnailHeight); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	outputFile, err := os.Open(outputPath)
	if err != nil {
		t.Fatalf("failed to open output thumbnail: %v", err)
	}
	defer outputFile.Close()

	output, _, err := image.Decode(outputFile)
	if err != nil {
		t.Fatalf("failed to decode output thumbnail: %v", err)
	}

	if got := output.Bounds().Dx(); got != DefaultThumbnailWidth {
		t.Fatalf("expected output width %d, got %d", DefaultThumbnailWidth, got)
	}
	if got := output.Bounds().Dy(); got != DefaultThumbnailHeight {
		t.Fatalf("expected output height %d, got %d", DefaultThumbnailHeight, got)
	}
}

func TestIsThumbnailSupportedOnlyIncludesImageAndVideoFiles(t *testing.T) {
	supported := []string{"photo.jpg", "clip.mp4", "movie.webm"}
	for _, filePath := range supported {
		if !IsThumbnailSupported(filePath) {
			t.Fatalf("expected %q to support thumbnails", filePath)
		}
	}

	unsupported := []string{"report.pdf", "song.mp3", "archive.zip"}
	for _, filePath := range unsupported {
		if IsThumbnailSupported(filePath) {
			t.Fatalf("expected %q to skip thumbnails", filePath)
		}
	}
}
