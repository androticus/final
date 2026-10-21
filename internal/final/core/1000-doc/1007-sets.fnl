/*
Documentation for Sets.
*/

/*
A set of unique comparable values (also known as AnySet).
Holds at most one of each value; values must be equatable
primitives such as Bool, Int, Uint, or String.
Prefer StringSet when all values are Strings.
Create empty with Set.new^ , or gather with ( … set) .
Supports foreach (order undefined). See list-maps for set:: ops.
*/
:doc Set

/*
A set of unique String values.
More efficient than Set when every value is a String.
Create empty with StringSet.new^ , or gather with ( … sset) .
See Set for shared behavior; see list-maps for set:: ops.
*/
:doc StringSet

/*
Returns the number of values in the Set.
( -- len:i )
*/
:doc Set#len

/*
Add the value if not already present; otherwise do nothing.
( value:a -- )
*/
:doc Set#put

/*
True if the Set contains the value.
( value:a -- has:b )
*/
:doc Set#has

/*
Set supports membership test via ?{} (same as has).
No @{} or !{} — those need a value in addition to the key.
*/
:doc Set#__items

/*
Create a new empty Set.
( -- set:Set )
*/
:doc Set#new

/*
Returns the number of values in the StringSet.
( -- len:i )
*/
:doc StringSet#len

/*
Add the String if not already present; otherwise do nothing.
( value:s -- )
*/
:doc StringSet#put

/*
True if the StringSet contains the String.
( value:s -- has:b )
*/
:doc StringSet#has

/*
StringSet supports membership test via ?{} (same as has).
No @{} or !{} — those need a value in addition to the key.
*/
:doc StringSet#__items

/*
Create a new empty StringSet.
( -- set:StringSet )
*/
:doc StringSet#new

/*
Union of two sets: every value that appears in a or b.
Both operands must be the same sealed type (Set or StringSet);
result is a new set of that type.
( a:Set|StringSet b:Set|StringSet -- c:Set|StringSet )
*/
:doc set::union

/*
Intersection of two sets: values that appear in both a and b.
Both operands must be the same sealed type (Set or StringSet);
result is a new set of that type.
( a:Set|StringSet b:Set|StringSet -- c:Set|StringSet )
*/
:doc set::intersect

/*
Difference of two sets: values in a that are not in b.
Both operands must be the same sealed type (Set or StringSet);
result is a new set of that type.
( a:Set|StringSet b:Set|StringSet -- c:Set|StringSet )
*/
:doc set::difference

/*
Symmetric difference: values in a or b but not in both.
Both operands must be the same sealed type (Set or StringSet);
result is a new set of that type.
( a:Set|StringSet b:Set|StringSet -- c:Set|StringSet )
*/
:doc set::symmetricDifference
