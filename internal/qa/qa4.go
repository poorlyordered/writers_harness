package qa

import "fmt"

// SceneDoc holds computed state for a QA-4 post-prose scene check.
type SceneDoc struct {
	ChapterNum          int
	SceneNum            int
	WordCountMet        bool // meets target word count
	HasOpeningHook      bool // first paragraph hooks the reader
	HasTurnPoint        bool // scene changes state
	HasClosingBeat      bool // closes on a resonant beat (not mid-action)
	NoNewUncardedChars  bool // no unnamed character introduced without a card
	POVConsistent       bool // POV does not shift within the scene
	VoiceConsistent     bool // character voice matches Voice Anchor in CHARACTER card
	NoContradictions    bool // no continuity errors with locked cards
	NoFlags             bool // no NEEDS-REVIEW or CARD-REQUIRED flags in the draft
	AdvancesChapterBeat bool // scene moves the chapter's story function forward
}

// PostProse runs the QA-4 post-prose checklist for a single scene (SPEC-004C §7).
func PostProse(doc SceneDoc) *Result {
	label := fmt.Sprintf("4 (Scene: Ch%02d-S%d)", doc.ChapterNum, doc.SceneNum)
	r := &Result{ChecklistID: label}

	check := func(id, desc string, pass bool, reason string) {
		r.Items = append(r.Items, CheckItem{
			ID:          id,
			Description: desc,
			Passed:      pass,
			FailReason:  reason,
		})
	}

	check("4.1", "Scene meets target word count", doc.WordCountMet, "under target word count")
	check("4.2", "Opening hook present", doc.HasOpeningHook, "scene opens without a hook")
	check("4.3", "Scene turn point present", doc.HasTurnPoint, "no clear scene-level turn")
	check("4.4", "Scene closes on resonant beat", doc.HasClosingBeat, "scene ends mid-action or unresolved")
	check("4.5", "No uncarded new characters introduced", doc.NoNewUncardedChars, "new character without a card")
	check("4.6", "POV consistent throughout scene", doc.POVConsistent, "POV shifts detected")
	check("4.7", "Character voice matches Voice Anchor", doc.VoiceConsistent, "voice inconsistency detected")
	check("4.8", "No continuity errors with locked cards", doc.NoContradictions, "contradiction with locked card")
	check("4.9", "No NEEDS-REVIEW or CARD-REQUIRED flags", doc.NoFlags, "flags found in draft")
	check("4.10", "Scene advances chapter story function", doc.AdvancesChapterBeat, "scene is tangential to chapter beat")

	return r
}

// PostProseChapter runs a chapter-level QA-4 check after all scenes are assembled.
func PostProseChapter(chapterNum int, allScenesPass bool, wordCount, targetWords int) *Result {
	label := fmt.Sprintf("4 (Chapter: Ch%02d)", chapterNum)
	r := &Result{ChecklistID: label}

	check := func(id, desc string, pass bool, reason string) {
		r.Items = append(r.Items, CheckItem{
			ID:          id,
			Description: desc,
			Passed:      pass,
			FailReason:  reason,
		})
	}

	check("4.C1", "All scenes in chapter pass QA-4", allScenesPass, "one or more scenes failed QA-4")
	withinRange := wordCount >= int(float64(targetWords)*0.85) && wordCount <= int(float64(targetWords)*1.15)
	check("4.C2", fmt.Sprintf("Chapter word count within 15%% of target (%d/%d)", wordCount, targetWords), withinRange, "")

	return r
}
