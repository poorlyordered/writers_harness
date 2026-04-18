// Package utils provides file naming, path construction, and version utilities
// following the conventions defined in SPEC-007 §4.
package utils

import (
	"fmt"
	"strings"
)

// Box folder path builders — SPEC-007 §3

func SeriesRoot(seriesTitle string) string {
	return title(seriesTitle)
}

func WorldBibleFolder(seriesTitle string) string {
	return SeriesRoot(seriesTitle) + "/World Bible"
}

func TrilogyFolder(seriesTitle string) string {
	return SeriesRoot(seriesTitle) + "/Trilogy"
}

func BookRoot(seriesTitle, bookTitle string) string {
	return SeriesRoot(seriesTitle) + "/" + title(bookTitle)
}

func SeedsFolder(seriesTitle, bookTitle string) string {
	return BookRoot(seriesTitle, bookTitle) + "/Seeds"
}

func ExpansionFolder(seriesTitle, bookTitle string) string {
	return BookRoot(seriesTitle, bookTitle) + "/Expansion"
}

func CardsFolder(seriesTitle, bookTitle string) string {
	return BookRoot(seriesTitle, bookTitle) + "/Cards"
}

func ProseFolder(seriesTitle, bookTitle string) string {
	return BookRoot(seriesTitle, bookTitle) + "/Prose"
}

func QAFolder(seriesTitle, bookTitle string) string {
	return BookRoot(seriesTitle, bookTitle) + "/QA"
}

// File name builders — SPEC-007 §4 naming table

func SeriesIndexFile(seriesTitle string) string {
	return "INDEX-" + title(seriesTitle) + ".md"
}

func SeedFile(bookTitle string, version int) string {
	return fmt.Sprintf("SEED-%s-v%d.md", title(bookTitle), version)
}

func ExpansionFile(bookTitle string, version int) string {
	return fmt.Sprintf("EXPAND-%s-v%d.md", title(bookTitle), version)
}

func TrilogyFile(version int) string {
	return fmt.Sprintf("TRILOGY-v%d.md", version)
}

func TrilogyNotesFile(version int) string {
	return fmt.Sprintf("TRILOGY-NOTES-v%d.md", version)
}

func WorldCardFile(cardType string, version int) string {
	return fmt.Sprintf("WORLD-%s-v%d.md", title(cardType), version)
}

func WorldCardNotesFile(cardType string, version int) string {
	return fmt.Sprintf("WORLD-%s-NOTES-v%d.md", title(cardType), version)
}

func WorldGeographyFile(locationName string, version int) string {
	return fmt.Sprintf("WORLD-Geography-%s-v%d.md", title(locationName), version)
}

func NovelCardFile(bookTitle string, version int) string {
	return fmt.Sprintf("NOVEL-%s-v%d.md", title(bookTitle), version)
}

func CharCardFile(characterName string, version int) string {
	return fmt.Sprintf("CHAR-%s-v%d.md", title(characterName), version)
}

func CharCardNotesFile(characterName string, version int) string {
	return fmt.Sprintf("CHAR-%s-NOTES-v%d.md", title(characterName), version)
}

func FactionCardFile(factionName string, version int) string {
	return fmt.Sprintf("FACTION-%s-v%d.md", title(factionName), version)
}

func FactionCardNotesFile(factionName string, version int) string {
	return fmt.Sprintf("FACTION-%s-NOTES-v%d.md", title(factionName), version)
}

func ThreatCardFile(threatName string, version int) string {
	return fmt.Sprintf("THREAT-%s-v%d.md", title(threatName), version)
}

func ThreatCardNotesFile(threatName string, version int) string {
	return fmt.Sprintf("THREAT-%s-NOTES-v%d.md", title(threatName), version)
}

func ChapterCardFile(chapterNum int, chapterTitle string, version int) string {
	return fmt.Sprintf("CH%02d-%s-v%d.md", chapterNum, title(chapterTitle), version)
}

func SceneCardFile(chapterNum, sceneNum, version int) string {
	return fmt.Sprintf("SCENE-CH%02d-S%d-v%d.md", chapterNum, sceneNum, version)
}

func SceneDraftFile(chapterNum int, chapterTitle string, sceneNum, version int) string {
	return fmt.Sprintf("CH%02d-%s-S%d-DRAFT-v%d.md", chapterNum, title(chapterTitle), sceneNum, version)
}

func SceneRejectedFile(chapterNum int, chapterTitle string, sceneNum, version int) string {
	return fmt.Sprintf("CH%02d-%s-S%d-REJECTED-v%d.md", chapterNum, title(chapterTitle), sceneNum, version)
}

func ChapterDraftFile(chapterNum int, chapterTitle string, version int) string {
	return fmt.Sprintf("CH%02d-%s-DRAFT-v%d.md", chapterNum, title(chapterTitle), version)
}

func ManuscriptFile(bookTitle string, version int) string {
	return fmt.Sprintf("%s-MANUSCRIPT-DRAFT-v%d.md", title(bookTitle), version)
}

func QALogFile(bookTitle string) string {
	return "QA-LOG-" + title(bookTitle) + ".md"
}

func PatternLogFile(bookTitle string) string {
	return "PATTERN-LOG-" + title(bookTitle) + ".md"
}

func CardQueueFile(bookTitle string) string {
	return "CARD-QUEUE-" + title(bookTitle) + ".json"
}

// title converts a human-readable name to the hyphenated title-case format
// used in Box file and folder names (spaces → hyphens).
func title(s string) string {
	return strings.ReplaceAll(s, " ", "-")
}
