// Package session manages harness session state — what phase is active,
// which series/book is loaded, and writing state to Box at session close.
package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/poorlyordered/writers_harness/internal/utils"
)

// Phase constants.
const (
	Phase1 = "phase1"
	Phase2 = "phase2"
	Phase3 = "phase3"
	Phase4 = "phase4"
)

// State holds all runtime information about the current session.
type State struct {
	Phase       string    `json:"phase"`
	SeriesTitle string    `json:"series_title"`
	BookTitle   string    `json:"book_title"`
	BookNum     int       `json:"book_num"`
	StartedAt   time.Time `json:"started_at"`
	LastSaved   time.Time `json:"last_saved"`

	// Phase 4 progress
	CurrentChapter int `json:"current_chapter,omitempty"`
	CurrentScene   int `json:"current_scene,omitempty"`
}

// Manager persists session state to the local config folder.
type Manager struct {
	configDir string
}

// NewManager creates a session manager backed by the given config directory.
func NewManager(configDir string) *Manager {
	return &Manager{configDir: configDir}
}

// Save writes the session state to disk.
func (m *Manager) Save(s *State) error {
	if err := os.MkdirAll(m.configDir, 0o755); err != nil {
		return err
	}
	s.LastSaved = time.Now().UTC()
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(m.configDir, "session.json"), data, 0o644)
}

// Load reads the most recent session state from disk.
func (m *Manager) Load() (*State, error) {
	path := filepath.Join(m.configDir, "session.json")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("no saved session found — use 'harness new-series' to start")
	}
	if err != nil {
		return nil, fmt.Errorf("reading session: %w", err)
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing session: %w", err)
	}
	return &s, nil
}

// New creates a fresh session state.
func New(phase, seriesTitle, bookTitle string, bookNum int) *State {
	return &State{
		Phase:       phase,
		SeriesTitle: seriesTitle,
		BookTitle:   bookTitle,
		BookNum:     bookNum,
		StartedAt:   time.Now().UTC(),
	}
}

// BoxSeedsPath returns the Box-relative folder path for seeds in this session.
func (s *State) BoxSeedsPath() string {
	return utils.SeedsFolder(s.SeriesTitle, s.BookTitle)
}

// BoxCardsPath returns the Box-relative folder path for cards.
func (s *State) BoxCardsPath() string {
	return utils.CardsFolder(s.SeriesTitle, s.BookTitle)
}

// BoxQAPath returns the Box-relative folder path for QA files.
func (s *State) BoxQAPath() string {
	return utils.QAFolder(s.SeriesTitle, s.BookTitle)
}
