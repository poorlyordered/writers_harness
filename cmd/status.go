package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/poorlyordered/writers_harness/internal/qa"
	"github.com/poorlyordered/writers_harness/internal/queue"
	"github.com/poorlyordered/writers_harness/internal/storage"
	"github.com/poorlyordered/writers_harness/internal/utils"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show active session and queue progress",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStatus()
	},
}

func init() { rootCmd.AddCommand(statusCmd) }

func runStatus() error {
	state, err := sessMgr.Load()
	if err != nil {
		return fmt.Errorf("no active session: %w\nRun 'harness new-series' to start", err)
	}

	fmt.Printf("\n═══════════════════════════════════════════════════════════\n")
	fmt.Printf("  WRITING HARNESS — STATUS\n")
	fmt.Printf("═══════════════════════════════════════════════════════════\n")
	fmt.Printf("  Series:  %s\n", state.SeriesTitle)
	fmt.Printf("  Book:    %s\n", state.BookTitle)
	fmt.Printf("  Phase:   %s\n", state.Phase)
	fmt.Printf("  Started: %s\n", state.StartedAt.Format("2006-01-02"))
	fmt.Printf("═══════════════════════════════════════════════════════════\n\n")

	local := storage.NewLocal(cfg.Local.SyncFolder)

	// Card Queue progress (Phase 3 and beyond).
	queueData, queueErr := local.Read(
		utils.QAFolder(state.SeriesTitle, state.BookTitle),
		utils.CardQueueFile(state.BookTitle),
	)
	if queueErr == nil {
		if q, err := queue.Unmarshal([]byte(queueData)); err == nil {
			p := q.Progress
			fmt.Printf("Card Queue (%s)\n", q.PhaseGateStatus)
			fmt.Printf("  %d/%d complete  |  %d remaining  |  %.0f%%\n\n",
				p.Complete, p.TotalCards, p.NotStarted, p.PercentComplete)
			if next := q.Next(); next != nil {
				tier := ""
				if next.Tier != nil {
					tier = " [" + *next.Tier + "]"
				}
				fmt.Printf("Next card:  #%d  %s%s — %s\n\n",
					next.Priority, next.CardType, tier, next.Name)
			} else if p.NotStarted == 0 {
				fmt.Println("All cards complete. Run 'harness phase4' to begin prose generation.")
			}
		}
	}

	// Locked card file count from local sync.
	if files, err := local.ListFiles(utils.CardsFolder(state.SeriesTitle, state.BookTitle), ""); err == nil && len(files) > 0 {
		fmt.Printf("Locked cards (local):  %d file(s)\n", len(files))
	}

	// QA Log summary.
	qaLogPath := local.AbsPath(
		utils.QAFolder(state.SeriesTitle, state.BookTitle),
		utils.QALogFile(state.BookTitle),
	)
	qaLog := qa.NewLog(qaLogPath)
	if pass, fail, err := qaLog.Summary(); err == nil && pass+fail > 0 {
		fmt.Printf("QA Log:                %d pass / %d fail (%d total runs)\n", pass, fail, pass+fail)
	}

	fmt.Println()
	return nil
}
