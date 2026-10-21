# Text and Byte Streams

Streams are sequential sources and sinks: you read or write **incrementally** (a rune, a line, a byte, a chunk) without loading everything at once. The same programming model applies whether the backing store is a file, an in-memory string or buffer, or the process’s stdin/stdout.

Final splits streams along two axes you already saw in the [Features](../quick/features.md) survey:

| Axis | Choices |
| ---- | ------- |
| **Unit** | **Text** (Utf-8 runes / strings / lines) vs **bytes** (raw octets in a **ByteList**) |
| **Direction** | **Read** (pull) vs **Write** (push) |

That gives four **role** types — the shapes APIs ask for:

| Type | Role |
| ---- | ---- |
| **TextRead** | read Utf-8 text |
| **TextWrite** | write Utf-8 text |
| **ByteRead** | read raw bytes |
| **ByteWrite** | write raw bytes |

You usually open a concrete handle with a **class factory** on **TextReader**, **TextWriter**, **ByteReader**, or **ByteWriter** (`TextReader.newFromFile^`, `ByteWriter.newFromBytes^`, and so on). Those handles implement the matching role. Other types can too — for example **StringBuilder** is a **TextWrite**.

Final does not expose a general “open File, then ask it for a stream” workflow for day-to-day I/O — the stream factories *are* the entry point. (A small **File** helper exists for checks like `fileExists?`; see [File Streams](file.md).)

**Already covered elsewhere**

* `$` prints to the session’s stdout ([Hello Final](../first/hello-final.md)).
* **ByteList** is the compact unsigned 8-bit buffer used with byte streams ([Numeric Lists](../compounds/lists/numeric.md)).
* Pips inherit stdout/stderr; prefer `write*` on a **TextWrite** / **ByteWrite** for real I/O ([Pips](../threading/pips.md)).

**In this section**

* [Stream Types](intro.md) — common API shapes, then Text and Byte members
* [File Streams](file.md) — path-based factories and open modes
* [String Streams](string.md) — in-memory String / ByteList / StringBuilder
* [stdin/out Streams](stdio.md) — session stdin, stdout, stderr

Use `list-streams` then `help TextRead` / `help TextReader` (and the other type names) in the UI for the full member list and stack effects.
