// Package storage manages the local file system fallback that mirrors the
// Box folder structure (SPEC-007 §12).
package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Local provides read/write access to the local sync folder.
type Local struct {
	root string
}

// NewLocal creates a Local storage backed by the given root directory.
func NewLocal(syncFolder string) *Local {
	return &Local{root: syncFolder}
}

// EnsureDir creates all directories in the path relative to the sync root.
func (l *Local) EnsureDir(boxPath string) error {
	full := l.fullPath(boxPath)
	return os.MkdirAll(full, 0o755)
}

// Write writes content to a file at the given Box-relative path.
func (l *Local) Write(boxPath, filename, content string) error {
	dir := l.fullPath(boxPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}
	full := filepath.Join(dir, filename)
	return os.WriteFile(full, []byte(content), 0o644)
}

// Read returns the content of a file at the given Box-relative path.
func (l *Local) Read(boxPath, filename string) (string, error) {
	full := filepath.Join(l.fullPath(boxPath), filename)
	data, err := os.ReadFile(full)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", full, err)
	}
	return string(data), nil
}

// Exists reports whether a file exists at the given Box-relative path.
func (l *Local) Exists(boxPath, filename string) bool {
	full := filepath.Join(l.fullPath(boxPath), filename)
	_, err := os.Stat(full)
	return err == nil
}

// ListFiles returns all filenames in the Box-relative directory that
// start with the given prefix.
func (l *Local) ListFiles(boxPath, prefix string) ([]string, error) {
	dir := l.fullPath(boxPath)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), prefix) {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

// AbsPath returns the absolute filesystem path for the given Box-relative folder and filename.
func (l *Local) AbsPath(boxFolder, filename string) string {
	return filepath.Join(l.fullPath(boxFolder), filename)
}

func (l *Local) fullPath(boxPath string) string {
	// Treat boxPath as relative to the sync root; strip leading slash if present.
	cleaned := strings.TrimPrefix(boxPath, "/")
	return filepath.Join(l.root, filepath.FromSlash(cleaned))
}
