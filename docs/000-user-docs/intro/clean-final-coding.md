# Writing Clean Final Code

Stack-based languages like Final have unique beneficial qualities, but also come with hazards and potential pitfalls.

Because Final uses stacks and stack manipulation operators, it is missing the symbolic anchors of the named parameters of procedural languages. And due to these stack manipulations, it is also fairly easy to create defective code, that does these potential things:
* underflow a stack: use or discard a value under the intended current context;
* leak values on a stack, failing to drop a no-longer needed value;
* confuse the order of operands;
* mistake a stack value for something it isn't.

## Use Local Variables

The best way to avoid stack confusion and bugs is to use [local variables](l-pls.md#word-local-variables-via-l). A local variable provides a simple mnemonic way to refer to parameters or subparameters, without having to remember what is currently on each stack.

Follow these guidelines
* consider moving stack-provided parameters to local variables, unless they are consumed only once;
* if the method is an object method, consider pulling fields you will access repeatedly from your o: instance and put them in a local;
* you don't have to define all your locals at the top, it can be clearer to define them "just in time" ahead of the code that needs them;
* if your subwords also access locals, you might consider defining those in a preliminary `init` subword, so they are defined when reading the the subword code. 
(Note: locals are resolved when used, and do not need to be lexically defined prior to use, so subwords are allowed to use locals that are only defined later in the file, either by other subwords, or the main word code.)

## Factor!

Final enables you to factor your code in two ways:
* by defining conventional Final words, and delegating work to them;
* using a [subword](control-flow/subwords.md) within a word.

Break long stretches of code, that has identifiable stages, into either subwords or another colon word.

"Clean coding" for procedural languages also recommends factoring larger procedures into small ones to which the main procedure delegates. But this often entails a fair bit of runtime overhead, insofar as you typically need to pass in a variety of parameters from the parent context to the delegate. At runtime, this is not cheap, since the runtime has to create new stack frames, copy parameter values, and so on.

Final by contrast enables much lower overhead factoring, especially for subwords. Since the values needed for an operation are already on the stacks, or also in locals for subwords, factoring is mostly a process of subdividing the work, and naming it clearly.

Factoring also lets you successively simplify your stack contexts, to narrow the values to just what a single step requires. This makes the factored code much easier to reason about, and to get correct.

## General Guidelines

Benchmarking has demonstrated that Final code runs very efficiently on modern CPUs. Because Final has a limited quantity of underlying machine code, its machine code tends to stay hot in local CPU caches, which provides a comfortable cushion to write the safest, most understandable code, not worrying about saving an operation here and there.

Another example, is code that moves an object to eg TS, for receiver operations, and that will need to be dropped at the end. Sometimes these operations have a pattern like:
```
// ( obj:a -- )
>ts
  t:peek /...
  t:peek /...
  // last operations, which do i choose?
  ts> / ...
  // or:
  t:peek
t:drop
```
You can save one instruction by moving the value on ts to ds on the last operation. But writing the code like the second example makes all stanzas identical, thus easy to reason about, and dropping ts.tos is an explicit end step, closing the >ts. The second is the far better form, because it makes every stanza identical, thus the reader (which will be you!) sees that every stanza does the same thing, that the last stanza is not some exception to the workflow, and the stack hygiene is separated from the logic of the operations.

In the last example, I also demonstrated a helpful convention, which is, where possible, to nest code relating to the content of an auxilliary stack, opening and closing with the value going on then off the stack. This provides clear visual highlighting that all the nested code is working from the same value on eg ts.

## Final Programming Conventions

Object operations that return a value, such as property reads, list item read, etc. do so on ds. If you start an attempted sequence of such operations against a single object, such as reading several properties to perform a check or perform an operation, you immediately hit a problem if you start with the object on ds. After the first read, the result is now ds.tos, so the instance is no longer available for the next operation. Rather than engaging in interminable ds juggling, which gets ever-more complex the more values you need to read, you should move the object to another stack, and root your operations there.

Rooted compound object operations always preserve the root object, so you never have to manually `dup` before a rooted operation. This makes the auxilliary stacks more convenient, and understable (writing and reading code).

Here are the stack usage guidelines that Final code follows, and that are suggested for user code as well. These guidelines suggest when and how to user auxilliary stacks, and how to make the choice of which stacks to use, and in what order. It should be noted that correct code operation does *NOT* depend on following these guidelines; all legal code that references the correct objects for its implementation will work; the guidelines just help provide semantic structure and familiarity of patterns.

**Do** continue to follow the Forth-like pattern of accepting the input parameters on ds, and placing the output parameters on ds as well, consuming the input parameters.

### Modal Result-shape Operations

For operations that either can return a result or not, depending on factors determined by that operation, provide two return effects:
* when the operation succeeds: ( ? -- result true)
* when the operation could not provide the result: ( ? -- false )
This pattern lets the consumer code be structured naturally, with a single test, whose false branch (`not if`) doesn't try to use the non-existent result, but whose true branch consumes the result.
* you typically mark this pattern by using ? affix in the name

If the result is truthy when it succeeds, but falsey when it doesn't, you can use this abbreviated pattern, along with ?dup that only duplicates a truthy result. This is suitable for a result that is an Int or String, where 0 or "" signal failure, but <>0 or <>"" signal success:
* when operation succeeds or fails, return the truthy or falsey result: ( ? -- result )
then you use code like this:
```
operation ?dup if
  // use the result, it was duped
else
  // no value to drop
then
```

For operations that don't produce a result, but need to signal an ok or an error, it is suggested to use the following pattern:
* when it succeeds: ( ? -- true );
* when it fails: ( ? -- err false ).

You can use the `nif` variant of `if` in these situations
```
operation nif
  ( err )
  // handle error
then
```

### Auxilliary Stack usage Conventions

Although all auxilliary stacks are identical in operation, by convention, ts is thought of as being the "Temp Stack", and xyz as "parking" stacks.

So ts is your first goto for a short sequence of operations that will typically only involve a single object operand. For example:

```forth
// MyObj
//   .prefix
//   .name
//   .affix

: printFullName ( inst:o -- )
  >ts
  ( t:.prefix@ t:.name@ t:.affix@ )
  "%s%s%s" fmt $
  t:drop
```

Sometimes an operation may be a bit more complex, and involve one or more input operands, and then the fetching of materializing of one or more objects in its code. While it is always desirable to factor Final operations for the goals of readability, testability, or reusability, such factoring can itself involve careful design of the factor word, plus dealing with the potential problems of use by other code, which turns the factor word from an internal-intended part of the parent word into a public artifact, which could cause client use of the word to break if the parent word itself needs to refactor the api or semantics of the factor word. (A planned feature of nested words and local-only namespacing will help, when that lands.)

Other scenarios involve a top-level word used to perform some operation using one or more operand words, one or more internally sourced or materialized words, and output results.

All these types of words benefit from continuing to use ts for localized operations, with a beginning-middle-end, and then x y and z in a loose priority order, depending on need.

**Auxiliary Stack Usage Heuristic**

The basic heuristic, as mentioned, not enforced in any way, is to use the three stacks in reverse order of permanence or importance. So operations then dealing with a second object whose lifetime is longer than the use of ts, can use xs. A lifetime longer than xs can use ys. Etc. 

In a situation where the top operation word is using or producing more than one word, and you want to keep those parked stably as you dispatch lower-level operations, you can follow a convention wherein input or source words are in order: xs, ys, zs; and outputs under construction parked in reverse order zs ys xs. Obviously this breaks down for more that three primary words, but rather than engaging in our unfriendly stack juggling, Final offers convenient help with Dynamic Objects.

If you are dealing with more state on input or output than you can reasonably juggle on txyz, you can easily create a DynamicObject to encapsulate the related items in a single object, which is then what you deal with, rather than disparate individual objects. This is also helpful to limit the number of parameters that need to be supplied on ds to an operation, or that are returned.
