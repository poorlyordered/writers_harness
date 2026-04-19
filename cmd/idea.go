package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/poorlyordered/writers_harness/internal/phases"
	"github.com/poorlyordered/writers_harness/internal/storage"
)

var ideaCmd = &cobra.Command{
	Use:   "idea",
	Short: "Idea Generation — brainstorm freely; nothing commits to canon unless you type SAVE",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runIdea(context.Background())
	},
}

func init() { rootCmd.AddCommand(ideaCmd) }

func runIdea(ctx context.Context) error {
	state, err := sessMgr.Load()
	if err != nil {
		return fmt.Errorf("no active session: %w\nRun 'harness new-series' first", err)
	}

	local := storage.NewLocal(cfg.Local.SyncFolder)

	return phases.RunIdea(ctx, phases.IdeaConfig{
		AI:          phases.NewAIConversation(aiClient),
		LocalWriter: phases.NewLocalFileWriter(local),
		LocalReader: local,
		Prompter:    prompter,
		SeriesTitle: state.SeriesTitle,
		BookTitle:   state.BookTitle,
	})
}
