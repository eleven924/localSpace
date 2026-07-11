package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateFileChecksum(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_file.txt")

	// Write known content
	testContent := []byte("Hello, World!")
	err := os.WriteFile(testFile, testContent, 0644)
	require.NoError(t, err)

	// Calculate checksum
	checksum, err := CalculateFileChecksum(testFile)
	require.NoError(t, err)

	// Verify checksum is not empty and has correct length
	assert.NotEmpty(t, checksum)
	assert.Equal(t, 64, len(checksum), "SHA-256 checksum should be 64 characters (hex)")

	// Calculate again to verify consistency
	checksum2, err := CalculateFileChecksum(testFile)
	require.NoError(t, err)
	assert.Equal(t, checksum, checksum2, "Same file should produce same checksum")
}

func TestCalculateFileChecksum_DifferentFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two different files
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")

	err := os.WriteFile(file1, []byte("Content 1"), 0644)
	require.NoError(t, err)

	err = os.WriteFile(file2, []byte("Content 2"), 0644)
	require.NoError(t, err)

	// Calculate checksums
	checksum1, err := CalculateFileChecksum(file1)
	require.NoError(t, err)

	checksum2, err := CalculateFileChecksum(file2)
	require.NoError(t, err)

	// Verify different files produce different checksums
	assert.NotEqual(t, checksum1, checksum2, "Different files should produce different checksums")
}

func TestCalculateFileChecksum_NonExistentFile(t *testing.T) {
	nonExistentFile := "/non/existent/file.txt"

	_, err := CalculateFileChecksum(nonExistentFile)
	assert.Error(t, err, "Should return error for non-existent file")
}

func TestVerifyFileChecksum(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_file.txt")

	// Write test content
	testContent := []byte("Verification test")
	err := os.WriteFile(testFile, testContent, 0644)
	require.NoError(t, err)

	// Calculate checksum
	checksum, err := CalculateFileChecksum(testFile)
	require.NoError(t, err)

	// Verify correct checksum
	isValid, err := VerifyFileChecksum(testFile, checksum)
	require.NoError(t, err)
	assert.True(t, isValid, "Should verify correct checksum")

	// Verify incorrect checksum
	isValid, err = VerifyFileChecksum(testFile, "wrong_checksum")
	require.NoError(t, err)
	assert.False(t, isValid, "Should reject incorrect checksum")
}

func TestCalculateFileChecksum_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	emptyFile := filepath.Join(tmpDir, "empty_file.txt")

	// Create empty file
	err := os.WriteFile(emptyFile, []byte{}, 0644)
	require.NoError(t, err)

	// Calculate checksum
	checksum, err := CalculateFileChecksum(emptyFile)
	require.NoError(t, err)

	// Empty file should have a specific checksum
	// SHA-256 of empty string is known value
	expectedEmptyChecksum := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	assert.Equal(t, expectedEmptyChecksum, checksum, "Empty file should have known checksum")
}

func TestCalculateFileChecksum_LargeFile(t *testing.T) {
	tmpDir := t.TempDir()
	largeFile := filepath.Join(tmpDir, "large_file.txt")

	// Create a larger file (1MB)
	largeContent := make([]byte, 1*1024*1024) // 1MB
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}

	err := os.WriteFile(largeFile, largeContent, 0644)
	require.NoError(t, err)

	// Calculate checksum
	checksum, err := CalculateFileChecksum(largeFile)
	require.NoError(t, err)

	// Verify checksum consistency
	checksum2, err := CalculateFileChecksum(largeFile)
	require.NoError(t, err)
	assert.Equal(t, checksum, checksum2, "Large file checksum should be consistent")
}