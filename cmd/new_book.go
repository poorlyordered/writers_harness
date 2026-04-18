package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/poorlyordered/writers_harness/internal/index"
	"github.com/poorlyordered/writers_harness/internal/phases"
	"github.com/poorlyordered/writers_harness/internal/session"
	"github.com/poorlyordered/writers_harness/internal/storage"
	"github.com/poorlyordered/writers_harness/internal/utils"
)

var newBookCmd = &cobra.Command{
	Use:   "new-book",
	Short: "Add a new book to an existing series and start Phase 1",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runNewBook(context.Background())
	},
}

func init() { rootCmd.AddCommand(newBookCmd) }

func runNewBook(ctx context.Context) error {
	// Load the current session to get series title and next book number.
	state, err := sessMgr.Load()
	if err != nil {
		return fmt.Errorf("no active session: %w\nRun 'harness new-series' first", err)
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("\n═══════════════════════════════════════════════════════════\n")
	fmt.Printf("  WRITING HARNESS — NEW BOOK\n")
	fmt.Printf("  Series: %s\n", state.SeriesTitle)
	fmt.Printf("═══════════════════════════════════════════════════════════\n\n")

	fmt.Print("New book title: ")
	bookTitle, _ := reader.ReadString('\n')
	bookTitle = strings.TrimSpace(bookTitle)
	if bookTitle == "" {
		return fmt.Errorf("book title cannot be empty")
	}

	bookNum := state.BookNum + 1
	fmt.Printf("\n[HARNESS] Creating Box folder structure for Book %d: %q...\n", bookNum, bookTitle)

	// Create Box folder tree for the new book.
	folders := []string{
		utils.SeedsFolder(state.SeriesTitle, bookTitle),
		utils.ExpansionFolder(state.SeriesTitle, bookTitle),
		utils.CardsFolder(state.SeriesTitle, bookTitle),
		utils.ProseFolder(state.SeriesTitle, bookTitle),
		utils.QAFolder(state.SeriesTitle, bookTitle),
	}
	folderList := "  - " + strings.Join(folders, "\n  - ")
	_, err = aiClient.Send(ctx,
		"You are a file system assistant. Use Box MCP tools to create folders as instructed.",
		fmt.Sprintf("Create these Box folders for book %d:\n%s\n\nConfirm when done.", bookNum, folderList))
	if err != nil {
		return fmt.Errorf("creating Box folders: %w", err)
	}
	fmt.Println("[HARNESS] Box folders created.")

	// Load and update the Series Index.
	local := storage.NewLocal(cfg.Local.SyncFolder)
	seriesRoot := utils.SeriesRoot(state.SeriesTitle)
	indexFile := utils.SeriesIndexFile(state.SeriesTitle)

	var idx *index.Index
	if _, err := local.Read(seriesRoot, indexFile); err == nil {
		// Existing index found — rebuild with both books preserved.
		idx = index.NewEmpty(state.SeriesTitle)
		idx.EnsureBook(state.BookTitle, state.BookNum)
		idx.EnsureBook(bookTitle, bookNum)
	} else {
		idx = index.NewEmpty(state.SeriesTitle)
		idx.EnsureBook(state.BookTitle, state.BookNum)
		idx.EnsureBook(bookTitle, bookNum)
	}
	indexContent := idx.Render()

	boxWriter := phases.NewBoxFileWriter(aiClient)
	if err := boxWriter.Write(seriesRoot, indexFile, indexContent); err != nil {
		return fmt.Errorf("updating series index in Box: %w", err)
	}
	if err := local.Write(seriesRoot, indexFile, indexContent); err != nil {
		fmt.Printf("[HARNESS] Warning: local index write failed: %v\n", err)
	}
	fmt.Printf("[HARNESS] Series Index updated: %s/%s\n", seriesRoot, indexFile)

	// Advance session to the new book.
	newState := session.New(session.Phase1, state.SeriesTitle, bookTitle, bookNum)
	if err := sessMgr.Save(newState); err != nil {
		return fmt.Errorf("saving session: %w", err)
	}

	fmt.Printf("\n[HARNESS] Book %d %q initialised.\n", bookNum, bookTitle)
	fmt.Println("[HARNESS] Run 'harness phase1' to begin Idea Generation for this book.")
	return nil
}
