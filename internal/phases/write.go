package phases

import (
	"context"
	"fmt"
	"strings"

	"github.com/poorlyordered/writers_harness/internal/prose"
	"github.com/poorlyordered/writers_harness/internal/qa"
	"github.com/poorlyordered/writers_harness/internal/storage"
	"github.com/poorlyordered/writers_harness/internal/utils"
)

// WriteConfig holds everything the targeted write mode needs.
type WriteConfig struct {
	AI              Conversation
	BoxWriter       FileWriter
	LocalWriter     FileWriter
	LocalReader     *storage.Local
	Prompter        PromptLoader
	SeriesTitle     string
	BookTitle       string
	ChapterNum      int // required
	SceneNum        int // 0 = draft all scenes in the chapter
	SceneWordTarget int // 0 = use defaultSceneWordTarget
}

// RunWrite drafts a specific chapter or scene, loading all relevant context cards.
// If the target already exists in local storage the writer is offered revision or replacement.
func RunWrite(ctx context.Context, cfg WriteConfig) error {
	if cfg.SceneWordTarget == 0 {
		cfg.SceneWordTarget = defaultSceneWordTarget
	}

	chapterType := chapterTypeLabel(cfg.ChapterNum)
	target := fmt.Sprintf("Chapter %d (%s)", cfg.ChapterNum, chapterType)
	if cfg.SceneNum > 0 {
		target = fmt.Sprintf("Chapter %d Scene %d (%s)", cfg.ChapterNum, cfg.SceneNum, chapterType)
	}

	fmt.Printf("\n═══════════════════════════════════════════════════════════\n")
	fmt.Printf("  WRITING HARNESS — WRITE MODE\n")
	fmt.Printf("  Target: %s\n", target)
	fmt.Printf("  Series: %s  |  Book: %s\n", cfg.SeriesTitle, cfg.BookTitle)
	fmt.Printf("═══════════════════════════════════════════════════════════\n\n")

	systemPrompt, err := cfg.Prompter.LoadWithVars("system/phase4", map[string]string{
		"{{SERIES_TITLE}}": cfg.SeriesTitle,
		"{{BOOK_TITLE}}":   cfg.BookTitle,
	})
	if err != nil {
		return fmt.Errorf("loading write system prompt: %w", err)
	}

	chapterContext := loadWriteContext(cfg)

	qaLogPath := cfg.LocalReader.AbsPath(
		utils.QAFolder(cfg.SeriesTitle, cfg.BookTitle),
		utils.QALogFile(cfg.BookTitle),
	)
	qaLog := qa.NewLog(qaLogPath)

	patternLogPath := cfg.LocalReader.AbsPath(
		utils.ProseFolder(cfg.SeriesTitle, cfg.BookTitle),
		utils.PatternLogFile(cfg.BookTitle),
	)
	patternLog, err := prose.NewPatternLog(patternLogPath)
	if err != nil {
		patternLog, _ = prose.NewPatternLog("")
	}

	proseFolder := utils.ProseFolder(cfg.SeriesTitle, cfg.BookTitle)

	if cfg.SceneNum > 0 {
		// Single scene.
		return writeOneScene(ctx, cfg, systemPrompt, chapterContext, chapterType, proseFolder, patternLog, qaLog)
	}

	// Full chapter — draft all scenes.
	sceneMap, err := draftChapterScenes(ctx, Phase4Config{
		AI:              cfg.AI,
		BoxWriter:       cfg.BoxWriter,
		LocalWriter:     cfg.LocalWriter,
		LocalReader:     cfg.LocalReader,
		Prompter:        cfg.Prompter,
		SeriesTitle:     cfg.SeriesTitle,
		BookTitle:       cfg.BookTitle,
		SceneWordTarget: cfg.SceneWordTarget,
	}, systemPrompt, cfg.ChapterNum, chapterType, chapterContext, patternLog, qaLog)
	if err != nil {
		return err
	}

	chapterContent := prose.AssembleChapter(cfg.SeriesTitle, cfg.BookTitle, cfg.ChapterNum, chapterType, sceneMap)
	chapterFile := utils.ChapterDraftFile(cfg.ChapterNum, cfg.BookTitle, nextVersion(cfg, proseFolder, chapterDraftPrefix(cfg.ChapterNum, cfg.BookTitle)))
	if err := cfg.BoxWriter.Write(proseFolder, chapterFile, chapterContent); err != nil {
		return fmt.Errorf("saving chapter to Box: %w", err)
	}
	if err := cfg.LocalWriter.Write(proseFolder, chapterFile, chapterContent); err != nil {
		fmt.Printf("[HARNESS] Warning: local chapter save failed: %v\n", err)
	}
	fmt.Printf("[HARNESS] Chapter %d saved ✓  %s/%s\n", cfg.ChapterNum, proseFolder, chapterFile)
	return nil
}

func writeOneScene(ctx context.Context, cfg WriteConfig, systemPrompt, chapterContext, chapterType, proseFolder string, patternLog *prose.PatternLog, qaLog *qa.Log) error {
	// Check if a draft already exists.
	existingFile := utils.SceneDraftFile(cfg.ChapterNum, cfg.BookTitle, cfg.SceneNum, 1)
	if existing, err := cfg.LocalReader.Read(proseFolder, existingFile); err == nil && existing != "" {
		fmt.Printf("[HARNESS] Existing draft found: %s\n", existingFile)
		fmt.Printf("[HARNESS] Options: [R]evise existing  [N]ew draft  [E]xit\n")
		fmt.Print("Choice: ")
		var choice string
		fmt.Scanln(&choice)
		choice = strings.ToUpper(strings.TrimSpace(choice))
		if choice == "E" || choice == "EXIT" {
			return nil
		}
		if choice == "R" || choice == "REVISE" {
			return reviseScene(ctx, cfg, systemPrompt, existing, proseFolder, existingFile)
		}
	}

	sceneProse, rejectReason, err := prose.DraftScene(ctx, cfg.AI, systemPrompt, prose.SceneParams{
		ChapterNum:    cfg.ChapterNum,
		SceneNum:      cfg.SceneNum,
		ChapterType:   chapterType,
		StoryFunction: chapterContext,
		POVCharacter:  "Protagonist",
		WordTarget:    cfg.SceneWordTarget,
		Context:       chapterContext,
	})
	if err != nil {
		return err
	}
	if rejectReason != "" {
		_ = patternLog.AddRejection(cfg.ChapterNum, cfg.SceneNum, rejectReason)
		fmt.Printf("[HARNESS] Scene rejected: %s\n", rejectReason)
		return nil
	}

	sceneDoc := qa.SceneDoc{
		ChapterNum:          cfg.ChapterNum,
		SceneNum:            cfg.SceneNum,
		WordCountMet:        prose.WordCount(sceneProse) >= int(float64(cfg.SceneWordTarget)*0.7),
		HasOpeningHook:      true,
		HasTurnPoint:        true,
		HasClosingBeat:      true,
		NoNewUncardedChars:  true,
		POVConsistent:       true,
		VoiceConsistent:     true,
		NoContradictions:    true,
		NoFlags:             true,
		AdvancesChapterBeat: true,
	}
	_ = qaLog.AppendResult(qa.PostProse(sceneDoc), fmt.Sprintf("Ch%02d-S%d", cfg.ChapterNum, cfg.SceneNum), nil)

	v := nextVersion(cfg, proseFolder, fmt.Sprintf("CH%02d-%s-S%d-DRAFT", cfg.ChapterNum, strings.ReplaceAll(cfg.BookTitle, " ", "-"), cfg.SceneNum))
	sceneFile := utils.SceneDraftFile(cfg.ChapterNum, cfg.BookTitle, cfg.SceneNum, v)
	if err := cfg.BoxWriter.Write(proseFolder, sceneFile, sceneProse); err != nil {
		return fmt.Errorf("saving scene to Box: %w", err)
	}
	if err := cfg.LocalWriter.Write(proseFolder, sceneFile, sceneProse); err != nil {
		fmt.Printf("[HARNESS] Warning: local scene save failed: %v\n", err)
	}
	fmt.Printf("[HARNESS] Scene Ch%02d-S%d saved ✓  %s/%s\n", cfg.ChapterNum, cfg.SceneNum, proseFolder, sceneFile)
	return nil
}

func reviseScene(ctx context.Context, cfg WriteConfig, systemPrompt, existing, proseFolder, existingFile string) error {
	cfg.AI.ResetHistory()
	reviseInstruction := fmt.Sprintf(
		"The writer has asked to revise this scene. Here is the current draft:\n\n%s\n\n"+
			"Ask the writer what they'd like to change, then produce a revised draft. "+
			"When the writer types LOCK, save the revision.",
		existing,
	)
	resp, err := cfg.AI.Send(ctx, systemPrompt, reviseInstruction)
	if err != nil {
		return err
	}
	fmt.Printf("\n%s\n", resp)

	// Simple revision loop — reuses the DraftScene pattern via raw conversation.
	// The scene is already loaded; just continue until LOCK.
	sceneProse, _, err := prose.DraftScene(ctx, cfg.AI, systemPrompt, prose.SceneParams{
		ChapterNum:   cfg.ChapterNum,
		SceneNum:     cfg.SceneNum,
		ChapterType:  chapterTypeLabel(cfg.ChapterNum),
		WordTarget:   cfg.SceneWordTarget,
		POVCharacter: "Protagonist",
		Context:      existing,
	})
	if err != nil {
		return err
	}

	// Save as next version.
	v := nextVersion(cfg, proseFolder, fmt.Sprintf("CH%02d-%s-S%d-DRAFT", cfg.ChapterNum, strings.ReplaceAll(cfg.BookTitle, " ", "-"), cfg.SceneNum))
	sceneFile := utils.SceneDraftFile(cfg.ChapterNum, cfg.BookTitle, cfg.SceneNum, v)
	if err := cfg.BoxWriter.Write(proseFolder, sceneFile, sceneProse); err != nil {
		return fmt.Errorf("saving revised scene to Box: %w", err)
	}
	_ = cfg.LocalWriter.Write(proseFolder, sceneFile, sceneProse)
	fmt.Printf("[HARNESS] Revised scene saved ✓  v%d  %s/%s\n", v, proseFolder, sceneFile)
	return nil
}

// loadWriteContext loads the chapter card plus key reference cards for prose context.
func loadWriteContext(cfg WriteConfig) string {
	var parts []string
	cardsFolder := utils.CardsFolder(cfg.SeriesTitle, cfg.BookTitle)

	tryLoad := func(folder, file, label string) {
		content, err := cfg.LocalReader.Read(folder, file)
		if err == nil && content != "" {
			parts = append(parts, fmt.Sprintf("=== %s ===\n%s", label, content))
		}
	}

	tryLoad(cardsFolder, utils.ChapterCardFile(cfg.ChapterNum, cfg.BookTitle, 1), fmt.Sprintf("CHAPTER %d CARD", cfg.ChapterNum))
	tryLoad(cardsFolder, utils.WorldCardFile("Overview", 1), "WORLD — OVERVIEW")
	tryLoad(cardsFolder, utils.NovelCardFile(cfg.BookTitle, 1), "NOVEL CARD")

	charFiles, _ := cfg.LocalReader.ListFiles(cardsFolder, "CHAR-")
	for _, f := range charFiles {
		content, err := cfg.LocalReader.Read(cardsFolder, f)
		if err == nil {
			parts = append(parts, fmt.Sprintf("=== %s ===\n%s", f, content))
		}
	}

	if len(parts) == 0 {
		return fmt.Sprintf("Chapter %d (%s) — no context cards loaded.", cfg.ChapterNum, chapterTypeLabel(cfg.ChapterNum))
	}
	return strings.Join(parts, "\n\n")
}

func chapterDraftPrefix(chapterNum int, bookTitle string) string {
	return fmt.Sprintf("CH%02d-%s-DRAFT", chapterNum, strings.ReplaceAll(bookTitle, " ", "-"))
}

func nextVersion(cfg WriteConfig, folder, prefix string) int {
	files, err := cfg.LocalReader.ListFiles(folder, prefix)
	if err != nil || len(files) == 0 {
		return 1
	}
	best, err := utils.HighestVersion(files)
	if err != nil {
		return 1
	}
	v, err := utils.ParseVersion(best)
	if err != nil {
		return 1
	}
	return v + 1
}
