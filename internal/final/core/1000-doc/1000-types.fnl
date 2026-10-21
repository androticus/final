// Basic Types

/*
A primitive whose only value is 'null'.
This is provided primarily for Json interop.
Its use otherwise is discouraged.
*/
:doc Null

/*
A primitive that can either be 'true' or 'false'.
Note that Final supports the concepts of 'truthy'
and 'falsey' in tests that take a Bool.
*/
:doc Bool

/*
A signed 64 bit integer.
Note that this is Final's preferred integral type.
*/
:doc Int

/*
An unsigned 64 bit integer.
This is provided mostly for interop purposes, typically
use Int unless the range of your value is > 2^63-1.
*/
:doc Uint

/*
An IEEE-754 64 bit floating point value.
*/
:doc Float 

/*
A Utf-8 encoded immutable string value.
Note: 7-bit ASCII is a subset of Utf-8, and is
encoded as 1 byte per character.
*/
:doc String

// Type probes — membership only; value is consumed

/*
True if the value is Null.
( v:a -- b )
*/
:doc Null?

/*
True if the value is a Bool.
( v:a -- b )
*/
:doc Bool?

/*
True if the value is an Int.
( v:a -- b )
*/
:doc Int?

/*
True if the value is a Uint.
( v:a -- b )
*/
:doc Uint?

/*
True if the value is a Float.
( v:a -- b )
*/
:doc Float?

/*
True if the value is a String.
( v:a -- b )
*/
:doc String?

// Representation casts — compatible primitives; no string parsing.
// Modal: success leaves value and true; failure leaves only false.

/*
Convert a Truthy value to Bool.
Numeric: zero is false, non-zero is true.
Does not accept String (use >>Bool?).
( v:a -- b true | false )
*/
:doc >Bool?

/*
Convert to Int.
Uint and Float (truncated toward zero) accepted.
Large Uint may wrap (Go bit-cast).
( v:a -- i true | false )
*/
:doc >Int?

/*
Convert to Uint.
Int and Float (truncated toward zero) accepted;
negatives wrap (Go bit-cast).
( v:a -- u true | false )
*/
:doc >Uint?

/*
Convert to Float from Int, Uint, or Float.
( v:a -- f true | false )
*/
:doc >Float?

/*
Same as >Bool?, but panics if conversion fails.
( v:a -- b )
*/
:doc >Bool

/*
Same as >Int?, but panics if conversion fails.
( v:a -- i )
*/
:doc >Int

/*
Same as >Uint?, but panics if conversion fails.
( v:a -- u )
*/
:doc >Uint

/*
Same as >Float?, but panics if conversion fails.
( v:a -- f )
*/
:doc >Float

// Semantic casts — representation plus string parsing where noted.
// Magnitude-safe for Int/Uint: no silent wrap of out-of-range values.

/*
Convert the string "null" to Null.
( v:a -- null true | false )
*/
:doc >>Null?

/*
Convert to Bool, including String "true" / "false"
(case-sensitive), plus >Bool? rules.
( v:a -- b true | false )
*/
:doc >>Bool?

/*
Convert to Int without wrapping.
Rejects Uint larger than Int.maxInt.
Also parses integral String (no float, no u/U suffix).
( v:a -- i true | false )
*/
:doc >>Int?

/*
Convert to Uint without wrapping.
Rejects negative Int/Float.
Also parses non-negative integral String.
( v:a -- u true | false )
*/
:doc >>Uint?

/*
Convert to Float, including parseable numeric String.
( v:a -- f true | false )
*/
:doc >>Float?

/*
Natural string form of Bool, Int, Uint, Float, or Null.
Does not accept String (use String? for that).
( v:a -- s true | false )
*/
:doc >>String?

/*
Same as >>Null?, but panics if conversion fails.
( v:a -- null )
*/
:doc >>Null

/*
Same as >>Bool?, but panics if conversion fails.
( v:a -- b )
*/
:doc >>Bool

/*
Same as >>Int?, but panics if conversion fails.
( v:a -- i )
*/
:doc >>Int

/*
Same as >>Uint?, but panics if conversion fails.
( v:a -- u )
*/
:doc >>Uint

/*
Same as >>Float?, but panics if conversion fails.
( v:a -- f )
*/
:doc >>Float

/*
Same as >>String?, but panics if conversion fails.
( v:a -- s )
*/
:doc >>String

/*
Parse a JSON number token into Int or Float.
( s:String -- n:a true | false )
*/
:doc String#parseJsonNumber

/*
Copy the Utf-8 bytes of this String into a new ByteList.
Pair with ByteList#asString for round-trip.
( -- bytes:ByteList )
*/
:doc String#asBytes

// String helpers (str::*)

/*
Byte length of a Utf-8 String (not rune count).
( s:String -- len:i )
*/
:doc str::len

/*
Trim leading whitespace.
( s:String -- s' )
*/
:doc str::trimStart

/*
Trim trailing whitespace.
( s:String -- s' )
*/
:doc str::trimEnd

/*
Trim leading and trailing whitespace.
( s:String -- s' )
*/
:doc str::trim

/*
Concatenate two Strings.
( a:String b:String -- ab:String )
*/
:doc str::cat

/*
Synonym for str::cat.
( a:String b:String -- ab:String )
*/
:doc str::+

/*
True if the String is empty.
( s:String -- b )
*/
:doc str::isEmpty

/*
True if s begins with prefix.
( s:String prefix:String -- b )
*/
:doc str::startsWith

/*
True if s ends with suffix.
( s:String suffix:String -- b )
*/
:doc str::endsWith

/*
Byte index of the first occurrence of sub in s,
or -1 if not found.
( s:String sub:String -- idx:i )
*/
:doc str::indexOf

/*
Split s on separator into a StringList.
( s:String sep:String -- parts:StringList )
*/
:doc str::split

/*
Split s into lines (on newline).
( s:String -- lines:StringList )
*/
:doc str::splitLines

/*
Join a List or StringList of Strings with sep.
( items:List|StringList sep:String -- s:String )
*/
:doc str::join

/*
Lowercase s (Unicode casing rules).
( s:String -- s' )
*/
:doc str::toLower

/*
Uppercase s (Unicode casing rules).
( s:String -- s' )
*/
:doc str::toUpper

/*
Replace every occurrence of find with replace.
( str:s find:s replace:s -- str' )
*/
:doc str::sub
