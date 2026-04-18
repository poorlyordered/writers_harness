package qa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// LogEntry records the outcome of a single QA checklist run.
type LogEntry struct {
	Timestamp string   `json:"timestamp"`
	Checklist string   `json:"checklist"` // "QA-1", "QA-2", etc.
	Subject   string   `json:"subject"`   // card name or phase name
	Passed    bool     `json:"passed"`
	FailCount int      `json:"fail_count"`
	Overrides []string `json:"overrides,omitempty"` // item IDs overridden by writer
}

// Log is an append-only QA log backed by a JSON-lines file.
// Entries are cached in memory after the first read.
type Log struct {
	path    string
	entries []LogEntry // in-memory cache; nil until first read
}

// NewLog creates a log backed by the given file path.
func NewLog(path string) *Log {
	return &Log{path: path}
}

// Append adds a new entry to the log file and updates the in-memory cache.
func (l *Log) Append(entry LogEntry) error {
	entry.Timestamp = time.Now().Format(time.RFC3339)
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("opening qa log: %w", err)
	}
	defer f.Close()
	if _, err = f.Write(append(data, '\n')); err != nil {
		return err
	}
	l.entries = append(l.entries, entry)
	return nil
}

// AppendResult is a convenience wrapper that builds a LogEntry from a Result.
func (l *Log) AppendResult(r *Result, subject string, overrides []string) error {
	checklistNum := "QA"
	if len(r.ChecklistID) > 0 {
		parts := strings.Fields(r.ChecklistID)
		if len(parts) > 0 {
			checklistNum = "QA-" + parts[0]
		}
	}
	return l.Append(LogEntry{
		Checklist: checklistNum,
		Subject:   subject,
		Passed:    r.Passed(),
		FailCount: len(r.FailedItems()),
		Overrides: overrides,
	})
}

// Summary returns pass/fail counts from the log.
// On the first call the file is read and cached; subsequent calls use the cache.
func (l *Log) Summary() (pass, fail int, err error) {
	if l.entries == nil {
		if loadErr := l.load(); loadErr != nil {
			return 0, 0, loadErr
		}
	}
	for _, e := range l.entries {
		if e.Passed {
			pass++
		} else {
			fail++
		}
	}
	return pass, fail, nil
}

func (l *Log) load() error {
	data, err := os.ReadFile(l.path)
	if os.IsNotExist(err) {
		l.entries = []LogEntry{}
		return nil
	}
	if err != nil {
		return err
	}
	l.entries = []LogEntry{}
	for _, line := range bytes.Split(data, []byte{'\n'}) {
		if len(line) == 0 {
			continue
		}
		var entry LogEntry
		if json.Unmarshal(line, &entry) != nil {
			continue
		}
		l.entries = append(l.entries, entry)
	}
	return nil
}
