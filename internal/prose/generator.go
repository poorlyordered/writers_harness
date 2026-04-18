package prose

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
)

// Conversation is the minimal AI interface the generator needs.
type Conversation interface {
	Send(ctx context.Context, systemPrompt, userMessage string) (string, error)
	ResetHistory()
}

// SceneParams describes the scene to draft.
type SceneParams struct {
	ChapterNum    int
	SceneNum      int
	ChapterType   string // e.g. "Really Bad Day", "Inciting Incident"
	StoryFunction string // what this scene must accomplish
	POVCharacter  string
	WordTarget    int
	Context       string // injected card content (NOVEL, CHARACTER, CHAPTER cards)
}

// DraftScene runs an AI conversation to draft a single scene.
// It loops until the writer types LOCK/CONFIRM or REJECT [reason].
// Returns the confirmed prose, the rejection reason (if rejected), and any error.
func DraftScene(ctx context.Context, ai Conversation, systemPrompt string, p SceneParams) (prose, rejectReason string, err error) {
	ai.ResetHistory()

	instruction := buildSceneInstruction(p)

	resp, err := ai.Send(ctx, systemPrompt, instruction)
	if err != nil {
		return "", "", fmt.Errorf("scene draft failed: %w", err)
	}
	printScene(resp)

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("\nYou (LOCK to accept, REJECT [reason] to reject): ")
		input, readErr := reader.ReadString('\n')
		if readErr != nil {
			return "", "", fmt.Errorf("reading input: %w", readErr)
		}
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}
		upper := strings.ToUpper(input)
		switch {
		case upper == "LOCK" || upper == "CONFIRM" || upper == "CONFIRMED":
			return resp, "", nil
		case strings.HasPrefix(upper, "REJECT"):
			reason := strings.TrimSpace(input[6:])
			if reason == "" {
				reason = "writer rejected without specific reason"
			}
			return "", reason, nil
		}
		resp, err = ai.Send(ctx, systemPrompt, input)
		if err != nil {
			return "", "", fmt.Errorf("revision error: %w", err)
		}
		printScene(resp)
	}
}

func buildSceneInstruction(p SceneParams) string {
	return fmt.Sprintf(
		"Draft Scene %d of Chapter %d (%s).\n\n"+
			"Story function: %s\n"+
			"POV character: %s\n"+
			"Target word count: ~%d words\n\n"+
			"Write the full scene in third-person close POV. "+
			"Include a strong opening hook, a clear turn point where the scene's state changes, "+
			"and a closing beat that resonates. "+
			"Stay consistent with the character's Voice Anchor and all locked cards.\n\n"+
			"Loaded context:\n%s",
		p.SceneNum, p.ChapterNum, p.ChapterType,
		p.StoryFunction, p.POVCharacter, p.WordTarget, p.Context,
	)
}

func printScene(s string) {
	fmt.Printf("\n═══ SCENE DRAFT ════════════════════════════════════════════\n%s\n════════════════════════════════════════════════════════════\n", s)
}

// WordCount returns a rough word count for the given text.
func WordCount(text string) int {
	return len(strings.Fields(text))
}
