package cards

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
)

// Conversation is the minimal AI interface the builder needs.
// phases.AIConversation satisfies this interface via structural typing.
type Conversation interface {
	Send(ctx context.Context, systemPrompt, userMessage string) (string, error)
	ResetHistory()
}

// Build drives an AI conversation to draft a single card.
// It resets AI history, sends the instruction, then loops on writer feedback
// until the writer types LOCK, CONFIRM, or APPROVED.
// Returns the final confirmed card body (without the lock header).
func Build(ctx context.Context, ai Conversation, systemPrompt, instruction string) (string, error) {
	ai.ResetHistory()

	resp, err := ai.Send(ctx, systemPrompt, instruction)
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
		resp, err = ai.Send(ctx, systemPrompt, input)
		if err != nil {
			return "", fmt.Errorf("revision error: %w", err)
		}
		printDraft(resp)
	}
}

func printDraft(s string) {
	fmt.Printf("\n─── CARD DRAFT ──────────────────────────────────────────────\n%s\n─────────────────────────────────────────────────────────────\n", s)
}
