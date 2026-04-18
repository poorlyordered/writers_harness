package phases

import (
	"context"
	"fmt"

	anthropic "github.com/poorlyordered/writers_harness/internal/anthropic"
	"github.com/poorlyordered/writers_harness/internal/storage"
)

// AIConversation adapts the internal anthropic.Client to the Conversation interface.
type AIConversation struct {
	client *anthropic.Client
}

// NewAIConversation wraps an anthropic.Client as a Conversation.
func NewAIConversation(c *anthropic.Client) *AIConversation {
	return &AIConversation{client: c}
}

func (a *AIConversation) Send(ctx context.Context, systemPrompt, userMessage string) (string, error) {
	return a.client.Send(ctx, systemPrompt, userMessage)
}

func (a *AIConversation) ResetHistory() {
	a.client.ResetHistory()
}

// BoxFileWriter adapts the AI client to write files to Box via a Claude tool call.
// The Go process asks Claude to write the file using Box MCP tools.
type BoxFileWriter struct {
	client *anthropic.Client
}

// NewBoxFileWriter creates a file writer that instructs Claude to use Box MCP tools.
func NewBoxFileWriter(c *anthropic.Client) *BoxFileWriter {
	return &BoxFileWriter{client: c}
}

func (b *BoxFileWriter) Write(folderPath, filename, content string) error {
	ctx := context.Background()
	prompt := fmt.Sprintf(
		"Please use the Box MCP tools to write the following file.\n"+
			"Folder path: %s\n"+
			"Filename: %s\n\n"+
			"File content (write exactly as provided):\n%s",
		folderPath, filename, content)

	_, err := b.client.Send(ctx,
		"You are a file management assistant. Use Box MCP tools to create files as instructed. "+
			"Confirm when the file has been written successfully.",
		prompt)
	return err
}

// LocalFileWriter adapts the storage.Local to the FileWriter interface.
type LocalFileWriter struct {
	local *storage.Local
}

// NewLocalFileWriter creates a file writer backed by local storage.
func NewLocalFileWriter(l *storage.Local) *LocalFileWriter {
	return &LocalFileWriter{local: l}
}

func (lw *LocalFileWriter) Write(folderPath, filename, content string) error {
	return lw.local.Write(folderPath, filename, content)
}
