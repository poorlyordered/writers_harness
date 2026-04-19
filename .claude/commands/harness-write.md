Draft prose for a specific chapter or scene.

**Arguments:** $ARGUMENTS
Examples: `--chapter 5` | `--chapter 5 --scene 2` | `--chapter 5 --scene 2 --words 2500`

## Setup

1. Parse arguments: `--chapter N` (required, 1–40), `--scene N` (optional, 0 = all scenes), `--words N` (default 2000).
2. Read `~/.writing-harness/session.json` and `config.json`. Hyphenate titles for paths.
3. Load context from `{SyncFolder}/{Series}/{Book}/Cards/` (highest version of each):
   - Chapter card: `CH{nn}-{Book}-v*.md` where nn = zero-padded chapter number
   - Novel card: `NOVEL-{Book}-v*.md`
   - World Overview: `WORLD-Overview-v*.md`
   - All `CHAR-*.md` files
4. Check for existing draft in `{SyncFolder}/{Series}/{Book}/Prose/`:
   - Scene: `CH{nn}-{Book}-S{n}-DRAFT-v*.md`
   - Chapter: `CH{nn}-{Book}-DRAFT-v*.md`
   - If found: "Existing draft found. [R]evise / [N]ew draft / [E]xit?" — wait for choice.

## Your Role as Prose Writer

You are writing vivid, character-driven science fiction prose. The chapter card is your blueprint — follow its scene breakdown and beat structure strictly.

Every scene requires:
- **Opening hook** — the first sentence earns the reader's attention; drop into action or tension
- **Turn point** — something changes: status, relationship, information, or understanding
- **Closing beat** — a line that propels the reader into the next scene

Target word count: `--words` value per scene. Write to at least 70% of target. Show word count after each scene draft.

Draft each scene fully. Show it to the writer. Iterate on feedback. When the chapter card specifies POV, voice, or mood — hold to it.

## Saving

When the writer types **LOCK**:
1. Prose folder: `{SyncFolder}/{Series}/{Book}/Prose/`
2. Find next version by listing files with matching prefix.
3. Write:
   - Single scene: `CH{nn}-{Book}-S{n}-DRAFT-v{N}.md`
   - Full chapter: `CH{nn}-{Book}-DRAFT-v{N}.md`
4. Confirm: "Scene/Chapter saved ✓  Prose/{filename}"

When the writer types **REJECT [reason]**:
- Acknowledge the reason, note it, and redraft from scratch.
- After 3 consecutive rejections on the same scene, surface the pattern and ask: "What's not working? Let's diagnose before redrafting."
