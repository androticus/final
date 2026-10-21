# Word- and Pip-local variables

Final provides the ability to define named variables within a `word`, called `local or l variables`, and in the context of a single [Pip (thread)](threading/index.md), called `pip-local or pls variables`. The two have different use cases, but operate almost indentically, so we cover both here.

## Word-local variables via `l`

Word-local variables are helpful in cases where a certain value, sometimes a flag, wants to be available to all words at all block nesting depths within a word. Maintaining such a control variable will typically require dedicating a stack to it, and modifying it requires pops/drops/repushes. They are also useful generally, as a way to provide named access to local values, rather than trying put them on auxiliary stacks, and then try to remember which stack holds which value. A more convenient alternative is a local variable, accessed via the keyword `l`.

`l` accesses a DynamicObject that is just-in-time instantiated when you first access l. The usual pattern is to initialize your variable or variables on entry to the word, and then query or mutate as needed within the word. The compiler provides special support for `l` by enabling you to just type eg. `false l.myFlag!` then `l.myFlag@` to access it anywhere in the word. 

## Pip-local storage via `pls`

Pip-local storage works almost identically to Word-local variables, except the scope of visibility and lifetime are that of the Pip, not a word. For the root Pip, they thus function much like a global variable such as var: . This feature should be used sparingly.

Since the visibility of `pls` is the entire Pip, and any words that run in it, you should take care to use a top-level field unlikely to be used by or conflict with other words. Consider adding a "namespace" prefix to your field names.

You access Pip-local storage in a similar way as locals, by using the `pls.` keyword.

Here is an example of moving the Pip Params object onto pls for availability to all words in your Pip:

```
/*
Launch word of the Pip.
Move pp to pls for Pip-global access.
( pp:PipParams -- )
*/
: myPip
  pls.pp!
  // remaining init logic
;
```
