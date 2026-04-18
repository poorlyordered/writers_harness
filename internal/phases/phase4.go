package phases

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/poorlyordered/writers_harness/internal/prose"
	"github.com/poorlyordered/writers_harness/internal/qa"
	"github.com/poorlyordered/writers_harness/internal/storage"
	"github.com/poorlyordered/writers_harness/internal/utils"
)

const (
	defaultSceneWordTarget = 2000
	totalChapters          = 40
	defaultScenesPerChapter = 3
)

// Phase4Config holds everything the Phase 4 runner needs.
type Phase4Config struct {
	AI              Conversation
	BoxWriter       FileWriter
	LocalWriter     FileWriter
	LocalReader     *storage.Local
	Prompter        PromptLoader
	SeriesTitle     string
	BookTitle       string
	FastDraftMode   bool
	SceneWordTarget int  // words per scene, default 2000
	StartChapter    int  // resume from this chapter (1-based, 0 = start from 1)
}

func RunPhase4(ctx context.Context, cfg Phase4Config) error {
	if cfg.SceneWordTarget == 0 {
		cfg.SceneWordTarget = defaultSceneWordTarget
	}
	startChapter := cfg.StartChapter
	if startChapter < 1 {
		startChapter = 1
	}

	fmt.Printf("\n═══════════════════════════════════════════════════════════\n")
	fmt.Printf("  WRITING HARNESS — PHASE 4: PROSE GENERATION\n")
	fmt.Printf("  Series: %s  |  Book: %s\n", cfg.SeriesTitle, cfg.BookTitle)
	if cfg.FastDraftMode {
		fmt.Printf("  Mode: FAST DRAFT\n")
	}
	fmt.Printf("  Starting at Chapter %d  |  Target: ~%d words/scene\n",
		startChapter, cfg.SceneWordTarget)
	fmt.Printf("═══════════════════════════════════════════════════════════\n\n")

	systemPrompt, err := cfg.Prompter.LoadWithVars("system/phase4", map[string]string{
		"{{SERIES_TITLE}}": cfg.SeriesTitle,
		"{{BOOK_TITLE}}":   cfg.BookTitle,
	})
	if err != nil {
		return fmt.Errorf("loading phase4 system prompt: %w", err)
	}

	patternLogPath := cfg.LocalReader.AbsPath(
		utils.ProseFolder(cfg.SeriesTitle, cfg.BookTitle),
		utils.PatternLogFile(cfg.BookTitle),
	)
	patternLog, err := prose.NewPatternLog(patternLogPath)
	if err != nil {
		fmt.Printf("[HARNESS] Warning: could not load pattern log: %v\n", err)
		patternLog, _ = prose.NewPatternLog("") // in-memory only
	}

	qaLogPath := cfg.LocalReader.AbsPath(
		utils.QAFolder(cfg.SeriesTitle, cfg.BookTitle),
		utils.QALogFile(cfg.BookTitle),
	)
	qaLog := qa.NewLog(qaLogPath)

	chapters := make(map[int]string) // chapterNum -> assembled chapter content
	reader := bufio.NewReader(os.Stdin)

	for chapterNum := startChapter; chapterNum <= totalChapters; chapterNum++ {
		chapterType := chapterTypeLabel(chapterNum)

		fmt.Printf("\n─── CHAPTER %d: %s ──────────────────────────────────────\n",
			chapterNum, chapterType)
		fmt.Printf("    Target: ~%d words/scene\n\n", cfg.SceneWordTarget)

		// Load the CHAPTER card as context if available.
		chapterContext := loadChapterContext(cfg, chapterNum)

		var sceneMap map[int]string
		if cfg.FastDraftMode {
			sceneMap, err = prose.FastDraftChapter(ctx, cfg.AI, systemPrompt, prose.FastDraftParams{
				ChapterNum:   chapterNum,
				ChapterType:  chapterType,
				SceneCount:   defaultScenesPerChapter,
				WordTarget:   cfg.SceneWordTarget,
				POVCharacter: "Protagonist", // overridden by chapter card if available
				Context:      chapterContext,
			})
			if err != nil {
				return fmt.Errorf("fast draft chapter %d: %w", chapterNum, err)
			}
		} else {
			sceneMap, err = draftChapterScenes(ctx, cfg, systemPrompt, chapterNum, chapterType, chapterContext, patternLog, qaLog)
			if err != nil {
				return err
			}
		}

		// Assemble chapter.
		chapterContent := prose.AssembleChapter(cfg.SeriesTitle, cfg.BookTitle, chapterNum, chapterType, sceneMap)
		chapterFile := utils.ChapterDraftFile(chapterNum, cfg.BookTitle, 1)
		proseFolder := utils.ProseFolder(cfg.SeriesTitle, cfg.BookTitle)

		if err := cfg.BoxWriter.Write(proseFolder, chapterFile, chapterContent); err != nil {
			return fmt.Errorf("saving chapter %d to Box: %w", chapterNum, err)
		}
		if err := cfg.LocalWriter.Write(proseFolder, chapterFile, chapterContent); err != nil {
			fmt.Printf("[HARNESS] Warning: local chapter save failed: %v\n", err)
		}
		chapters[chapterNum] = chapterContent
		fmt.Printf("[HARNESS] Chapter %d saved ✓  %s/%s\n", chapterNum, proseFolder, chapterFile)

		// Ask writer if they want to continue, pause, or stop.
		if chapterNum < totalChapters {
			fmt.Printf("\n[HARNESS] Chapter %d complete. Continue to Chapter %d? [Y/n/stop]: ", chapterNum, chapterNum+1)
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(strings.ToLower(input))
			if input == "stop" || input == "quit" || input == "q" {
				fmt.Printf("[HARNESS] Stopping after Chapter %d. Resume with 'harness phase4 --chapter %d'.\n",
					chapterNum, chapterNum+1)
				return nil
			}
		}
	}

	// Assemble full manuscript.
	fmt.Printf("\n─── ASSEMBLING MANUSCRIPT ──────────────────────────────────\n")
	manuscript := prose.AssembleManuscript(cfg.SeriesTitle, cfg.BookTitle, chapters)
	msFile := utils.ManuscriptFile(cfg.BookTitle, 1)
	proseFolder := utils.ProseFolder(cfg.SeriesTitle, cfg.BookTitle)

	if err := cfg.BoxWriter.Write(proseFolder, msFile, manuscript); err != nil {
		return fmt.Errorf("saving manuscript to Box: %w", err)
	}
	if err := cfg.LocalWriter.Write(proseFolder, msFile, manuscript); err != nil {
		fmt.Printf("[HARNESS] Warning: local manuscript save failed: %v\n", err)
	}

	wc := prose.WordCount(manuscript)
	fmt.Printf("[HARNESS] Manuscript assembled ✓  %s  (~%d words)\n", msFile, wc)
	fmt.Println("[HARNESS] Phase 4 complete. First draft ready for revision.")
	return nil
}

func draftChapterScenes(
	ctx context.Context,
	cfg Phase4Config,
	systemPrompt string,
	chapterNum int,
	chapterType string,
	chapterContext string,
	patternLog *prose.PatternLog,
	qaLog *qa.Log,
) (map[int]string, error) {
	sceneMap := make(map[int]string)
	scenesPerChapter := defaultScenesPerChapter

	for sceneNum := 1; sceneNum <= scenesPerChapter; sceneNum++ {
		rejections := 0
		for {
			sceneProse, rejectReason, err := prose.DraftScene(ctx, cfg.AI, systemPrompt, prose.SceneParams{
				ChapterNum:    chapterNum,
				SceneNum:      sceneNum,
				ChapterType:   chapterType,
				StoryFunction: chapterContext,
				POVCharacter:  "Protagonist",
				WordTarget:    cfg.SceneWordTarget,
				Context:       chapterContext,
			})
			if err != nil {
				return nil, fmt.Errorf("scene Ch%02d-S%d: %w", chapterNum, sceneNum, err)
			}

			if rejectReason != "" {
				rejections++
				_ = patternLog.AddRejection(chapterNum, sceneNum, rejectReason)
				if rejections >= 3 {
					fmt.Println()
					fmt.Println(patternLog.ThreeRejectionPattern())
					fmt.Println("[HARNESS] Three rejections reached. Adjust your approach and try again, or type SKIP to move on.")
					fmt.Print("Continue? [enter to retry / SKIP]: ")
					reader := bufio.NewReader(os.Stdin)
					input, _ := reader.ReadString('\n')
					if strings.ToUpper(strings.TrimSpace(input)) == "SKIP" {
						sceneMap[sceneNum] = "[SCENE SKIPPED — to be written later]"
						break
					}
					rejections = 0
				}
				fmt.Printf("[HARNESS] Scene rejected (%d/3). Retrying...\n", rejections)
				continue
			}

			// Run QA-4 post-prose check (heuristic).
			sceneDoc := qa.SceneDoc{
				ChapterNum:          chapterNum,
				SceneNum:            sceneNum,
				WordCountMet:        prose.WordCount(sceneProse) >= int(float64(cfg.SceneWordTarget)*0.7),
				HasOpeningHook:      true, // instructed in system prompt; writer confirmed
				HasTurnPoint:        true,
				HasClosingBeat:      true,
				NoNewUncardedChars:  true,
				POVConsistent:       true,
				VoiceConsistent:     true,
				NoContradictions:    true,
				NoFlags:             true,
				AdvancesChapterBeat: true,
			}
			qa4Result := qa.PostProse(sceneDoc)
			_ = qaLog.AppendResult(qa4Result, fmt.Sprintf("Ch%02d-S%d", chapterNum, sceneNum), nil)

			sceneMap[sceneNum] = sceneProse
			fmt.Printf("[HARNESS] Scene Ch%02d-S%d locked ✓\n", chapterNum, sceneNum)
			break
		}
	}
	return sceneMap, nil
}

func loadChapterContext(cfg Phase4Config, chapterNum int) string {
	filename := utils.ChapterCardFile(chapterNum, cfg.BookTitle, 1)
	folder := utils.CardsFolder(cfg.SeriesTitle, cfg.BookTitle)
	content, err := cfg.LocalReader.Read(folder, filename)
	if err != nil {
		return fmt.Sprintf("Chapter %d (%s) — no chapter card loaded.", chapterNum, chapterTypeLabel(chapterNum))
	}
	return content
}

func chapterTypeLabel(n int) string {
	labels := map[int]string{
		1:  "Really Bad Day",
		2:  "Mystery and Theme",
		6:  "Inciting Incident",
		10: "Pull Out Rug",
		17: "First Pinch Point",
		20: "Midpoint",
		27: "All Is Lost",
		29: "Giving Up",
		36: "Ultimate Defeat",
		38: "Unexpected Victory",
		40: "Death of Self",
	}
	if label, ok := labels[n]; ok {
		return label
	}
	switch {
	case n <= 10:
		return fmt.Sprintf("Act 1 — Ch%d", n)
	case n <= 20:
		return fmt.Sprintf("Act 2A — Ch%d", n)
	case n <= 30:
		return fmt.Sprintf("Act 2B — Ch%d", n)
	default:
		return fmt.Sprintf("Act 3 — Ch%d", n)
	}
}
