package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the Series Index summary for the active series",
	RunE: func(cmd *cobra.Command, args []string) error {
		state, err := sessMgr.Load()
		if err != nil {
			return fmt.Errorf("no active session: %w", err)
		}
		fmt.Printf("Series: %s\nBook:   %s\nPhase:  %s\n",
			state.SeriesTitle, state.BookTitle, state.Phase)
		fmt.Println("\n[Full Series Index display coming in next build iteration]")
		return nil
	},
}

func init() { rootCmd.AddCommand(statusCmd) }
