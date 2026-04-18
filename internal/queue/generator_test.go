package queue

import "testing"

func makeParams(isStandalone bool) GenerateParams {
	return GenerateParams{
		SeriesTitle:      "Shadows Awaken",
		BookTitle:        "Into the Dark",
		IsStandalone:     isStandalone,
		FullCharacters:   []string{"Jax Tarkin", "Cassius Drayton"},
		SketchCharacters: []string{"Mira Voss"},
		FullFactions:     []string{"The Conclave"},
		SketchFactions:   []string{},
		FullThreats:      []string{"The Silence"},
		SketchThreats:    []string{},
		Locations:        []string{"New Tharsis", "The Reach"},
	}
}

func TestGenerate_SeriesIncludesTrilogy(t *testing.T) {
	q := Generate(makeParams(false))

	first := q.Queue[0]
	if first.CardType != CardTypeTrilogy {
		t.Errorf("first card = %q, want TRILOGY", first.CardType)
	}
	if first.Priority != 1 {
		t.Errorf("trilogy priority = %d, want 1", first.Priority)
	}
}

func TestGenerate_StandaloneOmitsTrilogy(t *testing.T) {
	q := Generate(makeParams(true))

	for _, item := range q.Queue {
		if item.CardType == CardTypeTrilogy {
			t.Error("standalone queue should not contain a TRILOGY card")
		}
	}
}

func TestGenerate_WorldCardsPresent(t *testing.T) {
	q := Generate(makeParams(false))

	subtypes := map[string]bool{}
	for _, item := range q.Queue {
		if item.CardType == CardTypeWorld {
			subtypes[item.CardSubtype] = true
		}
	}

	// Constraints is its own CardTypeConstraints, not a WORLD subtype.
	required := []string{"Overview", "History", "Political", "Technology", "Culture"}
	for _, s := range required {
		if !subtypes[s] {
			t.Errorf("missing WORLD-%s card", s)
		}
	}

	// WORLD-Constraints uses CardTypeConstraints.
	hasConstraints := false
	for _, item := range q.Queue {
		if item.CardType == CardTypeConstraints {
			hasConstraints = true
		}
	}
	if !hasConstraints {
		t.Error("missing CONSTRAINTS card")
	}
	// Two geography cards for two locations.
	geoCount := 0
	for _, item := range q.Queue {
		if item.CardType == CardTypeWorld && item.CardSubtype == "Geography" {
			geoCount++
		}
	}
	if geoCount != 2 {
		t.Errorf("geography card count = %d, want 2", geoCount)
	}
}

func TestGenerate_PriorityIsMonotonic(t *testing.T) {
	q := Generate(makeParams(false))
	for i := 1; i < len(q.Queue); i++ {
		if q.Queue[i].Priority <= q.Queue[i-1].Priority {
			t.Errorf("priority not strictly increasing at index %d: %d <= %d",
				i, q.Queue[i].Priority, q.Queue[i-1].Priority)
		}
	}
}

func TestGenerate_ConstraintsIsLast(t *testing.T) {
	q := Generate(makeParams(false))
	last := q.Queue[len(q.Queue)-1]
	if last.CardType != CardTypeConstraints {
		t.Errorf("last card = %q, want CONSTRAINTS", last.CardType)
	}
}

func TestGenerate_NovelRequiresAllWorldCards(t *testing.T) {
	q := Generate(makeParams(false))

	var novelItem *QueueItem
	worldPriorities := map[int]bool{}
	for i := range q.Queue {
		item := &q.Queue[i]
		if item.CardType == CardTypeWorld {
			worldPriorities[item.Priority] = true
		}
		if item.CardType == CardTypeNovel {
			novelItem = item
		}
	}
	if novelItem == nil {
		t.Fatal("no NOVEL card found")
	}
	for _, pre := range novelItem.Prerequisites {
		if !worldPriorities[pre] {
			t.Errorf("NOVEL has prerequisite %d which is not a WORLD card", pre)
		}
	}
}

func TestGenerate_CharacterCardsHaveTier(t *testing.T) {
	q := Generate(makeParams(false))
	for _, item := range q.Queue {
		if item.CardType == CardTypeCharacter {
			if item.Tier == nil {
				t.Errorf("CHARACTER card %q has nil Tier", item.Name)
			}
		}
	}
}

func TestGenerate_ChapterAndSceneCards(t *testing.T) {
	q := Generate(makeParams(false))

	hasChapter, hasScene := false, false
	for _, item := range q.Queue {
		if item.CardType == CardTypeChapter {
			hasChapter = true
		}
		if item.CardType == CardTypeScene {
			hasScene = true
		}
	}
	if !hasChapter {
		t.Error("no CHAPTER card in queue")
	}
	if !hasScene {
		t.Error("no SCENE card in queue")
	}
}

func TestGenerate_Progress(t *testing.T) {
	q := Generate(makeParams(false))
	if q.Progress.TotalCards == 0 {
		t.Error("TotalCards = 0")
	}
	if q.Progress.NotStarted != q.Progress.TotalCards {
		t.Errorf("all cards should be NOT_STARTED; NotStarted=%d TotalCards=%d",
			q.Progress.NotStarted, q.Progress.TotalCards)
	}
}
