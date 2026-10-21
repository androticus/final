# Error and Structered Exception Handling (SEH)

Sometimes, processing cannot continue because some form of error has arisen. Sometimes these are developer-caused, such as integral divide by zero, or passing an incompatible type to an operator or method. Other times it can relate to unexpected service failures, such as a database or web call timeout. And other times it is caused by external input, not under the developer's control, that is nevertheless needed to continue. An example is missing or invalid input data, such as an invalid JSON string.

Final has three (well... two and a half...) ways in which it can deal with errors:
* `panic`, which is used by axioms or other runtime code
* in Final, an `error` call, which uses `panic` under the hood
* an exception, via the `throw` word

Of the three, `panic` (and thus `error`) is the nuclear option. When a `panic` occurs, the runtime does the following:
* if it occurs in a child Pip, it will virally cause all child Pips of the Session to end
* if in the main Session Pip, or propagated up from a child Pip, the main Pip is reset via a "Warm Boot". The stacks are cleared, and the interpreter is then invoked and entered as if from a new Session (the Dictionary remains intact however.)

Final's philosophy is to `panic` or `error` in any situation that could have been prevented; basically all situations of programmer error, such as requesting integral division by zero, or passing a number to a routine that required a String.

In situations involving non-preventable errors, particular those arising from external inputs, Final uses the `throw` statement, which is an error that you can trap by way of a `try`/`catch` block in your code. `throw` is particularly useful in deeply nested code, in which trying to carry some kind of termination flag, along with the extra code at every level also testing for this flag, results in overly complex code that intermixes normal processing with tests for the error state. A good example is JSON parsing, which can be deeply recursively nested as it processes arrays and objects within outer arrays or objects, or which can encounter an unexpected end of stream before parsing is complete. Final's `json>` operator uses the `throw` statement in many places, when it encounters invalid input, or unexpected end of stream.

You can use `throw` in your own code, instead of `error` when you want to be able to trap the error at a higher level, and can recover without causing the nuclear option.

To catch Exceptions, you use the `try`/`catch` block as follows; `finally` is optional, and will be explained below:
```forth
try
// words that might throw
catch // ( -- errmsg:s )
// your exception handling logic
finally // optional
// code that will always execute
/try
```

`try` signifies the beginning of an exception handling block. `try` blocks can be nested, both within the same word, as well as being nested by way of outer or inner words that may also have `try` blocks. In normal processing, the code between `try` and `catch` executes; the `catch` clause is skipped over, and if present, the `finally` clause executes.

If a word inside the `try` block, or any word it invokes, throws an exception via `throw`, then the following happens:
* the returns from every nested word are popped off of RS, to get the VM cleanly back up to the nesting level of the `try` block
* the data stack and all auxilliary stacks have their content restored to what it was at the entry to `try`
* the code in the `try` block stops executing
* the code in the `catch` block executes
* if the `catch` block does not then throw its own exception, the `finally` clause, if present, executes 
* normal execution then resumes after the `/try` statement

If a `throw` occurs in runtime code, and there is no `try` block active at that time, the situation is handled just like `panic`.

## How to use (and not use) exceptions and handlers

### When to use `throw`

Use `throw` for unexpected error conditions, typically those cause by invalid externally supplied data, over which you had no control. It is especially indicated in operations involving multiple steps, or that may become deeply nested.

Document in a word that has `throw`, and outer words that invoke it, that it throws. That provides some guidance that an exception handling block when invoking that operation may be helpful.

### Don't use `throw`...

* as a new "short-circuit" channel for communicating results; `throw` is about exceptional error situations, not normal processing situations. 
* as a replacement for having clear, sensible, understandable input requirements; there is not much point in using `throw` for programmer error; that deserves an `error` (or implicit runtime `panic`)

### When to use `try`/`catch`

Don't think you need to include a `try` block just because something you are calling can `throw`; the question you need to ask, is whether there is anything reasonable you can do at the point. In many cases there may be, such as sending an error result on a result pipe, or returning a modal `false` result to a caller. But if there is nothing you can do, just let the exception continue up the chain, where something else might catch it, or in the worst case, the system will have to treat it as a panic.

## Important Caveats and Limitations

SEH in a stack-based language is not as deterministic as in a procedural language. If your `try` block either consumes values from any stacks, or in normal execution leaves new or altered values on any stack, then this will likely NOT be the case if execution falls through to after `/try`. Therefore, you should base your handling logic on this fact. 

Typically, you might want your `try` block to span your entire word, or at least be the last code in the word. If your word follows a modal ( ? -- result true | false ) pattern, your `catch` block can drop any input values on eg. DS that were restored by the handler, then push a `false` and fall thru.

As with *all* languages that support SEH, the side effects caused by intermediate words of your normal chain are in an indefinite state from your viewpoint, since you may not know how much executed prior to the `throw`.