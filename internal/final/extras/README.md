# Cross-platform smoke extras

Opt-in smokes for OS-sensitive surfaces. Spec: [`docs/004-testing/03-win-mac-linux-tests.md`](../../../docs/004-testing/03-win-mac-linux-tests.md). Plan: [`.cursor/plans/cross_platform_smoke_0.0.6.plan.md`](../../../.cursor/plans/cross_platform_smoke_0.0.6.plan.md).

**Not** part of default `go test ./...`.

## Run (repo)

From the module root (Mac / Windows / Linux — same command):

```text
go run ./cmd/final-smoke -kind headless
go run ./cmd/final-smoke -kind pathing
go run ./cmd/final-smoke -kind pi-gpio
```

`-kind all` runs headless + pathing (skips pi-gpio unless `-pi` is set).

Exit 0 = pass. Stdout includes `SMOKE_OK …` markers (from Final `$` during load; `final -m` and `final-smoke` echo that captured stdout). Failures use `$$` + fault → non-zero exit.

## Manifests

| File | Purpose |
|------|---------|
| `headless-smoke.fnlm` | Boot + tiny arith; bootstrap/platform load sanity |
| `pathing-smoke.fnlm` | Create/read a text file (run when FS/path APIs change) |
| `pi-gpio-smoke.fnlm` | Soft GpioCtrl/GpioPin (`enabled` / `value`); needs `finalpi` |

## Checklists (manual / low-frequency)

| File | Purpose |
|------|---------|
| `CHECKLIST-tui.md` | TUI Send/Load; Windows **cmd.exe + PowerShell** |
| `CHECKLIST-pi-ssh.md` | On-device + minimal SSH regression |

## Zip

`build-install` packs this tree as `Final/internal/final/extras/`.
