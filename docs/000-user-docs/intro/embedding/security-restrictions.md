# Security Restrictions

**Security Restrictions (SR)** are themed capability groups a host can **disable** for a Final Session. When a theme bit is set in the Session’s disabled-set mask, tagged system words and type members that require that capability fail with a fixed `"operation not allowed"` — at compile time, interpret time, dictionary lookup, and as a safety net on native / compound invoke.

Policy is enforced in **Go at language boundaries**, not inferred from prompts. That makes SRs suitable for giving an AI (or any untrusted script source) access to Final **without** opening the filesystem, the network, or threading.

File factories, HTTP dialing, and pip spawn already participate in this model; see also [File Streams](../streams/file.md), [Web Dialing](../web/dialing.md), and [Threading](../threading/index.md).

---

## Model

| Idea | Meaning |
| ---- | ------- |
| **Disabled-set mask** | Integer bitfield on the Session / Task. **Bits set = themes denied.** |
| **Tag on affordance** | Privileged words and type members carry SR bits (e.g. file-open factories → `FileRead` / `FileWrite`). |
| **Violation** | `(taggedSR & Task.restrict) != 0` → deny |
| **Unrestricted** | Mask `0` — full vocabulary (normal TUI / trusted host default) |

Axioms such as arithmetic, stack ops, and `str::` / `dt::` families are **not** SR-gated. Denial starts at **system colon words** and **Type members** that are privileged affordances (open file, dial HTTP, spawn a pip, …).

On the Final side, theme constants live on the synthetic type **`SR`**:

```forth
SR.ModuleInstall
SR.ModuleImport
SR.FileRead
SR.FileWrite
SR.FileSystemRead
SR.HttpDial
SR.HttpListen
SR.Threads
SR.Db
SR.Crypto
SR.Mask
```

---

## Active catalog

| SR | Disallows |
| -- | --------- |
| **`ModuleInstall`** | The `install` axiom (and nested install from `loadProject`) |
| **`ModuleImport`** | The `import` axiom (and nested import from `loadProject`); pre-import in the host if scripts must still see a module |
| **`FileRead`** | Open-for-read / content stream factories from path strings |
| **`FileWrite`** | Open-for-write / create / append factories (includes directory creation) |
| **`FileSystemRead`** | Enumerate / metadata / exists-style filesystem queries (not the same as opening content) |
| **`HttpDial`** | Outbound HTTP client ops (`HttpDialer.send`, `httpDial`, …) |
| **`HttpListen`** | Creating an HTTP listener |
| **`Threads`** | Creating Pips / Pipes, `spawn:` / `(spawn)` |
| **`Db`** | Opening a database via `DbProvider.new` |
| **`Crypto`** | Digests (`Hash.hash`), Hex/Base64 encode/decode; later AEAD / sign / KDF / SecureRandom |

Combine themes with bitwise OR. A typical **AI tool Session** denies file / network / threads (and often **`Db`** / **`Crypto`**) so the agent can define words and run math / data logic, but cannot touch disk, network, or crypto primitives. Leave **`Crypto`** enabled if the agent needs content-id digests.

---

## Host usage

### Session default

The disabled-set is fixed when the Session is created (copied onto the Task). Every normal `Send` runs under that mask.

* Interactive CLI: `final-go -restrict <mask>` (hex-friendly), e.g. deny threads with `-restrict 0x200000` (`SR_Threads`)
* Embedders: `sess.SetSecurityRestrictions(mask)` after `NewSession` / `NewLockSession`, or `sender.SetSecurityRestrictions(mask)` on the handle from `GetSender()` / `GetLockedSender()` before `Send`. `finalgo.NewREPLSession(module, restrict)` still takes the mask at construction. `NewSession()` defaults to unrestricted (`0`). Prefer a restricted Session for any agent-facing tool. Prefer named `finalgo.SR_*` constants over raw hex (bit layout was remapped to uint32).

### Agent-shaped example

```go
package main

import (
	"fmt"
	"log"

	"github.com/androticus/final/finalgo"
)

// Disabled-set: bits set = denied. Prefer named SR_* consts (uint32 layout).
const agentRestrict = int64(
	finalgo.SR_FileRead |
		finalgo.SR_FileWrite |
		finalgo.SR_FileSystemRead |
		finalgo.SR_HttpDial |
		finalgo.SR_HttpListen |
		finalgo.SR_Threads,
)

func main() {
	// Persistent stacks are fine for a long-lived tool Session; dictionary state
	// is what agents usually need across tool calls.
	sess, err := finalgo.NewREPLSession(finalgo.REPLModuleLocal, agentRestrict)
	if err != nil {
		log.Fatal(err)
	}
	defer sess.Close()

	snd := sess.GetSender()
	out, err := snd.Send(`3 4 + $`, 0)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(out) // 7 — pure compute still works

	// File / HTTP / spawn affordances under this mask fail with
	// "operation not allowed" instead of performing the side effect.
	_, err = snd.Send(`"secret.txt" TextReader.newFromFile?^`, 0)
	if err == nil {
		log.Fatal("expected restriction denial")
	}
	fmt.Println("denied:", err)
}
```

With that mask, an LLM tool backed by `finalgo.NewFinalTool()` can still run postfix math, build Lists/Maps, parse Json in memory, and compile helper words — without enabling file ops, web dialing, or threading.

For the shortest unrestricted agent path, `finalgo.NewSession()` remains fine in trusted environments; add a restrict mask whenever the script source is model- or user-controlled.

### Per-Send extra restrictions

`sender.SetSecurityRestrictions(mask)` on the handle from `GetSender()` / `GetLockedSender()` **ORs** those bits with the Session floor for that `Send`. You can tighten a dispatch; you cannot lift a Session-level denial.

---

## What SRs are not

* **Not** a substitute for OS sandboxing, containerization, or network policy outside the process
* **Not** instruction-budget or hard wall-clock kill (those are separate runtime options; `timeoutS` on `Send` is advisory)
* **Not** path allowlists — host path roots, when added, complement SR bits rather than replace them

---

## Related

* [2 Embedding Models](two-embedding-models.md) — Session / Send and why agents prefer `finalgo`
* [Features](../quick/features.md) — SRs listed in the survey
* Normative design detail: `docs/003-design-notes/143-security-restrictions.md`
