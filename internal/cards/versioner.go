package cards

import (
	"fmt"
	"strings"
	"time"
)

// Lock prepends a status header to card content and returns the locked document.
func Lock(content, cardType, name string, version int) string {
	header := fmt.Sprintf(
		"# %s — %s\n\n**Status:** LOCKED\n**Version:** v%d\n**Locked:** %s\n\n---\n\n",
		cardType, name, version, time.Now().Format("2006-01-02"),
	)
	return header + strings.TrimSpace(content) + "\n"
}

// IsLocked reports whether the card content has STATUS: LOCKED in its header.
func IsLocked(content string) bool {
	lower := strings.ToLower(content)
	return strings.Contains(lower, "**status:** locked") ||
		strings.Contains(lower, "status: locked")
}

// ExtractBody returns the card content below the lock header (after the --- separator).
func ExtractBody(content string) string {
	idx := strings.Index(content, "\n---\n")
	if idx < 0 {
		return content
	}
	return strings.TrimSpace(content[idx+5:])
}
