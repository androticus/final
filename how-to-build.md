# How to build Final binaries

**Audience:** agents and maintainers. Prefer prebuilt zips from [final.androticus.ai/downloads](https://final.androticus.ai/downloads) for end users.

**Repo identity:** This file uses **relative** build commands only. The Go module path is whatever root `go.mod` declares (do not hard-code it in recipes). Private development trees may still use `github.com/androticus/final`; the public repo is `github.com/androticus/final`.

**Go version:** root `go.mod` requires **Go ≥ 1.25**. Do not use distro `golang` packages that ship older versions (e.g. Debian often lags). Install from [go.dev/dl](https://go.dev/dl/).

**Layout:**

| Binary | Module | Typical CGO |
|--------|--------|-------------|
| `final` | nested `cmd/final` | `CGO_ENABLED=0` OK (TUI) |
| `final-gui` | root | **CGO required** (Fyne) |
| `final-gui-pi` | root | **CGO required** (Fyne + Pi) |
| `final-pi` | nested `cmd/final-pi` | prefer native on Pi; `CGO_ENABLED=0` often OK for cross TUI |
| `final-go` | root | `CGO_ENABLED=0` OK |
| `final-ai` | nested `cmd/final-ai` | try `CGO_ENABLED=0`; enable CGO if oto/sound link fails |
| `lissajous` | root | usually `CGO_ENABLED=0` |

Nested modules: `cd cmd/<name>` then `go build .` (they `replace` the parent module with `../..`).

**Release-style flags** (optional for local work):

```text
go build -trimpath -ldflags="-s -w" …
```

**Output convention used below:** write into `bin/` at repo root (`mkdir -p bin`). On Windows, Go adds `.exe` when appropriate if you omit the suffix; examples show Unix names — append `.exe` on Windows if needed.

---

## Shared prerequisites

### All targets

```text
go version          # ≥ 1.25
mkdir -p bin
```

### Fyne / GUI (`final-gui`, `final-gui-pi`) — C toolchain

If `go env CGO_ENABLED` is `0` and no C compiler is on `PATH`, Fyne/OpenGL builds fail with errors like `gles2: build constraints exclude all Go files`.

**Linux / Pi OS (Debian family):**

```text
sudo apt-get update
sudo apt-get install gcc pkg-config libgl1-mesa-dev xorg-dev libwayland-dev libxkbcommon-dev libasound2-dev
```

Fyne/GLFW need OpenGL **and** Wayland headers even when running under X11.

**macOS:** Xcode Command Line Tools (`clang`). Native `CGO_ENABLED=1` is the default.

**Windows amd64:** MSYS2 **UCRT64** + gcc ([Fyne Windows](https://docs.fyne.io/started/)):

```text
pacman -S mingw-w64-ucrt-x86_64-gcc
```

Put `C:\msys64\ucrt64\bin` on `PATH`. Then in PowerShell:

```text
$env:CGO_ENABLED = "1"
```

**Windows arm64:** GCC does not target WoA. Use MSYS2 **CLANGARM64**:

1. Install [MSYS2](https://www.msys2.org/).
2. Open `clangarm64.exe` (not UCRT64/MINGW64).
3. `pacman -Suy` then `pacman -S mingw-w64-clang-aarch64-clang`
4. Put `C:\msys64\clangarm64\bin` on `PATH`.
5. PowerShell: `$env:CGO_ENABLED = "1"; $env:CC = "clang"`

Do not `go env -w CGO_ENABLED=1` until `clang`/`gcc` actually runs.

**Windows arm64 runtime (Parallels):** Fyne may need Microsoft Store *[OpenCL, OpenGL, and Vulkan Compatibility Pack](https://apps.microsoft.com/detail/9NQPSL29BFFF)* if you see `WGL_ARB_create_context_es2_profile is unavailable`. Enable 3D acceleration. Desktop **amd64** OpenGL builds usually do not need this pack.

**Pi: install Go** — prefer official tarball, not `apt install golang`:

```text
GOVER=1.25.0   # or newer stable ≥ 1.25
# linux-arm64 example:
curl -fsSL "https://go.dev/dl/go${GOVER}.linux-arm64.tar.gz" -o /tmp/go.tgz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf /tmp/go.tgz
export PATH=/usr/local/go/bin:$PATH
```

Verify: `go version`, `go env CGO_ENABLED` (must be `1` for GUI), `gcc --version`.

Maintainer note (private tree): non-darwin `final-gui` release mints may use cloud/golden builders — [179-cloud-release-builds.md](docs/003-design-notes/179-cloud-release-builds.md) (not copied to the public repo).

---

## 1. `final` (TUI product)

Nested module. From repo root:

```text
mkdir -p bin
cd cmd/final
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o ../../bin/final .
```

Cross-compile example (from a Mac, TUI only):

```text
cd cmd/final
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o ../../bin/final-linux-amd64 .
```

Windows consumers run `final.exe`.

---

## 2. `final-gui` (Fyne desktop GUI)

Root module. Requires [Fyne / GUI prerequisites](#fyne--gui-final-gui-final-gui-pi--c-toolchain). **Do not** set `CGO_ENABLED=0`.

```text
mkdir -p bin
# ensure CGO_ENABLED=1 and a C compiler on PATH
go build -trimpath -ldflags="-s -w" -o bin/final-gui ./cmd/final-gui
```

Cannot Mac→Linux/Windows CGO-cross with the host clang/gcc alone; build on the target OS (or a matching cloud image).

First Fyne compile is slow. Building does not need a display; running does (`DISPLAY` / `WAYLAND_DISPLAY` on Linux).

---

## 3. `final-gui-pi` (Fyne + Pi GPIO pack)

Root module. Same Fyne C deps as `final-gui`. Prefer **native build on Raspberry Pi OS** (64-bit).

```text
mkdir -p bin
go build -trimpath -ldflags="-s -w" -o bin/final-gui-pi ./cmd/final-gui-pi
```

Release zips often rename this to in-zip `final-gui`. Needs Pi hardware / `finalpi` for GPIO features at runtime.

---

## 4. `final-pi` (TUI + Pi GPIO)

Nested module. Prefer native on Pi OS. From repo root:

```text
mkdir -p bin
cd cmd/final-pi
go build -trimpath -ldflags="-s -w" -o ../../bin/final-pi .
```

Cross from Mac (TUI-oriented; validate on device):

```text
cd cmd/final-pi
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o ../../bin/final-pi .
```

Release Pi zips often ship this binary as in-zip `final`.

---

## 5. `final-go` (library-facing / slim CLI)

Root module:

```text
mkdir -p bin
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/final-go ./cmd/final-go
```

---

## 6. `final-ai` (OpenAI / agent host)

Nested module (extra deps). From repo root:

```text
mkdir -p bin
cd cmd/final-ai
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o ../../bin/final-ai .
```

If the build fails on audio/`oto`, retry with `CGO_ENABLED=1` and a C compiler. Needs network/API credentials at **runtime**, not at build time.

---

## 7. `lissajous` (demo)

Root module:

```text
mkdir -p bin
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/lissajous ./cmd/lissajous
```

---

## Quick checks

```text
go version
go env CGO_ENABLED CC GOOS GOARCH
ls bin/
```

Smoke: run the binary from a directory that has the expected Final working tree (examples/modules) when exercising language features — same idea as the release zip’s `Final/` folder.
