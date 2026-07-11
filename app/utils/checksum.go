package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// CalculateFileChecksum calculates SHA-256 checksum of a file
func CalculateFileChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file for checksum calculation: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to calculate checksum: %w", err)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// VerifyFileChecksum verifies if a file's checksum matches the expected value
func VerifyFileChecksum(filePath, expectedChecksum string) (bool, error) {
	actualChecksum, err := CalculateFileChecksum(filePath)
	if err != nil {
		return false, err
	}

	return actualChecksum == expectedChecksum, nil
}