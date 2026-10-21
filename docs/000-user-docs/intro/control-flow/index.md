# Control Flow

The [Quick Tour](../quick/control-flow.md) already introduces Final’s control-flow blocks (`if` / `switch`, loops, `break` / `continue`) and the context-effect comment convention.

This section goes deeper on a few of those items:

| Page | Focus |
| ---- | ----- |
| [`range` and `rangeby`](ranges.md) | Building counted iterators for `foreach` |
| [Iterators: `foreach`](iter-foreach.md) | Which types iterate, and what `i` is in each case |
| [Other Loops](loops-other.md) | Non-`foreach` loops: `forever`, `begin`…`while`…`repeat`, `begin`…`until`, legacy `begin`…`again`, plus `break` / `continue` |
| [Subwords](subwords.md) | Subwords are "word-lite" words defined within a colon word |

All of these words are **compile-only**: they belong inside a colon definition (`: … ;`). Trying them at the interpreter prompt raises an error.

`try` / `catch` / `finally` / `throw` are also compile-only; they are covered in [Error and Exception Handling](../error-handling.md).

For stack effects and short descriptions, use `help` on the word name (for example `help foreach`, `help forever`).
