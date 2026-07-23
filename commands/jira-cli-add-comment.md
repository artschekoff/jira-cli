---
name: jira-cli-add-comment
description: 'Add a comment to a Jira issue via jira-cli'
targets: ["*"]
---

# Add Jira Comment

Uses `jira-cli comment add`. JSON on stdout, errors on stderr.

## 1. Auth

```bash
jira-cli auth status
```

Not authenticated → stop: *Run `jira-cli auth login`.*

## 2. Resolve inputs

Need `$KEY` plus the comment content.

- **Key** — user provided a URL (`https?://.*/browse/([A-Z]+-\d+)`), a bare key (`^[A-Z]+-\d+$`), or digits only (`^\d+$` → prepend `JH-`). Anything else → ask for the key.
- **Content** — take the user's text verbatim. Missing/empty → ask once: *"Comment text?"* Do not paraphrase or invent content.

### Plain vs formatted — pick the mode

`acli` does NOT render Jira wiki markup (`h2.`, `||table||`, `* bullet`) or Markdown in comments — it posts them as literal characters. So:

- **Unformatted text** → `--body` (§4a). Never put wiki/markdown syntax in `--body`; it shows up raw.
- **Anything with headings, bold, bullet/numbered lists, tables, or code blocks** → build an **ADF** file and use `--adf-file` (§4b). This is the only path that renders.

## 3. Confirm (one screen)

```
Adding comment to $KEY:
  $BODY
```

**Post this comment? (yes / edit / cancel)**

Skip confirmation only if the user's original message clearly said "post" / "add" with the full body inline.

## 4a. Post — plain text

```bash
jira-cli comment add $KEY --body "$BODY"
```

Multi-line body → pass via `$(cat <<'EOF' ... EOF)` heredoc to preserve newlines and quotes.

## 4b. Post — formatted (ADF)

Write the comment as an [ADF](https://developer.atlassian.com/cloud/jira/platform/apis/document/structure/) document to a temp file, then pass its path. jira-cli posts a stub and updates it with `--body-adf` under the hood.

```bash
cat > /tmp/comment.adf.json <<'EOF'
{
  "version": 1,
  "type": "doc",
  "content": [
    { "type": "heading", "attrs": { "level": 2 },
      "content": [ { "type": "text", "text": "Implementation update" } ] },
    { "type": "paragraph",
      "content": [ { "type": "text", "text": "buildMetricValueIdMap() no longer loads the whole table." } ] },
    { "type": "bulletList", "content": [
      { "type": "listItem", "content": [ { "type": "paragraph",
        "content": [ { "type": "text", "text": "location_npi = 0" } ] } ] } ] },
    { "type": "table", "content": [
      { "type": "tableRow", "content": [
        { "type": "tableHeader", "content": [ { "type": "paragraph", "content": [ { "type": "text", "text": "Metric" } ] } ] },
        { "type": "tableHeader", "content": [ { "type": "paragraph", "content": [ { "type": "text", "text": "Before" } ] } ] },
        { "type": "tableHeader", "content": [ { "type": "paragraph", "content": [ { "type": "text", "text": "After" } ] } ] } ] },
      { "type": "tableRow", "content": [
        { "type": "tableCell", "content": [ { "type": "paragraph", "content": [ { "type": "text", "text": "rows loaded" } ] } ] },
        { "type": "tableCell", "content": [ { "type": "paragraph", "content": [ { "type": "text", "text": "58,113" } ] } ] },
        { "type": "tableCell", "content": [ { "type": "paragraph", "content": [ { "type": "text", "text": "1,420" } ] } ] } ] } ] }
  ]
}
EOF
jira-cli comment add $KEY --adf-file /tmp/comment.adf.json
```

Common nodes: `heading` (attrs.level 1–6), `paragraph`, `bulletList`/`orderedList` → `listItem` → `paragraph`, `codeBlock` (attrs.language), `table` → `tableRow` → `tableHeader`/`tableCell`. Bold/italic go on a text node via `"marks":[{"type":"strong"}]`. Preview a draft in the [ADF viewer](https://developer.atlassian.com/cloud/jira/platform/apis/document/viewer/) before posting.

**Success:** `✅ Comment added to $KEY` + `https://junahealth.atlassian.net/browse/$KEY`

## 5. Errors

| Error | Action |
|---|---|
| Auth failed | Run `jira-cli auth login` |
| Issue not found | Verify `$KEY` — check for typo, wrong project prefix |
| Empty body rejected | Ask user for comment text |
| `ADF file must be a document` | Top-level JSON must be `{"version":1,"type":"doc",...}` |
| `stub comment … created but ADF update failed` | A placeholder was posted; fix the ADF and re-run `jira-cli comment add` (or delete the stub manually) |
| Unknown flag | `jira-cli comment add --help` |
