# Subwords

Final lends itself well to factoring complicated words into smaller worker words. This can be done at a macro level by defining new colon words, and then delegating to them in a main word.

But such factoring can be inconvenient or have drawbacks for several reasons:
* Final has a global namespace, so mnemonic words like "init" are not good factored word names;
* a delegated colon word doesn't inherit locals set up in the calling word;
* when reading a file, it can be difficult to determine if a word is called in more than one place, or is a just a factor word;
* calling a colon word has overhead, which makes them less desirable for hot inner loop steps.

To address these issues, Final has subwords. A `subword` is a word defined inside a main colon word. Subwords are much lighter and more efficient than colon words, so are well suited for factoring code in inner loops or switch cases, to make the main flow more readable.

Here is an example use:
```
: main-word
  :s init
    // some factored init code, for clarity
  ;s
  
  // ( item -- )
  :s do-item
    // code to handle one item in a loop
  ;s

  init
  myList
  foreach
    i do-item
  /foreach
;
```

A subword has access to all the context of the main word, especially the local variables, and the i/j/k variables of `foreach` loops, and the a/b/c variables of `switch` statements. (And of course the stack contexts, which factored colon words also have access to.) Unlike a factored colon word, a subword may compile `i`/`j`/`k`/`a`/`b`/`c` without those blocks lexically around the use — they are checked at runtime against the shared frame. In the main body of the word, the usual static nesting rules still apply.

Subword definitions must be the first elements in a main word. There can be no main word code until the last subword definition. Subwords themselves cannot be nested; no `:s` in an outer `:s`.

You use subwords exactly like any colon word in your code, as demonstrated in the examples.

The subword names are only visible inside the main word, not outside it. You cannot invoke a subword inside one word from another. The same ordering rules that apply to words apply to subwords: a subword may call another subword that is lexically defined before it, but not after.

Subwords cannot call themselves. The `recurse` statement can be used, but it causes a new invocation of the main word, not the subword.

You can use most of the structured control blocks in a subword, such as `if/then`, `foreach`, etc. But the following restrictions apply:
* you may not define a `try/catch` block (but you can use `throw`)
* if you define a `foreach` loop, then it will nest inside any active outer `foreach`; you can access those using `j/k` but only if they are active when you call the subword
* the same rule applies for `switch` and `b/c/` accessors

Most importantly, you can access all the locals defined in the main word, or subwords, within a subword. A subword can also add locals. Since local variables are resolved when they are accessed, not when they are compiled, it doesn't matter if they are defined lexically later than where they are used. That said, for clarity, you may choose to initialize the locals that will be used throughout the words, in a subword like the `init` I gave in the example--this will make scanning the subwords prior to reading the main code more clear.

Subword invocation is extremely efficient. A subword is basically like a machine language subroutine, calling one is just a short `brsub` axiom, and returning is a `retsub` axiom, using a fast simple return stack in the ColonFrameMark of the main word. Using one will add negligible overhead even to hot inner loops, due to CPU caching. Therefore, subword factoring is encouraged even for hot paths, if it makes the code clearer.


