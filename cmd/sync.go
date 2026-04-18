package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/poorlyordered/writers_harness/internal/phases"
	"github.com/poorlyordered/writers_harness/internal/storage"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Push locally-written files to Box (re-syncs SYNC-PENDING log)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSync(context.Background())
	},
}

func init() { rootCmd.AddCommand(syncCmd) }

func runSync(ctx context.Context) error {
	sp, err := storage.LoadSyncPending(cfg.Local.ConfigFolder)
	if err != nil {
		return fmt.Errorf("loading sync-pending log: %w", err)
	}

	if len(sp.Entries) == 0 {
		fmt.Println("[HARNESS] Nothing pending — local and Box are in sync.")
		return nil
	}

	fmt.Printf("[HARNESS] %d file(s) pending sync to Box...\n", len(sp.Entries))

	local := storage.NewLocal(cfg.Local.SyncFolder)
	boxWriter := phases.NewBoxFileWriter(aiClient)

	failed := 0
	for _, entry := range sp.Entries {
		content, err := local.Read(entry.BoxPath, entry.Filename)
		if err != nil {
			fmt.Printf("[HARNESS] Warning: could not read local file %s/%s: %v\n",
				entry.BoxPath, entry.Filename, err)
			failed++
			continue
		}
		if err := boxWriter.Write(entry.BoxPath, entry.Filename, content); err != nil {
			fmt.Printf("[HARNESS] Warning: Box write failed for %s/%s: %v\n",
				entry.BoxPath, entry.Filename, err)
			failed++
			continue
		}
		fmt.Printf("[HARNESS] Synced: %s/%s\n", entry.BoxPath, entry.Filename)
	}

	if failed > 0 {
		return fmt.Errorf("sync completed with %d failure(s) — check warnings above", failed)
	}

	if err := sp.Clear(); err != nil {
		fmt.Printf("[HARNESS] Warning: could not clear sync-pending log: %v\n", err)
	}
	fmt.Printf("[HARNESS] Sync complete. %d file(s) pushed to Box.\n", len(sp.Entries))
	_ = ctx
	return nil
}
