package anthropic

import (
	"fmt"

	anthropicsdk "github.com/anthropics/anthropic-sdk-go"
)

// TokenCounter tracks approximate token usage and trims conversation history
// before the context window fills. The system prompt and loaded cards (pinned
// in the system prompt) are never trimmed.
type TokenCounter struct {
	contextWindow int
}

func NewTokenCounter(contextWindow int) *TokenCounter {
	return &TokenCounter{contextWindow: contextWindow}
}

// TrimIfNeeded removes the oldest user+assistant turn pairs from history
// when the estimated token count approaches the context limit.
func (tc *TokenCounter) TrimIfNeeded(history *[]anthropicsdk.BetaMessageParam, systemPrompt string) {
	threshold := int(float64(tc.contextWindow) * 0.80)
	critical := int(float64(tc.contextWindow) * 0.90)
	estimated := tc.estimate(systemPrompt, *history)

	if estimated > critical {
		if len(*history) >= 2 {
			*history = (*history)[2:]
			fmt.Println("\n[HARNESS] Context window approaching limit. Oldest conversation turns trimmed.")
			fmt.Println("[HARNESS] System prompt and loaded cards are retained.")
		}
	} else if estimated > threshold {
		pct := estimated * 100 / tc.contextWindow
		fmt.Printf("\n[HARNESS] Context window at ~%d%% capacity. Consider dropping optional cards.\n", pct)
	}
}

// estimate produces a rough token count (chars/4) for system prompt + history.
func (tc *TokenCounter) estimate(systemPrompt string, history []anthropicsdk.BetaMessageParam) int {
	total := len(systemPrompt) / 4
	for _, msg := range history {
		for _, block := range msg.Content {
			if block.OfText != nil {
				total += len(block.OfText.Text) / 4
			}
		}
	}
	return total
}
