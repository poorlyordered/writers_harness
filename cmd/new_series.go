package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/poorlyordered/writers_harness/internal/index"
	"github.com/poorlyordered/writers_harness/internal/phases"
	"github.com/poorlyordered/writers_harness/internal/session"
	"github.com/poorlyordered/writers_harness/internal/utils"
)

var newSeriesCmd = &cobra.Command{
	Use:   "new-series",
	Short: "Start a new series — creates Box folder structure and Series Index",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runNewSeries(context.Background())
	},
}

func init() {
	rootCmd.AddCommand(newSeriesCmd)
}

func runNewSeries(ctx context.Context) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n═══════════════════════════════════════════════════════════")
	fmt.Println("  WRITING HARNESS — NEW SERIES")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()

	fmt.Print("Series title: ")
	seriesTitle, _ := reader.ReadString('\n')
	seriesTitle = strings.TrimSpace(seriesTitle)
	if seriesTitle == "" {
		return fmt.Errorf("series title cannot be empty")
	}

	fmt.Print("First book title: ")
	bookTitle, _ := reader.ReadString('\n')
	bookTitle = strings.TrimSpace(bookTitle)
	if bookTitle == "" {
		return fmt.Errorf("book title cannot be empty")
	}

	fmt.Printf("\n[HARNESS] Creating Box folder structure for %q...\n", seriesTitle)

	// Ask Claude (via Box MCP) to create the folder tree.
	folders := []string{
		utils.SeriesRoot(seriesTitle),
		utils.WorldBibleFolder(seriesTitle),
		utils.TrilogyFolder(seriesTitle),
		utils.SeedsFolder(seriesTitle, bookTitle),
		utils.ExpansionFolder(seriesTitle, bookTitle),
		utils.CardsFolder(seriesTitle, bookTitle),
		utils.ProseFolder(seriesTitle, bookTitle),
		utils.QAFolder(seriesTitle, bookTitle),
	}
	folderList := "  - " + strings.Join(folders, "\n  - ")
	_, err := aiClient.Send(ctx,
		"You are a file system assistant. Use Box MCP tools to create folders as instructed.",
		fmt.Sprintf("Create these Box folders (create each as a subfolder under the series root):\n%s\n\nConfirm when done.", folderList))
	if err != nil {
		return fmt.Errorf("creating Box folders: %w", err)
	}
	fmt.Println("[HARNESS] Box folders created.")

	// Create the Series Index.
	idx := index.NewEmpty(seriesTitle)
	_ = idx.EnsureBook(bookTitle, 1)
	seriesRoot := utils.SeriesRoot(seriesTitle)
	indexContent := idx.Render()
	indexFile := utils.SeriesIndexFile(seriesTitle)

	boxWriter := phases.NewBoxFileWriter(aiClient)
	if err := boxWriter.Write(seriesRoot, indexFile, indexContent); err != nil {
		return fmt.Errorf("writing series index to Box: %w", err)
	}
	if err := localStore.Write(seriesRoot, indexFile, indexContent); err != nil {
		fmt.Printf("[HARNESS] Warning: local index write failed: %v\n", err)
	}

	// Save JSON state so future phases can load and update it without parsing Markdown.
	if stateJSON, err := json.MarshalIndent(idx, "", "  "); err == nil {
		stateFile := utils.SeriesIndexStateFile(seriesTitle)
		_ = localStore.Write(seriesRoot, stateFile, string(stateJSON))
		_ = boxWriter.Write(seriesRoot, stateFile, string(stateJSON))
	}

	fmt.Printf("[HARNESS] Series Index: %s/%s\n", seriesRoot, indexFile)

	// Persist session state.
	state := session.New(session.Phase1, seriesTitle, bookTitle, 1)
	if err := sessMgr.Save(state); err != nil {
		return fmt.Errorf("saving session state: %w", err)
	}

	fmt.Printf("\n[HARNESS] Series %q initialised successfully.\n", seriesTitle)
	fmt.Printf("[HARNESS] Run 'harness phase1' to begin Idea Generation.\n\n")
	return nil
}
