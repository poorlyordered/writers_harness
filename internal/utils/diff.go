package utils

import (
	"fmt"
	"strings"
)

// DiffView renders a two-column diff of old vs. new content for the Trilogy Card
// update review flow (SPEC-005 §6 Trilogy Card — Subsequent Books).
func DiffView(oldContent, newContent string) string {
	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")

	maxLines := len(oldLines)
	if len(newLines) > maxLines {
		maxLines = len(newLines)
	}

	const colWidth = 60
	sep := strings.Repeat("─", colWidth)
	header := fmt.Sprintf("%-*s  │  %-*s", colWidth, "CURRENT (LOCKED)", colWidth, "PROPOSED UPDATE")
	divider := sep + "──┼──" + sep

	var sb strings.Builder
	sb.WriteString(header + "\n")
	sb.WriteString(divider + "\n")

	for i := 0; i < maxLines; i++ {
		var left, right string
		if i < len(oldLines) {
			left = oldLines[i]
		}
		if i < len(newLines) {
			right = newLines[i]
		}
		marker := "  "
		if left != right {
			marker = "◄►"
		}
		sb.WriteString(fmt.Sprintf("%-*s %s  %-*s\n", colWidth, truncate(left, colWidth), marker, colWidth, truncate(right, colWidth)))
	}
	return sb.String()
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
