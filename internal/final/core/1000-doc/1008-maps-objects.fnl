// Final Maps and Objects

/*
A compound type for storing key:value pairs
(also known as AnyMap).
Keys must be Equatable primitives (String, Int, …);
keys cannot be a compound type.
Values can be any Final primitive or compound type.
Set with !{}; get with @{} (missing key panics);
test with ?{} or hasKey.
See also StringMap for string-keyed maps.
*/
:doc Map

/*
A map optimized for string keys.
Values can be of any Final type.
See Map for member descriptions.
*/
:doc StringMap

/*
Returns the number of key:value pairs in the Map.
( -- len:i )
*/
:doc Map#len

/*
Remove all key:value pairs from the Map.
( -- )
*/
:doc Map#clear

/*
True if the Map contains an entry for the key.
( key:a -- has:b )
*/
:doc Map#hasKey

/*
Remove the entry for the key, if present.
( key:a -- )
*/
:doc Map#deleteKey

/*
Map supports keyed read, write, and existence test
via @{} !{} ?{}.
Missing @{} key panics.
*/
:doc Map#__items

/*
Create a new empty Map.
( -- m:Map )
*/
:doc Map#new

/*
Create a new empty StringMap.
( -- m:StringMap )
*/
:doc StringMap#new

/*
An efficient dynamic map of named field/value pairs.
Names must be strings; values can be of any Final type.
Create with DynamicObject.new^ .
Access fields with .name@ / .name! or @{} / !{} / ?{}.
A field must be set before being read by .name@ or @{}.
Useful for PipParams, Json objects, and ad-hoc records.
Inherits Map-style keyed access from ProtoMap.
*/
:doc DynamicObject

/*
Create a new empty DynamicObject.
( -- o:DynamicObject )
*/
:doc DynamicObject#new

/*
TypedObject is the base for user-defined record types
with a fixed schema (named fields and optional class constants).
Unlike DynamicObject, fields are declared up front and
stored densely; each field has a declared Type.

Typical flow:
  1. TypedObject.newTypeDef^ — empty schema shell
  2. Fill .name, .fields (catalog Types), .constants (values)
  3. TypedObject.makeType^ — register the new Type
  4. : YourType#… ; / : YourType##… ; — Final methods
  5. YourType.new^ — mint an instance

Instances support .field@ / .field!, @{} / !{} / ?{},
and @() / !() by field ordinal, plus foreach over fields.
This help entry is only a sketch — see Final documentation
for TypedObject / TypeDef details and examples.
*/
:doc TypedObject

/*
Create an empty TypeDef shell (a DynamicObject) with
.name, .fields, and .constants ready to fill.
Field values must be live catalog Types (e.g. String),
not type-name strings. Constants use value .name!.
( -- def:DynamicObject )
*/
:doc TypedObject#newTypeDef

/*
Register a new Type from a filled TypeDef shell in the
current module.
Afterward, YourType.new^ creates instances.
( def:DynamicObject -- )
*/
:doc TypedObject#makeType

/*
Create a new instance of a user TypedObject type
(zeros for each declared field).
Call as YourType.new^ after TypedObject.makeType^ .
( -- o:TypedObject )
*/
:doc ProtoTypedObject#new

/*
A single key:value pair (Any key, Any value).
Produced when iterating a Map.
*/
:doc KeyValue

/*
The key of a KeyValue.
*/
:doc KeyValue#key

/*
The value of a KeyValue.
*/
:doc KeyValue#value

/*
Ordered list of string key / string value pairs.
Duplicate key values are allowed.
Used for HTTP headers and query parameters.
!{} upserts the first matching key; add always appends
(so duplicate keys are allowed on the wire).
@{} gets the first match (panics if missing);
?{} tests for any match.
*/
:doc StringKeyValueList

/*
Create an empty StringKeyValueList.
( -- list:StringKeyValueList )
*/
:doc StringKeyValueList#new

/*
Append a value/key pair (does not upsert).
Stack order: value then key.
( value:String key:String -- )
*/
:doc StringKeyValueList#add

/*
Keyed read/write/has via @{} !{} ?{}
(first matching key wins for get/set).
*/
:doc StringKeyValueList#__items
