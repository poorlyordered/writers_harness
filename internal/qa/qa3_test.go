package qa

import "testing"

func TestPreProse_AllPass(t *testing.T) {
	doc := PreProseDoc{
		NovelLocked:            true,
		WorldOverviewLocked:    true,
		WorldConstraintsLocked: true,
		ProtagonistLocked:      true,
		AntagonistLocked:       true,
		AllFullTierLocked:      true,
		ChapterCardsStarted:    true,
		NoOutstandingFlags:     true,
		QueueComplete:          true,
	}
	result := PreProse(doc)
	if !result.Passed() {
		t.Errorf("expected all pass, got failures: %s", result.FormatFailures())
	}
}

func TestPreProse_MissingNovel(t *testing.T) {
	doc := PreProseDoc{
		WorldOverviewLocked:    true,
		WorldConstraintsLocked: true,
		ProtagonistLocked:      true,
		AntagonistLocked:       true,
		AllFullTierLocked:      true,
		ChapterCardsStarted:    true,
		NoOutstandingFlags:     true,
		QueueComplete:          true,
	}
	result := PreProse(doc)
	if result.Passed() {
		t.Error("expected failure when NOVEL card not locked")
	}
	failed := result.FailedItems()
	found := false
	for _, f := range failed {
		if f.ID == "3.1" {
			found = true
		}
	}
	if !found {
		t.Error("expected item 3.1 (NOVEL card) to fail")
	}
}

func TestPreProse_ItemCount(t *testing.T) {
	doc := PreProseDoc{}
	result := PreProse(doc)
	// QA-3 has 9 items.
	if len(result.Items) != 9 {
		t.Errorf("QA-3 item count = %d, want 9", len(result.Items))
	}
}
