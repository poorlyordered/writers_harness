// Package ai defines shared interfaces used across the writing harness.
package ai

import "context"

// Conversation is the minimal interface for AI-backed interactive loops.
// phases.AIConversation satisfies this via structural typing.
type Conversation interface {
	Send(ctx context.Context, systemPrompt, userMessage string) (string, error)
	ResetHistory()
}
