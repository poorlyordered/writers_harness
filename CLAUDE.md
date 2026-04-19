# Writing Harness — Codebase Guide

Go CLI that guides a writer through 4 phases of science fiction story development using a card-driven, phase-gated, AI-assisted process. All persistent state lives in Box; the Go app injects Box MCP config into every Anthropic API request so Claude handles Box I/O natively.

## Commands

### Phase commands (first-pass pipeline)
```
harness new-series      # Create Box folder tree + Series Index, set session to Phase 1
harness new-book        # Add a second book to an existing series
harness phase1          # Phase 1: Idea Generation (seed prompt + deepening conversation)
harness phase2          # Phase 2: Snowflake expansion (7 steps) + Card Queue + templates
harness phase3          # Phase 3: Card Completion (queue-driven, AI drafts each card)
harness phase4          # Phase 4: Prose Generation (scene-by-scene, 40 chapters)
harness phase4 --fast   # Fast-draft mode: draft full chapter then review
harness phase4 --chapter 12  # Resume at chapter 12
harness continue        # Resume most recent session (any phase)
```

### Mode commands (ongoing, any phase)
```
harness idea                              # Brainstorm freely — nothing saves unless you type SAVE
harness build --card CHAR --name "Name"   # Create or update a specific card
harness build --card WORLD --subtype Overview
harness build --card NOVEL
harness write --chapter 5                 # Draft all scenes in chapter 5
harness write --chapter 5 --scene 2       # Draft only scene 2 of chapter 5
harness check <filename>                  # Post-edit consistency check against locked canon
```

### Utility commands
```
harness status          # Show queue progress, next card, QA log summary
harness sync            # Re-push SYNC-PENDING files to Box
harness qa-check 2 <file>  # Validate a card file from local sync
harness qa-check 3      # Run QA-3 pre-prose gate against current queue
```

Config is read from `config.json` in the working directory (see `config.example.json`).

## Architecture

### Conversation Loop

The Go CLI manages session state and conversation history. For each turn:
1. Append user message to `[]BetaMessageParam` history
2. Build API request: system prompt + history + Box MCP server config
3. Call Anthropic Beta Messages API — Claude may invoke Box tools mid-turn
4. Append assistant response to history
5. Present to writer; repeat

The Go app **never calls Box directly** — Claude does, via MCP tool use.

### Box MCP Passthrough

`internal/box/mcp.go` builds a `BetaRequestMCPServerURLDefinitionParam` injected into every `BetaMessageNewParams.MCPServers`. Beta header: `AnthropicBetaMCPClient2025_04_04`.

`BoxFileWriter` (in `internal/phases/adapters.go`) writes files by sending Claude a prompt instructing it to use Box MCP tools to create the file.

### Phase Flow

```
new-series ──► phase1 ──► phase2 ──► phase3 ──► phase4
               Seed       Snowflake  Card        Prose
               lock       expand     queue       draft
               QA-1       QA-1       QA-2/3      QA-4
```

QA gates block advancement. The writer can retry, override-with-log, or halt.

## Key Packages

### `internal/anthropic`

- `client.go` — `Client` wraps the Anthropic Go SDK; `Send()` manages history + MCP config;  `ResetHistory()` starts a clean conversation for each card/scene
- `tokens.go` — estimates token usage (chars/4), trims oldest 2 turns at 90% window, warns at 80%

### `internal/phases`

- `adapters.go` — `AIConversation`, `BoxFileWriter`, `LocalFileWriter`; `Conversation`/`FileWriter`/`PromptLoader` interfaces used across phase runners
- `phase1.go` — 3 entry paths (A: generate, B: import, C: freeform); deepening conversation loop; seed lock; QA-1 gate
- `phase2.go` — 7 Snowflake steps, each detected via `"STEP N STATUS: LOCKED"` in AI response; Card Queue generation; QA-1 Phase2→3 gate
- `phase3.go` — queue-driven card loop: `buildOneCard()` → `cards.Build()` → `cards.Validate()` → `qa.CardCompletion()` → `qa.PresentAndChoose()` → save; QA-3 gate at end
- `phase4.go` — 40-chapter loop; `draftChapterScenes()` per chapter; LOCK/REJECT detection; three-rejection diagnostic; chapter assembly; manuscript assembly; `--chapter` resume

### `internal/cards`

- `types.go` — card type/tier/status constants; `Failure`, `Section`, `Schema` types
- `schemas.go` — `allSchemas` slice with required `##` headings for every card type; `SchemaFor()` lookup; `RequiredSectionsList()` for AI instructions
- `builder.go` — `Build()`: reset history → AI drafts card → conversation loop until `LOCK`/`CONFIRM`
- `validator.go` — `Validate()`: checks `[TBD]`/`[TODO]`/`placeholder` text + presence and non-emptiness of all required `##` sections
- `versioner.go` — `Lock()` prepends status header; `IsLocked()` detects lock; `ExtractBody()` strips header

### `internal/queue`

- `queue.go` — `Queue`/`QueueItem` SPEC-007 §10 schema; `Next()` returns first NOT_STARTED item with all prereqs complete; `MarkComplete()` records locked filename and date; `recalcProgress()` stats
- `generator.go` — `Generate()` builds the full priority-ordered queue from Phase 2 roster: Trilogy → World cards → Geography(n) → Novel → Characters(FULL) → Factions(FULL) → Threats(FULL) → SKETCH tier → Chapter/Scene cards → CONSTRAINTS (always last)

### `internal/qa`

- `qa1.go` — `Phase1To2()` (18 items) and `Phase2To3()` (19 items) phase gate checklists; `Result`, `CheckItem`, `Passed()`, `FailedItems()`, `FormatFailures()`
- `qa2.go` — `CardCompletion()`: wraps `cards.Validate()` failures + existence/lock checks + card-type-specific items (pillars, psychology, Voice Anchor for CHARACTER; leadership for FACTION; escalation arc for THREAT; chapter map for NOVEL)
- `qa3.go` — `PreProse()` 9-item pre-prose readiness gate (all FULL tier locked, NOVEL/WORLD-Overview/Constraints locked, queue complete)
- `qa4.go` — `PostProse()` 10-item per-scene check + `PostProseChapter()` chapter-level check
- `resolution.go` — `PresentAndChoose()`: shows failures, returns `ResolveRetry`/`ResolveOverride`/`ResolveBlock` + overridden item IDs
- `log.go` — `Log`: append-only JSON-lines QA log; `AppendResult()` convenience wrapper; `Summary()` pass/fail counts

### `internal/prose`

- `generator.go` — `DraftScene()`: AI conversation with `LOCK`/`REJECT [reason]` detection; `WordCount()` helper
- `fast_draft.go` — `FastDraftChapter()`: drafts all scenes without pausing, presents assembled chapter for selective revision
- `assembler.go` — `AssembleChapter()` / `AssembleManuscript()` concatenation; `TransitionCheck()` builds a continuity-check prompt
- `pattern_log.go` — `PatternLog`: JSON rejection tracker; `ThreeRejectionPattern()` surfaces diagnostic when 3 consecutive rejections occur

### `internal/utils`

- `files.go` — all SPEC-007 §4 naming functions (folder paths + file names); `title()` helper converts "Shadows Awaken" → "Shadows-Awaken"
- `versions.go` — `ParseVersion()` extracts int from `-v3.md`; `IncrementVersion()` bumps it; `HighestVersion()` picks max from a list
- `prompts.go` — `Loader`: reads `{name}.txt` from prompts dir; `LoadWithVars()` replaces `{{KEY}}` placeholders
- `diff.go` — `DiffView()` two-column diff for Trilogy Card updates

### `internal/storage`

- `local.go` — `Local`: read/write local sync folder mirroring Box structure; `AbsPath()` for QA log path
- `sync.go` — `SyncPending`: JSON-lines log of locally-written files that need Box sync; `Add()`/`Clear()`; `harness sync` re-pushes entries

### `internal/session`

`State` persists phase/series/book/bookNum; `Manager` reads/writes `~/.writing-harness/session.json`

### `internal/index`

`Index`/`BookSection` SPEC-007 §11 Series Index; `Render()` → Markdown table; `EnsureBook()` / `AddBookCard()` / `AddSeriesCard()`

### `internal/config`

`Config` struct with `AnthropicConfig`, `BoxConfig`, `LocalConfig`, `DefaultsConfig`, `PreferencesConfig`; `Load()` validates and applies defaults

## File Naming (SPEC-007 §4)

| Card | Pattern |
|------|---------|
| Seed | `SEED-{Book}-v{n}.md` |
| Expansion | `EXPAND-{Book}-v{n}.md` |
| Card Queue | `CARD-QUEUE-{Book}.json` |
| WORLD | `WORLD-{Subtype}-v{n}.md` |
| WORLD-Geography | `WORLD-Geography-{Location}-v{n}.md` |
| NOVEL | `NOVEL-{Book}-v{n}.md` |
| CHARACTER | `CHAR-{Name}-v{n}.md` |
| FACTION | `FACTION-{Name}-v{n}.md` |
| THREAT | `THREAT-{Name}-v{n}.md` |
| CHAPTER | `CH{nn}-{Book}-v{n}.md` |
| SCENE | `SCENE-CH{nn}-S{n}-v{n}.md` |
| TRILOGY | `TRILOGY-v{n}.md` |
| QA Log | `QA-LOG-{Book}.md` |
| Pattern Log | `PATTERN-LOG-{Book}.md` |

All titles are hyphenated: "Shadows Awaken" → "Shadows-Awaken".

## Box Folder Structure (SPEC-007 §3)

```
{Series}/
  World Bible/
  Trilogy/
  INDEX-{Series}.md
  {Book}/
    Seeds/
    Expansion/
    Cards/
    Prose/
    QA/
```

## Prompt System

`prompts/system/phase{1-4}.txt` — system prompts loaded via `utils.Loader`; support `{{VAR}}` substitution.

`prompts/templates/` — reusable templates: `card_draft.txt`, `gap_resolution.txt`, `qa_failure.txt`, `scene_orient.txt`, `three_rejection.txt`.

## Testing

```
go test ./...                    # run all tests
go test ./internal/cards/...     # cards package only
go test ./internal/queue/...     # queue package only
```

Tests cover: file naming (utils), version parsing (utils), card validation (cards), card locking (cards), queue generation (queue), queue logic (queue), QA-1/2/3 checklists (qa).

AI-dependent code (phases, prose, anthropic client) is not unit-tested — test with a live config.

## Development Notes

- The Anthropic Go SDK returns `Client` by value from `NewClient()`, not a pointer
- `BetaMessageParam.Content` is `[]BetaContentBlockParamUnion` with `OfText *BetaTextBlockParam`
- MCP server config goes in `BetaMessageNewParams.MCPServers` (not `Params.Tools`)
- `param.NewOpt()` is at `github.com/anthropics/anthropic-sdk-go/packages/param`
- Model: `claude-sonnet-4-6` (set in config.example.json)
