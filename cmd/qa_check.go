package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/poorlyordered/writers_harness/internal/cards"
	"github.com/poorlyordered/writers_harness/internal/qa"
	"github.com/poorlyordered/writers_harness/internal/queue"
	"github.com/poorlyordered/writers_harness/internal/storage"
	"github.com/poorlyordered/writers_harness/internal/utils"
)

var qaCheckCmd = &cobra.Command{
	Use:   "qa-check [1-4] [args]",
	Short: "Run a QA checklist manually (2=card <file>, 3=pre-prose gate, 4=not applicable)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		n, err := strconv.Atoi(args[0])
		if err != nil || n < 1 || n > 4 {
			return fmt.Errorf("argument must be 1, 2, 3, or 4")
		}
		switch n {
		case 1:
			fmt.Println("QA-1 (phase gate) runs automatically at the end of Phase 1 and Phase 2.")
			fmt.Println("To re-run, use 'harness phase1' or 'harness phase2'.")
			return nil
		case 2:
			return runQA2(args[1:])
		case 3:
			return runQA3()
		case 4:
			fmt.Println("QA-4 (post-prose) runs automatically during Phase 4 after each scene.")
			return nil
		}
		return nil
	},
}

func init() { rootCmd.AddCommand(qaCheckCmd) }

// runQA2 validates a card file from local storage.
// Usage: harness qa-check 2 <card-filename>
// The card is looked up in the active session's cards folder.
func runQA2(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("qa-check 2 requires a card filename\nUsage: harness qa-check 2 <filename>")
	}

	state, err := sessMgr.Load()
	if err != nil {
		return fmt.Errorf("no active session: %w", err)
	}

	local := storage.NewLocal(cfg.Local.SyncFolder)
	filename := args[0]
	folder := utils.CardsFolder(state.SeriesTitle, state.BookTitle)

	content, err := local.Read(folder, filename)
	if err != nil {
		return fmt.Errorf("reading card %s: %w\n(looked in %s)", filename, err, folder)
	}

	// Infer card type and subtype from filename prefix.
	cardType, cardSubtype := inferCardTypeFromFilename(filename)

	failures := cards.Validate(cardType, cardSubtype, content)
	isLocked := cards.IsLocked(content)

	result := qa.CardCompletion(qa.CardQAInput{
		CardType:        cardType,
		CardSubtype:     cardSubtype,
		Name:            filename,
		ExistsInBox:     true,
		IsLocked:        isLocked,
		VersionInHeader: strings.Contains(content, "**Version:**"),
		Failures:        failures,
	})

	fmt.Printf("\nQA-2 check for: %s\n", filename)
	fmt.Println(result.FormatFailures())
	return nil
}

// runQA3 checks the pre-prose readiness gate against the current card queue.
func runQA3() error {
	state, err := sessMgr.Load()
	if err != nil {
		return fmt.Errorf("no active session: %w", err)
	}

	local := storage.NewLocal(cfg.Local.SyncFolder)
	data, err := local.Read(
		utils.QAFolder(state.SeriesTitle, state.BookTitle),
		utils.CardQueueFile(state.BookTitle),
	)
	if err != nil {
		return fmt.Errorf("card queue not found — run 'harness phase2' first: %w", err)
	}

	q, err := queue.Unmarshal([]byte(data))
	if err != nil {
		return fmt.Errorf("parsing card queue: %w", err)
	}

	doc := computePreProseDocFromQueue(q)
	result := qa.PreProse(doc)
	fmt.Println("\nQA-3: Pre-Prose Readiness Gate")
	fmt.Println(result.FormatFailures())
	return nil
}

// computePreProseDocFromQueue mirrors phases.computePreProseDoc without importing phases.
func computePreProseDocFromQueue(q *queue.Queue) qa.PreProseDoc {
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

// inferCardTypeFromFilename guesses the card type from a SPEC-007 filename.
func inferCardTypeFromFilename(filename string) (cardType, subtype string) {
	upper := strings.ToUpper(filename)
	switch {
	case strings.HasPrefix(upper, "WORLD-OVERVIEW"):
		return cards.TypeWorld, "Overview"
	case strings.HasPrefix(upper, "WORLD-HISTORY"):
		return cards.TypeWorld, "History"
	case strings.HasPrefix(upper, "WORLD-POLITICAL"):
		return cards.TypeWorld, "Political"
	case strings.HasPrefix(upper, "WORLD-TECHNOLOGY"):
		return cards.TypeWorld, "Technology"
	case strings.HasPrefix(upper, "WORLD-CULTURE"):
		return cards.TypeWorld, "Culture"
	case strings.HasPrefix(upper, "WORLD-GEO"):
		return cards.TypeWorld, "Geography"
	case strings.HasPrefix(upper, "WORLD-CONSTRAINTS"):
		return cards.TypeWorld, "Constraints"
	case strings.HasPrefix(upper, "CHAR-"):
		if strings.Contains(upper, "-SKETCH-") {
			return cards.TypeCharacter, cards.TierSketch
		}
		return cards.TypeCharacter, cards.TierFull
	case strings.HasPrefix(upper, "FACTION-"):
		if strings.Contains(upper, "-SKETCH-") {
			return cards.TypeFaction, cards.TierSketch
		}
		return cards.TypeFaction, cards.TierFull
	case strings.HasPrefix(upper, "THREAT-"):
		if strings.Contains(upper, "-SKETCH-") {
			return cards.TypeThreat, cards.TierSketch
		}
		return cards.TypeThreat, cards.TierFull
	case strings.HasPrefix(upper, "NOVEL-"):
		return cards.TypeNovel, ""
	case strings.HasPrefix(upper, "TRILOGY-"):
		return cards.TypeTrilogy, ""
	case strings.HasPrefix(upper, "CH"):
		return cards.TypeChapter, ""
	case strings.HasPrefix(upper, "SCENE-"):
		return cards.TypeScene, ""
	default:
		return "", ""
	}
}
