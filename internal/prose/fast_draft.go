package prose

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
)

// FastDraftParams describes a full chapter to fast-draft.
type FastDraftParams struct {
	ChapterNum    int
	ChapterType   string
	SceneCount    int
	WordTarget    int // per scene
	POVCharacter  string
	Context       string // loaded card content
}

// FastDraftChapter drafts all scenes in a chapter without per-scene pauses.
// After all scenes are drafted it presents the full chapter for review.
// Returns the map of scene number → prose content and any error.
func FastDraftChapter(ctx context.Context, ai Conversation, systemPrompt string, p FastDraftParams) (map[int]string, error) {
	scenes := make(map[int]string, p.SceneCount)

	fmt.Printf("\n[FAST DRAFT] Drafting Chapter %d (%s) — %d scenes, ~%d words each...\n",
		p.ChapterNum, p.ChapterType, p.SceneCount, p.WordTarget)

	for i := 1; i <= p.SceneCount; i++ {
		fmt.Printf("[FAST DRAFT] Scene %d/%d...\n", i, p.SceneCount)
		ai.ResetHistory()
		instruction := fmt.Sprintf(
			"Fast-draft Scene %d of Chapter %d (%s) in ~%d words. POV: %s.\n"+
				"No pauses. Write the complete scene: hook, turn point, closing beat.\n\n"+
				"Context:\n%s",
			i, p.ChapterNum, p.ChapterType, p.WordTarget, p.POVCharacter, p.Context,
		)
		resp, err := ai.Send(ctx, systemPrompt, instruction)
		if err != nil {
			return nil, fmt.Errorf("fast draft scene %d failed: %w", i, err)
		}
		scenes[i] = resp
	}

	// Present the assembled chapter for review.
	fmt.Printf("\n─── FAST DRAFT: CHAPTER %d COMPLETE ───────────────────────\n", p.ChapterNum)
	for i := 1; i <= p.SceneCount; i++ {
		fmt.Printf("\n--- Scene %d ---\n%s\n", i, scenes[i])
	}
	fmt.Println("────────────────────────────────────────────────────────────")
	fmt.Println("\n[FAST DRAFT] Chapter complete. Review above and revise specific scenes by scene number,")
	fmt.Println("or type LOCK to accept the chapter as-is.")

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("\nRevise scene # (or LOCK): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if strings.ToUpper(input) == "LOCK" {
			break
		}
		// Revise a specific scene.
		var sceneNum int
		if _, err := fmt.Sscanf(input, "%d", &sceneNum); err != nil || sceneNum < 1 || sceneNum > p.SceneCount {
			fmt.Printf("Enter a scene number 1–%d, or LOCK.\n", p.SceneCount)
			continue
		}
		fmt.Printf("[FAST DRAFT] Revising Scene %d...\n", sceneNum)
		fmt.Print("Your revision note: ")
		note, _ := reader.ReadString('\n')
		note = strings.TrimSpace(note)

		reviseInstruction := fmt.Sprintf(
			"Revise Scene %d of Chapter %d based on this note: %s\n\nCurrent draft:\n%s",
			sceneNum, p.ChapterNum, note, scenes[sceneNum],
		)
		ai.ResetHistory()
		revised, err := ai.Send(ctx, systemPrompt, reviseInstruction)
		if err != nil {
			fmt.Printf("[FAST DRAFT] Revision error: %v\n", err)
			continue
		}
		scenes[sceneNum] = revised
		printScene(revised)
	}

	return scenes, nil
}
