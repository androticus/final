# Stream Types

This page is the shared API for text and byte streams. **TextRead** / **TextWrite** / **ByteRead** / **ByteWrite** are the roles; **TextReader**, **TextWriter**, **ByteReader**, and **ByteWriter** are the usual concrete handles (factories below). Backing-specific openers (file path, in-memory string, ByteList) are illustrated in the following pages; the **read / write / close / flush** members are the same regardless of how you opened the stream.

## Common patterns

### Opening: bare vs `?`

File-backed factories come in two forms (same convention as casts and `File.readBytes` / `readBytes?`):

* **bare** — must succeed; panics on open failure (`( path:s -- stream )`)
* **`?`** — modal result for caller-handled failure

Modal shape (same as [Writing Clean Final Code](../clean-final-coding.md)):

* success: `( … -- stream true )`
* failure: `( … -- err:String false )`

```forth
"notes.txt" TextReader.newFromFile?^ not if
  $ nl                    // err string
else
  >ts
  // … use t: …
  t:.close^
  t:drop
then
```

When open **must** succeed (system / mandatory path):

```forth
"notes.txt" TextReader.newFromFile^ >ts
```

### Reading: value plus OK

Text and byte **next** / **line** reads use the Final modal apishape:

| Family | Effect |
| ------ | ------ |
| Text | `( -- rune:i true \| false )` or `( -- line:s true \| false )` |
| Byte | `( -- b:i true \| false )` |

**true** means you got a value; **false** means end of stream (no placeholder value). Chunked `readBytes` is different: it returns a **ByteList** (possibly empty) and does **not** use a flag — empty means no more data (or a non-positive count).

### Writing, flush, and close

* Writers buffer; call **`flush`** when you need data to reach the backing store before `close`.
* **`close`** releases the backing handle when the provider supports it (files do; many in-memory providers are no-ops). Safe to call when you are done.
* Put the stream on **ts** (or another aux stack) for a short session of method calls, then `t:drop` — same style as other objects ([Writing Clean Final Code](../clean-final-coding.md)).

### Iteration

**TextReader** is iterable. Put the reader on DS and use `foreach` … `/foreach`; `i` is each Unicode code point as **Int**. The loop consumes the reader (see [foreach](../control-flow/iter-foreach.md)).

---

## TextReader

Read Utf-8 text as runes or lines. A **TextRead** (`getNextRune`); **`readLine`** is on this concrete type.

### Factories (class)

| Name | Effect | Description |
| ---- | ------ | ----------- |
| `newFromFile` | `( path:s -- r )` | open path; panics on error |
| `newFromFile?` | `( path:s -- r true \| err:s false )` | open path for sequential text read |
| `newFromString` | `( s:String -- r true \| err:s false )` | read over an in-memory String |

### Instance methods

| Name | Effect | Description |
| ---- | ------ | ----------- |
| `getNextRune` | `( -- r:i true \| false )` | next Unicode code point |
| `readLine` | `( -- line:s true \| false )` | next line without the newline; strips a trailing `\r` |
| `close` | `( -- )` | release the reader |

```forth
"hi" TextReader.newFromString^ drop >ts
t:.getNextRune^ if
  r>s $               // prints h
then
t:.close^
t:drop
```

---

## TextWriter

Write Utf-8 text (runes, strings, or lines). A **TextWrite**.

### Factories (class)

| Name | Effect | Description |
| ---- | ------ | ----------- |
| `newFromFileAppendOrCreate` | `( path:s -- w )` | append; create if missing; panics |
| `newFromFileAppendOrCreate?` | `( path:s -- w true \| err:s false )` | append; create if missing |
| `newFromFileTruncateOrCreate` | `( path:s -- w )` | truncate or create; panics |
| `newFromFileTruncateOrCreate?` | `( path:s -- w true \| err:s false )` | truncate or create |

**StringBuilder** is a **TextWrite**: call its `writeRune` / `writeString` / `writeLine` members (same shape as below), then take `.toString^`. See [String Streams](string.md).

### Instance methods

| Name | Effect | Description |
| ---- | ------ | ----------- |
| `writeRune` | `( r:i -- )` | write one Unicode code point |
| `writeString` | `( s:String -- )` | write as-is (no added newline) |
| `writeLine` | `( s:String -- )` | write string then `\n` |
| `flush` | `( -- )` | push buffered output |
| `close` | `( -- )` | flush/release as needed |

```forth
"out.txt" TextWriter.newFromFileTruncateOrCreate^ >ts
"hello" t:.writeLine^
t:.flush^
t:.close^
t:drop
```

---

## ByteReader

Read raw bytes. A **ByteRead**. Values on DS are **Int** (0…255) for a single byte, or a **ByteList** for a chunk.

### Factories (class)

| Name | Effect | Description |
| ---- | ------ | ----------- |
| `newFromFile` | `( path:s -- r )` | open path; panics on error |
| `newFromFile?` | `( path:s -- r true \| err:s false )` | open path for sequential byte read |
| `newFromBytes` | `( bytes:ByteList -- r true \| err:s false )` | read over an existing buffer |

### Instance methods

| Name | Effect | Description |
| ---- | ------ | ----------- |
| `getNextByte` | `( -- b:i true \| false )` | next octet |
| `readBytes` | `( n:i -- bytes:ByteList )` | up to `n` bytes (may be shorter / empty) |
| `close` | `( -- )` | release the reader |

```forth
ByteList.new^ >xs
65 x:.append^
66 x:.append^
x:peek ByteReader.newFromBytes^ drop >ts
t:.getNextByte^ if
  $                       // 65
then
t:.close^
t:drop
x:drop
```

---

## ByteWriter

Write raw bytes into a file or append into a **ByteList**. A **ByteWrite**.

### Factories (class)

| Name | Effect | Description |
| ---- | ------ | ----------- |
| `newFromFileAppendOrCreate` | `( path:s -- w )` | append; create if missing; panics |
| `newFromFileAppendOrCreate?` | `( path:s -- w true \| err:s false )` | append; create if missing |
| `newFromFileTruncateOrCreate` | `( path:s -- w )` | truncate or create; panics |
| `newFromFileTruncateOrCreate?` | `( path:s -- w true \| err:s false )` | truncate or create |
| `newFromBytes` | `( bytes:ByteList -- w true \| err:s false )` | append writes into that list |

### Instance methods

| Name | Effect | Description |
| ---- | ------ | ----------- |
| `writeNextByte` | `( b:iu -- )` | write one octet (**Int** or **Uint**) |
| `writeBytes` | `( bytes:ByteList -- )` | write all bytes from the list |
| `flush` | `( -- )` | push buffered output |
| `close` | `( -- )` | flush/release as needed |

```forth
ByteList.new^ >xs
x:peek ByteWriter.newFromBytes^ drop >ts
65 t:.writeNextByte^
ByteList.new^ >ys
66 y:.append^
67 y:.append^
y:peek t:.writeBytes^
y:drop
t:.flush^
t:.close^
t:drop
x:.len@ $ nl          // 3
x:drop
```