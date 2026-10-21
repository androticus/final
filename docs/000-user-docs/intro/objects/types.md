# Types

A **Type** is Final's description of a value family: instances of which can have properties, methods, and collection accessors. A Type may also have class properties and class methods, both accessed through the Type itself.

Types are first-class objects. CapCase names such as `Int`, `List`, and `DynamicObject` are dictionary words that push the Type onto DS.

```forth
List $          // prints the Type
List.new^ $     // factory: empty List on DS
```

Primitives (`Int`, `String`, `Bool`, …) do not have object *instances*, but they still have a Type with useful **class** members (constants, conversions, factories). Reference types (`List`, `Map`, `Word`, …) have both class members and instance members.

## Class vs instance

| Receiver | What you can call | Examples |
| -------- | ----------------- | -------- |
| The **Type** itself | class properties, class methods | `Math.pi`, `Int.maxInt`, `List.new^` |
| An **instance** | instance properties, instance methods, items accessors | `t:.len@`, `t:.append^`, `t:@()` |

Class methods are never invoked from an instance (`myList.new` is invalid). Factories live on the Type: `List.new^`, `Person.new^`.

By convention:

* default factory is `new` — `( -- inst )` with default values or zeros in the properties
* parameterized factories use names like `newWithCapacity`, `newFromString`, etc.

```forth
16 List.newWithCapacity^ >ts
"Ada" StringBuilder.new^ >xs
```

## Exploring types

```forth
help List
help DynamicObject
help Word
```

`help` lists properties, methods, class members, and which item accessors the type supports.

## Prototype families

Final groups related concrete types under prototype parents such as **ProtoList** and **ProtoMap**. That is mainly a catalog / sharing story (for example Json rendering treats all list-like and map-like types uniformly). Day-to-day code still names the concrete type (`List`, `StringMap`, `DynamicObject`).

## Your own types

[TypedObject](../compounds/typedobject.md) registers a named schema in the catalog. After `makeType`, the new CapCase name behaves like any other Type: `YourType.new^`, dotted fields, optional Final methods.
