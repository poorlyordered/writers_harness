// Package index maintains the Series Index file (INDEX-[SeriesTitle].md) in Box.
// The index is auto-updated whenever any card is created, updated, or status-changed.
// Format defined in SPEC-007 §11.
package index

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// Entry represents one card row in the series index.
type Entry struct {
	Name        string `json:"name"`
	CardType    string `json:"card_type"`
	Tier        string `json:"tier"`
	Version     int    `json:"version"`
	Status      string `json:"status"`
	LastUpdated string `json:"last_updated"`
	BookTitle   string `json:"book_title"`
}

// Index holds the full state of a series index.
type Index struct {
	SeriesTitle string        `json:"series_title"`
	LastUpdated string        `json:"last_updated"`
	Books       []BookSection `json:"books"`
	SeriesCards []Entry       `json:"series_cards"`
}

// BookSection groups per-book entries under a book heading.
type BookSection struct {
	BookTitle string     `json:"book_title"`
	BookNum   int        `json:"book_num"`
	Phase     string     `json:"phase"`
	Progress  string     `json:"progress,omitempty"`
	Cards     []Entry    `json:"cards"`
	Flags     IndexFlags `json:"flags"`
}

// IndexFlags captures outstanding flag counts for a book.
type IndexFlags struct {
	NeedsReview  int `json:"needs_review"`
	CardRequired int `json:"card_required"`
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

// LoadState reads Index state from a JSON file. Returns (nil, nil) if the file doesn't exist.
func LoadState(path string) (*Index, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var idx Index
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, err
	}
	return &idx, nil
}

// SaveState writes Index state to a JSON file.
func SaveState(idx *Index, path string) error {
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
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
