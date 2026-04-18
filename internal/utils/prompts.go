package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Loader reads prompt templates from the prompts/ directory.
type Loader struct {
	promptsDir string
}

// NewLoader creates a Loader rooted at the given directory.
func NewLoader(promptsDir string) *Loader {
	return &Loader{promptsDir: promptsDir}
}

// Load reads a prompt file by name (e.g. "system/phase1" → prompts/system/phase1.txt).
func (l *Loader) Load(name string) (string, error) {
	path := filepath.Join(l.promptsDir, name+".txt")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("loading prompt %q: %w", name, err)
	}
	return string(data), nil
}

// LoadWithVars reads a prompt file and replaces {{KEY}} placeholders with the
// provided values.
func (l *Loader) LoadWithVars(name string, vars map[string]string) (string, error) {
	tmpl, err := l.Load(name)
	if err != nil {
		return "", err
	}
	for k, v := range vars {
		tmpl = strings.ReplaceAll(tmpl, k, v)
	}
	return tmpl, nil
}
