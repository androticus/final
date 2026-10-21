# Lists

A List is Final's growable ordered sequence. Items are addressed by zero-based index. Most list types share one API; the specialized types mainly change **what may be stored**, and a few add **`sort`** and **`shuffle`**.

## Creating a list

```forth
List.new^                    // empty list
16 List.newWithCapacity^     // empty, room reserved for 16 items
5 List.newWithLen^           // length 5, slots filled with zeros for that type
```

`newWithCapacity` is useful when you already know about how many items you will append. `newWithLen` is useful when you will fill the default zero slots by index with `!()`.

## Common API

These members are available on **List**, **StringList**, and the numeric list types.

| Kind | Name | Effect | Description |
| ---- | ---- | ------ | ----------- |
| class | `new` | `( -- list )` | empty list |
| class | `newWithCapacity` | `( cap:i -- list )` | empty list with reserved capacity |
| class | `newWithLen` | `( len:i -- list )` | list of the given length, zero/empty filled |
| prop | `len` | `( -- len:i )` | number of items |
| method | `append` | `( item:a -- )` | add at the end |
| method | `insertAt` | `( item:a at:i -- )` | insert; `0` = front, `len` = append |
| method | `hasItem` | `( item:a -- b )` | true if any element equals the item |
| method | `deleteAt` | `( at:i -- )` | remove at index; later items shift down |
| method | `clear` | `( -- )` | remove all items |
| method | `enqueue` | `( item:a -- )` | insert at the beginning (queue) |
| method | `dequeue` | `( -- item:a )` | remove from the end (queue) |
| method | `push` | `( item:a -- )` | same as `append` (stack) |
| method | `pop` | `( -- item:a )` | remove from the end (stack) |
| method | `peek` | `( -- item:a )` | end item, leave it in place |
| method | `depth` | `( -- len:i )` | synonym for `len` |
| method | `sort` | `( -- )` | sort the list items (ascending) |
| method | `shuffle` | `( -- )` | randomly reorder items |

### Indexed access

```forth
cs:@()   // ( index:i -- value )   read
cs:!()   // ( value index:i -- )   write existing slot
```

The receiver stays on the context stack. Writing with `!()` requires a valid index in `0 .. len-1`; use `append` / `insertAt` to grow the list.

### Iteration

Put the list on DS, then `foreach` … `/foreach`. The loop value `i` is each element, in order from index `0` to `len-1`.

## Typical pattern

```forth
List.new^ >ts
10 t:.append^
20 t:.append^
30 t:.append^
0 t:@() $ nl        // 10
1 0 t:!()           // replace index 0 with 1
t:.len@ $ nl        // 3
t:drop
```

The following pages cover what makes each list family different.
