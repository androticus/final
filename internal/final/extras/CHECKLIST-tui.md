# TUI checklist (manual)

Run on **macOS** and **Windows arm64**. Linux GUI optional.

**Windows:** complete this list **twice** — once in **cmd.exe**, once in **PowerShell**.

## Setup

1. Repo root (or unpacked `Final/` zip).
2. Build or use: `make final` then `./bin/final` (or `go run ./cmd/final`). Zip pack: `./final`.
3. Do **not** pass `--headless`.

## Steps

| # | Action | Pass? |
|---|--------|-------|
| 1 | App launches; chrome visible | |
| 2 | Send `20 22 + $` (or similar); Response shows `42` | |
| 3 | View ▸ Stdout/Stderr; same output in standing stdout; View ▸ Prompt/History restores Response | |
| 3b | While on Stdout/Stderr, EventTimer/`$` updates panes live (~250ms poll, GUI parity); scroll-back/selection not yanked to end | |
| 4 | Send `nosuchword`; one fault report; then Send `1 2 + $` still works (EventLoop warm-boot) | |
| 5 | File ▸ Load… a known `.fnl` (e.g. under `examples/`); compiles without fault | |
| 6 | Spot-check newline (**Ctrl-J**) and clipboard paste into prompt | |
| 7 | Optional: `EventTimer` / timer `$` while idle under Stdout/Stderr (proves `#run`) | |
| 8 | Quit cleanly (**Ctrl-Q** / Quit) | |

Failures → [`releases/BUGS.md`](../../../releases/BUGS.md).

## Windows hosts

| Host | Pass? | Notes |
|------|-------|-------|
| cmd.exe | | |
| PowerShell | | |
