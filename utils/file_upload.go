package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// AllowedImageTypes defines the allowed image MIME types
var AllowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/jpg":  true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// MaxFileSize defines the maximum file size (5MB)
const MaxFileSize = 5 * 1024 * 1024

// UploadFile saves an uploaded file to the specified directory
func UploadFile(file *multipart.FileHeader, directory string) (string, error) {
	// Validate file size
	if file.Size > MaxFileSize {
		return "", fmt.Errorf("file size exceeds maximum allowed size of 5MB")
	}

	// Validate file type
	contentType := file.Header.Get("Content-Type")
	if !AllowedImageTypes[contentType] {
		return "", fmt.Errorf("invalid file type. Allowed types: jpeg, jpg, png, gif, webp")
	}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Create directory if it doesn't exist
	if err := os.MkdirAll(directory, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	uniqueFilename := fmt.Sprintf("%d_%s%s", time.Now().Unix(), uuid.New().String(), ext)
	filePath := filepath.Join(directory, uniqueFilename)

	// Create the destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy the uploaded file to the destination
	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	return filePath, nil
}

// DeleteFile deletes a file from the filesystem
func DeleteFile(filePath string) error {
	if filePath == "" {
		return nil
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to delete
	}

	// Delete the file
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// GetFileExtension returns the file extension from a filename
func GetFileExtension(filename string) string {
	return strings.ToLower(filepath.Ext(filename))
}

// IsValidImageExtension checks if the file extension is valid for images
func IsValidImageExtension(filename string) bool {
	ext := GetFileExtension(filename)
	validExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}
	return validExtensions[ext]
}
