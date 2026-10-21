# Other Loops

The [Quick Tour](../quick/control-flow.md) shows `foreach` for iteration. Final also has indefinite loops that are **not** driven by an iterator. This page covers those, plus `break` and `continue`.

Like all control-flow blocks, these words are **compile-only** — use them inside `: … ;`.

| Form | Kind | Notes |
| ---- | ---- | ----- |
| `forever` … `/forever` | infinite | Prefer this over legacy `begin` … `again`; supports `continue` |
| `begin` … `while` … `repeat` | pretest | Body runs **zero or more** times |
| `begin` … `until` | post-test | Body runs **at least once**; exit when condition is truthy |
| `begin` … `again` | legacy infinite | Prefer `forever`; no `continue` |

## Infinite: `forever` … `/forever`

No parameters. Loop until `break`, a fault, or an interactive Break (Ctrl-B in final).

```forth
> // ( n:i -- )
: tickN
  forever
    dup 0 <= if drop break then
    dup $ sp
    1 -
  /forever
  nl
;
3 tickN

3 2 1
```

An empty `forever` … `/forever` spins forever — use Ctrl-B to stop it in the TUI.

## Pretest: `begin` … `while` … `repeat`

Words between `begin` and `while` run each attempt; `while` consumes a Truthy flag. Truthy → run the body through `repeat` and branch back to `begin`. Falsy → exit after `while` (skip the body).

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

If the first test is already falsy, the body never runs:

```forth
> // ( -- )
: never
  0
  begin
    dup while
      "should not print\n" $
  repeat
  drop
;
never
```

## Post-test: `begin` … `until`

The body runs, then `until` consumes a Truthy flag. **Falsy → loop back** to `begin`; **truthy → exit**. So the body always runs at least once.

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

## Legacy infinite: `begin` … `again`

Same idea as `forever`, but older spelling. Prefer `forever` … `/forever` in new code — `again` does **not** support `continue`.

```forth
> // ( n:i -- )
: tickAgain
  begin
    dup 0 <= if drop break then
    dup $ sp
    1 -
  again
  nl
;
3 tickAgain

3 2 1
```

## `break`

Exit the **innermost** enclosing loop among:

* `foreach` … `/foreach`
* `forever` … `/forever`
* `begin` … `until`
* `begin` … `while` … `repeat`
* legacy `begin` … `again`

Typical pattern: test inside the loop, then `if break then`.

```forth
> // ( -- )
: findThree
  1 10 range
  foreach
    i 3 == if break then
    i $ sp
  /foreach
  nl
;
findThree

1 2
```

When `i` is `3`, `break` exits before the print, so only `1 2` appear.

## `continue`

Skip the rest of the **current** iteration. Allowed only in:

* `foreach` … `/foreach`
* `forever` … `/forever`

Not in `until` or `while` … `repeat`.

```forth
> // ( -- )
: oddsOnly
  1 6 range
  foreach
    i 2 % 0 == if continue then
    i $ sp
  /foreach
  nl
;
oddsOnly

1 3 5
```

Inside `forever`, `continue` jumps back to the top of the forever body (here: skip printing `2`):

```forth
> // ( -- )
: demoSkip
  3
  forever
    dup 0 < if drop break then
    dup 2 == if 1 - continue then
    dup $ sp
    1 -
  /forever
  nl
;
demoSkip

3 1 0
```

Both `break` and `continue` may sit inside an `if` / `else` / `then` — they still apply to the **containing loop**, not to the conditional.

To affect an **outer** loop, set a flag or sentinel in the inner loop and test it in the outer loop after `/foreach` or the matching closer.
