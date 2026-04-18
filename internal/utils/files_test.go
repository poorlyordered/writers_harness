package utils

import "testing"

func TestSeriesRoot(t *testing.T) {
	got := SeriesRoot("Shadows Awaken")
	want := "Shadows-Awaken"
	if got != want {
		t.Errorf("SeriesRoot() = %q, want %q", got, want)
	}
}

func TestFolderPaths(t *testing.T) {
	series := "Shadows Awaken"
	book := "Into the Dark"

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"WorldBibleFolder", WorldBibleFolder(series), "Shadows-Awaken/World Bible"},
		{"TrilogyFolder", TrilogyFolder(series), "Shadows-Awaken/Trilogy"},
		{"BookRoot", BookRoot(series, book), "Shadows-Awaken/Into-the-Dark"},
		{"SeedsFolder", SeedsFolder(series, book), "Shadows-Awaken/Into-the-Dark/Seeds"},
		{"ExpansionFolder", ExpansionFolder(series, book), "Shadows-Awaken/Into-the-Dark/Expansion"},
		{"CardsFolder", CardsFolder(series, book), "Shadows-Awaken/Into-the-Dark/Cards"},
		{"ProseFolder", ProseFolder(series, book), "Shadows-Awaken/Into-the-Dark/Prose"},
		{"QAFolder", QAFolder(series, book), "Shadows-Awaken/Into-the-Dark/QA"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestFileNames(t *testing.T) {
	series := "Shadows Awaken"
	book := "Into the Dark"

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"SeriesIndexFile", SeriesIndexFile(series), "INDEX-Shadows-Awaken.md"},
		{"SeedFile", SeedFile(book, 1), "SEED-Into-the-Dark-v1.md"},
		{"ExpansionFile", ExpansionFile(book, 1), "EXPAND-Into-the-Dark-v1.md"},
		{"TrilogyFile", TrilogyFile(1), "TRILOGY-v1.md"},
		{"WorldCardFile-Overview", WorldCardFile("Overview", 1), "WORLD-Overview-v1.md"},
		{"WorldCardFile-History", WorldCardFile("History", 2), "WORLD-History-v2.md"},
		{"WorldGeographyFile", WorldGeographyFile("New Tharsis", 1), "WORLD-Geography-New-Tharsis-v1.md"},
		{"NovelCardFile", NovelCardFile(book, 1), "NOVEL-Into-the-Dark-v1.md"},
		{"CharCardFile", CharCardFile("Jax Tarkin", 1), "CHAR-Jax-Tarkin-v1.md"},
		{"FactionCardFile", FactionCardFile("The Conclave", 1), "FACTION-The-Conclave-v1.md"},
		{"ThreatCardFile", ThreatCardFile("The Silence", 1), "THREAT-The-Silence-v1.md"},
		{"ChapterCardFile", ChapterCardFile(1, book, 1), "CH01-Into-the-Dark-v1.md"},
		{"ChapterCardFile-ch10", ChapterCardFile(10, book, 1), "CH10-Into-the-Dark-v1.md"},
		{"SceneCardFile", SceneCardFile(1, 2, 1), "SCENE-CH01-S2-v1.md"},
		{"QALogFile", QALogFile(book), "QA-LOG-Into-the-Dark.md"},
		{"PatternLogFile", PatternLogFile(book), "PATTERN-LOG-Into-the-Dark.md"},
		{"CardQueueFile", CardQueueFile(book), "CARD-QUEUE-Into-the-Dark.json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}
