package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var qaCheckCmd = &cobra.Command{
	Use:   "qa-check [1-4]",
	Short: "Run a QA checklist manually (1=phase gate, 2=card, 3=pre-prose, 4=post-prose)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		n, err := strconv.Atoi(args[0])
		if err != nil || n < 1 || n > 4 {
			return fmt.Errorf("argument must be 1, 2, 3, or 4")
		}
		if n == 1 {
			return fmt.Errorf("QA-1 manual run not yet fully wired — run 'harness phase1' to trigger it")
		}
		return fmt.Errorf("QA-%d manual run not yet implemented", n)
	},
}

func init() { rootCmd.AddCommand(qaCheckCmd) }
