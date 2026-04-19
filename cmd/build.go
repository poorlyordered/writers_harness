package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/poorlyordered/writers_harness/internal/cards"
	"github.com/poorlyordered/writers_harness/internal/phases"
	"github.com/poorlyordered/writers_harness/internal/qa"
	"github.com/poorlyordered/writers_harness/internal/queue"
	"github.com/poorlyordered/writers_harness/internal/storage"
	"github.com/poorlyordered/writers_harness/internal/utils"
)

var (
	buildCardType    string
	buildCardSubtype string
	buildCardName    string
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Background Building — create or update a specific card",
	Long: `Build or update a control card for the active project.

Examples:
  harness build --card CHAR --name "Jax Tarkin"
  harness build --card WORLD --subtype Overview
  harness build --card NOVEL
  harness build --card FACTION --name "The Conclave"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runBuild(context.Background())
	},
}

func init() {
	buildCmd.Flags().StringVar(&buildCardType, "card", "", "Card type: CHAR, WORLD, NOVEL, FACTION, THREAT, CHAPTER, TRILOGY")
	buildCmd.Flags().StringVar(&buildCardSubtype, "subtype", "", "Card subtype (e.g. Overview, History) for WORLD cards, or FULL/SKETCH for character cards")
	buildCmd.Flags().StringVar(&buildCardName, "name", "", "Name of the character, faction, location, etc.")
	_ = buildCmd.MarkFlagRequired("card")
	rootCmd.AddCommand(buildCmd)
}

func runBuild(ctx context.Context) error {
	state, err := sessMgr.Load()
	if err != nil {
		return fmt.Errorf("no active session: %w", err)
	}

	local := storage.NewLocal(cfg.Local.SyncFolder)
	cardType := strings.ToUpper(buildCardType)
	cardSubtype := buildCardSubtype
	cardName := buildCardName

	// Resolve card filename and check for existing version.
	filename, err := resolveCardFilename(cardType, cardSubtype, cardName, state.BookTitle)
	if err != nil {
		return err
	}

	fmt.Printf("\n═══════════════════════════════════════════════════════════\n")
	fmt.Printf("  WRITING HARNESS — BUILD\n")
	fmt.Printf("  Card: %s%s  %s\n", cardType,
		func() string {
			if cardSubtype != "" {
				return " — " + cardSubtype
			}
			return ""
		}(),
		cardName)
	fmt.Printf("  File: %s\n", filename)
	fmt.Printf("═══════════════════════════════════════════════════════════\n\n")

	cardsFolder := utils.CardsFolder(state.SeriesTitle, state.BookTitle)

	// Load system prompt (reuse phase3 prompt).
	expandContent, _ := local.Read(
		utils.ExpansionFolder(state.SeriesTitle, state.BookTitle),
		utils.ExpansionFile(state.BookTitle, 1),
	)
	systemPrompt, err := prompter.LoadWithVars("system/phase3", map[string]string{
		"{{SERIES_TITLE}}":      state.SeriesTitle,
		"{{BOOK_TITLE}}":        state.BookTitle,
		"{{EXPANSION_CONTENT}}": expandContent,
	})
	if err != nil {
		return fmt.Errorf("loading system prompt: %w", err)
	}

	aiConv := phases.NewAIConversation(aiClient)

	// Check for existing locked card — if present, treat as a revision.
	existingContent, _ := local.Read(cardsFolder, filename)
	var content string

	instruction := buildInstruction(cardType, cardSubtype, cardName)

	if existingContent != "" && cards.IsLocked(existingContent) {
		fmt.Printf("[HARNESS] Existing locked card found — entering revision mode.\n\n")
		body := cards.ExtractBody(existingContent)
		content, err = cards.BuildFromTemplate(ctx, aiConv, systemPrompt, instruction, body)
	} else {
		// Check for a writer template.
		templateContent, _ := local.Read(
			utils.CardTemplatesFolder(state.SeriesTitle, state.BookTitle),
			utils.CardTemplateFile(filename),
		)
		if templateContent != "" && !cards.IsScaffoldOnly(templateContent) {
			fmt.Printf("[HARNESS] Template found — Claude will refine it.\n\n")
			content, err = cards.BuildFromTemplate(ctx, aiConv, systemPrompt, instruction, templateContent)
		} else {
			content, err = cards.Build(ctx, aiConv, systemPrompt, instruction)
		}
	}
	if err != nil {
		return fmt.Errorf("building card: %w", err)
	}

	// QA-2.
	failures := cards.Validate(cardType, cardSubtype, content)
	qa2Result := qa.CardCompletion(qa.CardQAInput{
		CardType:        cardType,
		CardSubtype:     cardSubtype,
		Name:            cardName,
		ExistsInBox:     true,
		IsLocked:        true,
		VersionInHeader: true,
		Failures:        failures,
	})

	if !qa2Result.Passed() {
		fmt.Printf("\n─── QA-2 FAILURES ──────────────────────────────────────────\n")
		choice, _ := qa.PresentAndChoose(qa2Result)
		if choice == qa.ResolveBlock {
			return fmt.Errorf("card blocked — resolve QA failures before locking")
		}
	}

	// Determine version for the saved file.
	existingFiles, _ := local.ListFiles(cardsFolder, strings.TrimSuffix(filename, ".md"))
	version := 1
	if len(existingFiles) > 0 {
		if best, err := utils.HighestVersion(existingFiles); err == nil {
			if v, err := utils.ParseVersion(best); err == nil {
				version = v + 1
			}
		}
	}

	locked := cards.Lock(content, cardType, cardName, version)
	boxWriter := phases.NewBoxFileWriter(aiClient)
	versionedFile := versionedFilename(filename, version)

	if err := boxWriter.Write(cardsFolder, versionedFile, locked); err != nil {
		return fmt.Errorf("writing card to Box: %w", err)
	}
	if err := local.Write(cardsFolder, versionedFile, locked); err != nil {
		fmt.Printf("[HARNESS] Warning: local card save failed: %v\n", err)
	}

	// Update queue status if this card is in the queue.
	updateQueueForCard(local, state.SeriesTitle, state.BookTitle, versionedFile, cardType, cardSubtype, cardName)

	fmt.Printf("\n[HARNESS] Card locked ✓  %s/%s\n", cardsFolder, versionedFile)
	return nil
}

func resolveCardFilename(cardType, cardSubtype, cardName, bookTitle string) (string, error) {
	switch cardType {
	case "CHAR", "CHARACTER":
		if cardName == "" {
			return "", fmt.Errorf("--name is required for CHARACTER cards")
		}
		return utils.CharCardFile(cardName, 1), nil
	case "WORLD":
		if cardSubtype == "" {
			return "", fmt.Errorf("--subtype is required for WORLD cards (e.g. Overview, History)")
		}
		if cardSubtype == "Geography" {
			if cardName == "" {
				return "", fmt.Errorf("--name is required for WORLD-Geography cards")
			}
			return utils.WorldGeographyFile(cardName, 1), nil
		}
		return utils.WorldCardFile(cardSubtype, 1), nil
	case "NOVEL":
		return utils.NovelCardFile(bookTitle, 1), nil
	case "TRILOGY":
		return utils.TrilogyFile(1), nil
	case "FACTION":
		if cardName == "" {
			return "", fmt.Errorf("--name is required for FACTION cards")
		}
		return utils.FactionCardFile(cardName, 1), nil
	case "THREAT":
		if cardName == "" {
			return "", fmt.Errorf("--name is required for THREAT cards")
		}
		return utils.ThreatCardFile(cardName, 1), nil
	case "CHAPTER":
		return "", fmt.Errorf("use 'harness write --chapter N' to draft chapter cards")
	default:
		return "", fmt.Errorf("unknown card type %q — try CHAR, WORLD, NOVEL, FACTION, THREAT, TRILOGY", cardType)
	}
}

func buildInstruction(cardType, cardSubtype, cardName string) string {
	label := cardType
	if cardSubtype != "" {
		label += " — " + cardSubtype
	}
	name := cardName
	if name == "" {
		name = cardType
	}
	sections := cards.RequiredSectionsList(cardType, cardSubtype)
	sectionsNote := ""
	if sections != "" {
		sectionsNote = "\n\nRequired sections (use these exact ## headings):\n" + sections
	}
	return fmt.Sprintf(
		"Build a complete %s card for: %s\n\n"+
			"Reference all locked cards and the expansion document. "+
			"Write complete, card-ready prose for every section — no placeholders.%s",
		label, name, sectionsNote,
	)
}

func versionedFilename(base string, version int) string {
	if version == 1 {
		return base
	}
	// Replace -v1.md with -v{n}.md.
	if idx := strings.LastIndex(base, "-v"); idx != -1 {
		return base[:idx] + fmt.Sprintf("-v%d.md", version)
	}
	return base
}

func updateQueueForCard(local *storage.Local, seriesTitle, bookTitle, filename, cardType, cardSubtype, cardName string) {
	queueData, err := local.Read(utils.QAFolder(seriesTitle, bookTitle), utils.CardQueueFile(bookTitle))
	if err != nil {
		return
	}
	q, err := queue.Unmarshal([]byte(queueData))
	if err != nil {
		return
	}
	for _, item := range q.Queue {
		if item.CardType == cardType && item.CardSubtype == cardSubtype &&
			strings.EqualFold(item.Name, cardName) && item.Status != queue.StatusComplete {
			_ = q.MarkComplete(item.Priority, filename)
			break
		}
	}
	data, err := q.Marshal()
	if err != nil {
		return
	}
	_ = local.Write(utils.QAFolder(seriesTitle, bookTitle), utils.CardQueueFile(bookTitle), string(data))
}
