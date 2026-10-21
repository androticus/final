# StringKeyValueList

**StringKeyValueList** is a specialized list of string key / string value pairs. Unlike **Map** / **StringMap**, the **same key may appear more than once**, and insertion order is kept. That matches HTTP headers and query parameters, where repeated keys are legal.

It does **not** use the ordinary List `append` / `@()` API. Use the members below instead.

## Creation

```forth
StringKeyValueList.new^
16 StringKeyValueList.newWithCapacity^
```

## API

| Kind | Name | Effect | Description |
| ---- | ---- | ------ | ----------- |
| class | `new` | `( -- list )` | empty list |
| class | `newWithCapacity` | `( cap:i -- list )` | empty, reserved capacity |
| class | `newWithLen` | `( len:i -- list )` | fixed length |
| method | `add` | `( value:s key:s -- )` | **always append** a new pair (duplicates allowed) |

### Keyed accessors

| Op | Effect | Description |
| -- | ------ | ----------- |
| `@{}` | `( key:s -- value:s )` | first match; missing key is an error |
| `!{}` | `( value:s key:s -- )` | upsert the **first** matching key (or append if new) |
| `?{}` | `( key:s -- b )` | true if any entry has that key |

Use **`add`** when you intentionally want a second header with the same name. Use **`!{}`** when you want to replace the first occurrence.

## Example

```forth
StringKeyValueList.new^ >ts
"text/html" "Content-Type" t:.add^
"1" "X-Count" t:.add^
"2" "X-Count" t:.add^          // second X-Count kept
"application/json" "Content-Type" t:!{}   // replaces first Content-Type

"Content-Type" t:@{} $ nl      // application/json

t:peek foreach
  i >xs
  x:.key@ $ ": " $ x:.value@ $ nl
  x:drop
/foreach
t:drop
```

In `foreach`, `i` is a **StringKeyValue** (same shape as **KeyValue**): read `.key@` and `.value@`.
