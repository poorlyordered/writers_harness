package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/poorlyordered/writers_harness/internal/phases"
	"github.com/poorlyordered/writers_harness/internal/session"
	"github.com/poorlyordered/writers_harness/internal/storage"
)

var phase1Cmd = &cobra.Command{
	Use:   "phase1",
	Short: "Phase 1 — Idea Generation (seed prompt and deepening conversation)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPhase1(context.Background())
	},
}

func init() {
	rootCmd.AddCommand(phase1Cmd)
}

func runPhase1(ctx context.Context) error {
	state, err := sessMgr.Load()
	if err != nil {
		return fmt.Errorf("no active session: %w\nRun 'harness new-series' first", err)
	}

	state.Phase = session.Phase1
	if err := sessMgr.Save(state); err != nil {
		return fmt.Errorf("saving session: %w", err)
	}

	pCfg := phases.Phase1Config{
		AI:           phases.NewAIConversation(aiClient),
		BoxWriter:    phases.NewBoxFileWriter(aiClient),
		LocalWriter:  phases.NewLocalFileWriter(storage.NewLocal(cfg.Local.SyncFolder)),
		Prompter:     prompter,
		SeriesTitle:  state.SeriesTitle,
		BookTitle:    state.BookTitle,
		IsStandalone: cfg.Defaults.Mode == "standalone",
		DefaultGenre: cfg.Defaults.SettingGenre,
	}

	return phases.Run(ctx, pCfg)
}
