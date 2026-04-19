package phases

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/poorlyordered/writers_harness/internal/storage"
	"github.com/poorlyordered/writers_harness/internal/utils"
)

// IdeaConfig holds everything the Idea Generation mode needs.
type IdeaConfig struct {
	AI          Conversation
	LocalWriter FileWriter
	LocalReader *storage.Local
	Prompter    PromptLoader
	SeriesTitle string
	BookTitle   string
}

// RunIdea runs an open-ended brainstorming session.
// Nothing is saved to Box unless the writer explicitly types SAVE.
// The session ends when the writer types EXIT or DONE.
func RunIdea(ctx context.Context, cfg IdeaConfig) error {
	fmt.Printf("\n═══════════════════════════════════════════════════════════\n")
	fmt.Printf("  WRITING HARNESS — IDEA GENERATION\n")
	fmt.Printf("  Series: %s  |  Book: %s\n", cfg.SeriesTitle, cfg.BookTitle)
	fmt.Printf("═══════════════════════════════════════════════════════════\n")
	fmt.Printf("  Type SAVE to capture output  |  EXIT or DONE to finish\n")
	fmt.Printf("═══════════════════════════════════════════════════════════\n\n")

	loadedContext := loadIdeaContext(cfg)

	systemPrompt, err := cfg.Prompter.LoadWithVars("system/idea", map[string]string{
		"{{SERIES_TITLE}}":   cfg.SeriesTitle,
		"{{BOOK_TITLE}}":     cfg.BookTitle,
		"{{LOADED_CONTEXT}}": loadedContext,
	})
	if err != nil {
		return fmt.Errorf("loading idea system prompt: %w", err)
	}

	cfg.AI.ResetHistory()

	_, err = cfg.AI.Send(ctx, systemPrompt,
		fmt.Sprintf("Idea Generation session started for %q. What would you like to explore?", cfg.BookTitle))
	if err != nil {
		return err
	}

	reader := bufio.NewReader(os.Stdin)
	saveCount := 0

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
			fmt.Println("\n[HARNESS] Idea session ended. Nothing was committed to canon.")
			return nil
		}

		if upper == "SAVE" || strings.HasPrefix(upper, "SAVE ") {
			saveCount++
			if err := saveIdeaNote(cfg, reader, saveCount); err != nil {
				fmt.Printf("[HARNESS] Save failed: %v\n", err)
			}
			continue
		}

		resp, err := cfg.AI.Send(ctx, systemPrompt, input)
		if err != nil {
			return err
		}
		fmt.Printf("\n%s\n", resp)
	}
}

// loadIdeaContext reads the Novel card, Trilogy card, and FULL character cards
// from local storage to inject as context into the idea session.
func loadIdeaContext(cfg IdeaConfig) string {
	var parts []string

	tryLoad := func(folder, file, label string) {
		content, err := cfg.LocalReader.Read(folder, file)
		if err == nil && content != "" {
			parts = append(parts, fmt.Sprintf("=== %s ===\n%s", label, content))
		}
	}

	cardsFolder := utils.CardsFolder(cfg.SeriesTitle, cfg.BookTitle)
	seriesRoot := utils.SeriesRoot(cfg.SeriesTitle)

	// Novel card
	tryLoad(cardsFolder, utils.NovelCardFile(cfg.BookTitle, 1), "NOVEL CARD")

	// Trilogy card (series-level)
	tryLoad(utils.TrilogyFolder(cfg.SeriesTitle), utils.TrilogyFile(1), "TRILOGY CARD")

	// World Overview
	tryLoad(cardsFolder, utils.WorldCardFile("Overview", 1), "WORLD — OVERVIEW")

	// FULL character cards (heuristic: try up to 8)
	charFiles, _ := cfg.LocalReader.ListFiles(cardsFolder, "CHAR-")
	for _, f := range charFiles {
		content, err := cfg.LocalReader.Read(cardsFolder, f)
		if err == nil {
			parts = append(parts, fmt.Sprintf("=== CHARACTER: %s ===\n%s", f, content))
		}
	}

	// Trilogy notes (series-level)
	tryLoad(seriesRoot, utils.TrilogyNotesFile(1), "TRILOGY NOTES")

	if len(parts) == 0 {
		return "(No locked cards found in local storage — working without canon context.)"
	}
	return strings.Join(parts, "\n\n")
}

func saveIdeaNote(cfg IdeaConfig, reader *bufio.Reader, count int) error {
	fmt.Print("\nSave as filename (or press Enter for auto-name): ")
	filename, _ := reader.ReadString('\n')
	filename = strings.TrimSpace(filename)
	if filename == "" {
		filename = fmt.Sprintf("IDEA-NOTE-%s-%s.md",
			utils.SeriesIndexStateFile(cfg.BookTitle)[:len(cfg.BookTitle)],
			time.Now().Format("20060102-150405"))
		filename = fmt.Sprintf("IDEA-NOTE-%s-%s.md",
			strings.ReplaceAll(cfg.BookTitle, " ", "-"),
			time.Now().Format("20060102-150405"))
	}

	fmt.Print("Content to save (paste it, then press Enter twice): ")
	var lines []string
	for {
		line, _ := reader.ReadString('\n')
		line = strings.TrimRight(line, "\r\n")
		if line == "" && len(lines) > 0 {
			break
		}
		lines = append(lines, line)
	}

	content := fmt.Sprintf("# Idea Note — %s\n\n**Saved:** %s\n\n---\n\n%s\n",
		cfg.BookTitle, time.Now().Format("2006-01-02 15:04"), strings.Join(lines, "\n"))

	notesFolder := utils.ExpansionFolder(cfg.SeriesTitle, cfg.BookTitle)
	if err := cfg.LocalWriter.Write(notesFolder, filename, content); err != nil {
		return err
	}
	fmt.Printf("[HARNESS] Saved ✓  %s/%s\n", notesFolder, filename)
	return nil
}
