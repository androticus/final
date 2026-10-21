# Rendering

The table below lists the Final types that are supported for Json rendering, and what Json type they render to. The entire string is in Utf-8 encoding.

| Final type | Json type | Note |
| ---------- | ---- | ---- |
| Null | null | |
| Bool | true / false | |
| Int | number | 1 |
| Uint | number | 1 |
| Float | number | 2 |
| String | string | 3 |
| List | array | 4 |
| Map, DynamicObject, TypedObject | object | 5 |
| DateTime | string | 6 |

1. No decimal.
2. With a decimal.
3. `\t`, `\r`, and `\n` are escaped.
4. The type of each list value determines the resulting Json type
5. The type of each Map value determines the resulting Json type
6. DateTime is formatted in ISO format, per RFC 3339

Nesting of structured objects in Lists and Maps is fully supported.

There are 2 words in Final used to create a Json string:

Create a compact json result, no unnecessary spaces, no newlines:
```
// ( root:a -- json:s )
>json
```
Create a "pretty" json result, with nesting and newlines, uses two space indenting:
```
// ( root:a -- json:s )
>jsonp
```

"Pro tip" use `>jsonp $` in Final when you want a quick display of collection objects like Lists, Maps, and Dyanamic/Typed Objects!

## Examples

### Primitives

```forth
null >json $ nl          // null
true >json $ nl          // true
-49 >json $ nl           // -49
3.14 >json $ nl          // 3.14
"hi\n" >json $ nl        // "hi\n"
```

### List (array)

`( … )` builds a temporary List on DS — convenient for small Json arrays:

```forth
( null true -49 "rain" ) >json $ nl
// [null,true,-49,"rain"]

( null true -49 "rain" ) >jsonp $
```

Pretty output uses two-space indent:

```json
[
  null,
  true,
  -49,
  "rain"
]
```

### DynamicObject (object)

```forth
DynamicObject.new^ >ts
null t:.Null!
true t:.Bool!
-49 t:.Int!
"hello" t:.msg!
ts>
>json $ nl
// {"Null":null,"Bool":true,"Int":-49,"msg":"hello"}
```

Field order follows **creation order** on the DynamicObject.

### Nesting

Lists and objects nest freely:

```forth
: makePayload
  DynamicObject.new^ >ts
  ( 1 2 3 ) t:.nums!
  DynamicObject.new^ >ys
  "world" y:.hello!
  ys> t:.nested!
  ts>
;

makePayload >jsonp $
```

Typical pretty result:

```json
{
  "nums": [
    1,
    2,
    3
  ],
  "nested": {
    "hello": "world"
  }
}
```
