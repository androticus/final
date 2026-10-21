# Context Selectors

Given that Final has six user-accessible stacks, a means is obviously needed for operations to specify to which stack you are referring. This is called Context Selection. It arises in two areas:

1. Many stack operations can be performed on any stack, so in this case, a Context Selector (cs) specifies which stack the operation should be performed on.

2. Compound Path operations can start with an object--the "receiver"--on most any stack, so the cs is required to specify which stack.

A cs is simply the stack letter, followed by a colon, eg
* t: prefix for Temp Stack
* x: prefix for X Stack  
etc.
* o: prefix for O Stack

The Data Stack does not have a cs selector; an operation that supports a CS, without one, is assumed to apply to DS.

Note that not all compound path operations are supported with the receiver on DS; most require the receiver to be located on an auxilliary or the O Stack.