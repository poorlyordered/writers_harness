Load the writing harness project context and all relevant cards to begin a co-creative session.

**Arguments:** $ARGUMENTS (optional — card type, chapter, or topic to focus on after loading)

## Steps

1. Read `~/.writing-harness/session.json` → extract SeriesTitle, BookTitle, BookNum, Phase, gate flags.
2. Read `config.json` in the current working directory → extract `local.sync_folder` (expand `~` to home).
3. Hyphenate all titles for file paths ("Shadows Awaken" → "Shadows-Awaken"). Store as `{Series}` and `{Book}`.
4. Cards folder: `{SyncFolder}/{Series}/{Book}/Cards/`
5. Load the following (use Glob for highest version, Read each, skip if missing):
   - Novel card: `NOVEL-{Book}-v*.md`
   - World Overview: `WORLD-Overview-v*.md`
   - All character cards: `CHAR-*.md`
   - Trilogy card: `{SyncFolder}/{Series}/Trilogy/TRILOGY-v*.md`
   - Expansion: `{SyncFolder}/{Series}/{Book}/Expansion/EXPAND-{Book}-v*.md` (highest version)
6. Read Card Queue: `{SyncFolder}/{Series}/{Book}/QA/CARD-QUEUE-{Book}.json`
   - Count total cards, complete cards, and identify next NOT_STARTED item.
7. Report to the user:
   - **Series / Book / Phase** and which phase gates are complete
   - **Queue:** N/M cards complete — next: {CardType} — {Name}
   - **Cards loaded:** list each file successfully read
8. If arguments were provided, proceed immediately to that task. Otherwise ask: "What would you like to work on?"

## File Naming Reference (from CLAUDE.md)
All titles hyphenated. Key patterns:
- `CHAR-{Name}-v{n}.md` | `WORLD-{Subtype}-v{n}.md` | `NOVEL-{Book}-v{n}.md`
- `CH{nn}-{Book}-v{n}.md` | `FACTION-{Name}-v{n}.md` | `THREAT-{Name}-v{n}.md`
