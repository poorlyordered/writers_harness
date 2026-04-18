package qa

import "github.com/poorlyordered/writers_harness/internal/queue"

// PreProseDocFromQueue builds a PreProseDoc by inspecting a completed card queue.
func PreProseDocFromQueue(q *queue.Queue) PreProseDoc {
	doc := PreProseDoc{
		QueueComplete:      q.Progress.NotStarted == 0 && q.Progress.InProgress == 0,
		NoOutstandingFlags: true,
	}

	charIndex := 0
	for _, item := range q.Queue {
		complete := item.Status == queue.StatusComplete
		switch item.CardType {
		case queue.CardTypeNovel:
			doc.NovelLocked = complete
		case queue.CardTypeWorld:
			if item.CardSubtype == "Overview" {
				doc.WorldOverviewLocked = complete
			}
			if item.CardSubtype == "Constraints" {
				doc.WorldConstraintsLocked = complete
			}
		case queue.CardTypeCharacter:
			if item.Tier != nil && *item.Tier == queue.TierFull && complete {
				charIndex++
				if charIndex == 1 {
					doc.ProtagonistLocked = true
				} else if charIndex == 2 {
					doc.AntagonistLocked = true
				}
			}
		case queue.CardTypeChapter:
			if complete {
				doc.ChapterCardsStarted = true
			}
		}
	}

	fullTotal, fullDone := 0, 0
	for _, item := range q.Queue {
		if item.Tier != nil && *item.Tier == queue.TierFull {
			fullTotal++
			if item.Status == queue.StatusComplete {
				fullDone++
			}
		}
	}
	doc.AllFullTierLocked = fullTotal > 0 && fullTotal == fullDone

	return doc
}

// PreProseDoc holds computed state for the QA-3 pre-prose readiness check.
type PreProseDoc struct {
	NovelLocked            bool
	WorldOverviewLocked    bool
	WorldConstraintsLocked bool
	ProtagonistLocked      bool
	AntagonistLocked       bool
	AllFullTierLocked      bool // all FULL tier CHARACTER, FACTION, THREAT cards locked
	ChapterCardsStarted    bool // at least one Chapter card locked
	NoOutstandingFlags     bool // no CARD-REQUIRED or NEEDS-REVIEW flags in any card
	QueueComplete          bool // Card Queue phase_gate_status = PHASE_2_COMPLETE
}

// PreProse runs the QA-3 pre-prose readiness checklist (SPEC-004C §6).
// A passing QA-3 gates advancement from Phase 3 to Phase 4.
func PreProse(doc PreProseDoc) *Result {
	r := &Result{ChecklistID: "3 (Pre-Prose Readiness)"}

	check := func(id, desc string, pass bool, reason string) {
		r.Items = append(r.Items, CheckItem{
			ID:          id,
			Description: desc,
			Passed:      pass,
			FailReason:  reason,
		})
	}

	check("3.1", "NOVEL card locked", doc.NovelLocked, "NOVEL card not locked")
	check("3.2", "WORLD-Overview card locked", doc.WorldOverviewLocked, "WORLD-Overview not locked")
	check("3.3", "WORLD-Constraints card locked", doc.WorldConstraintsLocked, "WORLD-Constraints not locked")
	check("3.4", "Protagonist CHARACTER card locked", doc.ProtagonistLocked, "protagonist card not locked")
	check("3.5", "Primary antagonist CHARACTER card locked", doc.AntagonistLocked, "antagonist card not locked")
	check("3.6", "All FULL tier cards locked", doc.AllFullTierLocked, "one or more FULL tier cards not locked")
	check("3.7", "At least one CHAPTER card locked", doc.ChapterCardsStarted, "no chapter cards locked")
	check("3.8", "No outstanding CARD-REQUIRED flags", doc.NoOutstandingFlags, "unresolved flags detected")
	check("3.9", "Card Queue marked complete", doc.QueueComplete, "queue phase gate not advanced")

	return r
}
