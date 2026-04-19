package phases

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/poorlyordered/writers_harness/internal/cards"
	"github.com/poorlyordered/writers_harness/internal/qa"
	"github.com/poorlyordered/writers_harness/internal/queue"
	"github.com/poorlyordered/writers_harness/internal/utils"
)

// Phase2Config holds everything the Phase 2 runner needs.
type Phase2Config struct {
	AI          Conversation
	BoxWriter   FileWriter
	LocalWriter FileWriter
	Prompter    PromptLoader

	SeriesTitle  string
	BookTitle    string
	IsStandalone bool
	SeedContent  string // loaded Locked Seed Prompt from Box/local
}

// snowflakeStep describes one of the 7 Snowflake steps.
type snowflakeStep struct {
	Number      int
	Name        string
	Instruction string // injected into the AI request
}

var snowflakeSteps = []snowflakeStep{
	{
		Number: 1,
		Name:   "One-Sentence Story Summary (Logline)",
		Instruction: `Draft a single sentence — no longer than 25 words — that captures: protagonist + conflict + stakes.
The sentence must imply the central thematic question from the seed without stating it directly.
It should reflect the protagonist's Want (conscious goal), not their Need.
Present your draft, explain how it maps to the seed's central question and Want, then ask the writer for feedback.
When confirmed, output: STEP 1 STATUS: LOCKED`,
	},
	{
		Number: 2,
		Name:   "One-Paragraph Story Summary",
		Instruction: `Draft a paragraph of 3-5 sentences covering beginning, middle, and end of the story.
Then explicitly map it against the Story Circle's 8 beats:
  Beats 1-2 (You/Need) → Setup
  Beats 3-4 (Go/Search) → Rising action
  Beats 5-6 (Find/Take) → Midpoint and cost
  Beats 7-8 (Return/Change) → Resolution
Also confirm the Four Acts are visible (Setup / Response / Attack / Resolution).
Flag any beat that is missing or compressed beyond recognition.
Present the paragraph, the Story Circle beat map, and the Four-Act check.
When confirmed, output: STEP 2 STATUS: LOCKED`,
	},
	{
		Number: 3,
		Name:   "Character Summaries",
		Instruction: `Draft a one-paragraph summary for each major character: protagonist, primary antagonist, and key supporting characters.
For each character cover: motivation, goal, conflict, and arc endpoint.
For each character, also surface the four pillars:
  Want: what they consciously pursue
  Need: what they actually require to grow
  Fear: what holds them back
  Misbelief: the lie they believe at the start
Assign each character an Archetype stage (Orphan / Wanderer / Warrior / Martyr) — where they start and where they end.
Explicitly map the antagonist's four pillars against the protagonist's — confirm the mirror relationship from the seed.
Characters appearing briefly in this book but important in later books get SKETCH tier treatment:
  core wound, misbelief, archetype stage, thematic role, estimated book of full appearance.
Challenge any character whose arc does not connect to the central thematic question.
Present all summaries with pillar tables and archetype assignments.
When confirmed, output: STEP 3 STATUS: LOCKED`,
	},
	{
		Number: 4,
		Name:   "Full Story Synopsis (One Page)",
		Instruction: `Draft a full-page synopsis (~500 words) covering the complete story from opening image to final image.
Every major plot point must be present.
Then build the 40-chapter beat map:
  Chapters 1-10:   Act 1 beats
  Chapters 11-20:  Act 2A beats
  Chapters 21-30:  Act 2B beats
  Chapters 31-40:  Act 3 beats
Identify the four act turning points with chapter numbers:
  Plot Point 1 (end of Act 1)
  Midpoint (end of Act 2A)
  Dark Night (end of Act 2B)
  Climax (Act 3 resolution)
Confirm all 8 Story Circle beats are present and correctly sequenced.
Flag any act that is over- or under-developed.
Present the synopsis and the 40-chapter map together.
When confirmed, output: STEP 4 STATUS: LOCKED`,
	},
	{
		Number: 5,
		Name:   "Character Arc Expansions",
		Instruction: `Expand each major FULL tier character to approximately one page each.
Cover: backstory, wound (specific formative event — not abstract), arc progression, relationship dynamics, story function.
For each character map the full Archetype Cycle progression:
  Where do they start? What stage do they reach by book end?
  What event triggers each stage transition?
Fully resolve all four pillars (Want/Need/Fear/Misbelief) — not just seed them.
Trace the misbelief back to the specific wound.
Connect the Need to the central theme.
For each antagonist: formally document their internal logic — the coherent worldview that makes them
not simply wrong. This section is REQUIRED. An antagonist without defensible internal logic
is not ready to proceed.
Run a pressure test on each character:
  Is the wound specific enough?
  Is the misbelief distinct from the fear?
  Does the arc endpoint connect to the thematic question?
  Does the antagonist have a genuinely defensible worldview?
Work through weak answers in dialogue before confirming.
SKETCH tier characters: abbreviated treatment (wound, archetype stage, thematic role, estimated book).
When confirmed, output: STEP 5 STATUS: LOCKED`,
	},
	{
		Number: 6,
		Name:   "Act-by-Act Scene Breakdown",
		Instruction: `Break all four acts into major scene blocks with chapter assignments from the 40-chapter structure.
Assign scene blocks to each of the 40 chapter types. Key structural chapters that must be explicitly covered:
  Ch 1:  Really Bad Day
  Ch 2:  Mystery and Theme
  Ch 6:  Inciting Incident
  Ch 10: Pull Out Rug
  Ch 17: First Pinch Point
  Ch 20: Midpoint
  Ch 27: All Is Lost
  Ch 29: Giving Up
  Ch 36: Ultimate Defeat
  Ch 38: Unexpected Victory
  Ch 40: Death of Self
Then recommend the Seven Basic Plots pattern. Present your top 2-3 options with reasoning:
  "This story most closely follows [Pattern] because [reason]. It could also be [Pattern] because [reason]."
Wait for the writer to select or override.
Weave subplots into the main chapter spine. Flag chapters carrying both a main beat and a subplot beat.
Surface any chapter that appears overloaded — suggest redistribution.
Apply trope weave based on the seed's trope combination:
  Western: personal stakes, pinch points, community consequences
  Noir: information reveals, moral ambiguity beats, trust erosion
  Political: faction dynamics, power shifts, institutional turning points
  Romance: proximity/resistance rhythm, earned vulnerability moments
Present the full act breakdown with chapter assignments and subplot weave map.
When confirmed (including Seven Basic Plots selection), output: STEP 6 STATUS: LOCKED`,
	},
	{
		Number: 7,
		Name:   "Character Detail Sheets",
		Instruction: `Build full pre-card detail for every FULL tier character. This is the bridge to Phase 3 card completion.
For each FULL tier character include:
  Enneagram type with reasoning
  Myers-Briggs type with reasoning
  Voice and mannerism notes (concrete enough to imitate in prose)
  At least one verbal tic
  At least two physical tells
  Silence behavior
  At least two skills (specific, not generic)
  Knowledge domain
  Physical condition and limitations
  At least two hard limits (what they CANNOT do)
  At least one blind spot
  Relationship dynamics with other characters
  Active chapters list (which chapters they appear in)
  Voice anchor (2-3 sentences written in their voice)
Then run a Card Readiness Check for each FULL tier character:
  "Is there enough detail here to populate a full Character Card in Phase 3?"
Flag any section with insufficient detail. Resolve gaps in dialogue before locking.
SKETCH tier characters: abbreviated (wound, archetype stage, thematic role, book of full appearance).
When every FULL tier character passes the Card Readiness Check,
output: STEP 7 STATUS: LOCKED`,
	},
}

// RunPhase2 executes a complete Phase 2 session (all 7 Snowflake steps).
func RunPhase2(ctx context.Context, cfg Phase2Config) error {
	cfg.AI.ResetHistory()

	fmt.Printf("\n═══════════════════════════════════════════════════════════\n")
	fmt.Printf("  WRITING HARNESS — PHASE 2: IDEA EXPANSION (SNOWFLAKE)\n")
	fmt.Printf("  Series: %s  |  Book: %s\n", cfg.SeriesTitle, cfg.BookTitle)
	fmt.Printf("═══════════════════════════════════════════════════════════\n\n")

	systemPrompt, err := cfg.Prompter.LoadWithVars("system/phase2", map[string]string{
		"{{SERIES_TITLE}}": cfg.SeriesTitle,
		"{{BOOK_TITLE}}":   cfg.BookTitle,
		"{{SEED_CONTENT}}": cfg.SeedContent,
	})
	if err != nil {
		return fmt.Errorf("loading phase2 system prompt: %w", err)
	}

	// Warm up the session by loading the seed.
	openingMsg := fmt.Sprintf(
		"We are beginning Phase 2: Idea Expansion for %q using the Snowflake Method.\n\n"+
			"The Locked Seed Prompt is already loaded in your context (system prompt).\n"+
			"We will work through all 7 Snowflake steps in sequence. "+
			"Each step must be confirmed and locked before advancing.\n\n"+
			"Begin with Step 1: the one-sentence logline.",
		cfg.BookTitle)

	resp, err := cfg.AI.Send(ctx, systemPrompt, openingMsg)
	if err != nil {
		return err
	}
	printAssistant(resp)

	// Track locked step content for the final expansion document.
	stepContent := make(map[int]string)
	reader := bufio.NewReader(os.Stdin)

	// Run all 7 steps.
	for stepIdx, step := range snowflakeSteps {
		fmt.Printf("\n─── STEP %d: %s ─────────────────────────────────────────\n",
			step.Number, step.Name)

		if stepIdx > 0 {
			// Advance to next step.
			advanceMsg := fmt.Sprintf(
				"Step %d is locked. Now proceed to Step %d: %s.\n\n%s",
				step.Number-1, step.Number, step.Name, step.Instruction)
			resp, err = cfg.AI.Send(ctx, systemPrompt, advanceMsg)
			if err != nil {
				return err
			}
			printAssistant(resp)
		}

		// Conversation loop for this step.
		locked := false
		var stepOutput strings.Builder
		stepOutput.WriteString(fmt.Sprintf("## Step %d — %s\n\nSTATUS: DRAFT\n\n", step.Number, step.Name))
		stepOutput.WriteString(resp)

		for !locked {
			fmt.Print("\nYou: ")
			input, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("reading input: %w", err)
			}
			input = strings.TrimSpace(input)
			if input == "" {
				continue
			}

			resp, err = cfg.AI.Send(ctx, systemPrompt, input)
			if err != nil {
				return err
			}
			printAssistant(resp)
			stepOutput.WriteString("\n\n")
			stepOutput.WriteString(resp)

			if strings.Contains(resp, fmt.Sprintf("STEP %d STATUS: LOCKED", step.Number)) {
				locked = true
				fmt.Printf("\n[HARNESS] Step %d locked ✓\n", step.Number)
			}
		}

		stepContent[step.Number] = stepOutput.String()
	}

	// Collect roster info for Card Queue generation.
	fmt.Println("\n─── CARD QUEUE GENERATION ──────────────────────────────────")
	fmt.Println("[HARNESS] Identifying characters, factions, and locations for the Card Queue...")

	rosterResp, err := cfg.AI.Send(ctx, systemPrompt,
		"Please provide a structured roster for the Card Queue. List exactly:\n"+
			"FULL CHARACTERS: [comma-separated names, protagonist first]\n"+
			"SKETCH CHARACTERS: [comma-separated names, or NONE]\n"+
			"FULL FACTIONS: [comma-separated names, or NONE]\n"+
			"SKETCH FACTIONS: [comma-separated names, or NONE]\n"+
			"FULL THREATS: [comma-separated names, or NONE]\n"+
			"SKETCH THREATS: [comma-separated names, or NONE]\n"+
			"LOCATIONS: [comma-separated major location names, or NONE]\n\n"+
			"Output only these labelled lines, nothing else.")
	if err != nil {
		return err
	}

	genParams := parseRoster(rosterResp, cfg.SeriesTitle, cfg.BookTitle, cfg.IsStandalone)
	cardQueue := queue.Generate(genParams)

	// Assemble and save the Expansion document.
	expandDoc := assembleExpansionDoc(cfg.BookTitle, stepContent)
	expandFile := utils.ExpansionFile(cfg.BookTitle, 1)
	expandFolder := utils.ExpansionFolder(cfg.SeriesTitle, cfg.BookTitle)

	if err := cfg.BoxWriter.Write(expandFolder, expandFile, expandDoc); err != nil {
		return fmt.Errorf("writing expansion doc to Box: %w", err)
	}
	if err := cfg.LocalWriter.Write(expandFolder, expandFile, expandDoc); err != nil {
		fmt.Printf("[HARNESS] Warning: local expansion backup failed: %v\n", err)
	}
	fmt.Printf("[HARNESS] Expansion document saved: %s/%s\n", expandFolder, expandFile)

	// Save Card Queue to Box.
	queueJSON, err := cardQueue.Marshal()
	if err != nil {
		return fmt.Errorf("marshalling card queue: %w", err)
	}
	queueFile := utils.CardQueueFile(cfg.BookTitle)
	queueFolder := utils.QAFolder(cfg.SeriesTitle, cfg.BookTitle)

	if err := cfg.BoxWriter.Write(queueFolder, queueFile, string(queueJSON)); err != nil {
		return fmt.Errorf("writing card queue to Box: %w", err)
	}
	if err := cfg.LocalWriter.Write(queueFolder, queueFile, string(queueJSON)); err != nil {
		fmt.Printf("[HARNESS] Warning: local queue backup failed: %v\n", err)
	}
	fmt.Printf("[HARNESS] Card Queue saved: %s/%s (%d cards)\n",
		queueFolder, queueFile, len(cardQueue.Queue))

	// Write blank scaffold templates so the writer can pre-fill them before Phase 3.
	fmt.Printf("\n─── CARD TEMPLATES ──────────────────────────────────────────\n")
	templateFolder := utils.CardTemplatesFolder(cfg.SeriesTitle, cfg.BookTitle)
	written := 0
	for _, item := range cardQueue.Queue {
		scaffold := cards.GenerateScaffold(item.CardType, item.CardSubtype, item.Name)
		tplFile := utils.CardTemplateFile(item.Filename)
		if err := cfg.BoxWriter.Write(templateFolder, tplFile, scaffold); err != nil {
			fmt.Printf("[HARNESS] Warning: template write failed for %s: %v\n", tplFile, err)
			continue
		}
		_ = cfg.LocalWriter.Write(templateFolder, tplFile, scaffold)
		written++
	}
	fmt.Printf("[HARNESS] %d card templates written to Cards/Templates/\n", written)
	fmt.Printf("[HARNESS] Fill in the templates before running 'harness phase3'.\n")

	// Run QA-1 Phase 2→3 gate.
	expState := parseExpansionQAState(expandDoc, queueJSON, cfg.IsStandalone)
	result := qa.Phase2To3(expState)

	fmt.Printf("\n─── QA-1: PHASE 2→3 GATE ──────────────────────────────────\n")
	fmt.Println(result.FormatFailures())
	fmt.Printf("───────────────────────────────────────────────────────────\n\n")

	if !result.Passed() {
		fmt.Println("[HARNESS] QA-1 FAILED. Resolve the items above before advancing to Phase 3.")
		return fmt.Errorf("QA-1 phase 2→3 gate failed: %d item(s) did not pass", len(result.FailedItems()))
	}

	// Gate passed: stamp the queue with Phase 2 complete status and re-save.
	cardQueue.PhaseGateStatus = queue.GatePhase2Complete
	queueJSON, err = cardQueue.Marshal()
	if err != nil {
		return fmt.Errorf("re-marshalling card queue after gate: %w", err)
	}
	if err := cfg.BoxWriter.Write(queueFolder, queueFile, string(queueJSON)); err != nil {
		return fmt.Errorf("updating card queue gate status in Box: %w", err)
	}
	if err := cfg.LocalWriter.Write(queueFolder, queueFile, string(queueJSON)); err != nil {
		fmt.Printf("[HARNESS] Warning: local queue gate update failed: %v\n", err)
	}

	fmt.Println("[HARNESS] QA-1 PASSED ✓  Phase 2 complete. Ready to advance to Phase 3.")
	return nil
}

// assembleExpansionDoc builds the full EXPAND-[BookTitle]-v1.md from step content.
func assembleExpansionDoc(bookTitle string, steps map[int]string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# EXPANSION — %s\n\n", bookTitle))
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n", time.Now().Format("2006-01-02")))
	sb.WriteString("**STATUS:** LOCKED\n\n---\n\n")
	for i := 1; i <= 7; i++ {
		content, ok := steps[i]
		if ok {
			// Replace DRAFT with LOCKED in step header.
			content = strings.ReplaceAll(content, "STATUS: DRAFT", "STATUS: LOCKED")
			sb.WriteString(content)
			sb.WriteString("\n\n---\n\n")
		}
	}
	return sb.String()
}

// parseRoster extracts character/faction/location lists from the AI's structured response.
func parseRoster(resp, seriesTitle, bookTitle string, isStandalone bool) queue.GenerateParams {
	p := queue.GenerateParams{
		SeriesTitle:  seriesTitle,
		BookTitle:    bookTitle,
		IsStandalone: isStandalone,
	}

	for _, line := range strings.Split(resp, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "FULL CHARACTERS:") {
			p.FullCharacters = splitList(strings.TrimPrefix(line, "FULL CHARACTERS:"))
		} else if strings.HasPrefix(line, "SKETCH CHARACTERS:") {
			p.SketchCharacters = splitList(strings.TrimPrefix(line, "SKETCH CHARACTERS:"))
		} else if strings.HasPrefix(line, "FULL FACTIONS:") {
			p.FullFactions = splitList(strings.TrimPrefix(line, "FULL FACTIONS:"))
		} else if strings.HasPrefix(line, "SKETCH FACTIONS:") {
			p.SketchFactions = splitList(strings.TrimPrefix(line, "SKETCH FACTIONS:"))
		} else if strings.HasPrefix(line, "FULL THREATS:") {
			p.FullThreats = splitList(strings.TrimPrefix(line, "FULL THREATS:"))
		} else if strings.HasPrefix(line, "SKETCH THREATS:") {
			p.SketchThreats = splitList(strings.TrimPrefix(line, "SKETCH THREATS:"))
		} else if strings.HasPrefix(line, "LOCATIONS:") {
			p.Locations = splitList(strings.TrimPrefix(line, "LOCATIONS:"))
		}
	}
	return p
}

func splitList(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" || strings.ToUpper(s) == "NONE" {
		return nil
	}
	var result []string
	for _, item := range strings.Split(s, ",") {
		item = strings.TrimSpace(item)
		if item != "" && strings.ToUpper(item) != "NONE" {
			result = append(result, item)
		}
	}
	return result
}

// parseExpansionQAState inspects the expansion document and card queue JSON
// to populate the QA-1 Phase2→3 checklist.
func parseExpansionQAState(doc string, queueJSON []byte, isStandalone bool) qa.ExpansionDocument {
	has := func(keyword string) bool {
		return strings.Contains(strings.ToLower(doc), strings.ToLower(keyword))
	}

	// Parse queue to check minimum card requirements.
	var q queue.Queue
	hasMinimum := false
	if err := json.Unmarshal(queueJSON, &q); err == nil {
		hasMinimum = checkQueueMinimum(q, isStandalone)
	}

	return qa.ExpansionDocument{
		ExistsInBox:            true,
		Step1Locked:            has("step 1 status: locked"),
		Step2Locked:            has("step 2 status: locked"),
		Step3Locked:            has("step 3 status: locked"),
		Step4Locked:            has("step 4 status: locked"),
		Step5Locked:            has("step 5 status: locked"),
		Step6Locked:            has("step 6 status: locked"),
		Step7Locked:            has("step 7 status: locked"),
		StoryCircleMapPresent:  has("beat 1") || has("beats 1-2") || has("story circle"),
		FortyChapterMapPresent: has("ch 1") || has("chapter 1") || has("40-chapter"),
		SevenPlotsConfirmed:    has("seven basic plots") || has("basic plots"),
		SubplotWeavePresent:    has("subplot"),
		FullTierCharsComplete:  has("card readiness") || has("enneagram"),
		SketchTierPresent:      has("sketch") || isStandalone,
		CardQueueGenerated:     len(queueJSON) > 10,
		CardQueueHasMinimum:    hasMinimum,
		IsStandalone:           isStandalone,
		NoNeedsReviewFlags:     !has("needs-review"),
		NoCardRequiredFlags:    !has("card-required"),
		NotContradictseed:      true, // AI is prompted to expand seed, not contradict it
	}
}

func checkQueueMinimum(q queue.Queue, isStandalone bool) bool {
	hasWorld, hasNovel, hasProtag, hasAntag, hasTrilogy := false, false, false, false, false
	for _, item := range q.Queue {
		switch item.CardType {
		case queue.CardTypeWorld:
			if item.CardSubtype == "Overview" {
				hasWorld = true
			}
		case queue.CardTypeNovel:
			hasNovel = true
		case queue.CardTypeTrilogy:
			hasTrilogy = true
		case queue.CardTypeCharacter:
			if !hasProtag {
				hasProtag = true
			} else {
				hasAntag = true
			}
		}
	}
	if !isStandalone {
		return hasTrilogy && hasWorld && hasNovel && hasProtag && hasAntag
	}
	return hasWorld && hasNovel && hasProtag && hasAntag
}
