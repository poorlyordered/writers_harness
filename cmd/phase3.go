package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/poorlyordered/writers_harness/internal/phases"
	"github.com/poorlyordered/writers_harness/internal/session"
	"github.com/poorlyordered/writers_harness/internal/storage"
	"github.com/poorlyordered/writers_harness/internal/utils"
)

var phase3Cmd = &cobra.Command{
	Use:   "phase3",
	Short: "Phase 3 — Card Completion (queue-driven)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPhase3(context.Background())
	},
}

func init() { rootCmd.AddCommand(phase3Cmd) }

func runPhase3(ctx context.Context) error {
	state, err := sessMgr.Load()
	if err != nil {
		return fmt.Errorf("no active session: %w\nRun 'harness new-series' then 'harness phase2' first", err)
	}

	local := storage.NewLocal(cfg.Local.SyncFolder)

	// Load expansion doc for AI context injection.
	expandContent, err := local.Read(
		utils.ExpansionFolder(state.SeriesTitle, state.BookTitle),
		utils.ExpansionFile(state.BookTitle, 1),
	)
	if err != nil {
		fmt.Printf("[HARNESS] Warning: expansion doc not found locally (%v); continuing without it.\n", err)
		expandContent = ""
	}

	state.Phase = session.Phase3
	if err := sessMgr.Save(state); err != nil {
		return fmt.Errorf("saving session: %w", err)
	}

	// Load existing index state so all prior card entries are preserved.
	idx := loadIndexState(local, state.SeriesTitle)
	book := idx.EnsureBook(state.BookTitle, state.BookNum)
	book.Phase = "3"

	pCfg := phases.Phase3Config{
		AI:               phases.NewAIConversation(aiClient),
		BoxWriter:        phases.NewBoxFileWriter(aiClient),
		LocalWriter:      phases.NewLocalFileWriter(local),
		LocalReader:      local,
		Prompter:         prompter,
		SeriesTitle:      state.SeriesTitle,
		BookTitle:        state.BookTitle,
		BookNum:          state.BookNum,
		IsStandalone:     cfg.Defaults.Mode == "standalone",
		ExpansionContent: expandContent,
		Idx:              idx,
	}

	return phases.RunPhase3(ctx, pCfg)
}
