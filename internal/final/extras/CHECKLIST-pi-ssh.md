# Pi SSH + on-device checklist

Hardware: CanaKit Pi 5, Pi OS 64-bit. Use the **Pi zip** / `final-pi` (shipped as `final`).

Soft GPIO (GpioCtrl stubs / conflict) can also run in-repo:

```text
go run ./cmd/final-smoke -kind pi-gpio
```

## A. On-device

| # | Action | Pass? |
|---|--------|-------|
| 1 | Unzip or repo; headless smoke: `go run ./cmd/final-smoke -kind headless` **or** zip `./final --headless` then load extras | |
| 2 | Soft GPIO: `go run ./cmd/final-smoke -kind pi-gpio` | |
| 3 | Hardware: header BCM `value!` / `raw` per `extras/platform/linux/pi/gpio.fnl` comments (**not** case power LED/button) | |

## B. SSH minimal regression

From another machine:

| # | Action | Pass? |
|---|--------|-------|
| 1 | `ssh` into Pi with a PTY | |
| 2 | Headless smoke exits 0 (launcher or `--headless -m internal/final/extras/headless-smoke.fnlm` then `bye`) | |
| 3 | Optional: start headed TUI, Send one line, quit — note known SSH TUI limits ([146](../../../docs/003-design-notes/_archived/146-final-ssh-work.md)) | |

Goal: catch regressions that break **remote** use, not full TUI parity over SSH.

Failures → [`releases/BUGS.md`](../../../releases/BUGS.md).
