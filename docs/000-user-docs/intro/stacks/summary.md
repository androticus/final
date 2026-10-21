# Stack Summary

Programmer accessible runtime stacks are a defining feature of stack-based languages like Forth and Final.

Final has seven(!) stacks, defined below. Note that this may seem to be a *LOT* of stacks, in comparison to Forth, which only had rs and ds. But the needs of a modern language, which will typically involve the use of several objects in an operation, along with the many shortcomings and dangers of Forth's need for complex ds juggling have led to this design.

## Return Stack (rs)

The Final Return Stack is similar to the singular "stack" in convenional languages and machine code. It's primary purpose is to hold return addresses when a nested defined word runs. It also holds some additional values, such as a ContextMark containing the depths of all other stacks, and Dictionary-related lengths for some operations. The foreach iterated loop stores the iterator on rs.

The Return Stack is not user accessible in Final; it is used solely by the runtime.

## Data Stack (ds)

This is the core stack used for holding values during an operation, and is almost directly comparable to Forth's "stack". Unlike procedure-based languages, parameters to, and results from most low and high level operations in Final are done via DS. This enables a system of operation decomposition and composition which provides numerous advantages over the procedural approach, in which a simple factoring for readability, testability, or reuse, requires a re-expression of all the parameters in the function call and return. In Final, you just factor, and the stacks hold what they have at the factor point.

Final allows compound operations, such as property and method access on objects, indexed accessors on Lists, and named accessors on Maps, to originate from objects on any stacks. But the convention for all those operations is for the parameters and results to be located on DS. DS thus serves as a common "hub", for the source of parameters, and target for results of most operations.

## Auxilliary Stacks (ts, xs, ys, zs)

Final provides four additional stacks, which all have the same capabilities. These can be used to help stage source and destination objects during a sequence of operations, without choking ds with multiple values that need to be constantly juggled. Doing this in Forth required constant juggling of ds via sundry dups drops overs nips tucks rots etc. which are difficult to reason about and get correct while writing, and almost impossible to track when trying to read code.

In addition to their general purpose uses, `ts`, the "Temp Stack", is, by convention, the stack for short-lived parking: compound handles during `t:`, and interpret/compile temps. Prefer `ts` over the other stacks for any word-local intermediates.

Guidelines for effective use of the auxilliary stacks is provided in [Writing Clean Final Code](../clean-final-coding.md)

## Object Stack (os)

The Object Stack (os) is used exclusively by the runtime, to push the instance of an object for a Final Method call.Therefore, the runtime of such methods always access their instance via os.tos; and the appearance of the instance on a stack during the call doesn't interfere with whatever are the context of objects and values already existing on the other stacks.  (Note: Class methods push the Type on os, not an instance. Thus, other class methods and properties are accessible in that calling context.) 

The Object stack can use the instance as a root for all OO accessor operations, by prefixing the operation or sequence with o: as with txyz.

However unlike the other stacks, you may not perform stack manipulation operations on os (e.g. o:drop, o: dup etc) nor may you push values onto os; there is no `>os` and `os>` operator pair. The runtime takes care of pushing the inst value to os before the method dispatch, and popping it after the method finishes. Note that this mechanism fully enables nesting Final method calls; the latest inst is always the one on os when that method runs.

