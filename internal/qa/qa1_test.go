package qa

import "testing"

func TestPhase1To2_AllPass(t *testing.T) {
	seed := SeedDocument{
		ExistsInBox:          true,
		Status:               "LOCKED",
		HasLayer1:            true,
		HasLayer2:            true,
		HasLayer3:            true,
		ProtagonistWant:      true,
		ProtagonistNeed:      true,
		ProtagonistFear:      true,
		ProtagonistMisbelief: true,
		AntagonistBelief:     true,
		AntagonistMirror:     true,
		CentralQuestion:      true,
		ThemeHint:            true,
		SeriesQuestions:      true,
		AntagonistLadder:     true,
		IsStandalone:         false,
		VersionRecorded:      true,
		NoNeedsReviewFlags:   true,
		NoCardRequiredFlags:  true,
	}
	result := Phase1To2(seed)
	if !result.Passed() {
		t.Errorf("expected all pass, got failures: %s", result.FormatFailures())
	}
}

func TestPhase1To2_MissingStatus(t *testing.T) {
	seed := SeedDocument{ExistsInBox: true, Status: "DRAFT"}
	result := Phase1To2(seed)
	if result.Passed() {
		t.Error("expected failure for DRAFT status")
	}
	hasItem := false
	for _, item := range result.FailedItems() {
		if item.ID == "1.2" {
			hasItem = true
		}
	}
	if !hasItem {
		t.Error("expected item 1.2 (status) to fail")
	}
}

func TestPhase1To2_StandaloneSkipsSeriesItems(t *testing.T) {
	seed := SeedDocument{
		ExistsInBox:          true,
		Status:               "LOCKED",
		HasLayer1:            true,
		HasLayer2:            true,
		HasLayer3:            true,
		ProtagonistWant:      true,
		ProtagonistNeed:      true,
		ProtagonistFear:      true,
		ProtagonistMisbelief: true,
		AntagonistBelief:     true,
		AntagonistMirror:     true,
		CentralQuestion:      true,
		ThemeHint:            true,
		IsStandalone:         true, // series items should be skipped
		VersionRecorded:      true,
		NoNeedsReviewFlags:   true,
		NoCardRequiredFlags:  true,
	}
	result := Phase1To2(seed)
	if !result.Passed() {
		t.Errorf("standalone should pass without series items: %s", result.FormatFailures())
	}
}

func TestPhase2To3_AllPass(t *testing.T) {
	exp := ExpansionDocument{
		ExistsInBox:           true,
		Step1Locked:           true,
		Step2Locked:           true,
		Step3Locked:           true,
		Step4Locked:           true,
		Step5Locked:           true,
		Step6Locked:           true,
		Step7Locked:           true,
		StoryCircleMapPresent: true,
		FortyChapterMapPresent: true,
		SevenPlotsConfirmed:   true,
		SubplotWeavePresent:   true,
		FullTierCharsComplete: true,
		SketchTierPresent:     true,
		CardQueueGenerated:    true,
		CardQueueHasMinimum:   true,
		IsStandalone:          false,
		NoNeedsReviewFlags:    true,
		NoCardRequiredFlags:   true,
		NotContradictseed:     true,
	}
	result := Phase2To3(exp)
	if !result.Passed() {
		t.Errorf("expected all pass, got failures: %s", result.FormatFailures())
	}
}

func TestPhase2To3_MissingStep(t *testing.T) {
	exp := ExpansionDocument{
		ExistsInBox: true,
		Step1Locked: false, // missing
	}
	result := Phase2To3(exp)
	if result.Passed() {
		t.Error("expected failure for missing Step 1")
	}
}

func TestResult_FailedItems(t *testing.T) {
	r := &Result{
		Items: []CheckItem{
			{ID: "1.1", Passed: true},
			{ID: "1.2", Passed: false},
			{ID: "1.3", Passed: false},
		},
	}
	failed := r.FailedItems()
	if len(failed) != 2 {
		t.Errorf("FailedItems() = %d, want 2", len(failed))
	}
}
