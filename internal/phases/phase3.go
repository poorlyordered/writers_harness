package phases

import (
	"context"
	"fmt"
	"strings"

	"github.com/poorlyordered/writers_harness/internal/cards"
	"github.com/poorlyordered/writers_harness/internal/qa"
	"github.com/poorlyordered/writers_harness/internal/queue"
	"github.com/poorlyordered/writers_harness/internal/storage"
	"github.com/poorlyordered/writers_harness/internal/utils"
)

// Phase3Config holds everything the Phase 3 runner needs.
type Phase3Config struct {
	AI               Conversation
	BoxWriter        FileWriter
	LocalWriter      FileWriter
	LocalReader      *storage.Local
	Prompter         PromptLoader
	SeriesTitle      string
	BookTitle        string
	IsStandalone     bool
	ExpansionContent string
}

// RunPhase3 executes the Phase 3 card completion loop.
// Cards are built in priority order from the Card Queue. Each card is drafted via
// AI conversation, validated with QA-2, saved to Box+local, and the queue updated.
// On completion, the QA-3 pre-prose gate is checked before advancing to Phase 4.
func RunPhase3(ctx context.Context, cfg Phase3Config) error {
	fmt.Printf("\n═══════════════════════════════════════════════════════════\n")
	fmt.Printf("  WRITING HARNESS — PHASE 3: CARD COMPLETION\n")
	fmt.Printf("  Series: %s  |  Book: %s\n", cfg.SeriesTitle, cfg.BookTitle)
	fmt.Printf("═══════════════════════════════════════════════════════════\n\n")

	q, err := loadCardQueue(cfg)
	if err != nil {
		return err
	}
	printQueueProgress(q)

	systemPrompt, err := cfg.Prompter.LoadWithVars("system/phase3", map[string]string{
		"{{SERIES_TITLE}}":      cfg.SeriesTitle,
		"{{BOOK_TITLE}}":        cfg.BookTitle,
		"{{EXPANSION_CONTENT}}": cfg.ExpansionContent,
	})
	if err != nil {
		return fmt.Errorf("loading phase3 system prompt: %w", err)
	}

	qaLogPath := cfg.LocalReader.AbsPath(
		utils.QAFolder(cfg.SeriesTitle, cfg.BookTitle),
		utils.QALogFile(cfg.BookTitle),
	)
	qaLog := qa.NewLog(qaLogPath)

	for {
		next := q.Next()
		if next == nil {
			break
		}
		if err := buildOneCard(ctx, cfg, systemPrompt, q, next, qaLog); err != nil {
			return err
		}
		if err := saveCardQueue(cfg, q); err != nil {
			return fmt.Errorf("saving card queue: %w", err)
		}
		printQueueProgress(q)
	}

	if q.Progress.NotStarted > 0 {
		fmt.Printf("[HARNESS] Warning: %d card(s) not started — prerequisite order may be blocked.\n",
			q.Progress.NotStarted)
	}

	q.PhaseGateStatus = queue.GatePhase3Complete
	if err := saveCardQueue(cfg, q); err != nil {
		return err
	}

	// QA-3 pre-prose gate.
	fmt.Printf("\n─── QA-3: PRE-PROSE READINESS GATE ────────────────────────\n")
	qa3Result := qa.PreProse(computePreProseDoc(q))
	fmt.Println(qa3Result.FormatFailures())
	_ = qaLog.AppendResult(qa3Result, "Phase3→4 gate", nil)

	if !qa3Result.Passed() {
		fmt.Println("[HARNESS] QA-3 FAILED. Resolve items above before advancing to Phase 4.")
		return fmt.Errorf("QA-3 pre-prose gate failed: %d item(s) did not pass", len(qa3Result.FailedItems()))
	}

	fmt.Println("[HARNESS] QA-3 PASSED ✓  Phase 3 complete. Ready to advance to Phase 4.")
	return nil
}

// buildOneCard runs the full draft/validate/save cycle for a single queue item.
// It is recursive once on ResolveRetry to allow the writer to restart the conversation.
func buildOneCard(ctx context.Context, cfg Phase3Config, systemPrompt string, q *queue.Queue, item *queue.QueueItem, qaLog *qa.Log) error {
	fmt.Printf("\n─── CARD %d/%d: %s — %s", item.Priority, q.Progress.TotalCards, item.CardType, item.Name)
	if item.CardSubtype != "" {
		fmt.Printf(" (%s)", item.CardSubtype)
	}
	if item.Tier != nil {
		fmt.Printf(" [%s]", *item.Tier)
	}
	fmt.Println()
	fmt.Printf("    Source: %s\n    File: %s\n\n", strings.Join(item.SourceSections, ", "), item.Filename)

	instruction := buildCardInstruction(item)
	content, err := cards.Build(ctx, cfg.AI, systemPrompt, instruction)
	if err != nil {
		return fmt.Errorf("building card %q: %w", item.Name, err)
	}

	failures := cards.Validate(item.CardType, item.CardSubtype, content)
	qa2Result := qa.CardCompletion(qa.CardQAInput{
		CardType:        item.CardType,
		CardSubtype:     item.CardSubtype,
		Name:            item.Name,
		ExistsInBox:     true,
		IsLocked:        true,
		VersionInHeader: true,
		Failures:        failures,
	})

	if !qa2Result.Passed() {
		fmt.Printf("\n─── QA-2 FAILURES ──────────────────────────────────────────\n")
		choice, ids := qa.PresentAndChoose(qa2Result)
		_ = qaLog.AppendResult(qa2Result, item.Name, ids)
		switch choice {
		case qa.ResolveBlock:
			return fmt.Errorf("advancement blocked at card %q", item.Name)
		case qa.ResolveRetry:
			return buildOneCard(ctx, cfg, systemPrompt, q, item, qaLog)
		}
		// ResolveOverride falls through and saves the card as-is.
	} else {
		_ = qaLog.AppendResult(qa2Result, item.Name, nil)
	}

	locked := cards.Lock(content, item.CardType, item.Name, 1)
	folder := utils.CardsFolder(cfg.SeriesTitle, cfg.BookTitle)
	if err := cfg.BoxWriter.Write(folder, item.Filename, locked); err != nil {
		return fmt.Errorf("writing card to Box: %w", err)
	}
	if err := cfg.LocalWriter.Write(folder, item.Filename, locked); err != nil {
		fmt.Printf("[HARNESS] Warning: local card backup failed: %v\n", err)
	}
	if err := q.MarkComplete(item.Priority, item.Filename); err != nil {
		return fmt.Errorf("updating queue: %w", err)
	}
	fmt.Printf("[HARNESS] Card locked ✓  %s/%s\n", folder, item.Filename)
	return nil
}

// buildCardInstruction constructs the AI draft instruction for one queue item.
func buildCardInstruction(item *queue.QueueItem) string {
	tierNote := ""
	if item.Tier != nil {
		tierNote = fmt.Sprintf(" (%s tier)", *item.Tier)
	}
	sections := cards.RequiredSectionsList(item.CardType, item.CardSubtype)
	sectionsNote := ""
	if sections != "" {
		sectionsNote = "\n\nRequired sections (use these exact ## headings):\n" + sections
	}
	return fmt.Sprintf(
		"Draft a %s%s card for: %s\n\nSource sections: %s%s\n\n"+
			"Reference the expansion document and any locked cards in your context. "+
			"Write complete, card-ready prose for every section — no placeholders.",
		item.CardType, tierNote, item.Name,
		strings.Join(item.SourceSections, ", "),
		sectionsNote,
	)
}

// loadCardQueue reads the card queue from local storage.
func loadCardQueue(cfg Phase3Config) (*queue.Queue, error) {
	data, err := cfg.LocalReader.Read(
		utils.QAFolder(cfg.SeriesTitle, cfg.BookTitle),
		utils.CardQueueFile(cfg.BookTitle),
	)
	if err != nil {
		return nil, fmt.Errorf("card queue not found (%v) — run 'harness phase2' first", err)
	}
	return queue.Unmarshal([]byte(data))
}

// saveCardQueue persists the queue to Box and local storage.
func saveCardQueue(cfg Phase3Config, q *queue.Queue) error {
	data, err := q.Marshal()
	if err != nil {
		return err
	}
	folder := utils.QAFolder(cfg.SeriesTitle, cfg.BookTitle)
	file := utils.CardQueueFile(cfg.BookTitle)
	if err := cfg.BoxWriter.Write(folder, file, string(data)); err != nil {
		return fmt.Errorf("saving queue to Box: %w", err)
	}
	if err := cfg.LocalWriter.Write(folder, file, string(data)); err != nil {
		fmt.Printf("[HARNESS] Warning: local queue save failed: %v\n", err)
	}
	return nil
}

// computePreProseDoc inspects the completed queue to populate the QA-3 checklist.
func computePreProseDoc(q *queue.Queue) qa.PreProseDoc {
	doc := qa.PreProseDoc{
		QueueComplete:      q.Progress.NotStarted == 0 && q.Progress.InProgress == 0,
		NoOutstandingFlags: true,
	}

	charIndex := 0
	for _, item := range q.Queue {
		complete := item.Status == queue.StatusComplete
		switch item.CardType {
		case queue.CardTypeNovel:
			doc.NovelLocked = complete
		case queue.CardTypeWorld:
			if item.CardSubtype == "Overview" {
				doc.WorldOverviewLocked = complete
			}
			if item.CardSubtype == "Constraints" {
				doc.WorldConstraintsLocked = complete
			}
		case queue.CardTypeCharacter:
			if item.Tier != nil && *item.Tier == "FULL" && complete {
				charIndex++
				if charIndex == 1 {
					doc.ProtagonistLocked = true
				} else if charIndex == 2 {
					doc.AntagonistLocked = true
				}
			}
		case queue.CardTypeChapter:
			if complete {
				doc.ChapterCardsStarted = true
			}
		}
	}

	// AllFullTierLocked: every FULL tier card is complete.
	fullTotal, fullDone := 0, 0
	for _, item := range q.Queue {
		if item.Tier != nil && *item.Tier == "FULL" {
			fullTotal++
			if item.Status == queue.StatusComplete {
				fullDone++
			}
		}
	}
	doc.AllFullTierLocked = fullTotal > 0 && fullTotal == fullDone

	return doc
}

// printQueueProgress shows a compact progress line.
func printQueueProgress(q *queue.Queue) {
	p := q.Progress
	fmt.Printf("[HARNESS] Queue: %d total | %d complete | %d remaining (%.0f%%)\n",
		p.TotalCards, p.Complete, p.NotStarted, p.PercentComplete)
}
