package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/poorlyordered/writers_harness/internal/phases"
	"github.com/poorlyordered/writers_harness/internal/storage"
	"github.com/poorlyordered/writers_harness/internal/utils"
)

var checkCmd = &cobra.Command{
	Use:   "check <filename>",
	Short: "Consistency check — compare an edited file against locked canon",
	Long: `After editing a card or prose file outside the harness, run a consistency
check to surface contradictions, broken references, and voice drift.

The file is looked up in the local sync folder for the active session.

Examples:
  harness check CHAR-Jax-Tarkin-v2.md
  harness check CH05-Into-the-Dark-DRAFT-v1.md`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCheck(context.Background(), args[0])
	},
}

func init() { rootCmd.AddCommand(checkCmd) }

func runCheck(ctx context.Context, filename string) error {
	state, err := sessMgr.Load()
	if err != nil {
		return fmt.Errorf("no active session: %w", err)
	}

	local := storage.NewLocal(cfg.Local.SyncFolder)

	// Try cards folder first, then prose folder.
	docType, content, err := locateFile(local, state.SeriesTitle, state.BookTitle, filename)
	if err != nil {
		return err
	}

	// Infer card type for display.
	cardType, _ := inferCardTypeFromFilename(filename)

	return phases.RunCheck(ctx, phases.CheckConfig{
		AI:           phases.NewAIConversation(aiClient),
		LocalReader:  local,
		Prompter:     prompter,
		SeriesTitle:  state.SeriesTitle,
		BookTitle:    state.BookTitle,
		Filename:     filename,
		Content:      content,
		DocumentType: docType,
		CardType:     cardType,
	})
}

func locateFile(local *storage.Local, seriesTitle, bookTitle, filename string) (docType, content string, err error) {
	upper := strings.ToUpper(filename)

	// Prose files.
	if strings.HasPrefix(upper, "CH") && (strings.Contains(upper, "DRAFT") || strings.Contains(upper, "SCENE")) ||
		strings.HasPrefix(upper, "SCENE-") {
		content, err = local.Read(utils.ProseFolder(seriesTitle, bookTitle), filename)
		if err == nil {
			return "PROSE", content, nil
		}
	}

	// Card files — try cards folder.
	content, err = local.Read(utils.CardsFolder(seriesTitle, bookTitle), filename)
	if err == nil {
		return "CARD", content, nil
	}

	// Series-level cards (Trilogy).
	if strings.HasPrefix(upper, "TRILOGY") {
		content, err = local.Read(utils.TrilogyFolder(seriesTitle), filename)
		if err == nil {
			return "CARD", content, nil
		}
	}

	return "", "", fmt.Errorf("file %q not found in local storage (tried cards and prose folders)", filename)
}
