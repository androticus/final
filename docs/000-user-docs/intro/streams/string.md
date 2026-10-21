# String Streams

In-memory streams let you use the same **Text** / **Byte** APIs without touching the filesystem — handy for tests, parsing buffers you already hold, and building text before a single print or send.

## Reading a String as text

```forth
// ( s:String -- r true | err:s false )
TextReader.newFromString^
```

```forth
: printRunes ( s:String -- )
  TextReader.newFromString^ drop >ts
  t:peek foreach
    i r>s $
  /foreach
  t:.close^
  t:drop
;
"αβ\nγ" printRunes
```

For line-oriented work, keep the reader on **ts** and call `.readLine^` in a loop (same pattern as [File Streams](file.md)).

Newlines in the source String are ordinary runes (`\n`, and `\r` if present). `readLine` still treats `\n` as the line terminator and strips a preceding `\r`.

## Writing text in memory: StringBuilder

**StringBuilder** is a **TextWrite** — the same write-shaped members as other text writers:

| Name | Effect | Description |
| ---- | ------ | ----------- |
| `new` | `( -- sb )` | empty builder |
| `writeRune` | `( r:i -- )` | append a code point |
| `writeString` | `( s:String -- )` | append as-is |
| `writeLine` | `( s:String -- )` | append string then newline |
| `toString` | `( -- s:String )` | snapshot of accumulated text |

```forth
StringBuilder.new^ >ts
"hello" t:.writeString^
32 t:.writeRune^
"world" t:.writeLine^
t:.toString^ $              // hello world\n
t:drop
```

You already saw this style under [Methods](../objects/methods.md). Pretty / indented builders (`newWithPretty`, nesting helpers) are for structured rendering (for example Json); the plain builder is enough for most stream-like accumulation.

## Bytes in memory: ByteList

**ByteList** is the buffer type for byte streams — a strongly typed list of unsigned 8-bit values ([Numeric Lists](../compounds/lists/numeric.md)).

| Factory | Effect | Description |
| ------- | ------ | ----------- |
| `ByteReader.newFromBytes` | `( bytes:ByteList -- r true \| err:s false )` | read sequentially from the list |
| `ByteWriter.newFromBytes` | `( bytes:ByteList -- w true \| err:s false )` | append writes into that same list |

```forth
ByteList.new^ >xs
x:peek ByteWriter.newFromBytes^ drop >ts
72 t:.writeNextByte^        // 'H'
105 t:.writeNextByte^       // 'i'
t:.flush^
t:.close^
t:drop

x:peek ByteReader.newFromBytes^ drop >ts
2 t:.readBytes^ >ys
y:.len@ $ nl                // 2
y:drop
t:.close^
t:drop
x:drop
```

The writer **appends** to the list you passed in; it does not replace earlier contents. The reader walks from the start of the list with its own offset — it does not remove items from the **ByteList**.
