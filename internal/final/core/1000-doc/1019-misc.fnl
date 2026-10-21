// Miscellaneous: random, color, ranges

/*
Class-only random helpers. No Rand instances —
use Rand.int^ etc.
*/
:doc Rand

/*
Random Int in 0 .. n-1 (n must be positive).
( n:i -- r:i )
*/
:doc Rand#int

/*
Random Float in [0.0, 1.0).
( -- f )
*/
:doc Rand#float

/*
Random Float from the standard normal distribution.
( -- f )
*/
:doc Rand#normFloat

/*
Perlin noise generator. Nearby coordinates yield related values
(useful for terrain and similar fields). Use Noise.new^ or
Noise.newWithSeed^.
*/
:doc Noise

/*
Create a Noise with a random seed.
( -- n:Noise )
*/
:doc Noise#new

/*
Create a Noise with a fixed seed (reproducible samples).
( seed:i -- n:Noise )
*/
:doc Noise#newWithSeed

/*
Sample 1D noise at x.
( x:f -- v:f )
*/
:doc Noise#at1

/*
Sample 2D noise at (x, y).
( x:f y:f -- v:f )
*/
:doc Noise#at2

/*
Sample 3D noise at (x, y, z). Use z as time for animation.
( x:f y:f z:f -- v:f )
*/
:doc Noise#at3

/*
Inclusive integer range with step 1.
Prefer the `range` word to build one.
See also Range#beg / Range#end.
*/
:doc Range

/*
Create a Range (same as the `range` word).
( beg:i end:i -- r:Range )
*/
:doc Range#new

/*
Start of the range (inclusive).
*/
:doc Range#beg

/*
End of the range (inclusive terminal).
*/
:doc Range#end

/*
Inclusive numeric range with an explicit step.
Prefer the `rangeby` word to build one.
beg/end/by are the same numeric type (Int, Uint, or Float).
*/
:doc RangeBy

/*
Create a RangeBy (same as the `rangeby` word).
( beg:n end:n step:n -- r:RangeBy )
*/
:doc RangeBy#new

/*
Start of the range (inclusive).
*/
:doc RangeBy#beg

/*
End of the range (inclusive terminal).
*/
:doc RangeBy#end

/*
Step size (must not be 0; same numeric type as beg/end).
*/
:doc RangeBy#by

/*
A Counter for runing tallies.
*/
:doc Counter


/*
Keeps an accurate runing average and variance, without storing samples.
The variance/stddev is population variance (/n).
Use getMeanAndN or getMeanStdDevAndN to atomically get all values.
Average is threadsafe.
*/
:doc Average

/*
The current running average.
*/
:doc Average#mean

/*
The current running (population) Variance.
*/
:doc Average#variance

/*
The current running (population) Standard Deviation.
*/
:doc Average#stdDev

/*
The number of samples averaged so far.
*/
:doc Average#n 

/*
Add a numeric value to the running average/variance.
( a:n -- )
*/
:doc Average#add

/*
Reset mean/variance to 0.0 and n to 0.
( -- )
*/
:doc Average#reset 

/*
Atomically get the current mean and n.
( -- mean:f n:i )
*/
:doc Average#getMeanAndN 

/*
Atomically get the current mean, standard deviation, and n.
( -- mean:f stdDev:f n:i )
*/
:doc Average#getMeanStdDevAndN 

/*
Create a new Average object, with all 0 values.
( -- avg:Average )
*/
:doc Average#new 


/*
Exponential moving average (EMA) of numeric samples. No variance.
Create with newWithSpan^ only; .span is fixed for the life of the object.
α = 2/(span+1) — span is the effective window in samples (about how many
recent samples matter). First add sets mean; later adds decay older weight.
Use Average when you need an exact mean/variance; MovingAverage when the
signal drifts (costs, rates, live telemetry). Threadsafe.
*/
:doc MovingAverage

/*
Current EMA mean.
*/
:doc MovingAverage#mean

/*
Number of samples added since create or reset.
*/
:doc MovingAverage#n

/*
Effective window (samples) fixed at create. Readonly.
*/
:doc MovingAverage#span

/*
Add a numeric sample to the EMA.
( a:n -- )
*/
:doc MovingAverage#add

/*
Reset mean to 0 and n to 0. Span is unchanged.
( -- )
*/
:doc MovingAverage#reset

/*
Atomically get the current mean and n.
( -- mean:f n:i )
*/
:doc MovingAverage#getMeanAndN

/*
Create a MovingAverage with the given span (clamped to at least 1).
( span:i -- ma:MovingAverage )
*/
:doc MovingAverage#newWithSpan


/*
State-smart XT of the next word in the input stream.
Interpret: push VirtXT. Compile (immediate): emit XT as a literal (same as &of,).
See also &of,
<word> ( -- xtword:XT ) | c( -- +lit xt )
*/
:doc &of

/*
Compile-only form of &of: compile the XT of the next word as a literal.
See also &of
<word> compiles XT of <word> as a literal
runtime: ( -- xtword:XT )
*/
:doc &of,

/*
Print the value as a String to stdout.
Accepts: String Int Uint Float Bool DateTime.
Other objects typically just print the type name.
( val:a -- )
*/
:doc $

/*
Print the value as a String to stderr.
Same natural-string rules as $.
( val:a -- )
*/
:doc $$

/*
Convert an Int Unicode code point to a one-character String.
Compose with $ or $$ to print (e.g. r>s $).
For fmt sandwiches, prefer %c with the Int code point.
( rune:i -- s )
*/
:doc r>s

/*
Format a string for printing, substituting values.
Typically used with a preceding ( ) with the values.
Works like C and Go Sprintf-style formats, with %
value type markers.
There must be one value per % marker, or panics.
Often-used markers:
%v - default value (a good default)
%s - String
%d - Int
%f - Float
%t - Bool
%c - Rune (Int that is a Unicode Rune)
Supports a variety of width and precision specifiers.
try: https://yourbasic.org/golang/fmt-printf-reference-cheat-sheet/
( values:List fmtString:s -- result:s )
*/
:doc fmt

/*
Delay the specified number of milliseconds.
Note: a millisecond is a LONG time on modern CPUs...
Caution: in final-gui, the EventLoop stops during the delay.
see also: delayUs for microseconds.
( delMs:i -- )
*/
:doc delayMs

/*
Delay the specified number of microseconds.
Caution: in final-gui, the EventLoop stops during the delay.
see also: delayMs for milliseconds.
( delUs:i -- )
*/
:doc delayUs
