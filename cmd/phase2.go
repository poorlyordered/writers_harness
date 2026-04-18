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

var phase2Cmd = &cobra.Command{
	Use:   "phase2",
	Short: "Phase 2 — Idea Expansion (Snowflake Method)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPhase2(context.Background())
	},
}

func init() { rootCmd.AddCommand(phase2Cmd) }

func runPhase2(ctx context.Context) error {
	state, err := sessMgr.Load()
	if err != nil {
		return fmt.Errorf("no active session: %w\nRun 'harness new-series' then 'harness phase1' first", err)
	}

	// Load locked seed from local sync folder. Warn but continue if missing —
	// Box MCP gives the AI direct access to the seed anyway.
	seedFolder := utils.SeedsFolder(state.SeriesTitle, state.BookTitle)
	seedFile := utils.SeedFile(state.BookTitle, 1)
	local := storage.NewLocal(cfg.Local.SyncFolder)
	seedContent, err := local.Read(seedFolder, seedFile)
	if err != nil {
		fmt.Printf("[HARNESS] Warning: could not read local seed (%v); proceeding without local copy.\n", err)
		seedContent = ""
	}

	state.Phase = session.Phase2
	if err := sessMgr.Save(state); err != nil {
		return fmt.Errorf("saving session: %w", err)
	}

	pCfg := phases.Phase2Config{
		AI:           phases.NewAIConversation(aiClient),
		BoxWriter:    phases.NewBoxFileWriter(aiClient),
		LocalWriter:  phases.NewLocalFileWriter(storage.NewLocal(cfg.Local.SyncFolder)),
		Prompter:     prompter,
		SeriesTitle:  state.SeriesTitle,
		BookTitle:    state.BookTitle,
		IsStandalone: cfg.Defaults.Mode == "standalone",
		SeedContent:  seedContent,
	}

	return phases.RunPhase2(ctx, pCfg)
}
