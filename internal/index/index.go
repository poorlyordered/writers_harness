// Package index maintains the Series Index file (INDEX-[SeriesTitle].md) in Box.
// The index is auto-updated whenever any card is created, updated, or status-changed.
// Format defined in SPEC-007 §11.
package index

import (
	"fmt"
	"strings"
	"time"
)

// Entry represents one card row in the series index.
type Entry struct {
	Name        string
	CardType    string
	Tier        string // "FULL", "SKETCH", or "" for non-character cards
	Version     int
	Status      string
	LastUpdated string
	BookTitle   string // empty for series-level cards
}

// Index holds the full state of a series index.
type Index struct {
	SeriesTitle string
	LastUpdated string
	Books       []BookSection
	SeriesCards []Entry
}

// BookSection groups per-book entries under a book heading.
type BookSection struct {
	BookTitle string
	BookNum   int
	Phase     string // "1", "2", "3", "4", "COMPLETE"
	Progress  string // e.g. "12 of 40 chapters complete" or ""
	Cards     []Entry
	Flags     IndexFlags
}

// IndexFlags captures outstanding flag counts for a book.
type IndexFlags struct {
	NeedsReview  int
	CardRequired int
}

// Render produces the full Markdown content of the series index.
func (idx *Index) Render() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# INDEX — %s\n", idx.SeriesTitle))
	sb.WriteString(fmt.Sprintf("**Last updated:** %s\n", idx.LastUpdated))
	sb.WriteString(fmt.Sprintf("**Books:** %d\n\n---\n\n", len(idx.Books)))

	if len(idx.SeriesCards) > 0 {
		sb.WriteString("## Series-Level Cards\n\n")
		sb.WriteString("| Card | Version | Status | Last Updated |\n")
		sb.WriteString("|------|---------|--------|:------------:|\n")
		for _, e := range idx.SeriesCards {
			sb.WriteString(fmt.Sprintf("| %s | v%d | %s | %s |\n",
				e.Name, e.Version, e.Status, e.LastUpdated))
		}
		sb.WriteString("\n---\n\n")
	}

	for _, book := range idx.Books {
		sb.WriteString(fmt.Sprintf("## %s — Book %d\n\n", book.BookTitle, book.BookNum))
		sb.WriteString(fmt.Sprintf("**Phase:** %s\n", book.Phase))
		if book.Progress != "" {
			sb.WriteString(fmt.Sprintf("**Phase 4 progress:** %s\n", book.Progress))
		}
		sb.WriteString("\n### Cards\n\n")
		sb.WriteString("| Card | Type | Tier | Version | Status | Last Updated |\n")
		sb.WriteString("|------|------|------|---------|--------|:------------:|\n")
		for _, e := range book.Cards {
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | v%d | %s | %s |\n",
				e.Name, e.CardType, e.Tier, e.Version, e.Status, e.LastUpdated))
		}
		sb.WriteString("\n### Outstanding Flags\n\n")
		sb.WriteString(fmt.Sprintf("**NEEDS-REVIEW:** %d cards\n", book.Flags.NeedsReview))
		sb.WriteString(fmt.Sprintf("**CARD-REQUIRED:** %d outstanding\n\n", book.Flags.CardRequired))
		sb.WriteString("---\n\n")
	}
	return sb.String()
}

// NewEmpty creates a fresh index for a series with no cards yet.
func NewEmpty(seriesTitle string) *Index {
	return &Index{
		SeriesTitle: seriesTitle,
		LastUpdated: time.Now().Format("2006-01-02"),
	}
}

// AddSeriesCard adds or updates a series-level card entry and refreshes LastUpdated.
func (idx *Index) AddSeriesCard(e Entry) {
	for i, existing := range idx.SeriesCards {
		if existing.Name == e.Name {
			idx.SeriesCards[i] = e
			idx.LastUpdated = time.Now().Format("2006-01-02")
			return
		}
	}
	idx.SeriesCards = append(idx.SeriesCards, e)
	idx.LastUpdated = time.Now().Format("2006-01-02")
}

// EnsureBook returns the BookSection for bookTitle, creating it if absent.
func (idx *Index) EnsureBook(bookTitle string, bookNum int) *BookSection {
	for i := range idx.Books {
		if idx.Books[i].BookTitle == bookTitle {
			return &idx.Books[i]
		}
	}
	idx.Books = append(idx.Books, BookSection{
		BookTitle: bookTitle,
		BookNum:   bookNum,
		Phase:     "1",
	})
	return &idx.Books[len(idx.Books)-1]
}

// AddBookCard adds or updates a card entry within a book section.
func (idx *Index) AddBookCard(bookTitle string, bookNum int, e Entry) {
	book := idx.EnsureBook(bookTitle, bookNum)
	for i, existing := range book.Cards {
		if existing.Name == e.Name {
			book.Cards[i] = e
			idx.LastUpdated = time.Now().Format("2006-01-02")
			return
		}
	}
	book.Cards = append(book.Cards, e)
	idx.LastUpdated = time.Now().Format("2006-01-02")
}
