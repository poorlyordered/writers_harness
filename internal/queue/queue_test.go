package queue

import (
	"encoding/json"
	"testing"
)

func TestNew(t *testing.T) {
	q := New("Shadows Awaken", "Into the Dark")
	if q.SeriesTitle != "Shadows Awaken" {
		t.Errorf("SeriesTitle = %q, want %q", q.SeriesTitle, "Shadows Awaken")
	}
	if q.PhaseGateStatus != GateInProgress {
		t.Errorf("PhaseGateStatus = %q, want %q", q.PhaseGateStatus, GateInProgress)
	}
}

func TestMarkComplete(t *testing.T) {
	q := New("S", "B")
	q.Queue = append(q.Queue, QueueItem{Priority: 1, Status: StatusNotStarted})

	if err := q.MarkComplete(1, "file-v1.md"); err != nil {
		t.Fatalf("MarkComplete() error = %v", err)
	}
	if q.Queue[0].Status != StatusComplete {
		t.Errorf("status = %q, want %q", q.Queue[0].Status, StatusComplete)
	}
	if q.Queue[0].LockedFilename == nil || *q.Queue[0].LockedFilename != "file-v1.md" {
		t.Error("LockedFilename not set correctly")
	}

	if err := q.MarkComplete(99, ""); err == nil {
		t.Error("MarkComplete with missing priority should return error")
	}
}

func TestNext_RespectsPrerequisites(t *testing.T) {
	q := New("S", "B")
	q.Queue = []QueueItem{
		{Priority: 1, Status: StatusNotStarted, Prerequisites: []int{}},
		{Priority: 2, Status: StatusNotStarted, Prerequisites: []int{1}},
	}

	next := q.Next()
	if next == nil {
		t.Fatal("Next() returned nil, expected item 1")
	}
	if next.Priority != 1 {
		t.Errorf("Next() priority = %d, want 1", next.Priority)
	}

	// Complete item 1; now item 2 should be available.
	_ = q.MarkComplete(1, "v1.md")
	next = q.Next()
	if next == nil {
		t.Fatal("Next() returned nil after completing prerequisite")
	}
	if next.Priority != 2 {
		t.Errorf("Next() priority = %d, want 2", next.Priority)
	}
}

func TestNext_NoReadyItems(t *testing.T) {
	q := New("S", "B")
	q.Queue = []QueueItem{
		{Priority: 1, Status: StatusNotStarted, Prerequisites: []int{2}}, // circular (unresolvable)
		{Priority: 2, Status: StatusNotStarted, Prerequisites: []int{1}},
	}
	if q.Next() != nil {
		t.Error("Next() should return nil when no prerequisites are met")
	}
}

func TestMarshalUnmarshal(t *testing.T) {
	q := New("Shadows Awaken", "Into the Dark")
	q.Queue = append(q.Queue, QueueItem{
		Priority:    1,
		CardType:    CardTypeTrilogy,
		Name:        "Test Trilogy",
		Status:      StatusNotStarted,
		Prerequisites: []int{},
	})

	data, err := q.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	q2, err := Unmarshal(data)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if q2.SeriesTitle != q.SeriesTitle {
		t.Errorf("SeriesTitle mismatch after roundtrip")
	}
	if len(q2.Queue) != 1 {
		t.Errorf("Queue length = %d, want 1", len(q2.Queue))
	}

	// Confirm it's valid JSON.
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Errorf("output is not valid JSON: %v", err)
	}
}

func TestRecalcProgress(t *testing.T) {
	q := New("S", "B")
	q.Queue = []QueueItem{
		{Priority: 1, Status: StatusComplete},
		{Priority: 2, Status: StatusInProgress},
		{Priority: 3, Status: StatusNotStarted},
		{Priority: 4, Status: StatusNotStarted},
	}
	q.recalcProgress()

	p := q.Progress
	if p.TotalCards != 4 {
		t.Errorf("TotalCards = %d, want 4", p.TotalCards)
	}
	if p.Complete != 1 {
		t.Errorf("Complete = %d, want 1", p.Complete)
	}
	if p.InProgress != 1 {
		t.Errorf("InProgress = %d, want 1", p.InProgress)
	}
	if p.NotStarted != 2 {
		t.Errorf("NotStarted = %d, want 2", p.NotStarted)
	}
	if p.PercentComplete != 25.0 {
		t.Errorf("PercentComplete = %.1f, want 25.0", p.PercentComplete)
	}
}
