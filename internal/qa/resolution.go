package qa

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ResolutionChoice is the writer's response to a QA failure.
type ResolutionChoice int

const (
	ResolveRetry    ResolutionChoice = iota // revise the card or scene and retry
	ResolveOverride                          // accept with flags logged
	ResolveBlock                             // halt advancement
)

// PresentAndChoose displays QA failures and asks the writer how to proceed.
// Returns the chosen action and any item IDs the writer overrides.
func PresentAndChoose(result *Result) (ResolutionChoice, []string) {
	fmt.Println()
	fmt.Println(result.FormatFailures())
	fmt.Println("─────────────────────────────────────────────────────────────")
	fmt.Println("How would you like to proceed?")
	fmt.Println("  1 — Revise to fix the issues (continue conversation)")
	fmt.Println("  2 — Override and accept-with-flags (failures logged)")
	fmt.Println("  3 — Block advancement (exit phase)")
	fmt.Print("\nChoice [1/2/3]: ")

	reader := bufio.NewReader(os.Stdin)
	for {
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		switch input {
		case "1":
			return ResolveRetry, nil
		case "2":
			ids := make([]string, 0, len(result.FailedItems()))
			for _, item := range result.FailedItems() {
				ids = append(ids, item.ID)
			}
			fmt.Printf("[HARNESS] Override accepted. Logged items: %s\n", strings.Join(ids, ", "))
			return ResolveOverride, ids
		case "3":
			return ResolveBlock, nil
		default:
			fmt.Print("Enter 1, 2, or 3: ")
		}
	}
}
