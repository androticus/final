# Compound Types

Final provides a useful set of compound objects. These are described briefly below, and covered in detail in the section [Compound Types](../compounds/index.md).

## Lists

List is Final's dyanamically sizable array type. A List starts out empty, after which items can be added, inserted, or deleted. A List can grow to any size.

Lists provide the @() and !() ordinal accessors for convenient reading/writing of existing elements. The List Type object provides the remaining API for adding, inserting, and deleting. 

Lists can be iterated in a `foreach` loop. Just put the List on DS prior to your `foreach` statement, and the iterator is then automatically provided. The `i` value is the current item in the iteration. Iteration is deteriministic over the range 0..Len-1.

### List

The type called List is Final's general purpose list. Its items are of type Any, so it can hold any Final basic or object type. Items do not have to be the same type.

### StringList

StringList is optimized for the common case of a List of String values. It may only hold values of type String.

### Numeric Lists

Final provides a strongly typed, memory-efficient list for all the following numeric types: int8, int16, int32, int64 (aka Int), unsigned 8-bit (**ByteList**), uint16, uint32, uint64 (aka Uint), float32, float64 (aka Float). Values are automatically boxed/unboxed to Any on read/write.

As an example, a **ByteList** (unsigned 8-bit bytes, 0…255) provides a 16x reduction in memory required to store byte-sized values, compared to List. ByteList also finds use in the Final runtime for interoperation with String, whose values are stored in an underlying array of bytes; ByteStreams, that are used for byte read/write of files; and in some Http scenarios.

### StringKeyValueList

This specialized list is used in Http classes for defining query parameters, and to hold Http headers, all of which are allowed to have multiple instances of the same key, therefore, Map-like objects cannot be used for these cases. The List stores values of type StringKeyValue in which .Key and .Value are both of type String. The list preserves the ordinal order of the source, such as in HttpRes headers.

The `i` for `foreach` iteration is a StringKeyValue object.

## Maps

A map is a collection of key:value pairs. Access is by key. The first write of a key value creates the entry. Subsequent writes to that key, overwrite the previous value.

Maps provide the @{} and !{} accessors for reading/writing a value by its key. For reads, the key must exist, otherwise an error occurs. The ?{} accessor can be used to determine if the key exists.

Maps can be iterated. Place the Map object on DS prior to a `foreach`. The order of iteration is indeterminate. `i` is an object of type KeyValue, providing the key and the value of the entry.

Final provides 2 Map variants.

### Map

Map holds keys and values of type Any. The values can be of any type, but the key must be a type that can be compared, such as String or Numbers. If you try to use an object as a key, an error will occur. Map is provided primarily to support the scenario of using numeric keys.

### StringMap

StringMap is specialized for the common case of a map whose keys are Strings. Only a String is accepted as a key. Like Map, values can be of any type.

### Set

Set holds values of any comparable type. It only contains one instance of any value added.

Sets have operations for producing a new Set from 2 input Sets: union, intersection, difference, and symmetric difference. Sets can be iterated

### StringSet

StringSet is specialized for the common case of a set whose values are Strings. .

## Vecs and Mats (Vectors and Matrices)

A Vec is a one dimensional ordinal collection of values with a fixed length set at construction time. Final provides VecNf which holds Float values, and can be created with any length. Vec use and its operators are covered in the [Linear Algebra](../la/index.md) section.

A Mat is a multi-dimensional ordinal collection of values, with fixed dimension sizes set at construction time. Final currently provides a variety of 2-dimensional Mats, addressed using the row/column convention. They are stored in row-major order into an underlying flat array of values. The Mat2f Mat type, in addition to being useful generally, also has special operators and roles in the Linear Algebra/Ordinary Least Squares regression system.

Mats and Vecs both use the `@[]` and `![]` accessors to read and write values. No additional method-based API is provided. For Vecs, you provide one index, For Mats you provide the indexes in order of outer-level to inner-level; for 2-d mats, your provide ( row:i col:i -- value ) for `@[]` and ( value row:i col:i -- ) for `![]`.

### Mat2TN

Where T is type, N is bits. A family of numeric Mats that hold signed/unsigned integral or float values, of different sizes. Some, in addition to being useful generally, also have special use in the [Drawing/Blitting](../draw-gui/index.md) framework.

### Mat2c

A Mat that stores Color values; also used in Blitting.

### Mat

A mat, comparable to List, that can store any Final basic, compound type, or object instance.

## DynamicObject

DynamicObject (DO) is a little dynamo of a type, and is used often by the Final runtime, and in Final code. A DO is similar to a Map, but has less overhead, and is faster to read and write than a Map for a number of entries in the range of less than ~64 (however it does not have a size restriction, and automatically optimizes access for field counts greater than 64.). You can also access the fields by name using the dotted .fieldName accessor.

You can use it like an object, creating fields dynamically using the .fieldName! notation, or like a map, using the !{} accessor. A field must be created before it may be read; if a read is attempted on a field that does not exist, a runtime error occurs. Like Map, DO provides the ?{} accessor.

Unlike Map, DO preserves the order of field entry. When you iterate a DO, it iterates in order of field creation. DO also provides a highly optimized set of ordinal accessors @() and !() which read and write a value by its ordinal order of creation. This form of access avoids the more expensive field name lookup, when you are certain of the order. Dyanamic vectors such as .x .y .z for example, can be accessed like an array[2] type of object.

DynamicObject instances are one-offs, but they can be (shallow) cloned, or used as a prototype to create instances of StaticObject.

## StaticObject

A StaticObject is used much like a DynamicObject. The major difference is that a StaticObject is created from a DynamicObject that serves as its prototype. Once created, a StaticObject has only the fields of the DynamicObject at the time of its creation. You cannot add new fields to a StaticObject.

StaticObject instances use much less memory than a DynamicObject, and are used for things like Database query result sets, and Table rows.

## TypedObject

TypedObject is similar to DynamicObject, but its fields are more formally defined, with a schema of fieldname:Type pairs. Once defined, a TypedObject becomes its own Type in the Final Type registry, and can be used to create new instances.

A TypedObject may also have methods, written in Final.

Runtime behavior of a TO instance is similar to a DO, including definite iteration in order of field definition and ordinal accessors for faster access.

When you use the .methodName^ syntax to invoke a Final method on a TO instance, the instance exists on the OS for the duration of the method, providing convenient access to its members via the o: prefix accessor. The instance remains on OS for the lifetime of the method, and is removed automatically by the runtime when the method returns.

## Table

A Table is a set of rows with a header, similar to a spreadsheet. It supports sorting of the rows by any column using progressive stable sorting, which preserves the order of the previous sort-step for ties.

A Table can be made from a DynamicObject, which then becomes a prototype for StaticObject rows. It can be created from a TypedObject, for instances of that type. It can be created from a Type that has fields, in which case, the fields become the column names. It can also be used manually by supplying a list of column names, and then adding any compatible objects as rows. A "compatible object" is any object that provides named or ordinal field access.

The `dbIter>table` word efficiently creates a table from the result set of a Database query.
