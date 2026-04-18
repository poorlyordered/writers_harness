package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/poorlyordered/writers_harness/internal/phases"
	"github.com/poorlyordered/writers_harness/internal/session"
	"github.com/poorlyordered/writers_harness/internal/storage"
)

var (
	phase4FastDraft  bool
	phase4WordTarget int
	phase4Chapter    int
)

var phase4Cmd = &cobra.Command{
	Use:   "phase4",
	Short: "Phase 4 — Prose Generation (scene-by-scene drafting)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPhase4(context.Background())
	},
}

func init() {
	rootCmd.AddCommand(phase4Cmd)
	phase4Cmd.Flags().BoolVar(&phase4FastDraft, "fast", false, "Fast draft mode (draft full chapter then review)")
	phase4Cmd.Flags().IntVar(&phase4WordTarget, "words", 0, "Target words per scene (default: config preference)")
	phase4Cmd.Flags().IntVar(&phase4Chapter, "chapter", 0, "Resume from this chapter number (default: 1)")
}

func runPhase4(ctx context.Context) error {
	state, err := sessMgr.Load()
	if err != nil {
		return fmt.Errorf("no active session: %w\nRun 'harness new-series' then complete phases 1-3 first", err)
	}

	local := storage.NewLocal(cfg.Local.SyncFolder)

	wordTarget := phase4WordTarget
	if wordTarget == 0 {
		wordTarget = cfg.Preferences.SceneWordTargetDefault
	}
	if wordTarget == 0 {
		wordTarget = 2000
	}

	fastDraft := phase4FastDraft || cfg.Preferences.FastDraftDefault

	state.Phase = session.Phase4
	if err := sessMgr.Save(state); err != nil {
		return fmt.Errorf("saving session: %w", err)
	}

	pCfg := phases.Phase4Config{
		AI:              phases.NewAIConversation(aiClient),
		BoxWriter:       phases.NewBoxFileWriter(aiClient),
		LocalWriter:     phases.NewLocalFileWriter(local),
		LocalReader:     local,
		Prompter:        prompter,
		SeriesTitle:     state.SeriesTitle,
		BookTitle:       state.BookTitle,
		FastDraftMode:   fastDraft,
		SceneWordTarget: wordTarget,
		StartChapter:    phase4Chapter,
	}

	return phases.RunPhase4(ctx, pCfg)
}
