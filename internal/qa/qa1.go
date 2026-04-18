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

// ExpansionDocument represents the parsed state of a locked Expansion document
// for the Phase 2→3 gate check.
type ExpansionDocument struct {
	ExistsInBox           bool
	Step1Locked           bool // Logline
	Step2Locked           bool // Story paragraph
	Step3Locked           bool // Character summaries
	Step4Locked           bool // Full synopsis
	Step5Locked           bool // Character arcs
	Step6Locked           bool // Act breakdown
	Step7Locked           bool // Character detail sheets
	StoryCircleMapPresent bool
	FortyChapterMapPresent bool
	SevenPlotsConfirmed   bool
	SubplotWeavePresent   bool
	FullTierCharsComplete bool
	SketchTierPresent     bool  // may be vacuously true if no sketch chars
	CardQueueGenerated    bool
	CardQueueHasMinimum   bool  // Trilogy (if series), World, Novel, Protagonist, Antagonist
	IsStandalone          bool
	NoNeedsReviewFlags    bool
	NoCardRequiredFlags   bool
	NotContradictseed     bool
}

// Phase2To3 runs the QA-1 Phase 2→3 gate checklist (SPEC-004C §4).
func Phase2To3(exp ExpansionDocument) *Result {
	r := &Result{ChecklistID: "1 (Phase 2→3)"}

	check := func(id, desc string, pass bool, reason string) {
		r.Items = append(r.Items, CheckItem{
			ID:          id,
			Description: desc,
			Passed:      pass,
			FailReason:  reason,
		})
	}

	check("2.1",  "Expansion document exists in Box at correct path", exp.ExistsInBox, "file not found in Box")
	check("2.2",  "Step 1 (Logline) present and locked", exp.Step1Locked, "")
	check("2.3",  "Step 2 (Story paragraph + Story Circle map) present and locked", exp.Step2Locked, "")
	check("2.4",  "Step 3 (Character summaries) present and locked", exp.Step3Locked, "")
	check("2.5",  "Step 4 (Full synopsis + 40-chapter beat map) present and locked", exp.Step4Locked, "")
	check("2.6",  "Step 5 (Character arc expansions) present and locked", exp.Step5Locked, "")
	check("2.7",  "Step 6 (Act breakdown + subplot weave) present and locked", exp.Step6Locked, "")
	check("2.8",  "Step 7 (Character detail sheets) present and locked", exp.Step7Locked, "")
	check("2.9",  "Story Circle map attached to Step 2", exp.StoryCircleMapPresent, "")
	check("2.10", "40-chapter beat map attached to Step 4", exp.FortyChapterMapPresent, "")
	check("2.11", "Seven Basic Plots pattern confirmed in Step 6", exp.SevenPlotsConfirmed, "")
	check("2.12", "Subplot weave map present in Step 6", exp.SubplotWeavePresent, "")
	check("2.13", "All FULL tier character detail sheets card-ready", exp.FullTierCharsComplete, "")
	check("2.14", "Sketch tier character sheets present", exp.SketchTierPresent, "")
	check("2.15", "Phase 3 Card Queue generated and saved to Box", exp.CardQueueGenerated, "")

	if exp.IsStandalone {
		check("2.16", "Card Queue minimum: World, Novel, Protagonist, Antagonist (standalone)", exp.CardQueueHasMinimum, "")
	} else {
		check("2.16", "Card Queue minimum: Trilogy, World, Novel, Protagonist, Antagonist (series)", exp.CardQueueHasMinimum, "")
	}

	check("2.17", "No NEEDS-REVIEW flags outstanding on Expansion document", exp.NoNeedsReviewFlags, "")
	check("2.18", "No CARD-REQUIRED flags outstanding", exp.NoCardRequiredFlags, "")
	check("2.19", "Expansion document does not contradict Seed Prompt", exp.NotContradictseed, "")

	return r
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "S"
}
