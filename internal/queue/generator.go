package queue

import (
	"fmt"
	"strings"

	"github.com/poorlyordered/writers_harness/internal/utils"
)

// GenerateParams holds everything needed to build a Card Queue from Phase 2 output.
type GenerateParams struct {
	SeriesTitle      string
	BookTitle        string
	IsStandalone     bool
	FullCharacters   []string // names of FULL tier characters, protagonist first
	SketchCharacters []string // names of SKETCH tier characters
	FullFactions     []string // names of FULL tier factions
	SketchFactions   []string // names of SKETCH tier factions
	FullThreats      []string // names of FULL tier threats
	SketchThreats    []string // names of SKETCH tier threats
	Locations        []string // major locations needing Geography cards
}

// Generate builds a Card Queue following the strict priority order from SPEC-005 §4.
func Generate(p GenerateParams) *Queue {
	q := New(p.SeriesTitle, p.BookTitle)

	priority := 1

	// Priority 1: Trilogy Card (series only)
	if !p.IsStandalone {
		q.Queue = append(q.Queue, QueueItem{
			Priority:       priority,
			CardType:       CardTypeTrilogy,
			CardSubtype:    "",
			Name:           p.SeriesTitle + " Trilogy",
			Filename:       utils.TrilogyFile(1),
			Status:         StatusNotStarted,
			Prerequisites:  []int{},
			SourceSections: []string{"Step 3", "Step 5"},
		})
		priority++
	}

	worldStart := priority

	// Priority 2: WORLD-Overview
	q.Queue = append(q.Queue, QueueItem{
		Priority:       priority,
		CardType:       CardTypeWorld,
		CardSubtype:    "Overview",
		Name:           "WORLD-Overview",
		Filename:       utils.WorldCardFile("Overview", 1),
		Status:         StatusNotStarted,
		Prerequisites:  prereqs(!p.IsStandalone, 1),
		SourceSections: []string{"Step 4", "Step 6"},
	})
	priority++

	// Priority 3: WORLD-History
	q.Queue = append(q.Queue, QueueItem{
		Priority:       priority,
		CardType:       CardTypeWorld,
		CardSubtype:    "History",
		Name:           "WORLD-History",
		Filename:       utils.WorldCardFile("History", 1),
		Status:         StatusNotStarted,
		Prerequisites:  []int{worldStart},
		SourceSections: []string{"Step 4"},
	})
	priority++

	// Priority 4: WORLD-Political
	q.Queue = append(q.Queue, QueueItem{
		Priority:       priority,
		CardType:       CardTypeWorld,
		CardSubtype:    "Political",
		Name:           "WORLD-Political",
		Filename:       utils.WorldCardFile("Political", 1),
		Status:         StatusNotStarted,
		Prerequisites:  []int{worldStart + 1}, // History
		SourceSections: []string{"Step 4", "Step 6"},
	})
	priority++

	// Priority 5: WORLD-Technology
	q.Queue = append(q.Queue, QueueItem{
		Priority:       priority,
		CardType:       CardTypeWorld,
		CardSubtype:    "Technology",
		Name:           "WORLD-Technology",
		Filename:       utils.WorldCardFile("Technology", 1),
		Status:         StatusNotStarted,
		Prerequisites:  []int{worldStart},
		SourceSections: []string{"Step 4"},
	})
	priority++

	// Priority 6: WORLD-Culture
	q.Queue = append(q.Queue, QueueItem{
		Priority:       priority,
		CardType:       CardTypeWorld,
		CardSubtype:    "Culture",
		Name:           "WORLD-Culture",
		Filename:       utils.WorldCardFile("Culture", 1),
		Status:         StatusNotStarted,
		Prerequisites:  []int{worldStart + 1, worldStart + 2}, // History, Political
		SourceSections: []string{"Step 4"},
	})
	priority++

	// Priority 7+: WORLD-Geography cards (one per location)
	geoStart := priority
	for _, loc := range p.Locations {
		q.Queue = append(q.Queue, QueueItem{
			Priority:       priority,
			CardType:       CardTypeWorld,
			CardSubtype:    "Geography",
			Name:           "WORLD-Geography-" + loc,
			Filename:       utils.WorldGeographyFile(loc, 1),
			Status:         StatusNotStarted,
			Prerequisites:  []int{worldStart, worldStart + 3, worldStart + 4}, // Overview, Culture, Political
			SourceSections: []string{"Step 4", "Step 6"},
		})
		priority++
	}

	// Novel Card — requires all World Bible cards
	worldPrereqs := makeRange(worldStart, geoStart-1)
	if len(p.Locations) > 0 {
		worldPrereqs = makeRange(worldStart, priority-1)
	}
	novelPriority := priority
	q.Queue = append(q.Queue, QueueItem{
		Priority:       priority,
		CardType:       CardTypeNovel,
		CardSubtype:    "",
		Name:           fmt.Sprintf("NOVEL-%s", p.BookTitle),
		Filename:       utils.NovelCardFile(p.BookTitle, 1),
		Status:         StatusNotStarted,
		Prerequisites:  worldPrereqs,
		SourceSections: []string{"Step 4", "Step 6", "Step 7"},
	})
	priority++

	// Character Cards — FULL tier, protagonist first
	charStart := priority
	charPrereqs := []int{novelPriority}
	for i, char := range p.FullCharacters {
		prereq := charPrereqs
		if i > 0 {
			// Each subsequent character requires the previous one (mirror relationship)
			prereq = []int{charStart} // protagonist must be locked first
		}
		q.Queue = append(q.Queue, QueueItem{
			Priority:       priority,
			CardType:       CardTypeCharacter,
			CardSubtype:    "FULL",
			Name:           char,
			Filename:       utils.CharCardFile(char, 1),
			Status:         StatusNotStarted,
			Tier:           fullTier(),
			Prerequisites:  prereq,
			SourceSections: []string{"Step 7"},
		})
		priority++
	}

	// Faction Cards — FULL tier
	factionStart := priority
	factionPrereqs := makeRange(charStart, charStart+len(p.FullCharacters)-1)
	for _, faction := range p.FullFactions {
		q.Queue = append(q.Queue, QueueItem{
			Priority:       priority,
			CardType:       CardTypeFaction,
			CardSubtype:    "FULL",
			Name:           faction,
			Filename:       utils.FactionCardFile(faction, 1),
			Status:         StatusNotStarted,
			Tier:           fullTier(),
			Prerequisites:  factionPrereqs,
			SourceSections: []string{"Step 5", "Step 7"},
		})
		priority++
	}

	// Threat Cards — FULL tier
	threatPrereqs := makeRange(factionStart, factionStart+len(p.FullFactions)-1)
	if !p.IsStandalone {
		threatPrereqs = append([]int{1}, threatPrereqs...) // requires Trilogy Card
	}
	for _, threat := range p.FullThreats {
		q.Queue = append(q.Queue, QueueItem{
			Priority:       priority,
			CardType:       CardTypeThreat,
			CardSubtype:    "FULL",
			Name:           threat,
			Filename:       utils.ThreatCardFile(threat, 1),
			Status:         StatusNotStarted,
			Tier:           fullTier(),
			Prerequisites:  threatPrereqs,
			SourceSections: []string{"Step 5"},
		})
		priority++
	}

	// SKETCH tier: Characters, Factions, Threats
	sketchPrereqs := []int{charStart} // after protagonist card exists
	for _, char := range p.SketchCharacters {
		q.Queue = append(q.Queue, QueueItem{
			Priority:       priority,
			CardType:       CardTypeCharacter,
			CardSubtype:    "SKETCH",
			Name:           char,
			Filename:       utils.CharCardFile(char, 1),
			Status:         StatusNotStarted,
			Tier:           sketchTier(),
			Prerequisites:  sketchPrereqs,
			SourceSections: []string{"Step 3", "Step 5"},
		})
		priority++
	}
	for _, faction := range p.SketchFactions {
		q.Queue = append(q.Queue, QueueItem{
			Priority:       priority,
			CardType:       CardTypeFaction,
			CardSubtype:    "SKETCH",
			Name:           faction,
			Filename:       utils.FactionCardFile(faction, 1),
			Status:         StatusNotStarted,
			Tier:           sketchTier(),
			Prerequisites:  sketchPrereqs,
			SourceSections: []string{"Step 3"},
		})
		priority++
	}
	for _, threat := range p.SketchThreats {
		q.Queue = append(q.Queue, QueueItem{
			Priority:       priority,
			CardType:       CardTypeThreat,
			CardSubtype:    "SKETCH",
			Name:           threat,
			Filename:       utils.ThreatCardFile(threat, 1),
			Status:         StatusNotStarted,
			Tier:           sketchTier(),
			Prerequisites:  sketchPrereqs,
			SourceSections: []string{"Step 3"},
		})
		priority++
	}

	// Chapter Cards (40) — auto-generated from Novel Card
	chapterStart := priority
	chapterPrereqs := []int{novelPriority}
	q.Queue = append(q.Queue, QueueItem{
		Priority:       priority,
		CardType:       CardTypeChapter,
		CardSubtype:    "All 40",
		Name:           "Chapter Cards (auto-generated)",
		Filename:       fmt.Sprintf("CH01..CH40-%s-v1.md", strings.ReplaceAll(p.BookTitle, " ", "-")),
		Status:         StatusNotStarted,
		Prerequisites:  chapterPrereqs,
		SourceSections: []string{"Step 4", "Step 6"},
	})
	priority++

	// Scene Cards — structural beat chapters
	q.Queue = append(q.Queue, QueueItem{
		Priority:       priority,
		CardType:       CardTypeScene,
		CardSubtype:    "Structural beats",
		Name:           "Scene Cards (Ch 1,2,6,10,17,20,27,29,36,38,40)",
		Filename:       fmt.Sprintf("SCENE-CH*-S*-%s-v1.md", strings.ReplaceAll(p.BookTitle, " ", "-")),
		Status:         StatusNotStarted,
		Prerequisites:  []int{chapterStart},
		SourceSections: []string{"Step 6", "Step 7"},
	})
	priority++

	// WORLD-Constraints — always last
	allPrereqs := makeRange(worldStart, priority-1)
	q.Queue = append(q.Queue, QueueItem{
		Priority:       priority,
		CardType:       CardTypeConstraints,
		CardSubtype:    "",
		Name:           "WORLD-Constraints",
		Filename:       utils.WorldCardFile("Constraints", 1),
		Status:         StatusNotStarted,
		Prerequisites:  allPrereqs,
		SourceSections: []string{"All World Bible cards", "All Character Cards"},
	})

	q.recalcProgress()
	return q
}

// makeRange returns a slice of ints from start to end inclusive.
func makeRange(start, end int) []int {
	if end < start {
		return []int{}
	}
	r := make([]int, end-start+1)
	for i := range r {
		r[i] = start + i
	}
	return r
}

// prereqs conditionally includes priority 1 (Trilogy Card) in prerequisites.
func prereqs(includeTrilogyCard bool, extra ...int) []int {
	if includeTrilogyCard {
		return append([]int{1}, extra...)
	}
	return extra
}
