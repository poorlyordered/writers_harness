Upload a local card or prose file to Box, splitting automatically if the file exceeds the ~10 KB Box MCP limit.

**Arguments:** $ARGUMENTS
Examples: `WORLD-Geography-Desolations-Edge-v2.md` | `CH03-Shadows-Awaken-DRAFT-v1.md`

## Setup

1. Parse arguments: filename (required). Resolve to absolute path under the local sync folder.
2. Confirm the file exists locally. If not, report and stop.
3. Run `wc -c <filepath>` to get byte count.

## Box Folder Routing

Determine target Box folder from the TYPE field in the file header:

- TYPE contains `CHAPTER PROSE` → Prose folder ID: `377580093518`
- All other types → Cards folder ID: `377580623644`

## Size Check and Upload Strategy

**If file is 10,000 bytes or under:** Upload directly as a single file using `upload_file`. Report the Box file ID and confirm done.

**If file is over 10,000 bytes:** Split into parts. Follow the procedure below.

## Split Procedure (files over 10 KB)

1. Run `wc -c` to get total bytes and `grep -n "^## " <filepath>` to get all section header line numbers and names.
2. Plan splits so each part is under 10,000 bytes. Split ONLY at `## ` section boundaries — never mid-section. Aim for roughly equal parts.
3. For each part, determine the line range using the section boundary lines identified in step 1.
4. Extract each part using `sed -n '{start},{end}p' <filepath>`.
5. Part 1 filename: original filename (e.g. `WORLD-Geography-Desolations-Edge-v2.md`)
6. Part 2+ filenames: insert `-p2`, `-p3` etc. before `.md` (e.g. `WORLD-Geography-Desolations-Edge-v2-p2.md`)
7. Part 2+ headers: prepend a brief continuation header so each file is self-identifying:
   ```
   STATUS: LOCKED
   TYPE: {original TYPE} (continued)
   VERSION: {original VERSION} — Part {N} of {total}
   LOCKED: {original LOCKED date}
   LOCATION: {original LOCATION if present}
   ---
   ```
8. Upload all parts in parallel using `upload_file` calls in a single message, all targeting the same Box folder ID.

## Reporting

After all uploads complete, report:

| File | Box ID | Size |
|------|--------|------|
| filename | ID | KB |

Confirm total parts uploaded and that the file is complete in Box.

## Error Handling

- If any upload returns an error or times out: report which part failed, do not retry automatically, ask the user whether to retry that part.
- If the file has no `## ` section headers and exceeds 10 KB: report the problem and ask the user how to split it before proceeding.
