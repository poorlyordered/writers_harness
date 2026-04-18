package prose

import (
	"fmt"
	"strings"
	"time"
)

// AssembleChapter concatenates scenes into a chapter document.
func AssembleChapter(seriesTitle, bookTitle string, chapterNum int, chapterType string, scenes map[int]string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Chapter %d — %s\n\n", chapterNum, chapterType))
	sb.WriteString(fmt.Sprintf("**Series:** %s  |  **Book:** %s\n", seriesTitle, bookTitle))
	sb.WriteString(fmt.Sprintf("**Assembled:** %s\n\n---\n\n", time.Now().Format("2006-01-02")))

	for i := 1; i <= len(scenes); i++ {
		if content, ok := scenes[i]; ok {
			sb.WriteString(strings.TrimSpace(content))
			sb.WriteString("\n\n")
		}
	}

	return sb.String()
}

// AssembleManuscript concatenates all chapters into a full manuscript document.
func AssembleManuscript(seriesTitle, bookTitle string, chapters map[int]string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", bookTitle))
	sb.WriteString(fmt.Sprintf("**Series:** %s\n", seriesTitle))
	sb.WriteString(fmt.Sprintf("**Draft assembled:** %s\n\n", time.Now().Format("2006-01-02")))
	sb.WriteString("---\n\n")

	for i := 1; i <= len(chapters); i++ {
		if content, ok := chapters[i]; ok {
			sb.WriteString(strings.TrimSpace(content))
			sb.WriteString("\n\n---\n\n")
		}
	}

	return sb.String()
}

// TransitionCheck returns a prompt to ask the AI to verify scene transitions.
func TransitionCheck(chapterNum, fromScene, toScene int, fromProse, toProse string) string {
	return fmt.Sprintf(
		"Check the transition between Scene %d and Scene %d of Chapter %d. "+
			"Confirm: time/space continuity, emotional continuity, and no character position errors.\n\n"+
			"End of Scene %d:\n%s\n\nStart of Scene %d:\n%s\n\n"+
			"Report any continuity errors, or confirm the transition is clean.",
		fromScene, toScene, chapterNum,
		fromScene, lastNWords(fromProse, 150),
		toScene, firstNWords(toProse, 150),
	)
}

func lastNWords(s string, n int) string {
	words := strings.Fields(s)
	if len(words) <= n {
		return s
	}
	return "..." + strings.Join(words[len(words)-n:], " ")
}

func firstNWords(s string, n int) string {
	words := strings.Fields(s)
	if len(words) <= n {
		return s
	}
	return strings.Join(words[:n], " ") + "..."
}
