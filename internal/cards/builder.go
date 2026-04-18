package cards

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/poorlyordered/writers_harness/internal/ai"
)

func Build(ctx context.Context, conv ai.Conversation, systemPrompt, instruction string) (string, error) {
	conv.ResetHistory()

	resp, err := conv.Send(ctx, systemPrompt, instruction)
	if err != nil {
		return "", fmt.Errorf("drafting card: %w", err)
	}
	printDraft(resp)

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("\nYou (LOCK to confirm): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("reading input: %w", err)
		}
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}
		switch strings.ToUpper(input) {
		case "LOCK", "CONFIRM", "CONFIRMED", "APPROVE", "APPROVED":
			return resp, nil
		}
		resp, err = conv.Send(ctx, systemPrompt, input)
		if err != nil {
			return "", fmt.Errorf("revision error: %w", err)
		}
		printDraft(resp)
	}
}

func printDraft(s string) {
	fmt.Printf("\n─── CARD DRAFT ──────────────────────────────────────────────\n%s\n─────────────────────────────────────────────────────────────\n", s)
}
