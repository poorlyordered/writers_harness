// Package prose handles Phase 4 prose generation: scene drafting, assembly, and pattern tracking.
package prose

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// RejectionEntry records one writer rejection of a prose draft.
type RejectionEntry struct {
	Timestamp  string `json:"timestamp"`
	ChapterNum int    `json:"chapter"`
	SceneNum   int    `json:"scene"`
	Reason     string `json:"reason"`
}

// PatternLog tracks writer rejections to surface recurring patterns.
type PatternLog struct {
	path    string
	Entries []RejectionEntry
}

// NewPatternLog loads or creates a pattern log at the given path.
func NewPatternLog(path string) (*PatternLog, error) {
	pl := &PatternLog{path: path}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return pl, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading pattern log: %w", err)
	}
	return pl, json.Unmarshal(data, &pl.Entries)
}

// AddRejection records a new rejection and saves the log.
func (pl *PatternLog) AddRejection(chapterNum, sceneNum int, reason string) error {
	pl.Entries = append(pl.Entries, RejectionEntry{
		Timestamp:  time.Now().Format(time.RFC3339),
		ChapterNum: chapterNum,
		SceneNum:   sceneNum,
		Reason:     reason,
	})
	return pl.save()
}

// RecentRejections returns the last N rejection reasons as a formatted string.
func (pl *PatternLog) RecentRejections(n int) string {
	start := len(pl.Entries) - n
	if start < 0 {
		start = 0
	}
	var sb strings.Builder
	for i, e := range pl.Entries[start:] {
		sb.WriteString(fmt.Sprintf("  %d. Ch%02d-S%d: %s\n", i+1, e.ChapterNum, e.SceneNum, e.Reason))
	}
	return sb.String()
}

// ThreeRejectionPattern returns a diagnostic summary when 3 consecutive rejections occur.
func (pl *PatternLog) ThreeRejectionPattern() string {
	if len(pl.Entries) < 3 {
		return ""
	}
	last3 := pl.Entries[len(pl.Entries)-3:]
	var reasons []string
	for _, e := range last3 {
		reasons = append(reasons, e.Reason)
	}
	return fmt.Sprintf(
		"THREE-REJECTION DIAGNOSTIC\n"+
			"Three consecutive drafts rejected. Common thread in rejections:\n%s\n"+
			"Suggestion: adjust the scene orientation or approach before retrying.",
		strings.Join(reasons, "\n"))
}

func (pl *PatternLog) save() error {
	data, err := json.MarshalIndent(pl.Entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(pl.path, data, 0o644)
}
