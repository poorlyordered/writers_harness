// Package phases implements the four phase session flows.
// phase1.go is the Idea Generation session (SPEC-002).
package phases

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/poorlyordered/writers_harness/internal/ai"
	"github.com/poorlyordered/writers_harness/internal/qa"
	"github.com/poorlyordered/writers_harness/internal/utils"
)

// Phase1Config holds everything the Phase 1 runner needs.
type Phase1Config struct {
	AI          ai.Conversation
	BoxWriter   FileWriter
	LocalWriter FileWriter
	Prompter    PromptLoader

	SeriesTitle  string
	BookTitle    string
	IsStandalone bool
	DefaultGenre string
}

// Conversation is an alias kept for backward compatibility within this package.
type Conversation = ai.Conversation

// FileWriter is the interface for writing a file (to Box or local storage).
type FileWriter interface {
	Write(folderPath, filename, content string) error
}

// PromptLoader loads a named prompt template from the prompts/ directory.
type PromptLoader interface {
	Load(name string) (string, error)
	LoadWithVars(name string, vars map[string]string) (string, error)
}

// Run executes a complete Phase 1 session.
// It presents a menu, runs the chosen entry path, conducts the deepening
// conversation, locks the seed, saves it to Box, and runs QA-1.
func Run(ctx context.Context, cfg Phase1Config) error {
	cfg.AI.ResetHistory()

	fmt.Printf("\n═══════════════════════════════════════════════════════════\n")
	fmt.Printf("  WRITING HARNESS — PHASE 1: IDEA GENERATION\n")
	fmt.Printf("  Series: %s  |  Book: %s\n", cfg.SeriesTitle, cfg.BookTitle)
	fmt.Printf("═══════════════════════════════════════════════════════════\n\n")

	systemPrompt, err := cfg.Prompter.LoadWithVars("system/phase1", map[string]string{
		"{{SERIES_TITLE}}": cfg.SeriesTitle,
		"{{BOOK_TITLE}}":   cfg.BookTitle,
		"{{DEFAULT_GENRE}}": cfg.DefaultGenre,
	})
	if err != nil {
		return fmt.Errorf("loading phase1 system prompt: %w", err)
	}

	// Entry path selection.
	path, seedInput := selectEntryPath()

	// Build opening message to the AI.
	var openingMsg string
	switch path {
	case "A":
		openingMsg = fmt.Sprintf(
			"Please generate a seed for a new story. Series: %q, Book: %q.\n"+
				"Default setting: %s. Use a cross-genre trope (non-sci-fi).\n"+
				"Present all three layers and wait for my feedback before we deepen.",
			cfg.SeriesTitle, cfg.BookTitle, cfg.DefaultGenre)
	case "B":
		openingMsg = fmt.Sprintf(
			"I have a seed to import. Series: %q, Book: %q.\n"+
				"Please map it to the three-layer structure and flag any missing components.\n\nSeed:\n%s",
			cfg.SeriesTitle, cfg.BookTitle, seedInput)
	case "C":
		openingMsg = fmt.Sprintf(
			"I have a rough idea for a new story. Series: %q, Book: %q.\n"+
				"Please conduct a structured intake, extract the seed components, and then "+
				"present what you've captured for my confirmation.\n\nIdea:\n%s",
			cfg.SeriesTitle, cfg.BookTitle, seedInput)
	}

	// First AI response: seed assembly.
	fmt.Printf("\n[HARNESS] Starting seed assembly...\n\n")
	resp, err := cfg.AI.Send(ctx, systemPrompt, openingMsg)
	if err != nil {
		return err
	}
	printAssistant(resp)

	// Deepening conversation loop.
	fmt.Printf("\n[HARNESS] Seed assembled. Beginning deepening conversation.\n")
	fmt.Printf("[HARNESS] Type your responses. When you're ready to lock the seed, type LOCK.\n\n")

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("You: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("reading input: %w", err)
		}
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}
		if strings.ToUpper(input) == "LOCK" {
			break
		}

		resp, err = cfg.AI.Send(ctx, systemPrompt, input)
		if err != nil {
			return err
		}
		printAssistant(resp)
	}

	// Ask the AI to produce the final locked seed document.
	fmt.Printf("\n[HARNESS] Generating locked seed document...\n\n")
	lockPrompt := fmt.Sprintf(
		"Please generate the complete Locked Seed Prompt document now.\n"+
			"Format it as a Markdown document with:\n"+
			"- CARD TYPE: SEED\n"+
			"- BOOK TITLE: %s\n"+
			"- STATUS: LOCKED\n"+
			"- LOCKED DATE: %s\n"+
			"- VERSION: 1\n\n"+
			"Then include all three layers fully populated, plus protagonist core "+
			"(want/need/fear/misbelief), antagonist logic (belief/case/mirror), "+
			"thematic architecture (central question, theme hint), "+
			"and (since this is a series) the series questions and antagonist ladder.\n\n"+
			"Output only the document, no surrounding commentary.",
		cfg.BookTitle, time.Now().Format("2006-01-02"))

	seedDoc, err := cfg.AI.Send(ctx, systemPrompt, lockPrompt)
	if err != nil {
		return err
	}

	fmt.Printf("\n─── LOCKED SEED DOCUMENT ──────────────────────────────────\n")
	fmt.Println(seedDoc)
	fmt.Printf("───────────────────────────────────────────────────────────\n\n")

	// Confirm before saving.
	fmt.Print("Save and lock this seed? (yes/no): ")
	confirm, _ := reader.ReadString('\n')
	if strings.ToLower(strings.TrimSpace(confirm)) != "yes" {
		fmt.Println("[HARNESS] Seed not locked. Continue the conversation or restart.")
		return nil
	}

	// Save to Box and local.
	seedFilename := utils.SeedFile(cfg.BookTitle, 1)
	seedsFolder := utils.SeedsFolder(cfg.SeriesTitle, cfg.BookTitle)

	if err := cfg.BoxWriter.Write(seedsFolder, seedFilename, seedDoc); err != nil {
		return fmt.Errorf("writing seed to box: %w", err)
	}
	if err := cfg.LocalWriter.Write(seedsFolder, seedFilename, seedDoc); err != nil {
		// Non-fatal: local write failure is logged but does not block.
		fmt.Printf("[HARNESS] Warning: local backup write failed: %v\n", err)
	}

	fmt.Printf("[HARNESS] Seed saved: %s/%s\n", seedsFolder, seedFilename)

	// Run QA-1 Phase 1→2 gate.
	seedState := parseSeedQAState(seedDoc, cfg.IsStandalone)
	result := qa.Phase1To2(seedState)

	fmt.Printf("\n─── QA-1: PHASE 1→2 GATE ──────────────────────────────────\n")
	fmt.Println(result.FormatFailures())
	fmt.Printf("───────────────────────────────────────────────────────────\n\n")

	if !result.Passed() {
		fmt.Println("[HARNESS] QA-1 FAILED. Resolve the items above before advancing to Phase 2.")
		fmt.Println("[HARNESS] The seed document has been saved as LOCKED. Update it and re-run QA if needed.")
		return fmt.Errorf("QA-1 phase gate failed: %d item(s) did not pass", len(result.FailedItems()))
	}

	fmt.Println("[HARNESS] QA-1 PASSED ✓  Phase 1 complete. Ready to advance to Phase 2.")
	return nil
}

// parseSeedQAState inspects the seed document text to populate a SeedDocument
// for the QA-1 checklist. Uses simple keyword presence checks — the AI is
// responsible for producing a well-structured document.
func parseSeedQAState(doc string, isStandalone bool) qa.SeedDocument {
	has := func(keyword string) bool {
		return strings.Contains(strings.ToLower(doc), strings.ToLower(keyword))
	}
	return qa.SeedDocument{
		ExistsInBox:          true, // already saved before calling this
		Status:               "LOCKED",
		HasLayer1:            has("setting genre") && has("time period") && has("scale"),
		HasLayer2:            has("primary trope") && has("structural weight"),
		HasLayer3:            has("protagonist seed") && has("central conflict") && has("inciting moment"),
		ProtagonistWant:      has("want:"),
		ProtagonistNeed:      has("need:"),
		ProtagonistFear:      has("fear:"),
		ProtagonistMisbelief: has("misbelief:"),
		AntagonistBelief:     has("belief:"),
		AntagonistMirror:     has("mirror:") || has("mirror function"),
		CentralQuestion:      has("central question"),
		ThemeHint:            has("theme hint"),
		SeriesQuestions:      has("book 1 question") || has("series questions"),
		AntagonistLadder:     has("book 1 antagonist") || has("antagonist ladder"),
		IsStandalone:         isStandalone,
		VersionRecorded:      has("version: 1") || has("version:"),
		NoNeedsReviewFlags:   !has("needs-review"),
		NoCardRequiredFlags:  !has("card-required"),
	}
}

// selectEntryPath prompts the writer to choose Path A, B, or C and
// returns the path letter and any user-provided seed text.
func selectEntryPath() (string, string) {
	fmt.Println("Choose your entry path:")
	fmt.Println("  A — Generate a new seed (harness creates a randomized seed)")
	fmt.Println("  B — Import a seed (paste in a seed from another source)")
	fmt.Println("  C — Freeform (type a rough idea in natural language)")
	fmt.Print("\nPath (A/B/C): ")

	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')
	choice = strings.ToUpper(strings.TrimSpace(choice))

	switch choice {
	case "A":
		return "A", ""
	case "B", "C":
		label := "Paste your seed (end with a line containing only END):"
		if choice == "C" {
			label = "Describe your idea (end with a line containing only END):"
		}
		fmt.Println(label)
		var lines []string
		for {
			line, _ := reader.ReadString('\n')
			line = strings.TrimRight(line, "\r\n")
			if strings.ToUpper(strings.TrimSpace(line)) == "END" {
				break
			}
			lines = append(lines, line)
		}
		return choice, strings.Join(lines, "\n")
	default:
		fmt.Println("[HARNESS] Invalid choice. Defaulting to Path A.")
		return "A", ""
	}
}

// printAssistant formats and prints an AI response.
func printAssistant(resp string) {
	fmt.Printf("\nHarness: %s\n", resp)
}
