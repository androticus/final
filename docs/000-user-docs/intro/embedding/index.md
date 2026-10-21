# Embedding Final in Go

Final was designed first as a language for **AI tools**, and second as a general-purpose stack language you can drive from Go. The [Introduction](../introduction.md) and [Features](../quick/features.md) survey already mentioned embedding; this section is the **host API** guide.

You embed Final by creating a **Session**, then calling **`Send`** with a string of Final source. Each `Send` is one interpret/compile batch: the script runs to completion (or faults), captured stdout comes back as a string, and dictionary state in that Session persists for later Sends.

| Page | Contents |
| ---- | -------- |
| [2 Embedding Models](two-embedding-models.md) | `finalgo` (simple) vs `finalembed` (host types); Session / Send / async; `LockSession`; agent sample app |
| [Security Restrictions](security-restrictions.md) | Disabled-set capability model for sandboxed AI and other hosts |

**Already covered elsewhere**

* Interactive use without embedding — [The Final TUI](../first/final-tui.md)
* File / HTTP / threading affordances that SRs can deny — [File Streams](../streams/file.md), [Web Dialing](../web/dialing.md), [Threading](../threading/index.md)
* How a Session-local dictionary relates to the VM — [How Final Works](../how-final-works.md)

**Which package?**

* Prefer **`finalgo`** for agentic and request/response hosts: create a Session, `Send` scripts, optionally expose `FinalTool` to an LLM. That is the path used by the sample chat app.
* Use **`finalembed`** when the host must register its own Go types so Final scripts can construct and drive them (widgets, domain objects, generators).

Details and tradeoffs: [2 Embedding Models](two-embedding-models.md).
