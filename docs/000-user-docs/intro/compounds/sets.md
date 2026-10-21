# Sets

A Set is similar to a Map without values. It holds a single instance of any key entered. Sets, like Maps are **unordered**, so you can query for the existence of a value, but not access the values in ordinal order.

Set operators are provided as follows:
* set::union -- produces a set that contains the elements of both sets
* set::intersection -- produces a set with only the items in both sets
* set::difference -- produces a set of the items of the first set, with those in the second set removed from it
* set::symmetric difference -- is like an XOR, it returns the unique elements in each set, but not those shared by both sets

Final provides two Set objects:
* Set, which can take any comparable type, like numerics and String
* StringSet, which holds String values

A Set can be iterated with `foreach`; `i` returns the value.

A Set can be created using the List operators `( ... set)` for a Set and `( ... sset)` for a StringSet.

A Set is rendered in JSON as an array; the items are in an undefined order. A Set cannot be produced by a JSON string.