Build or update a specific control card for the active project.

**Arguments:** $ARGUMENTS
Examples: `--card CHAR --name "Jax Tarkin"` | `--card WORLD --subtype Overview` | `--card NOVEL` | `--card FACTION --name "The Conclave"`

## Setup

1. Parse arguments: `--card TYPE` (required), `--name NAME` (optional), `--subtype SUBTYPE` (optional).
2. Read `~/.writing-harness/session.json` and `config.json`. Hyphenate titles for paths.
3. Compute the base filename from card type:
   - `CHAR` / `CHARACTER` → `CHAR-{Name}-v1.md` (requires --name)
   - `WORLD` (non-Geography) → `WORLD-{Subtype}-v1.md` (requires --subtype)
   - `WORLD` + `--subtype Geography` → `WORLD-Geography-{Name}-v1.md` (requires --name)
   - `NOVEL` → `NOVEL-{Book}-v1.md`
   - `TRILOGY` → `TRILOGY-v1.md`
   - `FACTION` → `FACTION-{Name}-v1.md` (requires --name)
   - `THREAT` → `THREAT-{Name}-v1.md` (requires --name)
   - `CHAPTER` → error: use `harness-write` instead
4. Cards folder: `{SyncFolder}/{Series}/{Book}/Cards/`

## Mode Detection (check in this order)

**Revision mode** — existing file contains `STATUS: LOCKED`:
- Load the card body (strip the status header).
- Tell the user: "Found existing locked card — entering revision mode."

**Template mode** — file exists in `{SyncFolder}/{Series}/{Book}/CardTemplates/` and contains writer content beyond `[Write content here]` placeholders:
- Load the template.
- Tell the user: "Writer template found — I'll refine it."

**Build mode** — no existing locked card, no template:
- Build fresh from scratch.

## Context Loading

Load all locked cards from the Cards folder plus the Expansion document as background canon before drafting.

## Card Conversation

Use the required `##` section headings for the card type (see CLAUDE.md schemas). Draft each section fully — no `[TBD]` or placeholder text. Show the complete draft. Iterate based on writer feedback until the writer is satisfied.

## Locking

When the writer types **LOCK** or **CONFIRM**:
1. List existing versioned files matching the base name to find the next version number.
2. Prepend the lock header:
   ```
   STATUS: LOCKED
   TYPE: {CardType}
   VERSION: v{N}
   LOCKED: {YYYY-MM-DD}
   ---
   ```
3. Write to `{CardsFolder}/{base-name-with-vN}.md` (e.g. `CHAR-Jax-Tarkin-v2.md`).
4. Confirm: "Card locked ✓  {Cards folder}/{filename}"
5. Note: run `harness sync` if Box sync is needed.
