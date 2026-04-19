package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/poorlyordered/writers_harness/internal/phases"
	"github.com/poorlyordered/writers_harness/internal/storage"
)

var (
	writeChapter int
	writeScene   int
	writeWords   int
)

var writeCmd = &cobra.Command{
	Use:   "write",
	Short: "Prose Generation — draft a specific chapter or scene",
	Long: `Draft prose for a specific chapter or scene, loading all relevant context cards.

Examples:
  harness write --chapter 5            # draft all scenes in chapter 5
  harness write --chapter 5 --scene 2  # draft only scene 2 of chapter 5`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWrite(context.Background())
	},
}

func init() {
	writeCmd.Flags().IntVar(&writeChapter, "chapter", 0, "Chapter number to draft (required)")
	writeCmd.Flags().IntVar(&writeScene, "scene", 0, "Scene number within the chapter (0 = all scenes)")
	writeCmd.Flags().IntVar(&writeWords, "words", 0, "Target word count per scene (default from config)")
	_ = writeCmd.MarkFlagRequired("chapter")
	rootCmd.AddCommand(writeCmd)
}

func runWrite(ctx context.Context) error {
	state, err := sessMgr.Load()
	if err != nil {
		return fmt.Errorf("no active session: %w", err)
	}
	if writeChapter < 1 || writeChapter > 40 {
		return fmt.Errorf("--chapter must be between 1 and 40")
	}

	local := storage.NewLocal(cfg.Local.SyncFolder)

	wordTarget := writeWords
	if wordTarget == 0 {
		wordTarget = cfg.Preferences.SceneWordTargetDefault
	}

	return phases.RunWrite(ctx, phases.WriteConfig{
		AI:              phases.NewAIConversation(aiClient),
		BoxWriter:       phases.NewBoxFileWriter(aiClient),
		LocalWriter:     phases.NewLocalFileWriter(local),
		LocalReader:     local,
		Prompter:        prompter,
		SeriesTitle:     state.SeriesTitle,
		BookTitle:       state.BookTitle,
		ChapterNum:      writeChapter,
		SceneNum:        writeScene,
		SceneWordTarget: wordTarget,
	})
}
