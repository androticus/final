# TypedObject

**TypedObject** is how you define your own record types in Final: a fixed set of named fields, each with a declared Type, registered in the type catalog. After registration you create instances with `YourType.new^`, read and write fields like other objects, attach **class constants**, and define **Final methods** with a simple naming convention.

Unlike **DynamicObject**, the schema is declared up front and shared by every instance of the type.

## Define and register a type

Typical flow:

1. `TypedObject.newTypeDef^` — empty schema shell (`.name`, `.fields`, `.constants`)
2. Fill `.name` and `.fields` (each field value must be a live catalog **Type**, such as `String` or `List`, not a type-name string)
3. Optional: fill `.constants` with `value .name!` (readonly class consts)
4. `TypedObject.makeType^` — seal and register the type
5. Optional: `: YourType#method … ;` / `: YourType##classMethod … ;` — Final methods
6. `YourType.new^` — create instances

### Example schema

The schema for a TypedObject is most conveniently defined using paren operators, as in this example:

```forth
// you will be building DynamicObjects, with nesting
{
  "Person" "name" // your type name

  // now define your fields
  // you provide the Type object of each field (not the name in quotes)
  {
    String "firstName"
    String "lastName"
    Int    "age"
  } "fields"

  // optional: define class readonly constant values
  {
    111 "defaultAge"
  } "constants"
}
TypedObject.makeType^ // your type is now available to use or add methods to

Person.defaultAge@ $ nl    // 111
```

### Create and fill an instance

```forth
Person.new^ >ts
"Ada" t:.firstName!
"Lovelace" t:.lastName!
36 t:.age!

t:.firstName@ $ sp t:.age@ $ nl
0 t:@() $ sp                  // firstName by declaration ordinal
1 t:@() $ nl
t:drop
```

Field stores are checked against the declared type. Accessors:

| Form | Notes |
|------|--------|
| `.field@` / `.field!` | Named property |
| `@{}` / `!{}` / `?{}` | Keyed by field name |
| `@()` / `!()` | Ordinal by **field declaration order** (faster in hot loops) |

`foreach` on a TypedObject walks fields in **declaration order**; `i` is a **KeyValue**.

## Final methods

After `makeType`, define instance methods with `: TypeName#methodName … ;` and class methods with `: TypeName##methodName … ;`. The `;` registers the method on the type; **no dictionary word** remains.

The instance (or, for `##`, the Type) is on the **Object Stack**; use the `o:` prefix. (You do not push or drop OS yourself, the runtime takes care of that.)

```forth
/*
Print the fieldnames and values for the instance.
( -- )  o( inst:Person )
*/
: Person#print
  ( o:.firstName@ o:.lastName@ o:.age@ ) "%s %s (%d)\n" fmt $
;

/*
Create a new Person with defined default values.
( -- person:Person )  o( type:Person )
*/
: Person##newDefault
  o:.new^ >ts
  "anon" t:.firstName!
  "" t:.lastName!
  Person.defaultAge@ t:.age!
  ts>
;
```

```forth
Person.newDefault^ >ys
y:.print^
"Bilbo" y:.firstName!
"Baggins" y:.lastName!
y:.print^
y:drop
```

Member names are unique across fields, constants, and instance/class methods on that type.

If you include the recommended prefix /* */ comments before your method, it gets included in your type catalog, for later listing using `help MyType`.

While a Final method runs, `o:` keeps the receiver in view for nested property and method paths.

## Base Types

Some Final Types are designed to be used with an instance of a TypedObject, so that the base object can invoke Final callback methods and event handlers on the TypedObject, called the `delegate`. The documentation for such types will specify the need for your delegate instance. When you define such a TypedObject, you include an additional specifier, `baseType`. The following example uses the final-gui SimGame type as an example:

```forth
// you will be building DynamicObjects, with nesting
{
  "Person" "name" // your type name
  SimGame  "baseType" // this will create a `base` field that will be a SimGame
// same as a regular type: fields and optional constants
```

When you create an instance of your type, it automatically creates an instance of the base type, and sets that into the `base` field that is automatically added to your type. The base type will have a field named `del` (for delegate) that it initializes with your TypeDef object. It uses that to forward callback and event method invocations.

A Base Type will typically require you to implement certain named methods on your type. For example, SimGame requires a #preInit and #postInit pair, which it calls during its #init. A base type may include optional methods that are required if you enable a specific feature of the base type. For example, SimGame has options for keyboard and mouse event handling, and when you enable those features, it requires you to define handler methods for those events. See the documentation for individual base types for the properties and methods of the base type, and the methods, and possibly fields it requires you to implement on your delegate TypedObject type.

Most Base Types will have a template in the examples folder that provide stub implementations of the required and optional events, and a stub TypeObject schema. You can copy this template, rename the file and the type name instances in the file, and you will be ready to add your own custom code.
