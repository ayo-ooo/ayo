package ayod

import (
	"fmt"
	"os"
)

// FileHandler handles file read/write operations.
type FileHandler struct{}

// NewFileHandler creates a new file handler.
func NewFileHandler() *FileHandler {
	return &FileHandler{}
}

// ReadFile reads the contents of a file.
func (h *FileHandler) ReadFile(req ReadFileRequest) (*ReadFileResponse, error) {
	info, err := os.Stat(req.Path)
	if err != nil {
		return nil, fmt.Errorf("stat %q: %w", req.Path, err)
	}

	if info.IsDir() {
		return nil, fmt.Errorf("%q is a directory", req.Path)
	}

	content, err := os.ReadFile(req.Path)
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", req.Path, err)
	}

	return &ReadFileResponse{
		Content: content,
		Mode:    info.Mode(),
	}, nil
}

// WriteFile writes content to a file.
func (h *FileHandler) WriteFile(req WriteFileRequest) error {
	// Ensure parent directory exists
	if err := os.MkdirAll(parentDir(req.Path), 0755); err != nil {
		return fmt.Errorf("create parent dir: %w", err)
	}

	mode := req.Mode
	if mode == 0 {
		mode = 0644
	}

	if err := os.WriteFile(req.Path, req.Content, mode); err != nil {
		return fmt.Errorf("write %q: %w", req.Path, err)
	}

	return nil
}

// parentDir returns the parent directory of a path.
func parentDir(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			if i == 0 {
				return "/"
			}
			return path[:i]
		}
	}
	return "."
}
