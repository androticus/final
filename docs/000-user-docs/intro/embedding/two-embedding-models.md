# 2 Embedding Models

Final exposes two public Go packages for embedding. Both boot the runtime on import, own a long-lived **Session**, and run scripts with **`Send`**. They differ in how much of the host you expose into Final.

| Package | Role | Choose when |
| ------- | ---- | ----------- |
| **`finalgo`** | Minimal Session / Send / Close API; optional `FinalTool` for LLM tool calling | AI agents, chat backends, sandboxed script runners, most “run this Final string” hosts |
| **`finalembed`** | Same Session/Send model **plus** host type registration (`RegisterTypes`) and re-exported Final cell/collection types | Embedding Final as a scripting language for *your* Go types (UI, domain objects, generators) |

For **AI agentic support, prefer the simple model (`finalgo`)**. Agents need a stable interpret tool, captured stdout, and a restricted Session — not a catalog of host widgets. Reach for `finalembed` only when scripts must mint and drive types that live in your application.

---

## Concepts: Session, Send, and async

### Session

A **Session** is a long-lived Final execution context:

* Private **module dictionary** (Session-local DC) where colon words you define with `:` persist across Sends
* One dedicated **VM worker** goroutine that runs script batches
* Optional **Security Restrictions** disabled-set (see [Security Restrictions](security-restrictions.md))

Create one Session per logical agent conversation, web client, or embed unit. Close it when that unit ends.

### Send

```text
out, err := sess.GetSender().Send(script, timeoutS)
```

| Argument / result | Meaning |
| ----------------- | ------- |
| `script` | Final source (interpret and/or compile). Empty script is an error. |
| `timeoutS` | Advisory run budget in whole seconds (`0` = none). Used for inward I/O defaults; **not** a hard wall-clock kill of blocking natives. |
| `out` | Captured **stdout** text from the batch (`$`, `.s`, stream writes to stdout, …) |
| `err` | Fault or interpret error; may be accompanied by partial `out` |

Default Sessions use **ephemeral stacks**: each `Send` starts with empty DS/RS/… Dictionary and compiled words **persist**. That matches request/batch and agent-tool semantics. Persistent stacks (REPL-like) exist for the TUI path (`finalembed.NewPersistentSession` / `finalgo.NewREPLSession`) — not the usual agent default.

`Send` **blocks** until the worker finishes the batch. Concurrent `Send` calls on the same Session **serialize**.

Call `sess.SetSecurityRestrictions(mask)` as a Session-wide floor, or `sender.SetSecurityRestrictions(mask)` on the handle from `GetSender()` before that `Send` (those bits are **ORed** with the floor — you can tighten, not lift). Bits **set** deny the matching [SR](security-restrictions.md) themes.

### Making Send asynchronous

The VM already runs on a **worker goroutine** inside the Session. Your host call to `Send` is the synchronous rendezvous with that worker.

To keep an HTTP handler, chat turn, or UI thread responsive, run `Send` on **your own** goroutine (or job queue) and deliver `out`/`err` when ready:

```go
go func() {
    out, err := sess.GetSender().Send(script, 30)
    // deliver out/err on your completion channel, websocket, etc.
    _ = out
    _ = err
}()
```

Do not start multiple overlapping `Send`s and expect parallel VM progress on one Session — they queue. For cancel, call `sess.Break()`; the VM stops at the next opcode boundary (natives may still finish first).

### `LockSession` (exclusive agent turns)

`finalgo.NewLockSession()` / `finalembed.NewLockSession()` returns a **`LockSession`**. Its only Send path is **`GetLockedSender()`**, which holds until that handle’s **`Unlock()`**. A second `GetLockedSender` **blocks** until Unlock. `final-ai` uses this so a Dev console cannot interleave Sends in the middle of an AI turn.

You may `Send` several times on the same locked sender (compile helpers, then a follow-up) and then `Unlock`. After Unlock the handle is invalid (`ErrLockedSenderInvalid`).

---

## Simple model: `finalgo`

Import path: `github.com/androticus/final/finalgo`.

### Core API

| API | Effect |
| --- | ------ |
| `NewSession() (*Session, error)` | Local module Session; ephemeral stacks; unrestricted by default |
| `NewLockSession() (*LockSession, error)` | Same, but Send only via `GetLockedSender` |
| `(*Session) GetSender() *SessionSender` | Handle for one or more `Send`s; optional `SetSecurityRestrictions` |
| `(*LockSession) GetLockedSender() *LockedSender` | Exclusive hold until `Unlock()` |
| `(*SessionSender) Send(script string, timeoutS int) (string, error)` | Run one batch; return stdout |
| `(*Session) Break()` | Cooperative stop of in-flight VM run |
| `(*Session) Close() error` | Stop worker; idempotent |
| `NewFinalTool() *FinalTool` | OpenAI-compat tool wrapper (`Name` / `Description` / `ParametersJSON` / `Dispatch`) |
| `Sender` interface | `SetSecurityRestrictions` + `Send` — what `FinalTool.Dispatch` uses |

### Minimal embed (agent-friendly)

```go
package main

import (
	"fmt"
	"log"

	"github.com/androticus/final/finalgo"
)

func main() {
	sess, err := finalgo.NewSession()
	if err != nil {
		log.Fatal(err)
	}
	defer sess.Close()

	snd := sess.GetSender()

	// Interpret a postfix expression; $ prints the TOS number to stdout.
	out, err := snd.Send(`3 4 + $`, 0)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(out) // 7

	// Definitions persist for later Sends in this Session.
	if _, err := snd.Send(`: sq dup * ;`, 0); err != nil {
		log.Fatal(err)
	}
	out, err = snd.Send(`5 sq $`, 0)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(out) // 25
}
```

That is the whole loop for most agentic hosts: **NewSession → Send → Close**.

### Final as an LLM tool

`FinalTool` packages the Session for providers that accept OpenAI-style tools. `Dispatch` parses `{"script":"…"}`, calls `Sender.Send`, and returns captured stdout (or an `[error] …` content string).

```go
sess, err := finalgo.NewSession()
if err != nil {
	log.Fatal(err)
}
defer sess.Close()

ft := finalgo.NewFinalTool()
// Register ft.Name(), ft.Description(), ft.ParametersJSON() with your AI client.
// On tool_calls:
content, err := ft.Dispatch(sess.GetSender(), toolCallID, ft.Name(), argumentsJSON)
```

### Sample chat application

This repo ships **`cmd/final-ai`**: a small web chat UI that registers Final as a tool the model can call. Early tests showed the agent successfully calling Final for **mathematical calculations** (postfix scripts via the `final` tool). Run it with your provider API key (see `cmd/final-ai`); use `-dev` for the tabbed AI Chat + Final Dev console that shares one Session per browser cookie.

---

## Extensive model: `finalembed`

Import path: `github.com/androticus/final/finalembed`.

Use this when Final scripts must create and operate on **your** Go types. You register each host type once after `NewSession`, before any `Send`. Registration is app-global for that Session’s module.

### Core API (beyond `finalgo`)

| API | Effect |
| --- | ------ |
| `NewSession() (*Session, error)` | Local module; ephemeral stacks |
| `NewLockSession() (*LockSession, error)` | Exclusive Send via `GetLockedSender` |
| `NewPersistentSession() (*Session, error)` | Same, but stacks survive across Sends (REPL / TUI style) |
| `(*Session) RegisterTypes([]TypeReg) error` | Install host types; call once before Send |
| `NewTypeReg[T](name string, reg func(*Type)) TypeReg` | Build a registration entry for Go type `T` |
| `(*Type) RegisterProperty` / `RegisterNativeMethod` / `RegisterNativeClassMethod` / … | Describe members on the Final type |
| Re-exports in `types.go` / `cells.go` | `Int`, `Float`, `String`, `AnyList`, `ToCell`, … for DS-safe boundaries |

### Host type sketch

```go
sess, err := finalembed.NewSession()
if err != nil {
	return err
}
defer sess.Close()

err = sess.RegisterTypes([]finalembed.TypeReg{
	finalembed.NewTypeReg[*Widget]("Widget", func(t *finalembed.Type) {
		t.RegisterNativeClassMethod("new", func(f *finalembed.RunFrame, ths *finalembed.TypeObjPair) {
			f.RP.DataStack.Push(NewWidget())
		})
		t.RegisterProperty("title", "String",
			func(inst any) any { return finalembed.String(inst.(*Widget).Title) },
			func(inst any, v any) { inst.(*Widget).Title = string(v.(finalembed.String)) },
			false,
		)
		t.RegisterNativeMethod("render", func(f *finalembed.RunFrame, ths *finalembed.TypeObjPair) {
			ths.Obj.(*Widget).Render()
		})
	}),
})
if err != nil {
	return err
}

_, err = sess.GetSender().Send(`Widget.new^ >ts "Hi" t:.title! t:.render^ t:drop`, 0)
```

**Cell rule:** values that cross the data stack (property get/set, method args/results) must be Final cells — `finalembed.Int`, `Float`, `String`, … — or pass through `finalembed.ToCell`. Raw Go `int32`/`float32` will eventually panic under arithmetic.

A fuller worked example lives in `cmd/lissajous` (register a generator type, drive it from a Final script).

### Tradeoffs vs `finalgo`

| Concern | `finalgo` | `finalembed` |
| ------- | --------- | ------------ |
| Lines to first useful `Send` | Very few | Registration boilerplate per host type |
| API for an LLM tool | Excellent (`FinalTool`) | Possible, but usually unnecessary |
| Script your own Go structs | No | Yes |
| Re-use Final Lists/Maps from Go natives | Limited | First-class via re-exported types |
| Typical security posture | Tight SR mask; math / pure logic | Often needs the affordances your types wrap (files, UI, …) |

**Recommendation:** start with `finalgo` for agents. Graduate to `finalembed` only when a script must say `YourType.new^` and call methods on host objects.

---

## Related

* [Security Restrictions](security-restrictions.md) — deny file I/O, HTTP dial, threading for sandboxed Sessions
* [Features](../quick/features.md) — embedding, `LockSession`, and SRs listed in the survey
* [AI Tool Guidelines](../ai-tools.md) — sandboxing and exclusive turns for LLM tools
* Design depth (not required for host authors): `docs/003-design-notes/134-embedding-final.md`
