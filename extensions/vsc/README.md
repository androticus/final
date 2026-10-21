# Final (VS Code / Cursor)

Language support for **Final** source files (`.fnl`).

**Status (v0.0.7):** local install on your machines only — not on the Marketplace. A later release may offer a limited download.

## Features

- **Highlighting** — `//` and `/* */` comments; `"…"` (escaped) and `` `…` `` (raw) strings; `'…'` runes (character scope, not string); definition **name** after `:`
- **Indent** — 2-space defaults; Tab / Shift-Tab; Enter indent after sandwich openers
- **Comment toggle** — `Cmd+/` / `Ctrl+/`
- **Snippets** (prefix + Tab) — colon-def, sandwiches, flush-left block comment and raw string
- **`"…"` auto-close** — typing `"` inserts the closing quote on the same line

### Not in this pack

- No keyword / axiom coloring (Final has no special keyword class for ordinary words)
- No rune / unicode-character autocomplete (type `'…'` yourself; it still highlights)
- No LSP, completion from the dictionary, hover, or diagnostics yet

## Install (local)

### From this repo (recommended while developing)

1. Open **`extensions/vsc`** as the workspace folder in VS Code or Cursor.
2. Run **Extension** from the Run and Debug view (F5) — opens an Extension Development Host with Final loaded.
3. Or package and install a VSIX:

```bash
cd extensions/vsc
npx --yes @vscode/vsce package --allow-missing-repository
# then in VS Code / Cursor:
#   Extensions → … → Install from VSIX… → final-*.vsix
```

Cursor: same VSIX flow (Extensions view → Install from VSIX).

### After install

Open any `.fnl` file — language mode should be **Final**.

## Snippets

| Prefix | Inserts |
|--------|---------|
| `def` or `:` | `: name` / indented body / `;` |
| `if` / `nif` | sandwich closed by `then` |
| `switch` | … `/switch` |
| `foreach` | … `/foreach` |
| `forever` | … `/forever` |
| `begin` | … closer choice (`again` / `until` / `repeat`) |
| `/*` | block comment, **flush left** |
| `raw` or `` ` `` | raw string, **flush left** |

Colon-def and sandwiches indent the body by 2 spaces. Block comments and raw strings stay column 0.

## Colors

Scopes are theme-driven. Typical scopes:

| Form | Scope |
|------|--------|
| `//` | `comment.line.double-slash.final` |
| `/* */` | `comment.block.final` |
| `"…"` | `string.quoted.double.final` |
| `` `…` `` | `string.quoted.other.raw.final` |
| `'…'` | `constant.character.final` |
| name after `:` | `entity.name.function.final` |

## Requirements

- VS Code or Cursor meeting `engines.vscode` in `package.json`
- No Final runtime binary required for this declarative pack

## Design notes

- [159 — Phase 1 declarative pack](../../docs/003-design-notes/159-vsc-declarative.md)
- [141 — Anchor plugin / LPC](../../docs/003-design-notes/141-vscode-final-plugin.md)

## License

Private / unlicensed for now — not for public redistribution in v0.0.7.
