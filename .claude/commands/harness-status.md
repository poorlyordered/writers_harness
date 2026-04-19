Show quick writing harness project status without loading full card content.

## Steps

1. Read `~/.writing-harness/session.json` → extract SeriesTitle, BookTitle, Phase, GatePhase1Complete, GatePhase2Complete, GatePhase3Complete.
2. Read `config.json` → extract `local.sync_folder`. Hyphenate titles for paths.
3. Read Card Queue: `{SyncFolder}/{Series}/{Book}/QA/CARD-QUEUE-{Book}.json`
   - Count total, complete, in-progress, not-started items.
   - Find next NOT_STARTED item (lowest Priority with all prereqs met).
4. Read last 5 lines of QA log: `{SyncFolder}/{Series}/{Book}/QA/QA-LOG-{Book}.md`
5. Display a clean status panel:

```
═══════════════════════════════════════
  WRITING HARNESS — STATUS
═══════════════════════════════════════
  Series : {SeriesTitle}
  Book   : {BookTitle}  (Book {BookNum})
  Phase  : {Phase}
  Gates  : P1={✓/✗}  P2={✓/✗}  P3={✓/✗}

  Queue  : {N}/{M} cards complete
  Next   : {CardType} — {Name}  (Priority {P})

  Recent QA:
  {last 3 QA log entries}
═══════════════════════════════════════
```

6. Ask: "What would you like to work on next?"
