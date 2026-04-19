Run a consistency check on an edited file against all locked canon.

**Arguments:** $ARGUMENTS — filename to check
Examples: `CHAR-Jax-Tarkin-v2.md` | `CH05-Shadows-Awaken-DRAFT-v1.md`

## Setup

1. Read `~/.writing-harness/session.json` and `config.json`. Hyphenate titles for paths.
2. Locate the target file (try in order):
   - Cards folder: `{SyncFolder}/{Series}/{Book}/Cards/{filename}`
   - Prose folder: `{SyncFolder}/{Series}/{Book}/Prose/{filename}`
   - Series Trilogy folder: `{SyncFolder}/{Series}/Trilogy/{filename}`
   - Error if not found.
3. Load all locked cards from `{SyncFolder}/{Series}/{Book}/Cards/` as reference canon.
4. For prose files (CH* or SCENE-*): also load the chapter card for that chapter number.

## Analysis

Compare the target file line by line against all locked canon. Look for:
- **Name / alias inconsistencies** — character names, place names, faction names
- **Factual contradictions** — ages, timelines, established history, technology rules
- **Relationship drift** — character relationships that contradict locked cards
- **Voice inconsistency** — POV drift, tone shifts for named characters
- **World rule violations** — setting rules established in World cards

Classify every finding:
- `CONFLICT` — direct contradiction with locked canon; must resolve before locking
- `WARNING` — possible inconsistency; writer should review
- `NOTE` — minor observation, suggestion, or question; no action required
- `CONSISTENT` — no issues found

## Report Format

```
## Consistency Check: {filename}
Checked against {N} locked cards.

### CONFLICTS
- [{section/line}]: {description — what contradicts what}

### WARNINGS
- [{section/line}]: {description}

### NOTES
- {observation}

### VERDICT
{CONSISTENT | N conflict(s), N warning(s) — resolve before locking}
```

## Follow-up Conversation

After the report, enter a conversation loop:
- Writer can ask about specific sections, request suggested fixes, or ask "how do I resolve X?"
- Propose concrete fixes for each CONFLICT
- Type **DONE** to end the check session
