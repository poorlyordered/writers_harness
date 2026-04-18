package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var phase2Cmd = &cobra.Command{
	Use:   "phase2",
	Short: "Phase 2 — Idea Expansion (Snowflake Method) [not yet implemented]",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("Phase 2 not yet implemented — coming in next build iteration (SPEC-003)")
	},
}

func init() { rootCmd.AddCommand(phase2Cmd) }
