# File Streams

File streams are opened by **class factories** on **TextReader**, **TextWriter**, **ByteReader**, and **ByteWriter**. Each file factory has two forms:

* **bare** — must succeed; panics on open failure (`( path:s -- stream )`)
* **`?`** — modal result (`( path:s -- stream true | err:s false )`)

Everyday sequential I/O goes through these factories. **File** also has whole-file `readBytes` / `writeBytes` / `readText` / `writeText` (same bare / `?` pairing).

## Text from a file

| Factory | Effect | Opens |
| ------- | ------ | ----- |
| `TextReader.newFromFile` | `( path:s -- r )` | read-only; panics on error |
| `TextReader.newFromFile?` | `( path:s -- r true \| err:s false )` | read-only |
| `TextWriter.newFromFileAppendOrCreate` | `( path:s -- w )` | append; create if missing; panics |
| `TextWriter.newFromFileAppendOrCreate?` | `( path:s -- w true \| err:s false )` | append; create if missing |
| `TextWriter.newFromFileTruncateOrCreate` | `( path:s -- w )` | truncate or create; panics |
| `TextWriter.newFromFileTruncateOrCreate?` | `( path:s -- w true \| err:s false )` | truncate or create |

```forth
"log.txt" TextWriter.newFromFileAppendOrCreate?^ not if
  $ nl
else
  >ts
  "another line" t:.writeLine^
  t:.flush^
  t:.close^
  t:drop
then
```

```forth
: printFileLines ( path:s -- )
  TextReader.newFromFile?^ not if
    $ nl
  else
    >ts
    forever
      t:.readLine^ not if
        break
      then
      $ nl
    /forever
    t:.close^
    t:drop
  then
;
"log.txt" printFileLines
```

`readLine` returns the line **without** the terminating newline. A trailing `\r` (CRLF files) is stripped. A final line that has no newline is still returned once; the next call returns **false**.

## Bytes from a file

| Factory | Effect | Opens |
| ------- | ------ | ----- |
| `ByteReader.newFromFile` | `( path:s -- r )` | read-only; panics on error |
| `ByteReader.newFromFile?` | `( path:s -- r true \| err:s false )` | read-only |
| `ByteWriter.newFromFileAppendOrCreate` | `( path:s -- w )` | append; create if missing; panics |
| `ByteWriter.newFromFileAppendOrCreate?` | `( path:s -- w true \| err:s false )` | append; create if missing |
| `ByteWriter.newFromFileTruncateOrCreate` | `( path:s -- w )` | truncate or create; panics |
| `ByteWriter.newFromFileTruncateOrCreate?` | `( path:s -- w true \| err:s false )` | truncate or create |

```forth
"data.bin" ByteWriter.newFromFileTruncateOrCreate^ >ts
ByteList.new^ >xs
0 x:.append^
255 x:.append^
x:peek t:.writeBytes^
x:drop
t:.flush^
t:.close^
t:drop
```

```forth
"data.bin" ByteReader.newFromFile^ >ts
4 t:.readBytes^ >xs          // up to 4 bytes as ByteList
x:.len@ $ nl
x:drop
t:.close^
t:drop
```

Choose **Text\*** when the file is Utf-8 text you want as runes or lines. Choose **Byte\*** for opaque binary (images, protocols, hashes). Mixing layers in one handle is not how these APIs are meant to be used — open the unit you need.

## Whole-file helpers on File

| Method | Effect |
| ------ | ------ |
| `File.readBytes` / `readBytes?` | entire file → `ByteList` |
| `File.writeBytes` / `writeBytes?` | create/truncate; write a `ByteList` |
| `File.readText` / `readText?` | entire file → `String` |
| `File.writeText` / `writeText?` | create/truncate; write UTF-8 of a `String` |

Bare forms panic on I/O error; `?` forms use `( … -- val true | err:s false )` (writes omit `val`).

## File existence

```forth
// ( path:s -- b )
"notes.txt" File.fileExists?^ $ nl
```

Unexpected filesystem errors from `fileExists?` are treated as faults (they panic), not as a modal `false`. Missing paths return `false`.

## Security note

File open factories and whole-file read/write are gated by session **Security Restrictions** (read vs write). Restricted Sessions used for embedding or AI tools may refuse file access. Details: [Security Restrictions](../embedding/security-restrictions.md).
