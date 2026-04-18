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
	case "phase2":
		return runPhase2(ctx)
	case "phase3":
		return runPhase3(ctx)
	case "phase4":
		return runPhase4(ctx)
	default:
		return fmt.Errorf("unknown phase %q in saved session — use 'harness phase1' to restart", state.Phase)
	}
}
