# Objects and Types

Final provides first-class support for objects, including a concise, expressive syntax and operator set for accessing the members of an object.

All objects in Final are instances of a Type, which is Final's object that holds the definition of an object's members.

An object may have any of the following:
* instance properties, which can be read only
* instance methods, which perform an operation
* class properties, often used for constants
* class methods, often used for instance factories
* ordinal accessors, such as for lists: @() and !()
* name accessors, such as for Maps: @{} !{} and ?{}
* vec and matrix accessors: @[] and ![]

You can get detail on all the properties and accessors provided by a type, by entering `help <typename>` in the UI.

Final's type system is hierarchical, and supports a form of inheritance called prototype inheritance. This is primarily a conceptual concept, that groups members common to a family of concrete instance types. Such types have names like ProtoList and ProtoMap. It is also used by a number of Final capabilities, such as Json rendering, to provide Final methods that will apply to all the child types. Json uses this to generate Json arrays for all List types, and Json objects for all Map types, even those that have no implementation relation, such as Map and DynamicObject.

Many of the Types that Final uses and exposes for your use are written in Go, Final's implementation language. But all Objects and Types, whether they are go-native or instances of your defined objects, such as Dynamic and Typed Objects, share the same means of access and operation, introduced next.

## Compound Path Operations

Legacy Forth had a basic policy that all operations be expressed by single, whitespace-separated tokens. During development of Final's Type and Object system, this approach to object operations was found to be lacking in terms of user understanding, and convenience. So a special sub-language of composable object operations was created. These "Compound Path Operations" are a core part of Final, and used extensively. We cover the basics here, but invite you to dive deeper in the [Objects](../objects/index.md) deep-dive.

The simplest example of an object operation is accessing a readonly constant value of a Type. This convention is used to provide enum values, Type-related values, and math constants. To do this, you simply go TypeName.property. Here are common examples:

```forth
Math.pi // pushes 3.14 etc on DS
Int.maxInt
Float.NaN / pushes the IEEE-754 NaN value
```

Many types provide factory methods for creating a new instance of the type. Here is an example of creating a DynamicObject, and then moving it to the Temp Stack for further operations. (Final's extended set of auxilliary stack are covered in-depth in the topic [Stacks](../stacks/index.md)). The following sequence shows the steps in creating an instance of a DynamicObject, setting some field values, then reading them.

```forth
DynamicObject.new^ // ( -- do:DynamicObject )
>ts // ( do:DO -- ) t( -- do:DO ) move it to Temp Stack
1 t:.one! // ( 1 -- ) t( do:DO ) creates/sets .one property to 1
t:.one@ // ( -- 1 ) t( do:DO ) fetches the value of .one property
".one is: " $ $ nl // nl prints a newline

.one is: 1
```

Common Path operations:
* moving values to/from auxilliary stacks use eg for Temp Stack, >ts ts>
* the "Context Selector", which specifies on which stack the object resides, uses the cs: prefix, where cs can be: t x y z o If you see `cs:` that means that "cs" will have one of those actual values
* Methods are invoked with .name then using the ^ affix operator.
* Property values are referenced with .name  
  - to store a value, the value should be on DS, then use the ! "store" affix
  - to retrieve a value, which goes to DS, use the @ ("fetch") affix
  - readonly properties can omit the @ affix

Due to the logic of how stack-based languages use the Data Stack for the results of operations, it was decided in Final that most compound path operations must start with the object, called the "receiver" (i.e. receives the access commands), on an auxilliary stack. So a common sequence in Final is creating a new object instance using a factory method, which results in the object on DS, then moving it to an auxilliary stack for all further operations.

Final provides convenience operators for the most common ordinal and by-name operations on objects that support those operations, like lists and maps. Here is a quick summary:
* cs:@() // ( index:i -- value ) cs( obj )
* cs:!() // ( value index:i -- ) cs( obj )
* cs:@{} // ( name:s -- value ) cs( obj )
* cs:!{} // ( value name:s -- ) cs( obj )
* cs:?{} // ( name:s -- exists:b ) cs( obj )
Matrix operators, examples for 2d mats
* cs@[]  // ( row:i col:i -- value )
* cs![]  // ( value row:i col:i -- )

The `help` for every Type lists the accessors that are supported by instances of that type. Usually an object only supports the `()` set or `{}` set, but some, like DynamicObject, support both. Mats support the `()` accessors for flat indexed access of the values, and the `[]` accessors for eg row/col in 2d Mats.

Paths can be used to walk object hierachies, that include nested instances of other objects. You just use .propertyname for the contained instances. The intermediate objects fetched during such a "path walk" are automatically instantiated and elided by the path walker, so never appear on any stack. For example, to find the length of your current dictionary, you walk a path starting with the DictionaryContext DC value:

```forth
dc.dict.len $

24
```
Indicating you have defined 24 words in your local dictionary so far.
