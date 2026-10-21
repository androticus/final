# Methods

A **method** is a named operation on a Type. Invoke it with the **`^`** ("run") affix. Bare `.method` without `^` is an error.

Operands and results use **DS**. The receiver stays on the context stack for rooted calls.

## Instance methods

```forth
List.new^ >ts
10 t:.append^
20 t:.append^
t:.len@ $ nl      // 2
t:.clear^
t:drop
```

| Form | Effect |
| ---- | ------ |
| `cs:.name^` | invoke instance method; fused one-shot |
| `.name~` then `^` | materialize MethodRef, then invoke (ref consumed) |

```forth
StringBuilder.new^ >ts
"hi" t:.writeString^
t:.toString^ $ nl    // hi
t:drop
```

## Class methods

Class methods run with the **Type** as receiver. Factories are the common case:

```forth
List.new^                         // ( -- list )
16 List.newWithCapacity^          // ( cap:i -- list )
5 List.newWithLen^                // ( len:i -- list )
```

Other class methods query or compute without an instance handle:

```forth
Float.NaN Float.isNaN^ $ nl    // true
2 10 Float.pow^ $ nl           // 1024.0
```

For string→number conversion, use the cast operators from the Quick Tour (`"42" >>Int?`), not a Type method.

## Final methods (TypedObject)

Methods written in Final are defined on a TypedObject type (see [TypedObject](../compounds/typedobject.md)) with `: Type#name … ;` (instance) or `: Type##name … ;` (class). Registration happens at `;` — no dict word is left behind. While the method runs, the receiver is on the **Object Stack**. Use the `o:` prefix; you may not push or drop OS yourself; the runtime takes care of pushing the instance prior to invoking the method, and then dropping it when the method returns.

```forth
: Person#describe
  o:.name@
  o:.age@
  "%s (%d)" fmt
;

Person.new^ >ts
"Ada" t:.name!
36 t:.age!
t:.describe^ $ nl    // Ada (36)
t:drop
```

Nested `o:.…` paths behave like any other rooted compound path for the duration of the call.
