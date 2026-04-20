# Writing Harness — Claude Code Guide

Go CLI that guides a writer through 4 phases of science fiction story development. Card-driven, phase-gated, AI-assisted. Claude Code is the primary co-creative writing partner; the harness CLI manages state, QA gates, and Box sync.

---

## Claude Code Skills (Slash Commands)

Start every session with `/harness-context`. It loads your session state and all locked cards so Claude has full project context before any creative work begins.

| Skill | Purpose |
|-------|---------|
| `/harness-context` | Load session state + all cards — **start here** |
| `/harness-status` | Quick queue/phase status |
| `/harness-idea [topic]` | Brainstorm freely — nothing saves unless you type `SAVE` |
| `/harness-build --card TYPE [--name N] [--subtype S]` | Build or revise a control card |
| `/harness-write --chapter N [--scene N] [--words N]` | Draft prose scene-by-scene |
| `/harness-check <filename>` | Consistency check against all locked canon |

**Session keywords (work in any skill):**

| Keyword | Effect |
|---------|--------|
| `LOCK` or `CONFIRM` | Save and lock the current card or scene draft |
| `SAVE` | Write current idea as a note (idea mode only) |
| `REJECT [reason]` | Discard draft, note the reason, redraft |
| `EXIT` or `DONE` | End the session cleanly |

**Typical session flow:** `/harness-context` → create → `LOCK` → `harness sync`

---

## Your Role as Co-Author

You are a skilled science fiction co-author, not a writing assistant. That means:

- **Generate boldly.** Propose complete drafts, not suggestions. The writer's job is to react and redirect, not to fill in blanks you left empty.
- **Hold canon absolutely.** Once a card is locked (`STATUS: LOCKED`), its content is inviolable. Every character name, relationship, world rule, and timeline fact in locked cards is ground truth. Never contradict it without a `harness-check` conversation first.
- **Write prose, not prose about prose.** No meta-commentary ("here is a scene where..."), no summaries in place of scenes. Every prose output should be ready to read.
- **Scene structure is non-negotiable.** Every scene: opening hook → turn point → closing beat. The hook earns the reader's attention in the first sentence. The turn changes something (status, relationship, information, understanding). The closing beat propels forward.
- **Voice consistency.** Each character has a Voice Anchor in their CHARACTER card. Hold to it across every scene they appear in.
- **Word count targets are real targets.** Aim for at least 70% of the specified word count. A 2000-word scene target means 1400+ words minimum.

---

## Card System

Cards are the project's source of truth. They accumulate across phases and are loaded as context for every creative decision. A card that passes QA and is locked with `STATUS: LOCKED` becomes permanent canon.

### Card Lock Header

Every locked card begins with:
```
STATUS: LOCKED
TYPE: {CardType}
VERSION: v{N}
LOCKED: {YYYY-MM-DD}
---
```

### Card Schemas — Required Sections

All required `##` headings must be present and fully written. No `[TBD]`, `[TODO]`, or `placeholder` text in locked cards.

**TRILOGY**
- `## Series Arc`
- `## Book-by-Book Progression`
- `## Series-Level Antagonist Ladder`
- `## Series Thematic Questions`

**WORLD — Overview**
- `## Setting`
- `## Tone and Genre`
- `## Core Concept`
- `## Thematic Backdrop`
- `## Reader Experience`

**WORLD — History**
- `## Timeline` / `## Key Events` / `## World-Shaping Conflicts` / `## Current State`

**WORLD — Political**
- `## Power Structure` / `## Governing Bodies` / `## Faction Landscape` / `## Political Tensions`

**WORLD — Technology**
- `## Technology Level` / `## Key Technologies` / `## Social Impact` / `## Constraints and Limits`

**WORLD — Culture**
- `## Social Norms` / `## Cultural Values` / `## Class Structure` / `## Language Notes`

**WORLD — Geography** (one card per named location)
- `## Physical Description` / `## Climate and Conditions` / `## Strategic Significance` / `## Notable Features`

**WORLD — Constraints**
- `## Continuity Rules` / `## Established Facts` / `## Series Constraints`

**NOVEL**
- `## Story Summary`
- `## Act Structure`
- `## 40-Chapter Map`
- `## Protagonist Journey`
- `## Central Thematic Question`

**CHARACTER — FULL** (protagonist, primary antagonist, key secondaries)
- `## Identity`
- `## Want` / `## Need` / `## Fear` / `## Misbelief` / `## Wound`
- `## Arc Progression`
- `## Psychology`
- `## Voice and Mannerisms`
- `## Skills and Knowledge`
- `## Physical Description`
- `## Hard Limits`
- `## Relationship Dynamics`
- `## Active Chapters`
- `## Voice Anchor`

**CHARACTER — SKETCH** (minor/background characters)
- `## Identity` / `## Wound` / `## Misbelief` / `## Archetype Stage` / `## Thematic Role` / `## Estimated Full Appearance`

**FACTION — FULL**
- `## Overview` / `## Leadership` / `## Goals and Methods` / `## Internal Tensions` / `## Relationship to Protagonist` / `## Relationship to Threats`

**FACTION — SKETCH**
- `## Overview` / `## Leadership` / `## Goals` / `## Estimated Full Appearance`

**THREAT — FULL**
- `## Nature and Scope` / `## Origin` / `## Methods` / `## Escalation Arc` / `## Connection to Antagonist`

**THREAT — SKETCH**
- `## Overview` / `## Nature` / `## Estimated Full Appearance`

**CHAPTER**
- `## Chapter Type`
- `## Story Circle Beats`
- `## POV Character`
- `## Scene Beats`
- `## Subplot Threads` *(optional)*

**SCENE**
- `## Chapter Reference` / `## Story Function` / `## Opening Hook` / `## Turn Point` / `## Closing Beat` / `## Characters Present`

---

## QA Standards (Plain Language)

**Card QA (QA-2):** A card passes when every required `##` section exists, is non-empty, and contains no placeholder text (`[TBD]`, `[TODO]`, `[placeholder]`, `[Write content here]`). Card-type-specific checks: CHARACTER needs a Voice Anchor; FACTION needs Leadership; THREAT needs an Escalation Arc; NOVEL needs a Chapter Map.

**Pre-Prose Gate (QA-3):** All FULL-tier cards locked. NOVEL, WORLD-Overview, and WORLD-Constraints cards locked. Card queue 100% complete.

**Prose QA (QA-4):** Each scene hits 70%+ of word target, has a clear opening hook, a turn point, and a closing beat. No new unintroduced characters. POV and voice consistent throughout. No contradictions with locked canon. No `[FLAG]` markers left in the text.

---

## File Naming

All titles hyphenated: "Shadows Awaken" → `Shadows-Awaken`.

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
| SCENE draft | `CH{nn}-{Book}-S{n}-DRAFT-v{n}.md` |
| Chapter draft | `CH{nn}-{Book}-DRAFT-v{n}.md` |
| TRILOGY | `TRILOGY-v{n}.md` |
| QA Log | `QA-LOG-{Book}.md` |
| Pattern Log | `PATTERN-LOG-{Book}.md` |
| Index state | `INDEX-STATE-{Series}.json` |
| Card template | `{base-filename}-template.md` |

## Box Folder Structure

```
{Series}/
  World Bible/
  Trilogy/
  INDEX-{Series}.md
  INDEX-STATE-{Series}.json
  {Book}/
    Seeds/
    Expansion/
    Cards/
    CardTemplates/
    Prose/
    QA/
```

---

## CLI Reference

### Phase pipeline (first-pass, run once per book)
```
harness new-series      # Box folder tree + Series Index, sets Phase 1
harness new-book        # Add a second book to an existing series
harness phase1          # Idea Generation: seed prompt + deepening conversation
harness phase2          # Snowflake expansion (7 steps) + Card Queue + templates
harness phase3          # Card Completion: queue-driven, AI drafts each card
harness phase4          # Prose Generation: scene-by-scene, 40 chapters
harness phase4 --fast   # Fast-draft mode
harness phase4 --chapter 12  # Resume at chapter 12
harness continue        # Resume most recent session
```

### Utility
```
harness status          # Queue progress, next card, QA log summary
harness sync            # Re-push SYNC-PENDING files to Box
harness qa-check 2 <file>   # Validate a card file from local sync
harness qa-check 3          # Run QA-3 pre-prose gate
```

Config is read from `config.json` in the working directory (see `config.example.json`).

---

## Codebase Reference

### Architecture

The Go CLI manages session state and conversation history. For each turn:
1. Append user message to `[]BetaMessageParam` history
2. Build API request: system prompt + history + Box MCP server config
3. Call Anthropic Beta Messages API — Claude may invoke Box tools mid-turn
4. Append assistant response to history
5. Present to writer; repeat

The Go app **never calls Box directly** — Claude does, via MCP tool use.

Box MCP: `internal/box/mcp.go` builds `BetaRequestMCPServerURLDefinitionParam` injected into `BetaMessageNewParams.MCPServers`. Beta header: `AnthropicBetaMCPClient2025_04_04`.

### Phase Flow

```
new-series ──► phase1 ──► phase2 ──► phase3 ──► phase4
               Seed       Snowflake  Card        Prose
               lock       expand     queue       draft
               QA-1       QA-1       QA-2/3      QA-4
```

### Key Packages

**`internal/anthropic`** — `Client` wraps SDK; `Send()` manages history + MCP; `ResetHistory()` per card/scene; token budget trims at 90%, warns at 80%.

**`internal/phases`** — `adapters.go`: `AIConversation`, `BoxFileWriter`, `LocalFileWriter`, shared interfaces. Phase runners: `phase1.go` (3 entry paths, deepening loop, seed lock), `phase2.go` (7 Snowflake steps, queue generation), `phase3.go` (queue-driven card loop), `phase4.go` (40-chapter loop, fast draft). Mode runners: `idea.go`, `check.go`, `write.go`.

**`internal/cards`** — `types.go`: constants. `schemas.go`: required `##` headings per card type. `builder.go`: `Build()` and `BuildFromTemplate()` conversation loops. `validator.go`: `Validate()` checks placeholders + section presence. `versioner.go`: `Lock()` / `IsLocked()` / `ExtractBody()`. `scaffold.go`: `GenerateScaffold()` / `IsScaffoldOnly()`.

**`internal/queue`** — `Queue`/`QueueItem` schema; `Next()` returns first NOT_STARTED item; `MarkComplete()` records filename + date. `generator.go`: priority order: Trilogy → World → Geography → Novel → Characters(FULL) → Factions → Threats → SKETCH tier → Chapters → CONSTRAINTS.

**`internal/qa`** — `qa1.go`: Phase1→2 (18 items) and Phase2→3 (19 items) gates. `qa2.go`: card completion. `qa3.go`: pre-prose gate + `PreProseDocFromQueue()`. `qa4.go`: per-scene + chapter-level checks. `resolution.go`: `PresentAndChoose()` → Retry/Override/Block. `log.go`: append-only JSON-lines log.

**`internal/prose`** — `generator.go`: `DraftScene()` with LOCK/REJECT detection. `fast_draft.go`: `FastDraftChapter()`. `assembler.go`: chapter + manuscript assembly + transition check. `pattern_log.go`: three-rejection diagnostic.

**`internal/utils`** — `files.go`: all path/filename functions. `versions.go`: `ParseVersion()`, `IncrementVersion()`, `HighestVersion()`. `prompts.go`: `Loader` with `{{VAR}}` substitution. `diff.go`: Trilogy Card two-column diff.

**`internal/storage`** — `local.go`: local sync folder read/write + `AbsPath()`. `sync.go`: SYNC-PENDING log.

**`internal/session`** — `State`: phase/series/book/bookNum/gates. `Manager`: reads/writes `~/.writing-harness/session.json`.

**`internal/index`** — `Index`/`BookSection` Series Index; `Render()` → Markdown; `EnsureBook()` / `AddBookCard()` / `AddSeriesCard()`; `LoadState()` / `SaveState()` JSON round-trip.

**`internal/config`** — `Config` with `AnthropicConfig`, `BoxConfig`, `LocalConfig`, `DefaultsConfig`, `PreferencesConfig`; `Load()` validates and applies defaults.

### Testing

```bash
go test ./...                    # all tests
go test ./internal/cards/...     # cards only
go test ./internal/queue/...     # queue only
```

Tests cover: file naming, version parsing, card validation, card locking, queue generation, queue logic, QA-1/2/3 checklists. AI-dependent code is not unit-tested — use a live config.

### Development Notes

- Anthropic Go SDK returns `Client` by value from `NewClient()`, not a pointer
- `BetaMessageParam.Content` is `[]BetaContentBlockParamUnion` with `OfText *BetaTextBlockParam`
- MCP server config goes in `BetaMessageNewParams.MCPServers` (not `Params.Tools`)
- `param.NewOpt()` is at `github.com/anthropics/anthropic-sdk-go/packages/param`
- Model: `claude-sonnet-4-6` (set in `config.example.json`)
- `internal/ai/interface.go` — canonical shared `Conversation` interface; all packages import from here, not from each other
- **Stream idle timeout fix:** When writing large file content (synopses, full card drafts), use Bash heredocs in small sections (one act or section at a time) appended sequentially — never generate large content as text output or in a single Write tool call. Box uploads over ~18 KB must be split into parts.
