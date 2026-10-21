# AI Tool Guidelines

Final is designed to be a **tool an agent can call**: the host creates a Session, the model emits a `script` string, and Final returns captured stdout. The sample is **`cmd/final-ai`** (web chat; `-dev` adds a Final Dev console that shares the Session).

Details of Session / Send live in [Embedding Final](embedding/index.md). This page is the short policy for that use.

## Sandbox first

Give the model a **restricted** Session. Set [Security Restrictions](embedding/security-restrictions.md) so untrusted scripts cannot open files, dial HTTP, listen, spawn Pips, open a database, use crypto, or `install` / `import` modules unless you intend that.

```go
sess, err := finalgo.NewLockSession()
// ...
sess.SetSecurityRestrictions(
	finalgo.SR_FileRead | finalgo.SR_FileWrite | finalgo.SR_FileSystemRead |
		finalgo.SR_HttpDial | finalgo.SR_HttpListen | finalgo.SR_Threads |
		finalgo.SR_Db | finalgo.SR_Crypto |
		finalgo.SR_ModuleInstall | finalgo.SR_ModuleImport,
)
```

Prefer **`LockSession`** for agent turns: `GetLockedSender()`, run the tool `Send`(s), then `Unlock()`. That keeps a Dev REPL or a second HTTP handler from interleaving dictionary changes mid-turn.

## Tool shape

`finalgo.NewFinalTool()` is a static OpenAI-style tool named `final`. `Dispatch` parses `{"script":"…"}` and calls `Sender.Send`. Register `Name` / `Description` / `ParametersJSON` with your provider; pass the **sender handle**, not the Session, into `Dispatch`.

Keep the tool description honest about what the Session can do. If you denied files and HTTP, do not advertise “read this path” or “fetch that URL.”

## Scripts the model writes

* Prefer short postfix **`$`** results over dumping huge structures.
* Invalid JSON and similar input faults **`throw`**. An uncaught throw looks like a panic to the host — wrap parse of model- or user-supplied text in `try` / `catch` inside the script, or treat `Send` errors as the tool result.
* Do not rely on values left on stacks between tool calls unless you created a **persistent** Session. The default embed Session clears stacks each `Send`; **words** persist.

## Related

* [2 Embedding Models](embedding/two-embedding-models.md) — `GetSender` / `GetLockedSender`, `FinalTool`
* [Security Restrictions](embedding/security-restrictions.md) — SR catalog including `ModuleInstall` / `ModuleImport`
* [Error and Exception Handling](error-handling.md) — `throw` vs panic
