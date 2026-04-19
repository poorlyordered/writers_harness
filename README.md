# Writing Harness

A Go CLI that guides a writer through four phases of science fiction story development. Card-driven, phase-gated, AI-assisted. All persistent state lives in Box; Claude handles Box I/O natively via MCP tool use. Claude Code is the primary co-creative writing partner.

## How It Works

The harness structures the writing process as a pipeline with four phases:

```
new-series → phase1 → phase2 → phase3 → phase4
              Seed     Snowflake  Cards    Prose
              lock     expand     queue    draft
              QA-1     QA-1       QA-2/3   QA-4
```

Each phase gate runs a QA checklist that must pass before advancing. Cards (character, world, faction, threat, novel, chapter) accumulate as locked canon that Claude loads as context for every subsequent AI call. Prose is drafted scene-by-scene against that canon.

For ongoing co-creative work, Claude Code skills (slash commands) replace the CLI — see [Claude Code Skills](#claude-code-skills) below.

## Requirements

- Go 1.24+
- [Anthropic API key](https://console.anthropic.com)
- Box account with MCP server access (for cloud storage)

## Setup

```bash
git clone https://github.com/poorlyordered/writers_harness
cd writers_harness
go build -o harness .
cp config.example.json config.json
# Edit config.json — add your Anthropic API key and Box MCP credentials
```

## Configuration

Copy `config.example.json` to `config.json` and fill in:

```json
{
  "anthropic": {
    "model": "claude-sonnet-4-6",
    "max_tokens": 8192
  },
  "box": {
    "mcp_server_url": "...",
    "mcp_auth_token": "...",
    "root_folder_id": "..."
  },
  "local": {
    "sync_folder": "~/writing-harness-sync"
  },
  "preferences": {
    "scene_word_target_default": 2000
  }
}
```

`config.json` is gitignored — never commit it.

## CLI Commands

### First-pass pipeline

```bash
harness new-series      # Create Box folder tree + Series Index
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

```bash
harness status          # Queue progress, next card, QA log summary
harness sync            # Re-push SYNC-PENDING files to Box
harness qa-check 2 <file>   # Validate a card file
harness qa-check 3          # Run QA-3 pre-prose gate
```

## Claude Code Skills

Open this project in Claude Code and use these slash commands for all ongoing creative work. Claude loads your session state and card files, then works as a direct co-author.

| Command | Purpose |
|---------|---------|
| `/harness-context` | Load session + all cards — start every writing session here |
| `/harness-status` | Quick queue and phase status |
| `/harness-idea` | Brainstorm freely — nothing saves unless you type `SAVE` |
| `/harness-build --card CHAR --name "Name"` | Build or revise a control card |
| `/harness-write --chapter 5 [--scene 2]` | Draft prose with full card context |
| `/harness-check <filename>` | Consistency check against locked canon |

**Session keywords:**
- `LOCK` / `CONFIRM` — save and lock the current card or scene
- `SAVE` — (idea mode) write current idea as a note
- `REJECT [reason]` — discard draft and redraft
- `EXIT` / `DONE` — end the session

## Card Types

| Card | Description |
|------|-------------|
| SEED | The originating story premise (Phase 1) |
| NOVEL | Book-level story bible |
| TRILOGY | Series-spanning arc card |
| WORLD | Setting cards (Overview, History, Geography, etc.) |
| CHARACTER | Full or Sketch character profiles |
| FACTION | Organization/group cards |
| THREAT | Antagonist/conflict arc cards |
| CHAPTER | Per-chapter scene breakdown and beat map |

Cards are versioned (`CHAR-Jax-Tarkin-v2.md`) and locked with a status header. Locked cards feed into every subsequent AI conversation as established canon.

## Box Folder Structure

```
{Series}/
  World Bible/
  Trilogy/
  INDEX-{Series}.md
  {Book}/
    Seeds/
    Expansion/
    Cards/
    CardTemplates/
    Prose/
    QA/
```

## Testing

```bash
go test ./...
```

Unit tests cover file naming, version parsing, card validation, card locking, queue generation, and QA checklists. AI-dependent code (phases, prose, API client) requires a live config to test.

## Project Layout

```
cmd/                    # Cobra subcommands
internal/
  anthropic/            # API client + token budget
  cards/                # Card schemas, builder, validator, versioner, scaffold
  config/               # Config struct + loader
  index/                # Series Index (Markdown + JSON state)
  phases/               # Phase runners (1–4) + mode runners (idea, check, write)
  prose/                # Scene drafting, fast draft, assembly, pattern log
  qa/                   # QA-1/2/3/4 checklists + resolution + log
  queue/                # Card queue schema, generator, progress tracking
  session/              # Session state persistence
  storage/              # Local sync folder + Box sync pending log
  utils/                # File naming, version parsing, prompt loader, diff view
prompts/
  system/               # System prompts for each phase and mode
  templates/            # Reusable AI prompt templates
.claude/commands/       # Claude Code slash command skill definitions
```
