# Intro PDF — release checklist (Mac / Chrome)

**0.0.6 approach:** one PDF per release from mdBook’s `print.html` via Chrome. Manual is fine; automate only if it stays this simple.

## Manual (recommended for now)

1. Build the book:
   ```bash
   make intro-book-serve
   # or: make intro-book   # HTML only, no local server
   ```
2. Open the print page (either works):
   - Served: http://localhost:3000/print.html
   - Or file: `docs/000-user-docs/intro-book/print.html`
3. **⌘P** → Destination: **Save as PDF**.
4. Print dialog:
   - **Headers and footers:** Off (we supply title + page # via `theme/css/print.css` `@page` boxes; Chrome 131+).
   - **Background graphics:** optional (usually Off).
   - Paper: Letter (or A4 if you standardize on that).
   - Margins: Default is fine — `@page { margin: 0.75in; }` owns the page box.
5. Save as `final-guide.pdf` into `releases/<token>/` (basename expected by `build-install`).

Spot-check: no TOC column / blank left gutter; custom header + page numbers; chapter breaks look sane.

## Optional one-shot headless (same machine)

Only worth it if you want a Makefile hook later. Chrome’s headless PDF does **not** always honor `@page` margin boxes the same as the interactive print dialog — **verify the PDF** before relying on it for a release.

```bash
# After mdbook build; adjust Chrome path if needed
"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" \
  --headless=new --disable-gpu --no-pdf-header-footer \
  --print-to-pdf="$PWD/releases/v0.0.8-rc1/final-guide.pdf" \
  "file://$PWD/docs/000-user-docs/intro-book/print.html"
```

If the headless PDF lacks your CSS headers/footers, stick to the manual ⌘P path for 0.0.6.

## Out of scope for 0.0.6

- Pandoc/Typst pipeline (still a longer-term option in [144](../../003-design-notes/144-docs-platform.md)).
- CI PDF on Linux (different Chrome flags / font story).
