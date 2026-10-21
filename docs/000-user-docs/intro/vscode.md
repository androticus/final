# VS Code and Cursor — Final extension

Final ships a **language extension** for **VS Code** and **Cursor** (same VSIX; no Cursor-specific fork). It registers language id **`final`** for **`*.fnl`** files and provides TextMate highlighting, 2-space indent defaults, comment toggle, snippets for sandwiches, and same-line `"` auto-close.

Package home in the repo: `extensions/vsc/` (extension id **final-lang**, version **0.0.7**).

Install is **local** on your machines (F5 Extension Development Host, or Install from VSIX). It is not published to the Marketplace. No Final runtime binary is required for this pack — it is declarative editor config plus grammar and snippets.

---

## Features

| Area | What you get |
| ---- | ------------ |
| **Highlighting** | `//` and `/* */` comments; `"…"` (escaped) and `` `…` `` (raw) strings; `'…'` runes; definition **name** after `:` |
| **Indent** | 2-space defaults for `[final]`; Tab / Shift-Tab; Enter indent after sandwich / colon openers; outdent on closers |
| **Line endings** | LF (`\n`) preferred for `.fnl` |
| **Comment toggle** | `Cmd+/` / `Ctrl+/` |
| **Snippets** | colon-def, control-flow sandwiches, flush-left block comment and raw string (prefix + Tab) |
| **`"…"` pair** | typing `"` inserts the closing quote on the same line |

### Not in this pack

* No coloring of ordinary words as “keywords” (Final has no special keyword class for axioms)
* No rune / unicode-character autocomplete (type `'…'` yourself; it still highlights)
* No language server yet — no dictionary completion, stack-effect hover, go-to-definition, or diagnostics

---

## Install

### From the Final tree (recommended while developing)

1. Open **`extensions/vsc`** as the workspace folder in VS Code or Cursor.
2. Run **Run Extension** from the Run and Debug view (**F5**) — opens an Extension Development Host with Final loaded.
3. Or package and install a VSIX:

```bash
cd extensions/vsc
npx --yes @vscode/vsce package --allow-missing-repository
```

Then in VS Code or Cursor: **Extensions → … → Install from VSIX…** → choose `final-lang-0.0.7.vsix` (or the `final-*.vsix` just built). A packaged VSIX may already be present under `extensions/vsc/`.

### After install

Open any `.fnl` file — language mode should be **Final**.

Requirements: VS Code or Cursor meeting `engines.vscode` in `extensions/vsc/package.json` (currently `^1.85.0`).

---

## Snippets

Type the prefix and press Tab:

| Prefix | Inserts |
| ------ | ------- |
| `def` or `:` | `: name` / indented body / `;` |
| `if` / `nif` | sandwich closed by `then` |
| `switch` | … `/switch` |
| `foreach` | … `/foreach` |
| `forever` | … `/forever` |
| `begin` | … closer choice (`again` / `until` / `repeat`) |
| `/*` | block comment, **flush left** |
| `raw` or `` ` `` | raw string, **flush left** |

Colon-def and sandwiches indent the body by 2 spaces. Block comments and raw strings stay at column 0.

---

## Colors (theme scopes)

Scopes are theme-driven. Typical scopes:

| Form | Scope |
| ---- | ----- |
| `//` | `comment.line.double-slash.final` |
| `/* */` | `comment.block.final` |
| `"…"` | `string.quoted.double.final` |
| `` `…` `` | `string.quoted.other.raw.final` |
| `'…'` | `constant.character.final` |
| name after `:` | `entity.name.function.final` |

---

## Notes

* Indent is cosmetic — the Final parser does not care about indentation.
* Prefer associating `.fnl` with this extension (or your editor of choice) when working through [Installation](first/installation.md) and [Hello Final!](first/hello-final.md).
* Extension README in-tree: `extensions/vsc/README.md`.
