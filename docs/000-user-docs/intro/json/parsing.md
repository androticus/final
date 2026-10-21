# Reading/Parsing

Final can parse any legal Json. The table below shows the Final Types used for each Json type.

| Json type | Final type | Note |
| --------- | ---------- | ---- |
| null | Null | |
| bool | Bool | |
| number | Int or Float | 1 |
| string | String | 2 |
| array | List | 3 |
| object | DynamicObject | 4 |

1. **Int** when the value has no decimal and fits in **Int**; otherwise **Float**.
2. All json escapes are processed, including hex Unicode
3. The value of each item will follow the Json->Final mapping; full nesting of arrays and objects is supported.
4. The value of each object value will follow the Json->Final mapping; full nesting of arrays and objects is supported.

The everyday parse word is:

```
// ( json:s -- root:a )
json>
```

Invalid Json **`throw`s** a string message rather than panicking. Wrap `json>` in `try` / `catch` when the text is untrusted (an HTTP body, a file, a tool argument). If nothing catches, an uncaught `throw` is handled like a panic. See [Error and Exception Handling](../error-handling.md).

There is also a lower-level `parseJson` with the same stack effect; prefer `json>` in application code.

## Examples

### Primitives

```forth
"null" json> $ nl        // null
"true" json> $ nl        // true
"-49" json> $ nl         // -49
"3.14" json> $ nl        // 3.14
`"hello"` json> $ nl     // hello
```

### Array → List

```forth
`[1,2,3]` json> >ts
0 t:@() $ sp             // 1
1 t:@() $ sp             // 2
2 t:@() $ nl             // 3
t:drop
```

Or walk with `foreach`:

```forth
`["a","b","c"]` json>
foreach
  i $ sp
/foreach
nl
// a b c
```

### Object → DynamicObject

```forth
`{"x":1,"y":2,"msg":"hi"}` json> >ts
t:.x@ $ sp               // 1
t:.y@ $ sp               // 2
t:.msg@ $ nl             // hi
t:drop
```

String-key access works the same as any DynamicObject:

```forth
`{"host":"wttr.in"}` json> >ts
"host" t:@{} $ nl        // wttr.in
t:drop
```

### Nesting and round-trip

```forth
: sampleJson
  `{"nums":[1,2,3],"nested":{"hello":"world"}}`
;

sampleJson json> >ts
t:.nums@ foreach i $ sp /foreach nl    // 1 2 3
t:.nested@ >ys
y:.hello@ $ nl                          // world
y:drop
ts> >jsonp $                            // pretty round-trip
```

A compact round-trip is just `json> >json`. Values that were numbers without a decimal come back as **Int** when they fit; otherwise **Float**.

## Json Parsing Limits

The following limits are checked when parsing Json, to protect against malicious or unreasonable json. If a limit is reached during parsing, `json>` **throws** a string stating the limit type and the limit value.

| Limit | Value | Note |
| ----- | ----- | ---- |
| Input length | 8,388,608 | Maximum size of the JSON text (about 8 MB). |
| Nesting depth | 64 | How deeply arrays and objects may nest inside each other. |
| Node count | 1,000,000 | Maximum number of JSON values — objects, arrays, strings, numbers, true, false, and null. Field names are not counted. An empty object or array counts as one value. |
| Identifier length | 2,048 | Maximum length of an object field name. |
| String length | 1,048,576 | Maximum length of a string value (about 1 MB), after JSON escapes are decoded. |
| Number length | 64 | Maximum length of a number as written in the JSON (sign, digits, decimal point, and exponent). |

