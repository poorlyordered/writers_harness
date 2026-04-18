package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SyncPendingEntry records a local file write that needs to sync to Box.
type SyncPendingEntry struct {
	BoxPath   string    `json:"box_path"`
	Filename  string    `json:"filename"`
	WrittenAt time.Time `json:"written_at"`
}

// SyncPending tracks files written locally while Box was unavailable.
type SyncPending struct {
	Entries []SyncPendingEntry `json:"entries"`
	logPath string
}

// LoadSyncPending reads (or initialises) the SYNC-PENDING log at the given path.
func LoadSyncPending(logDir string) (*SyncPending, error) {
	path := filepath.Join(logDir, "SYNC-PENDING.json")
	sp := &SyncPending{logPath: path}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return sp, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading sync-pending: %w", err)
	}
	if err := json.Unmarshal(data, sp); err != nil {
		return nil, fmt.Errorf("parsing sync-pending: %w", err)
	}
	return sp, nil
}

// Add records a new pending sync entry and persists the log.
func (sp *SyncPending) Add(boxPath, filename string) error {
	sp.Entries = append(sp.Entries, SyncPendingEntry{
		BoxPath:   boxPath,
		Filename:  filename,
		WrittenAt: time.Now().UTC(),
	})
	return sp.save()
}

// Clear removes all entries after a successful sync.
func (sp *SyncPending) Clear() error {
	sp.Entries = nil
	return sp.save()
}

func (sp *SyncPending) save() error {
	if err := os.MkdirAll(filepath.Dir(sp.logPath), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(sp, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(sp.logPath, data, 0o644)
}
