// Package anthropic wraps the Anthropic Beta Messages API, managing the
// multi-turn conversation loop and Box MCP server injection.
package anthropic

import (
	"context"
	"fmt"
	"strings"

	anthropicsdk "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// Client manages a stateful conversation with the Anthropic API.
type Client struct {
	sdk       anthropicsdk.Client
	model     string
	maxTokens int
	mcpServer *anthropicsdk.BetaRequestMCPServerURLDefinitionParam

	history []anthropicsdk.BetaMessageParam
	counter *TokenCounter
}

// New creates an Anthropic client. apiKey may be empty to use ANTHROPIC_API_KEY env var.
func New(apiKey, model string, maxTokens int) *Client {
	opts := []option.RequestOption{}
	if apiKey != "" {
		opts = append(opts, option.WithAPIKey(apiKey))
	}
	sdk := anthropicsdk.NewClient(opts...)
	return &Client{
		sdk:       sdk,
		model:     model,
		maxTokens: maxTokens,
		counter:   NewTokenCounter(200_000),
	}
}

// SetMCPServer configures the Box MCP server injected into every API request.
func (c *Client) SetMCPServer(srv anthropicsdk.BetaRequestMCPServerURLDefinitionParam) {
	c.mcpServer = &srv
}

// ResetHistory clears the conversation history for a new session.
func (c *Client) ResetHistory() {
	c.history = nil
}

// History returns the current conversation history.
func (c *Client) History() []anthropicsdk.BetaMessageParam {
	return c.history
}

// Send appends a user message, calls the API, appends the assistant response,
// and returns the assistant's text. The system prompt has card content injected
// by the caller before this is invoked.
func (c *Client) Send(ctx context.Context, systemPrompt, userMessage string) (string, error) {
	// Append user turn.
	c.history = append(c.history, anthropicsdk.NewBetaUserMessage(
		anthropicsdk.NewBetaTextBlock(userMessage),
	))

	// Trim oldest turns if approaching token budget.
	c.counter.TrimIfNeeded(&c.history, systemPrompt)

	params := anthropicsdk.BetaMessageNewParams{
		Model:     anthropicsdk.Model(c.model),
		MaxTokens: int64(c.maxTokens),
		System: []anthropicsdk.BetaTextBlockParam{
			{Text: systemPrompt},
		},
		Messages: c.history,
		Betas: []anthropicsdk.AnthropicBeta{
			anthropicsdk.AnthropicBetaMCPClient2025_04_04,
		},
	}
	if c.mcpServer != nil {
		params.MCPServers = []anthropicsdk.BetaRequestMCPServerURLDefinitionParam{*c.mcpServer}
	}

	resp, err := c.sdk.Beta.Messages.New(ctx, params)
	if err != nil {
		return "", fmt.Errorf("anthropic api: %w", err)
	}

	text := extractText(resp)

	// Append assistant turn.
	c.history = append(c.history, anthropicsdk.BetaMessageParam{
		Role: anthropicsdk.BetaMessageParamRoleAssistant,
		Content: []anthropicsdk.BetaContentBlockParamUnion{
			{OfText: &anthropicsdk.BetaTextBlockParam{Text: text}},
		},
	})

	return text, nil
}

// extractText concatenates all text blocks from a BetaMessage response.
func extractText(msg *anthropicsdk.BetaMessage) string {
	var parts []string
	for _, block := range msg.Content {
		if block.Type == "text" {
			parts = append(parts, block.Text)
		}
	}
	return strings.Join(parts, "")
}
