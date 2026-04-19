package phases

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/poorlyordered/writers_harness/internal/storage"
	"github.com/poorlyordered/writers_harness/internal/utils"
)

// CheckConfig holds everything the consistency check mode needs.
type CheckConfig struct {
	AI           Conversation
	LocalReader  *storage.Local
	Prompter     PromptLoader
	SeriesTitle  string
	BookTitle    string
	Filename     string // file to check (in local storage)
	Content      string // pre-loaded file content
	DocumentType string // "CARD" or "PROSE"
	CardType     string // inferred from filename for display
}

// RunCheck loads canon context and asks Claude to surface inconsistencies
// in the target document. After the report the writer can ask follow-up
// questions in a short conversation loop.
func RunCheck(ctx context.Context, cfg CheckConfig) error {
	fmt.Printf("\n═══════════════════════════════════════════════════════════\n")
	fmt.Printf("  WRITING HARNESS — CONSISTENCY CHECK\n")
	fmt.Printf("  File: %s\n", cfg.Filename)
	fmt.Printf("═══════════════════════════════════════════════════════════\n\n")

	canonContext := loadCheckContext(cfg)

	systemPrompt, err := cfg.Prompter.LoadWithVars("system/check", map[string]string{
		"{{SERIES_TITLE}}":      cfg.SeriesTitle,
		"{{BOOK_TITLE}}":        cfg.BookTitle,
		"{{LOADED_CONTEXT}}":    canonContext,
		"{{DOCUMENT_TYPE}}":     cfg.DocumentType,
		"{{DOCUMENT_FILENAME}}": cfg.Filename,
		"{{DOCUMENT_CONTENT}}":  cfg.Content,
	})
	if err != nil {
		return fmt.Errorf("loading check system prompt: %w", err)
	}

	cfg.AI.ResetHistory()

	resp, err := cfg.AI.Send(ctx, systemPrompt,
		fmt.Sprintf("Run a consistency check on %s against the loaded canon. Produce the full report.", cfg.Filename))
	if err != nil {
		return err
	}
	fmt.Printf("\n%s\n", resp)

	// Follow-up conversation loop.
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("\n─── Follow-up questions (EXIT to finish) ───────────────────\n")
	for {
		fmt.Print("\nYou: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("reading input: %w", err)
		}
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}
		upper := strings.ToUpper(input)
		if upper == "EXIT" || upper == "DONE" || upper == "QUIT" {
			fmt.Println("\n[HARNESS] Consistency check complete.")
			return nil
		}

		resp, err = cfg.AI.Send(ctx, systemPrompt, input)
		if err != nil {
			return err
		}
		fmt.Printf("\n%s\n", resp)
	}
}

// loadCheckContext loads all locked canon cards relevant to the document being checked.
func loadCheckContext(cfg CheckConfig) string {
	var parts []string
	cardsFolder := utils.CardsFolder(cfg.SeriesTitle, cfg.BookTitle)

	tryLoad := func(folder, file, label string) {
		content, err := cfg.LocalReader.Read(folder, file)
		if err == nil && content != "" {
			parts = append(parts, fmt.Sprintf("=== %s ===\n%s", label, content))
		}
	}

	// Always load world and novel cards.
	tryLoad(cardsFolder, utils.WorldCardFile("Overview", 1), "WORLD — OVERVIEW")
	tryLoad(cardsFolder, utils.WorldCardFile("Constraints", 1), "WORLD — CONSTRAINTS")
	tryLoad(cardsFolder, utils.NovelCardFile(cfg.BookTitle, 1), "NOVEL CARD")
	tryLoad(utils.TrilogyFolder(cfg.SeriesTitle), utils.TrilogyFile(1), "TRILOGY CARD")

	// Load all character, faction, and threat cards.
	for _, prefix := range []string{"CHAR-", "FACTION-", "THREAT-"} {
		files, _ := cfg.LocalReader.ListFiles(cardsFolder, prefix)
		for _, f := range files {
			content, err := cfg.LocalReader.Read(cardsFolder, f)
			if err == nil {
				label := strings.TrimSuffix(f, ".md")
				parts = append(parts, fmt.Sprintf("=== %s ===\n%s", label, content))
			}
		}
	}

	// If checking prose, also load the relevant chapter card.
	if cfg.DocumentType == "PROSE" {
		chapterFiles, _ := cfg.LocalReader.ListFiles(cardsFolder, "CH")
		for _, f := range chapterFiles {
			content, err := cfg.LocalReader.Read(cardsFolder, f)
			if err == nil {
				label := strings.TrimSuffix(f, ".md")
				parts = append(parts, fmt.Sprintf("=== %s ===\n%s", label, content))
			}
		}
	}

	if len(parts) == 0 {
		return "(No locked canon cards found — check running without canon context.)"
	}
	return strings.Join(parts, "\n\n")
}
