# Basic Types

Final has a set of basic, primitive types, and a rich set of collection and other compound object types. We'll cover the basic types here.

Basic types are immutable, meaning that once created, their value does not change. In operations using the value and returning a value, a new value is created; you cannot alter or change an instance of a basic type.

## Null

The Null type has only one value: `null`. It exists primarily to support Json. It's use otherwise is discouraged. Final has a pattern it widely uses for indicating there is "no value" returned from an operation, so Null is not used by Final, except when generating or parsing Json, or initializing strongly typed fields with no value yet.

## Bool

The Bool type has the values `true` and `false`. It is not a numeric type. However, Final has the concept of "Truthiness", which enables some of the other types to be used as if they were Bool.

## Numeric types

Final has two integral types, and one real type.

* `Int` - a 64 bit signed integral value;
* `Uint` - a 64 bit unsigned integral value;
* `Float` - a 64 bit IEEE-754 floating point value;
* `Rune` - a synonym for Int, representing a 32 bit unicode character.

`Int` is the type you should use for most integral operations and representations, even in contexts such as an ordinal index, in which negatives do not occur. Almost all of Final's apis that use or return an integral value use Int.

`Uint` is provided primarily for contexts which require integral values larger than 2^63-1, or for flags or bitfields that use bit 63, in which you don't want the natural printing to default to displaying a negative, 2's complement interpretation. Note that `Int` and `Uint` do not interoperate in dyadic operations; type coercion is covered in the next topic. You create a Uint literal by appending an integral value with the `u` suffix.

`Rune` is a synonym for Int, and is used in contexts involving Unicode code points, called `runes`. A Unicode rune is a 32 bit integer value.

## String

A Final `String` is an immutable sequence of Unicode characters using Utf-8 encoding. All mutating operations on strings accept immutable string inputs, and return a new immutable string value. (Final provides mutable `StringBuilder` to successively build a `String` from parts.) Strings are not 0-terminated. Strings are copied by reference throughout the runtime, so there is only one underlying value, regardless how many times it is moved or copied.

Utf-8 encoding provides an excellent compromise between byte size, efficiency, and interoperability with the Web and Json, as well as other benefits. ASCII-7, i.e. 7-bit ASCII is a proper subset of Utf-8, so an ASCII string in bytes is identical in Utf-8. Unicode runes beyond ASCII are encoded in either 2, 3, or 4 byte sequences, with the majority of Western/Latin characters only needing 2 or 3 bytes. It is thus more storage efficient than Utf-32, which uses a fixed 4 bytes per character.

One caution: the concept of "length" in Unicode in general, and especially Utf-8 requires decoding the entire string, so Final does not provide a true string length. The String.len property is the length in bytes of the encoding. If the string is known to contain only ASCII-7, then that is also the length. Final provides an efficient mechanism to iterate the Runes of a String, which is explained in the [String](../strings/index.md) topic.

## `Any`

`Any` is a type that can hold any other Final type, basic or compound/object. The DS is composed of a growable stack of `Any` values. So when a concrete type goes on DS, it will always be wrapped in an Any. 

## "Truthiness"

For convenience, you can use a value of the following types in most operations expecting a `Bool`, such as boolean logic operators, and `if` and `case` control structures.

* Numeric types: exact 0 is false, non-0 is true
* String: empty "" is false, anything else is true

No other types are accepted, and will raise a runtime error if used.

## DateTime and Interval

These are also basic types, but discussion is deferred to [DateTime and Interval](../datetime.md).

Let's move on to a discussion and demonstrations of the Final Interpreter, which includes details on expressing literal values of the basic types in the interpreter and compiled words.

