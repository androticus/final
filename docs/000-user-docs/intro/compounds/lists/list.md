# List and StringList

Both types share the [common List API](index.md). They differ only in what they store, and **StringList** adds sorting.

## List

**List** is the general-purpose sequence. Each cell holds **Any**: numbers, strings, Bool, Null, or other objects (including nested lists and maps). Items do do not need to be the same type.

```forth
List.new^ >ts
"hello" t:.append^
42 t:.append^
true t:.append^
null t:.append^
t:.len@ $ nl    // 4
t:drop
```

Use **List** when the element type varies, or when you are building a temporary list of mixed values (for example before rendering Json).

## StringList

**StringList** stores only **String** values. That keeps storage compact for the common “list of text” case, and rejects non-string appends.

```forth
StringList.new^ >ts
"alpha" t:.append^
"Beta" t:.append^
"gamma" t:.append^
t:.sort^
t:peek foreach
  i $ sp
/foreach
nl
t:drop
```

### `sort`

```forth
// ( -- )
.sort^
```

**StringList#sort** orders by Utf-8 byte values. In plain ASCII that puts uppercase before lowercase (`"Beta"` before `"alpha"`). It is **not** locale-aware alphabetic sorting.

**List** (Any cells) does not provide `sort`.
