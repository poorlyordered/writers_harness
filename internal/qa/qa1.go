// Package qa implements the four QA checklists defined in SPEC-004C.
// qa1.go covers the Phase Gate checklists (QA-1).
package qa

import (
	"fmt"
	"strings"
)

// CheckItem is a single QA checklist item.
type CheckItem struct {
	ID          string
	Description string
	Passed      bool
	FailReason  string
}

// Result holds the outcome of a QA check run.
type Result struct {
	ChecklistID string
	Items       []CheckItem
}

// Passed returns true only when every item passed.
func (r *Result) Passed() bool {
	for _, item := range r.Items {
		if !item.Passed {
			return false
		}
	}
	return true
}

// FailedItems returns only the items that failed.
func (r *Result) FailedItems() []CheckItem {
	var failed []CheckItem
	for _, item := range r.Items {
		if !item.Passed {
			failed = append(failed, item)
		}
	}
	return failed
}

// FormatFailures returns a human-readable failure report.
func (r *Result) FormatFailures() string {
	failed := r.FailedItems()
	if len(failed) == 0 {
		return fmt.Sprintf("QA-%s: ALL PASS ✓", r.ChecklistID)
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("QA-%s: %d FAIL%s\n\n", r.ChecklistID, len(failed), plural(len(failed))))
	for _, item := range failed {
		sb.WriteString(fmt.Sprintf("  FAIL [%s] %s\n", item.ID, item.Description))
		if item.FailReason != "" {
			sb.WriteString(fmt.Sprintf("       Reason: %s\n", item.FailReason))
		}
	}
	return sb.String()
}

// SeedDocument represents the parsed state of a Locked Seed Prompt document.
// Fields are populated by the Phase 1 runner before QA-1 is invoked.
type SeedDocument struct {
	ExistsInBox    bool
	Status         string // "DRAFT" | "LOCKED"
	HasLayer1      bool   // World layer
	HasLayer2      bool   // Story DNA layer
	HasLayer3      bool   // Spark layer
	ProtagonistWant      bool
	ProtagonistNeed      bool
	ProtagonistFear      bool
	ProtagonistMisbelief bool
	AntagonistBelief     bool
	AntagonistMirror     bool
	CentralQuestion      bool
	ThemeHint            bool
	SeriesQuestions      bool // may be skipped for standalone
	AntagonistLadder     bool // may be skipped for standalone
	IsStandalone         bool
	VersionRecorded      bool
	NoNeedsReviewFlags   bool
	NoCardRequiredFlags  bool
}

// Phase1To2 runs the QA-1 Phase 1→2 gate checklist (SPEC-004C §4).
func Phase1To2(seed SeedDocument) *Result {
	r := &Result{ChecklistID: "1 (Phase 1→2)"}

	check := func(id, desc string, pass bool, reason string) {
		r.Items = append(r.Items, CheckItem{
			ID:          id,
			Description: desc,
			Passed:      pass,
			FailReason:  reason,
		})
	}

	check("1.1",  "Seed Prompt document exists in Box at correct path", seed.ExistsInBox, "file not found in Box")
	check("1.2",  "Seed status field = LOCKED", seed.Status == "LOCKED", fmt.Sprintf("status is %q", seed.Status))
	check("1.3",  "Layer 1 (World) fully populated", seed.HasLayer1, "")
	check("1.4",  "Layer 2 (Story DNA) fully populated", seed.HasLayer2, "")
	check("1.5",  "Layer 3 (Spark) fully populated", seed.HasLayer3, "")
	check("1.6",  "Protagonist want defined", seed.ProtagonistWant, "")
	check("1.7",  "Protagonist need defined", seed.ProtagonistNeed, "")
	check("1.8",  "Protagonist fear defined", seed.ProtagonistFear, "")
	check("1.9",  "Protagonist misbelief defined", seed.ProtagonistMisbelief, "")
	check("1.10", "Primary antagonist belief defined", seed.AntagonistBelief, "")
	check("1.11", "Primary antagonist mirror relationship defined", seed.AntagonistMirror, "")
	check("1.12", "Central thematic question defined", seed.CentralQuestion, "")
	check("1.13", "Theme hint defined", seed.ThemeHint, "")

	// Series-only items.
	if seed.IsStandalone {
		check("1.14", "Series questions defined (skipped — standalone)", true, "")
		check("1.15", "Antagonist ladder sketched (skipped — standalone)", true, "")
	} else {
		check("1.14", "Series questions defined", seed.SeriesQuestions, "required for series")
		check("1.15", "Antagonist ladder sketched", seed.AntagonistLadder, "required for series")
	}

	check("1.16", "Seed version number recorded", seed.VersionRecorded, "")
	check("1.17", "No NEEDS-REVIEW flags outstanding on Seed document", seed.NoNeedsReviewFlags, "")
	check("1.18", "No CARD-REQUIRED flags outstanding", seed.NoCardRequiredFlags, "")

	return r
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "S"
}
