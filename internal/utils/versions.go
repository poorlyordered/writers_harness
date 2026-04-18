package utils

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

// ParseVersion extracts the version integer from a filename ending in -v{n}.md.
// Returns 0 and an error if the pattern is not found.
func ParseVersion(filename string) (int, error) {
	base := strings.TrimSuffix(filepath.Base(filename), ".md")
	idx := strings.LastIndex(base, "-v")
	if idx < 0 {
		return 0, fmt.Errorf("no version suffix found in %q", filename)
	}
	n, err := strconv.Atoi(base[idx+2:])
	if err != nil {
		return 0, fmt.Errorf("non-integer version in %q", filename)
	}
	return n, nil
}

// IncrementVersion returns a new filename with the version number incremented by 1.
func IncrementVersion(filename string) (string, error) {
	v, err := ParseVersion(filename)
	if err != nil {
		return "", err
	}
	base := strings.TrimSuffix(filepath.Base(filename), ".md")
	idx := strings.LastIndex(base, "-v")
	prefix := base[:idx]
	return fmt.Sprintf("%s-v%d.md", prefix, v+1), nil
}

// HighestVersion returns the filename with the highest version number from
// a list of candidate filenames sharing the same prefix.
func HighestVersion(filenames []string) (string, error) {
	if len(filenames) == 0 {
		return "", fmt.Errorf("no filenames provided")
	}
	best := ""
	bestV := -1
	for _, f := range filenames {
		v, err := ParseVersion(f)
		if err != nil {
			continue
		}
		if v > bestV {
			bestV = v
			best = f
		}
	}
	if best == "" {
		return "", fmt.Errorf("no versioned files found")
	}
	return best, nil
}
