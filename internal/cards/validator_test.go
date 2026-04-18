package cards

import (
	"strings"
	"testing"
)

func TestValidate_NoFailuresForCompleteCard(t *testing.T) {
	content := `## Setting
A frontier planet on the edge of the Reach, dust-choked and lawless.

## Tone and Genre
Sci-Fi Western with noir undertones. Morally grey.

## Core Concept
Power without accountability always corrupts.

## Thematic Backdrop
The collapse of the old order and what rises in its place.

## Reader Experience
Tense, propulsive, with moments of brutal stillness.`

	failures := Validate(TypeWorld, "Overview", content)
	if len(failures) != 0 {
		t.Errorf("expected 0 failures, got %d: %+v", len(failures), failures)
	}
}

func TestValidate_DetectsPlaceholder(t *testing.T) {
	content := `## Setting
[TBD]

## Tone and Genre
Sci-Fi Western.

## Core Concept
Power without accountability.

## Thematic Backdrop
Old order collapse.

## Reader Experience
Tense.`

	failures := Validate(TypeWorld, "Overview", content)
	hasPlaceholder := false
	for _, f := range failures {
		if f.ID == "2.3" {
			hasPlaceholder = true
		}
	}
	if !hasPlaceholder {
		t.Error("expected placeholder failure (2.3), got none")
	}
}

func TestValidate_DetectsMissingSection(t *testing.T) {
	// Omit "## Core Concept"
	content := `## Setting
A frontier planet.

## Tone and Genre
Sci-Fi Western.

## Thematic Backdrop
Old order collapse.

## Reader Experience
Tense.`

	failures := Validate(TypeWorld, "Overview", content)
	var missing []string
	for _, f := range failures {
		if f.ID == "2.2" {
			missing = append(missing, f.Section)
		}
	}
	if !contains(missing, "## Core Concept") {
		t.Errorf("expected missing section '## Core Concept', got: %v", missing)
	}
}

func TestValidate_DetectsEmptySection(t *testing.T) {
	content := `## Setting

## Tone and Genre
Sci-Fi Western.

## Core Concept
Power.

## Thematic Backdrop
Old order.

## Reader Experience
Tense.`

	failures := Validate(TypeWorld, "Overview", content)
	var emptyOrMissing []string
	for _, f := range failures {
		if f.ID == "2.2" {
			emptyOrMissing = append(emptyOrMissing, f.Section)
		}
	}
	if !contains(emptyOrMissing, "## Setting") {
		t.Errorf("expected empty section '## Setting', got: %v", emptyOrMissing)
	}
}

func TestValidate_UnknownCardType(t *testing.T) {
	// No schema → only placeholder check applies. Content has no placeholders.
	failures := Validate("UNKNOWN", "", "some content here, fully written out")
	if len(failures) != 0 {
		t.Errorf("expected 0 failures for unknown card type, got %d", len(failures))
	}
}

func TestValidate_CharacterFull(t *testing.T) {
	// Build a card with all required CHARACTER FULL sections.
	sections := []string{
		"## Identity", "## Want", "## Need", "## Fear", "## Misbelief",
		"## Wound", "## Arc Progression", "## Psychology", "## Voice and Mannerisms",
		"## Skills and Knowledge", "## Physical Description", "## Hard Limits",
		"## Relationship Dynamics", "## Active Chapters", "## Voice Anchor",
	}
	var sb strings.Builder
	for _, s := range sections {
		sb.WriteString(s + "\nSome content here.\n\n")
	}
	failures := Validate(TypeCharacter, TierFull, sb.String())
	if len(failures) != 0 {
		t.Errorf("expected 0 failures for complete CHARACTER FULL, got %d: %+v", len(failures), failures)
	}
}

func contains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}
