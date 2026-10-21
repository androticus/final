// Final Lists

/*
A general purpose list that can hold any primitive or 
compound Final type in each cell.
*/
:doc List

/*
A strongly typed List that efficiently holds String values.
See List for member descriptions.
*/
:doc StringList

/*
A strongly typed List of unsigned 8-bit bytes (0…255).
Useful for interop with String, Byte Stream, and DB cols whose
underlying storage is a Go []byte slice.
See List for member descriptions.
*/
:doc ByteList

/*
Decode the byte list as a Utf-8 String (copies the bytes).
Pair with String#asBytes for round-trip.
Invalid Utf-8 is accepted as Go does for string([]byte).
( -- s:String )
*/
:doc ByteList#asString

/*
A strongly typed List that efficiently holds uint16 values.
See List for member descriptions.
*/
:doc Uint16List

/*
A strongly typed List that efficiently holds uint32 values.
See List for member descriptions.
*/
:doc Uint32List

/*
A strongly typed List that efficiently holds Uint values (64 bits, Go uint).
See List for member descriptions.
*/
:doc UintList

/*
A strongly typed List that efficiently holds signed int8 values.
Note: NOT a ByteList — Final bytes are unsigned 8-bit (0…255).
See List for member descriptions.
*/
:doc Int8List

/*
A strongly typed List that efficiently holds int16 values.
See List for member descriptions.
*/
:doc Int16List

/*
A strongly typed List that efficiently holds int32 values
(including Unicode code points when used that way).
See List for member descriptions.
*/
:doc Int32List

/*
A strongly typed List that efficiently holds Int values (64 bits, Go int)
See List for member descriptions.
*/
:doc IntList

/*
A strongly typed List that efficiently holds float32 values
(widened to Float on the stack).
See List for member descriptions.
*/
:doc Float32List

/*
A strongly typed List that efficiently holds Float values (64 bits, Go float64).
See List for member descriptions.
*/
:doc FloatList

/*
Returns the number of items in the List.
( -- len:i )
*/
:doc List#len

/*
Add an item to the end of the List.
( item:a -- )
*/
:doc List#append

/*
Insert an item at the indicated position.
0=beginning; len=end (append).
Existing items are preserved, and moved if needed.
( item:a at:i -- )
*/
:doc List#insertAt

/*
True if any list element equals the given item
(identity/equality of the stored values).
( item:a -- b )
*/
:doc List#hasItem

/*
Remove the first element equal to the given item
(same identity/equality as hasItem). False if absent;
does not panic.
( item:a -- removed:b )
*/
:doc List#deleteItem

/*
Delete the item at the index, moving remaining
items down.
( at:i -- )
*/
:doc List#deleteAt

/*
Remove all items from the list.
Note: it is faster to just create a new List
via List.new^
*/
:doc List#clear

/*
Insert an item at the beginning of the list,
treating the list as a queue.
( item:a -- )
*/
:doc List#enqueue

/*
Remove the item at the end of the list,
treating the list as a queue.
( -- item:a )
*/
:doc List#dequeue

/*
Push an item at the end of the list,
treating the list as a stack.
( item:a -- )
*/
:doc List#push

/*
Remove the item at the end of the list,
treating the list as a stack.
( -- item:a )
*/
:doc List#pop

/*
Get the item at the end of the list, leaving
the item in place, treating the list as a stack.
( -- item:a )
*/
:doc List#peek

/*
Synonym for len, treating the list as a stack.
( -- len:i )
*/
:doc List#depth

/*
(Pseudo-)Randomly reorder the items in the list.
( -- )
*/
:doc List#shuffle

/*
List supports indexed read and write.
*/
:doc List#__items

/*
Create a new empty List.
( -- )
*/
:doc List#new

/*
Sort the list in the natural order for the items.
Note: not all list types support sorting. It is
supported if it shows `sort` in its base method list.
( -- )
*/
:doc ProtoList#sort

/*
Sort the list in order of the Utf-8 byte values.
This differs from conventional alphabetic sorting.
In ASCII this results in 'A' before 'a'.
In Unicode, "caffeine" is before "café"; and
"Österreich" will sort at the end of a list.
( -- )
*/
:doc StringList#sort
