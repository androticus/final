# Control Flow

A set of Control Flow words are used to create control flow blocks. Final has a rich set of control flow blocks. All Final Control Flow blocks must be used in a colon definition; you will get an error if you try to use any of these in the interpreter.

Before diving in to the blocks, this is a good time to introduce Final's convention for describing the stack effects of a word, called its "context effects". The convention follows simple rules for writing a comment that describe the state of the stack on input, and its shape on output. When placing a comment before a word, it is a good idea to use the /* */ form, and place the context effect line last. The compiler picks up a last comment before a word, and stores it in the .doc field of the Word. This lets you print it by going `help wordname`. This is the comment for +

```
/*
Add two numbers, and return the sum.
( a:n b:n -- c:n )
*/
: + ... etc.
```
The ( ) indicate a stack effect. The left hand side lists the expected input values, if any. you can optionally give them a mnemonic name, eg `a` and `b` here, a colon, and the type of the value. There are codes for common values:
* `b` or `t` Bool or Truthy, most Bool contexts take Truthy (note: Json requires Bool only)
* `i` Int 
* `u` Uint
* `iu` integral, Int or Uint
* `f` Float
* `n` numeric, any of Int Uint Float
* `s` String
* `a` Any

For other types, such as objects, collections, etc. the convention is to write out the entire typename. There are a few shortcuts for common types, such as `DO` for `DynamicObject`. This is covered in more detail in [Word Doc Guidelines](../word-docs.md)

Let's dive in to the blocks. We introduce each with an example that you can enter and try.

* Definite:
  - `if` `else` `then`
  - `switch` … `/switch`
* Loops:
  - infinite: `forever` … `/forever`
  - iterated: `foreach` … `/foreach`
  - pretest: `begin` … `while` … `repeat`
  - post-test: `begin` … `until` (also: `begin` … `again` for a legacy infinite loop)
* Break and Continue
  - break: `break`
  - continue: `continue`

## Definite Blocks

### `if` / `else` / `then`

Takes a Truthy value. The `else` arm is optional — skip it when you only care about the true path.

```forth
> // ( yes:b -- )
: announce
  if "yes!\n" $ else "nope.\n" $ then
;
true announce

yes!
> false announce

nope.
```

### `switch`

The `switch` statement is Final's pattern matching statement. It takes a value, called the target, and any number of `case` clauses that expect a truthy value. You may use any kind of comparison or matching operation for each case. The target value provided at the start is available anywhere within the `switch` using the keyword `a`. If you nest switches, you can access the target value of up to two outer containing switches with `b` and `c`.

Typically, each case expression will start with `a` followed by an equality or similar matching statement. However you are not required to use the target in any arm, enabling you to mix pattern matching with other logic. For strings, in addition to simple `==` statements, you could also use Regex, or any of the str:: expressions, such as `str::indexOf`.

The case clauses are executed in lexical order, until a case matches; when that `case` clause reaches `/case` the `switch` ends, and control is passed to the next statement after `/switch`. An optional `default` clause will execute if no case matches.

```forth
> // ( choice:i -- )
: pick
  switch
    a 1 == case "got 1\n" $ /case
    a 2 == case "got 2\n" $ /case
    default
      "got something else\n" $
    /default
  /switch
;
1 pick

got 1
> 99 pick

got something else
```

Tip: if arms don’t need a target value, push a dummy (`null`, `false`, …) before `switch`; your `case` expressions can be any legal truthy expression.

## Loops

### Infinite: `forever`

An empty forever loop just spins. Press **Ctrl-B** (Break) in final or Run/Break in final-gui to stop it.

```forth
> // ( -- )
: hang
  "This loop runs forever! Press Ctrl-B to Break!\n" $ 
  forever
  /forever
;
hang

```

### Iterated: `foreach`

The `foreach` block is Final's workhorse iterated loop. It can be driven from a variety of compound data structures, such as Lists, Maps, and Objects. You can also use it to easily iterate counted ranges. In all cases, the keyword `i` in the loop refers to the current value of the type being iterated. You can even iterate the Unicode runes in a String this way!

`range` builds an inclusive numeric range. Try a number from 1 to 10 — here `5` prints `1` through `5`, each followed by a space, then a newline after the loop.

```forth
> // ( end:i -- )
: countTo
  1 swap range
  foreach
    i $ sp
  /foreach
  nl
;
5 countTo

1 2 3 4 5
```

### Pretest: `begin` … `while` … `repeat`

Condition is tested at the top of every iteration. Zero or more times.

```forth
> // ( n:i -- )
: drain
  begin
    dup while
      dup $ sp
      1 -
  repeat
  drop nl
;
3 drain

3 2 1
```

### Post-test: `begin` … `until`

Body runs at least once; the test comes after. Loop continues while the condition is falsy, and exits when it becomes truthy.

```forth
> // ( n:i -- )
: blastoff
  begin
    dup $ sp
    1 -
    dup 0 == until
  drop nl
;
3 blastoff

3 2 1
```
### `break` and `continue`

`break` can be used to exit completely from the innermost enclosing loop among:
* foreach … /foreach
* forever … /forever
* begin … until
* begin … while … repeat
* also legacy begin … again

`continue` — only in:
* foreach … /foreach
* forever … /forever
* (Not on until or while … repeat.)

Both words can be used in a conditional `if` block--the `break` or `continue` applies to the containing loop.

To percolate either in an outer loop, pass a flag or sentinel from the inner loop, checked by the outer.