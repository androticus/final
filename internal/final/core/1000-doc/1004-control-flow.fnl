// Control flow constructs

/*
Opens a conditional block, with optional 'else'.
Closed with 'then'. Consumes a Truthy condition.
Truthy → run the if-body; falsy → skip to else/then.
( b -- )
*/
:doc if

/*
Like 'if', but inverted: run the body when the
condition is falsy (sugar for 'not if').
( b -- )
*/
:doc nif

/*
Optional clause to execute if original test was falsey
(for 'if') or truthy (for 'nif').
Closed with 'then'.
( -- )
*/
:doc else

/*
Closes an 'if'/'nif'/'else' block.
( -- )
*/
:doc then

/*
Marks the start of a conditional indefinite loop
(post-test 'until' or pretest 'while'/'repeat').
For infinite loops prefer 'forever' … '/forever'.
( -- )
*/
:doc begin

/*
Legacy infinite closer for 'begin' … 'again'.
Prefer 'forever' … '/forever' (supports 'continue').
( -- )
*/
:doc again

/*
Post-test loop closer: body runs at least once.
Consumes a Truthy test; falsy → back to 'begin';
truthy → exit the loop.
begin ... done? until
( b -- )
*/
:doc until

/*
Pretest mid-word: after 'begin', words before 'while'
run each iteration; truthy → enter body until 'repeat';
falsy → exit after 'repeat'.
begin ... cond while ... repeat
( b -- )
*/
:doc while

/*
Pretest loop closer: branch back to matching 'begin'.
( -- )
*/
:doc repeat

/*
Infinite loop opener (no parameters or flags).
Closed with '/forever'. Supports 'break' and 'continue'.
forever ... /forever
( -- )
*/
:doc forever

/*
Closes a 'forever' loop; unconditional branch back
to the matching 'forever'.
( -- )
*/
:doc /forever

/*
Exit the innermost enclosing loop
('foreach', 'forever', or 'begin' … until/while/repeat).
Typical use: cond if break then
( -- )
*/
:doc break

/*
Skip the rest of the current iteration in
'foreach' or 'forever' (not in until/while/repeat).
Typical use: skip? if continue then
( -- )
*/
:doc continue

/*
Iterated loop over an Iterable (Range, List, Map,
TextReader, …). Consumes the provider; dup it if
needed during the loop. Current item is 'i'
(nested: 'j', 'k').
Closed with '/foreach'.
( o:Iterable -- )
*/
:doc foreach

/*
Closes a 'foreach' loop; branches back for the
next item, or exits when exhausted.
( -- )
*/
:doc /foreach

/*
Current item of the innermost 'foreach'.
Only valid in the same colon definition as that
foreach (not from factored callees).
( -- item:a )
*/
:doc i

/*
Current item of the next-outer 'foreach'
(when nested two deep).
( -- item:a )
*/
:doc j

/*
Current item of the third-outer 'foreach'
(when nested three deep). Deeper nesting has no
accessor; pass values on the data stack.
( -- item:a )
*/
:doc k

/*
Inclusive integer range for 'foreach'.
For indexed walks: beg=0, end=len-1
(empty: 0 -1).
( beg:i end:i -- r:Range )
*/
:doc range

/*
Inclusive range with step (forward or backward).
beg/end/step must be the same numeric type (Int Uint Float).
Stops when the next step would pass end (inclusive if landed on).
`i` will be of the same numeric type as the rangeby values.
( beg:n end:n step:n -- r:RangeBy )
*/
:doc rangeby

/*
Opens a target-bearing multi-way switch (first-match wins).
Parks DS target on SwitchTargets; re-fetch with a/b/c in arms.
Each arm: Condition case ... /case
Optional: default ... /default
Closed with '/switch' (drops target).
( target:a -- )
*/
:doc switch

/*
Arm guard: consumes a Truthy condition.
Truthy → run case body then exit switch at '/case';
falsy → skip to next arm.
( b -- )
*/
:doc case

/*
Closes a 'case' arm. Matched path exits the switch;
falsy path continues to the next arm.
( -- )
*/
:doc /case

/*
Push innermost / outer switch targets (like i/j/k).
Compile-only; require 1/2/3 enclosing switches.
( -- target )
*/
:doc a

/*
( -- target )
*/
:doc b

/*
( -- target )
*/
:doc c

/*
Optional default section when no 'case' matched.
Closed with '/default'.
( -- )
*/
:doc default

/*
Closes a 'default' section.
( -- )
*/
:doc /default

/*
Closes a 'switch'; merge point for exit branches.
( -- )
*/
:doc /switch


// SEH

/*
Start a Structured Exception Handling block.
`throw` will be caught by `catch` block.
( **snapshot all data stacks as "S1"** )
*/
:doc try

/*
Start the Catch block, executed if a `throw` occurs.
( **restore all data stacks to "S1"** )
*/
:doc catch

/*
Optional clause that will always execute.
( -- )
*/
:doc finally

/*
End the `try` block.
( -- )
*/
:doc /try

/*
Throw an exception, with the given error text.
( msg:s -- )
*/
:doc throw

/*
Word-local storage. A DynamicObject local to just
the currently executing word. Useful for flags,
counters, etc.
( -- l:DynamicObject )
*/
:doc l
