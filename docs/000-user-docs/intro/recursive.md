# Recursive Techniques

Final provides support for writing recursive code, and code that uses forward references. A good example is Final's own json parsing code. The words that parse individual Json types, needs to be able to deal with any Json type. For example, a Json array can contain items of any Json type. 

This is re-entrant recursion, and is an appropriate technique when the depth is not expected to be too deep. (What is "too deep"? We'll get to limits later.)

Other times, an algorithm can be succinctly expressed as a step that subdivides or simplifies the work, then calls itself. Factorial is an often cited example.

## Simple Recursion

During compilation of a Final word, the word itself is not yet visible in the dictionary. This presents a challenge in trying to use the word itself within itself. In this situation, Final provides a keyword, `recurse`; this has the same effect as naming the word itself; at runtime, it will nest into a new instance of the word.

```forth
// ( N:i -- N!:i )
: factorial
  dup 1 > if
    dup 1 - recurse *
  else
    drop 1
  then
;
```

When `N > 1`, the word keeps a copy of `N`, computes `(N-1)!` via `recurse`, then multiplies. For `0` and `1`, the base case drops the input and leaves `1`.

## Forward References

Json parsing is a good example of re-entrant recursion. In this type of code, a word may not itself be recursive, but will indirectly call code that causes the nesting level of the overall algorithm to increase. 

In this type of code, there is typically a chicken/egg in which earlier defined code must be able to call later-defined code. However Final, as it is defined, cannot directly support forward-references--a Word needs to be defined before it can be compiled. 

In this case, Final provides `defer:`. It creates a word with no behavior yet. You assign the behavior later:

```forth
&of realWord is deferredWord
```

`&of` names the word that does the work. `is` makes that the behavior of the deferred word. You can compile calls to the deferred name before the real word exists. Until `is` runs, calling the deferred word panics.

You can also change that behavior later. A program can start with a basic version, then replace it once more of the system is defined.

Here is a short piece of Final's own Json parser. Arrays and objects both contain values, and a value may itself be an array or object, so `jsonParseArray` and `jsonParseObject` call `jsonParseValue` before that word's body is written:

```forth
// sole cycle: object/array members → value → object/array
defer: jsonParseValue

// ( -- list:List )
: jsonParseArray
  …
  jsonParseValue    // may nest into another array or object
  …
;

// ( -- obj:DynamicObject )
: jsonParseObject
  …
  jsonParseValue
  …
;

// ( -- v:a )
: (jsonParseValue)
  …
  // '"' → string | '{' → object | '[' → array | …
;

&of (jsonParseValue) is jsonParseValue
```

`defer: jsonParseValue` creates the name so the array and object parsers can call it. The real work is `(jsonParseValue)`. The last line assigns that word's behavior to `jsonParseValue`. Until that line runs, calling `jsonParseValue` panics.

## Limits

Recursive and Re-entrant code can often provide a concise, easy to reason about algorithm. However these techniques are only suitable within limits of the level of recursion that happens at runtime. Final places depth limits on all its stacks, including the Return Stack. These limits are well above what almost any typical code would stress. But a recursive algorithm that results in excessive nesting can hit a data stack or return stack limit. In Final, that is a bit over 1,000.

So recursive/reentrant code needs to itself have limits. Recursion can always be replaced with iteration, so if the data set or recursion depth will exceed Final's limits, then it should be reworked using an iterated type of approach.
