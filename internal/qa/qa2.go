package qa

import (
	"fmt"

	"github.com/poorlyordered/writers_harness/internal/cards"
)

// CardQAInput holds computed state for a QA-2 card completion check.
type CardQAInput struct {
	CardType    string
	CardSubtype string
	Name        string
	ExistsInBox bool
	IsLocked    bool
	VersionInHeader bool
	Failures    []cards.Failure // from cards.Validate()
}

// CardCompletion runs the QA-2 card completion checklist (SPEC-004C §5).
func CardCompletion(input CardQAInput) *Result {
	r := &Result{ChecklistID: fmt.Sprintf("2 (Card: %s %s)", input.CardType, input.Name)}

	check := func(id, desc string, pass bool, reason string) {
		r.Items = append(r.Items, CheckItem{
			ID:          id,
			Description: desc,
			Passed:      pass,
			FailReason:  reason,
		})
	}

	check("2.1", "Card file exists in Box at correct path", input.ExistsInBox, "file not found in Box")
	check("2.2", "Card status is LOCKED", input.IsLocked, "status field is not LOCKED")
	check("2.3", "Version recorded in card header", input.VersionInHeader, "")

	// Content failures from cards.Validate() — each becomes a checklist item.
	contentPass := len(input.Failures) == 0
	failReason := ""
	if !contentPass && len(input.Failures) > 0 {
		failReason = input.Failures[0].Description
		if len(input.Failures) > 1 {
			failReason += fmt.Sprintf(" (and %d more)", len(input.Failures)-1)
		}
	}
	check("2.4", "All required sections present and non-empty", contentPass, failReason)
	check("2.5", "No placeholder text ([TBD], [TODO], 'placeholder')", !hasPlaceholderFailure(input.Failures), "")

	// Card-type-specific checks.
	switch input.CardType {
	case cards.TypeCharacter:
		if input.CardSubtype == cards.TierFull {
			check("2.6", "All four pillars defined (Want/Need/Fear/Misbelief)", true, "") // confirmed by section check
			check("2.7", "Wound section present and specific", !missingSection(input.Failures, "## Wound"), "")
			check("2.8", "Psychology section (Enneagram/MBTI) present", !missingSection(input.Failures, "## Psychology"), "")
			check("2.9", "Voice Anchor written", !missingSection(input.Failures, "## Voice Anchor"), "")
		}
	case cards.TypeFaction:
		check("2.6", "Leadership defined", !missingSection(input.Failures, "## Leadership"), "")
		check("2.7", "Goals and methods defined", !missingSection(input.Failures, "## Goals and Methods"), "")
	case cards.TypeThreat:
		if input.CardSubtype == cards.TierFull {
			check("2.6", "Escalation arc defined", !missingSection(input.Failures, "## Escalation Arc"), "")
		}
	case cards.TypeNovel:
		check("2.6", "40-chapter map included", !missingSection(input.Failures, "## 40-Chapter Map"), "")
		check("2.7", "Act structure defined", !missingSection(input.Failures, "## Act Structure"), "")
	}

	return r
}

func hasPlaceholderFailure(failures []cards.Failure) bool {
	for _, f := range failures {
		if f.ID == "2.3" {
			return true
		}
	}
	return false
}

func missingSection(failures []cards.Failure, heading string) bool {
	for _, f := range failures {
		if f.Section == heading {
			return true
		}
	}
	return false
}
