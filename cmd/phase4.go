package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var phase4Cmd = &cobra.Command{
	Use:   "phase4",
	Short: "Phase 4 — Prose Generation [not yet implemented]",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("Phase 4 not yet implemented — coming in next build iteration (SPEC-006)")
	},
}

func init() { rootCmd.AddCommand(phase4Cmd) }
