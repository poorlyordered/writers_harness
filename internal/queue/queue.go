// Package queue manages the Phase 3 Card Queue — the ordered list of cards
// that must be built in Phase 3, stored as JSON per SPEC-007 §10.
package queue

import (
	"encoding/json"
	"fmt"
	"time"
)

// CardType constants match SPEC-007 §10 card_type values.
const (
	CardTypeTrilogy    = "TRILOGY"
	CardTypeWorld      = "WORLD"
	CardTypeCharacter  = "CHARACTER"
	CardTypeFaction    = "FACTION"
	CardTypeThreat     = "THREAT"
	CardTypeNovel      = "NOVEL"
	CardTypeChapter    = "CHAPTER"
	CardTypeScene      = "SCENE"
	CardTypeConstraints = "CONSTRAINTS"
)

// CardStatus constants.
const (
	StatusNotStarted = "NOT_STARTED"
	StatusInProgress = "IN_PROGRESS"
	StatusComplete   = "COMPLETE"
)

// Tier constants for character/faction/threat cards.
const (
	TierFull   = "FULL"
	TierSketch = "SKETCH"
)

// PhaseGateStatus constants.
const (
	GateInProgress       = "IN_PROGRESS"
	GatePhase2Complete   = "PHASE_2_COMPLETE"
	GatePhase3Complete   = "PHASE_3_COMPLETE"
)

// QueueItem represents one card in the Card Queue.
type QueueItem struct {
	Priority        int      `json:"priority"`
	CardType        string   `json:"card_type"`
	CardSubtype     string   `json:"card_subtype"`
	Name            string   `json:"name"`
	Filename        string   `json:"filename"`
	Status          string   `json:"status"`
	Tier            *string  `json:"tier"` // "FULL", "SKETCH", or null
	Prerequisites   []int    `json:"prerequisites"`
	SourceSections  []string `json:"source_sections"`
	CompletedDate   *string  `json:"completed_date"`
	LockedFilename  *string  `json:"locked_filename"`
}

// OutstandingFlag tracks a CARD-REQUIRED flag.
type OutstandingFlag struct {
	Type      string         `json:"type"` // "CHARACTER" or "LOCATION"
	Name      string         `json:"name"`
	FirstUsed map[string]int `json:"first_used"` // {"chapter": n, "scene": n}
	Status    string         `json:"status"`      // "OUTSTANDING" or "RESOLVED"
}

// OutstandingFlags groups all outstanding flags.
type OutstandingFlags struct {
	NeedsReview  []string          `json:"needs_review"`
	CardRequired []OutstandingFlag `json:"card_required"`
}

// Progress tracks completion statistics.
type Progress struct {
	TotalCards      int     `json:"total_cards"`
	Complete        int     `json:"complete"`
	InProgress      int     `json:"in_progress"`
	NotStarted      int     `json:"not_started"`
	PercentComplete float64 `json:"percent_complete"`
}

// Queue is the full Card Queue document (SPEC-007 §10 schema).
type Queue struct {
	BookTitle        string           `json:"book_title"`
	SeriesTitle      string           `json:"series_title"`
	GeneratedDate    string           `json:"generated_date"`
	PhaseGateStatus  string           `json:"phase_gate_status"`
	Queue            []QueueItem      `json:"queue"`
	OutstandingFlags OutstandingFlags `json:"outstanding_flags"`
	Progress         Progress         `json:"progress"`
}

// New creates an empty Queue.
func New(seriesTitle, bookTitle string) *Queue {
	return &Queue{
		BookTitle:       bookTitle,
		SeriesTitle:     seriesTitle,
		GeneratedDate:   time.Now().Format("2006-01-02"),
		PhaseGateStatus: GateInProgress,
		OutstandingFlags: OutstandingFlags{
			NeedsReview:  []string{},
			CardRequired: []OutstandingFlag{},
		},
	}
}

// Marshal serialises the queue to JSON.
func (q *Queue) Marshal() ([]byte, error) {
	q.recalcProgress()
	return json.MarshalIndent(q, "", "  ")
}

// Unmarshal parses a JSON queue document.
func Unmarshal(data []byte) (*Queue, error) {
	var q Queue
	if err := json.Unmarshal(data, &q); err != nil {
		return nil, fmt.Errorf("parsing card queue: %w", err)
	}
	return &q, nil
}

// MarkComplete sets an item's status to COMPLETE and records the locked filename.
func (q *Queue) MarkComplete(priority int, lockedFilename string) error {
	for i := range q.Queue {
		if q.Queue[i].Priority == priority {
			q.Queue[i].Status = StatusComplete
			now := time.Now().Format("2006-01-02")
			q.Queue[i].CompletedDate = &now
			q.Queue[i].LockedFilename = &lockedFilename
			q.recalcProgress()
			return nil
		}
	}
	return fmt.Errorf("no queue item with priority %d", priority)
}

// Next returns the first NOT_STARTED item whose prerequisites are all COMPLETE.
func (q *Queue) Next() *QueueItem {
	completed := make(map[int]bool)
	for _, item := range q.Queue {
		if item.Status == StatusComplete {
			completed[item.Priority] = true
		}
	}
	for i := range q.Queue {
		if q.Queue[i].Status != StatusNotStarted {
			continue
		}
		ready := true
		for _, pre := range q.Queue[i].Prerequisites {
			if !completed[pre] {
				ready = false
				break
			}
		}
		if ready {
			return &q.Queue[i]
		}
	}
	return nil
}

func (q *Queue) recalcProgress() {
	total := len(q.Queue)
	complete, inProg, notStarted := 0, 0, 0
	for _, item := range q.Queue {
		switch item.Status {
		case StatusComplete:
			complete++
		case StatusInProgress:
			inProg++
		default:
			notStarted++
		}
	}
	pct := 0.0
	if total > 0 {
		pct = float64(complete) / float64(total) * 100
	}
	q.Progress = Progress{
		TotalCards:      total,
		Complete:        complete,
		InProgress:      inProg,
		NotStarted:      notStarted,
		PercentComplete: pct,
	}
}

// ptr helpers for optional string fields.
func strPtr(s string) *string { return &s }
func fullTier() *string       { return strPtr("FULL") }
func sketchTier() *string     { return strPtr("SKETCH") }
