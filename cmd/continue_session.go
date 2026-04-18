package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var continueCmd = &cobra.Command{
	Use:   "continue",
	Short: "Continue the most recent session",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runContinue(context.Background())
	},
}

func init() { rootCmd.AddCommand(continueCmd) }

func runContinue(ctx context.Context) error {
	state, err := sessMgr.Load()
	if err != nil {
		return fmt.Errorf("no saved session: %w\nRun 'harness new-series' to start", err)
	}
	fmt.Printf("[HARNESS] Resuming %s session — Series: %q  Book: %q\n",
		state.Phase, state.SeriesTitle, state.BookTitle)
	switch state.Phase {
	case "phase1":
		return runPhase1(ctx)
	case "phase2", "phase3", "phase4":
		return fmt.Errorf("phase %s resume not yet implemented", state.Phase)
	default:
		return fmt.Errorf("unknown phase %q in saved session", state.Phase)
	}
}
