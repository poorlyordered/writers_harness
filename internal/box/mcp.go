// Package box builds the Box MCP server configuration for Anthropic API requests
// and probes Box connectivity. Actual Box I/O is performed by Claude via MCP
// tool use — the Go process never calls Box directly.
package box

import (
	"context"
	"fmt"
	"net/http"
	"time"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
)

// Client holds the Box MCP connection parameters.
type Client struct {
	serverURL string
	authToken string
}

// New creates a Box client from config values.
func New(serverURL, authToken string) *Client {
	return &Client{serverURL: serverURL, authToken: authToken}
}

// MCPServerParam returns the BetaRequestMCPServerURLDefinitionParam to inject
// into every Anthropic Beta Messages API request.
func (c *Client) MCPServerParam() anthropic.BetaRequestMCPServerURLDefinitionParam {
	p := anthropic.BetaRequestMCPServerURLDefinitionParam{
		Name: "box",
		URL:  c.serverURL,
	}
	if c.authToken != "" {
		p.AuthorizationToken = param.NewOpt(c.authToken)
	}
	return p
}

// Ping checks that the Box MCP server URL is reachable via HTTP HEAD.
// Returns nil if reachable, an error otherwise.
func (c *Client) Ping(ctx context.Context) error {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, c.serverURL, nil)
	if err != nil {
		return fmt.Errorf("building ping request: %w", err)
	}
	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("box mcp unreachable: %w", err)
	}
	resp.Body.Close()
	return nil
}
