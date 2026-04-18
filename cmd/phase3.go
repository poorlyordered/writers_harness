package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var phase3Cmd = &cobra.Command{
	Use:   "phase3",
	Short: "Phase 3 — Card Completion [not yet implemented]",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("Phase 3 not yet implemented — coming in next build iteration (SPEC-005)")
	},
}

func init() { rootCmd.AddCommand(phase3Cmd) }
