# Iterating Runes

A **Rune** is a Unicode code point carried as an **`Int`** ([Basic Types](../quick/basic-types.md)). To walk the characters of a String — not its raw bytes — put the String on **DS** and use **`foreach`** … **`/foreach`**. The loop variable **`i`** is each successive Rune.

You already met `foreach` for lists and maps in [Iterators: foreach](../control-flow/iter-foreach.md). **String** participates in the same protocol: `foreach` obtains a fresh rune iterator and drives the loop until the text is exhausted.

```forth
"aé🙂日" foreach
  i r>s $ sp
/foreach
nl
// prints: a é 🙂 日
```

Print a single Rune with **`r>s $`** (`r>s` is `( rune:i -- s )`; `$` prints the String). Ordinary **`$`** on an `Int` prints the decimal code point, which is rarely what you want for text. Inside a **`fmt`** string, use **`%c`** with the Int code point instead of converting first.

**TextReader** is also iterable by rune ([Stream Types](../streams/intro.md)). Prefer `foreach` on a **String** when the text is already in memory; prefer a reader when the source is a file or you want line APIs as well.

---

## Byte length vs rune count

```forth
: runeCount ( s:String -- n:i )
  0 swap
  foreach
    drop
    1 +
  /foreach
;

"é" dup str::len $ sp    // 2
runeCount $ nl           // 1
```

That gap is exactly the Utf-8 story from [Unicode and Utf-8](utf8.md): `str::len` is storage; `foreach` is characters.

---

## Example: source file rune census

This small program reads a Utf-8 source file into a String, walks every Rune, and prints a compact summary: how many code points, how the Utf-8 widths break down, and which non-ASCII characters appear. It is a handy sanity check when a file “looks fine” in an editor but something downstream complains about encoding — and it shows `foreach`, `str::len`, and **StringBuilder** working together.

```forth
: utf8Width ( r:i -- n:i )
  dup 0x80 < if
    drop 1
  else
    dup 0x800 < if
      drop 2
    else
      0x10000 < if
        3
      else
        4
      then
    then
  then
;

: readFileString ( path:s -- s:String true | err:s false )
  TextReader.newFromFile?^ not if
    false
  else
    >ts
    StringBuilder.new^ >xs
    t:peek foreach
      i x:.writeRune^
    /foreach
    t:.close^
    t:drop
    x:.toString^
    x:drop
    true
  then
;

: printRuneCensus ( path:s -- )
  dup >ys
  readFileString not if
    "cannot read: " y:peek str::+ $ nl
    "  " swap str::+ $ nl
    y:drop
  else
    >xs
    newdo >ts
    0 t:.runes!
    0 t:.w1!
    0 t:.w2!
    0 t:.w3!
    0 t:.w4!
    StringBuilder.new^ >zs
    x:peek foreach
      t:.runes@ 1 + t:.runes!
      i utf8Width
      switch
        a 1 == case
          t:.w1@ 1 + t:.w1!
        /case
        a 2 == case
          t:.w2@ 1 + t:.w2!
        /case
        a 3 == case
          t:.w3@ 1 + t:.w3!
        /case
        default
          t:.w4@ 1 + t:.w4!
        /default
      /switch
      i 0x80 >= if
        i z:.writeRune^
        32 z:.writeRune^
      then
    /foreach
    "File: " y:peek str::+ $ nl
    "  bytes (str::len): " x:peek str::len "%d" fmt str::+ $ nl
    "  runes:            " t:.runes@ "%d" fmt str::+ $ nl
    "  utf-8 widths:     " $
      t:.w1@ "%d×1 " fmt $
      t:.w2@ "%d×2 " fmt $
      t:.w3@ "%d×3 " fmt $
      t:.w4@ "%d×4" fmt $ nl
    "  non-ASCII runes:  " $
    z:.toString^ dup str::isEmpty if
      drop "(none)" $ nl
    else
      $ nl
    then
    z:drop
    t:drop
    x:drop
    y:drop
  then
;

"examples/stringRune-iter.fnl" printRuneCensus
```

Sample shape of the report (counts depend on the file):

```
File: examples/stringRune-iter.fnl
  bytes (str::len): 218
  runes:            210
  utf-8 widths:     204×1 2×2 0×3 4×4
  non-ASCII runes:  é 🙂 日
```

The interesting lines are the ones where **bytes ≠ runes** and the width histogram is not “all `×1`”: that is Utf-8 doing its job, and `foreach` making the character stream easy to measure without hand-rolled decoding.

Next: [StringBuilder](stringbuilder.md) for the mutable side of building text.
