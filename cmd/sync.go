package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Manually trigger Box ↔ local sync",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("manual sync not yet implemented — coming in next build iteration")
	},
}

func init() { rootCmd.AddCommand(syncCmd) }
