// Basic Operator Documentation

/*
Add 2 numbers and return the result.
( a:n b:n -- c:n )
*/
:doc +

/*
Subtract b from a and return the result.
( a:n b:n -- c:n )
*/
:doc -

/*
Multiply 2 numbers and return the result.
( a:n b:n -- c:n )
*/
:doc *

/*
Divide a by b and return the result.
Panics on divide by zero for integral types.
( a:n b:n -- c:n )
*/
:doc /

/*
Absolute value of a number.
( a:n -- b:n )
*/
:doc abs

/*
Logical shift left.
Operands must be Integral (Int or Uint).
( value:iu by:iu -- res:iu )
*/
:doc <<

/*
Shift right: arithmetic if Int, logical if Uint.
( value:iu by:iu -- res:iu )
*/
:doc >>

/*
Rotate left; result type matches value.
( value:iu by:iu -- res:iu )
*/
:doc rotl

/*
Rotate right; result type matches value.
( value:iu by:iu -- res:iu )
*/
:doc rotr

/*
Truncated remainder (dividend % divisor).
( dividend:iu divisor:iu -- rem:iu )
*/
:doc %

/*
Integer quotient and truncated remainder (same as / then %).
( dividend:iu divisor:iu -- quot:iu rem:iu )
*/
:doc modrem

/*
Bitwise and. Operands must be Integral.
( a:iu b:iu -- c:iu )
*/
:doc &

/*
Bitwise or. Operands must be Integral.
( a:iu b:iu -- c:iu )
*/
:doc |

/*
Bitwise xor. Operands must be Integral.
( a:iu b:iu -- c:iu )
*/
:doc xor

/*
Bitwise not (ones' complement).
( a:iu -- c:iu )
*/
:doc bitnot

/*
Set bits in flags where mask is 1.
( flags:iu mask:iu -- c:iu )
*/
:doc bitset

/*
Clear bits in flags where mask is 1.
( flags:iu mask:iu -- c:iu )
*/
:doc bitclear

/*
Toggle bits in flags where mask is 1.
( flags:iu mask:iu -- c:iu )
*/
:doc bittoggle

/*
Boolean and on Truthy values; result is Bool.
( a:t b:t -- c:b )
*/
:doc &&

/*
Boolean or on Truthy values; result is Bool.
( a:t b:t -- c:b )
*/
:doc ||

/*
Boolean not on a Truthy value; result is Bool.
( a:t -- c:b )
*/
:doc not

/*
Equality. Operands must be Equatable to each other
(Bool, Numeric, or String within family).
( a:e b:e -- c:b )
*/
:doc ==

/*
Inequality. Same type rules as ==.
( a:e b:e -- c:b )
*/
:doc <>

/*
Less than. Operands must be Comparable to each other.
( a:c b:c -- c:b )
*/
:doc <

/*
Less than or equal.
( a:c b:c -- c:b )
*/
:doc <=

/*
Greater than.
( a:c b:c -- c:b )
*/
:doc >

/*
Greater than or equal.
( a:c b:c -- c:b )
*/
:doc >=

/*
True if TOS equals zero / is falsey-zero for its type.
Same as 0 == .
( a:e -- c:b )
*/
:doc ==0

/*
True if TOS is not equal to zero.
Same as 0 <> .
( a:e -- c:b )
*/
:doc <>0

/*
True if TOS is less than zero.
Same as 0 < .
( a:e -- c:b )
*/
:doc <0

/*
True if TOS is less than or equal to zero.
Same as 0 <= .
( a:e -- c:b )
*/
:doc <=0

/*
True if TOS is greater than zero.
Same as 0 > .
( a:e -- c:b )
*/
:doc >0

/*
True if TOS is greater than or equal to zero.
Same as 0 >= .
( a:e -- c:b )
*/
:doc >=0

/*
Returns the smaller of two numbers.
( a:n b:n -- n )
*/
:doc min

/*
Returns the larger of two numbers.
( a:n b:n -- n )
*/
:doc max

/*
True if val is between min and max inclusive.
( val:n min:n max:n -- flag:b )
*/
:doc within

// Stack and data movement

/*
Stop the current task.
Mostly used in tests and low-level control.
( -- )
*/
:doc stop

/*
Duplicate the top of the data stack.
( a -- a a )
*/
:doc dup

/*
Duplicate the top of the data stack if it is truthy.
( a:t -- a a | a )
*/
:doc ?dup

/*
Duplicate the top two items.
( a b -- a b a b )
*/
:doc dup2

/*
Discard the top of the data stack.
( a -- )
*/
:doc drop

/*
Discard the top two items.
( a b -- )
*/
:doc drop2

/*
Swap the top two items.
( a b -- b a )
*/
:doc swap

/*
Copy the second item to the top.
( a b -- a b a )
*/
:doc over

/*
Discard the second item, keep the top.
( a b -- b )
*/
:doc nip

/*
Copy the top under the second item.
( a b -- b a b )
*/
:doc tuck

/*
Copy the nth item to the top.
0 is TOS; n must be non-negative.
( ... +n:iu -- ... [n] )
*/
:doc pick

/*
Rotate the top three items.
( a b c -- b c a )
*/
:doc rot

/*
Move the top of the data stack to the temp stack.
( a -- ) t( -- a )
*/
:doc >ts

/*
Move the top of the temp stack to the data stack.
( -- a ) t( a -- )
*/
:doc ts>

/*
Discard the top of the temp stack.
t( a -- )
*/
:doc tdrop

/*
Move the top of the x-stack to the data stack.
( -- a ) xs( a -- )
*/
:doc xs>

/*
Move the top of the data stack to the x-stack.
( a -- ) xs( -- a )
*/
:doc >xs

/*
Discard the top of the x-stack.
xs( a -- )
*/
:doc xdrop

/*
Move the top of the y-stack to the data stack.
( -- a ) ys( a -- )
*/
:doc ys>

/*
Move the top of the data stack to the y-stack.
( a -- ) ys( -- a )
*/
:doc >ys

/*
Discard the top of the y-stack.
ys( a -- )
*/
:doc ydrop

/*
Move the top of the z-stack to the data stack.
( -- a ) zs( a -- )
*/
:doc zs>

/*
Move the top of the data stack to the z-stack.
( a -- ) zs( -- a )
*/
:doc >zs

/*
Discard the top of the z-stack.
zs( a -- )
*/
:doc zdrop

/*
How many items are on the current stack
(data stack by default).
( -- n:i )
*/
:doc depth

/*
Look at the top of the current stack without removing it,
and push that value onto the data stack.
(Data stack by default.)
See also List#peek when treating a List as a stack.
( -- a )
*/
:doc peek

/*
Move the top of the current stack to the data stack.
Useful when working with another stack context.
cs( x -- ) ( -- x )
*/
:doc ds>

/*
Move the top of the data stack to the current stack.
( x -- ) cs( -- x )
*/
:doc >ds

/*
True if the value is a typed Final object
(has a Final Type), leaving the value in place.
( value:a -- value b )
*/
:doc isTyped

// Numeric class constants

/*
Smallest Int (signed 64-bit minimum).
( -- i )
*/
:doc Int#minInt

/*
Largest Int (signed 64-bit maximum).
( -- i )
*/
:doc Int#maxInt

/*
Smallest Uint (always 0).
( -- u )
*/
:doc Uint#minUint

/*
Largest Uint (unsigned 64-bit maximum).
( -- u )
*/
:doc Uint#maxUint

/*
Largest finite Float.
( -- f )
*/
:doc Float#maxFloat

/*
Smallest positive non-zero Float.
( -- f )
*/
:doc Float#minNonZero

/*
The IEEE-754 +Inf value.
use Float.isInf to check for Inf
( -- +inf:f )
*/
:doc Float#+Inf

/*
The IEEE-754 -Inf value.
use Float.isInf to check for Inf
( -- -inf:f )
*/
:doc Float#-Inf

/*
The IEEE-754 NaN (not a number) value.
Note: NaN is viral; every operation with a 
NaN as an operand results in NaN.
Also, NaN cannot be compared to any value;
use Float.isNan to check for Nan
( -- nan:f )
*/
:doc Float#NaN

/*
Returns true if the value is an IEE-754
+Inf or -Inf value (usually the result
of dividing by 0.)
Accepts any numeric, but only Floats can 
be Inf; others will return false.
( a:n -- b )
*/
:doc Float#isInf

/*
Returns true if the value is an IEE-754
NaN)
see: Float.NaN
Accepts any numeric, but only Floats can 
be NaN; others will return false.
( a:n -- b )
*/
:doc Float#isNaN

/*
Returns true if the absolute value of the
difference betwen value and target is
<= epsilon. 
Useful to check float results while ignoring
small rounding errors.
Accepts any numeric, but primarily designed
for floats.
( value:n target:n epsilon:n -- b )
*/
:doc Float#isClose

/*
Returns x raised to the power of y. 
Accepts any numeric, but primarily designed
for floats. Returns a float.
( x:n y:n -- f )
*/
:doc Float#_pow
