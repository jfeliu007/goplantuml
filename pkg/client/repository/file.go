package repository

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// FileRepository handles file system operations
type FileRepository interface {
	ValidateDirectory(path string) error
	ListGoFiles(directory string, recursive bool) ([]string, error)
	WriteFile(path, content string) error
	ReadFile(path string) (string, error)
}

// fileRepository implements FileRepository
type fileRepository struct{}

// NewFileRepository creates a new FileRepository
func NewFileRepository() FileRepository {
	return &fileRepository{}
}

// ValidateDirectory validates that a directory exists and is accessible
func (r *fileRepository) ValidateDirectory(path string) error {
	if fi, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", path)
	} else if err != nil {
		return fmt.Errorf("cannot access directory %s: %w", path, err)
	} else if !fi.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}
	return nil
}

// ListGoFiles lists all .go files in a directory
func (r *fileRepository) ListGoFiles(directory string, recursive bool) ([]string, error) {
	var goFiles []string

	if recursive {
		err := filepath.WalkDir(directory, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && strings.HasSuffix(path, ".go") {
				goFiles = append(goFiles, path)
			}
			return nil
		})
		return goFiles, err
	} else {
		entries, err := os.ReadDir(directory)
		if err != nil {
			return nil, err
		}

		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
				goFiles = append(goFiles, filepath.Join(directory, entry.Name()))
			}
		}
		return goFiles, nil
	}
}

// WriteFile writes content to a file
func (r *fileRepository) WriteFile(path, content string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", path, err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %w", path, err)
	}

	return nil
}

// ReadFile reads content from a file
func (r *fileRepository) ReadFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", path, err)
	}
	return string(content), nil
}
