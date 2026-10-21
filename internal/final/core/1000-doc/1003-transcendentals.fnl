// Transcendentals and Math constants

/*
Class-only Math namespace for common constants.
There are no Math instances — use Math.pi etc.
*/
:doc Math

/*
Base of natural logarithms (e).
( -- f )
*/
:doc Math#e

/*
Pi (circumference / diameter).
( -- f )
*/
:doc Math#pi

/*
Natural log of 2.
( -- f )
*/
:doc Math#ln2

/*
Natural log of 10.
( -- f )
*/
:doc Math#ln10

/*
Sine of an angle in radians.
Accepts Numeric; result is Float.
( a:n -- f )
*/
:doc sin

/*
Cosine of an angle in radians.
( a:n -- f )
*/
:doc cos

/*
Tangent of an angle in radians.
( a:n -- f )
*/
:doc tan

/*
Arcsine; result in radians.
( a:n -- f )
*/
:doc asin

/*
Arccosine; result in radians.
( a:n -- f )
*/
:doc acos

/*
Arctangent; result in radians.
( a:n -- f )
*/
:doc atan

/*
Hyperbolic sine.
( a:n -- f )
*/
:doc sinh

/*
Hyperbolic cosine.
( a:n -- f )
*/
:doc cosh

/*
Hyperbolic tangent.
( a:n -- f )
*/
:doc tanh

/*
Inverse hyperbolic sine.
( a:n -- f )
*/
:doc asinh

/*
Inverse hyperbolic cosine.
( a:n -- f )
*/
:doc acosh

/*
Inverse hyperbolic tangent.
( a:n -- f )
*/
:doc atanh

/*
Square root.
( a:n -- f )
*/
:doc sqrt

/*
Cube root.
( a:n -- f )
*/
:doc cbrt

/*
e raised to the power a.
( a:n -- f )
*/
:doc exp

/*
Natural logarithm (base e).
( a:n -- f )
*/
:doc log

/*
Base-2 logarithm.
( a:n -- f )
*/
:doc log2

/*
Base-10 logarithm.
( a:n -- f )
*/
:doc log10

/*
Smallest integer Float not less than a.
( a:n -- f )
*/
:doc ceil

/*
Largest integer Float not greater than a.
( a:n -- f )
*/
:doc floor

/*
Nearest integer Float (half away from zero,
matching Go math.Round).
( a:n -- f )
*/
:doc round

/*
Truncate toward zero to an integer Float.
( a:n -- f )
*/
:doc trunc
