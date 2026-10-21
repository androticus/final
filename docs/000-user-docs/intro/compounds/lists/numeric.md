# Numeric Lists

Numeric lists use the [common List API](index.md), but store a single numeric width in each cell. Values are widened to **Int**, **Uint**, or **Float** on the Data Stack when you read them, and narrowed again when you write.

They exist so large sequences of numbers do not pay the cost of a fully general **List** cell. A **ByteList**, for example, is far more compact than a **List** of small integers, and is the natural home for raw byte buffers used with strings and byte streams.

## Types

| Type | Stores | On DS |
| ---- | ------ | ----- |
| `Int8List` | signed 8-bit | `Int` |
| `Int16List` | signed 16-bit | `Int` |
| `Int32List` | signed 32-bit | `Int` |
| `IntList` | Final `Int` | `Int` |
| `ByteList` | unsigned 8-bit (0…255) | `Int` |
| `Uint16List` | unsigned 16-bit | `Int` |
| `Uint32List` | unsigned 32-bit | `Int` |
| `UintList` | Final `Uint` | `Uint` |
| `Float32List` | 32-bit float | `Float` |
| `FloatList` | Final `Float64` | `Float` |

**ByteList** is the unsigned 8-bit list — not a signed “char” width. Use **Int8List** when you need signed 8-bit storage. Unicode code points fit in **Int32List** when you want a dense list of them; day-to-day text work usually streams runes instead.

## Example

```forth
5 FloatList.newWithLen^ >ts
3.14 0 t:!()
2.71 1 t:!()
1.41 2 t:!()
0 t:@() $ nl     // 3.14
t:.sort^
t:drop
```

```forth
ByteList.new^ >ts
65 t:.append^    // 'A'
66 t:.append^
t:.len@ $ nl
t:drop
```

## `sort`

Every numeric list type supports:

```forth
// ( -- )
.sort^
```

Sorting uses the natural order for that numeric type.

**`VecNf` / `Mat2f`** are fixed-shape dense Float arrays for linear algebra, not Lists — they do not grow. See [Linear Algebra](../../la/index.md).
