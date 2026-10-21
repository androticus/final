# Iterators: `foreach`

Final supports the modern concept of Iterators. An iterator is an object produced from a source object, that knows how to iterate over values of the source. Many Final Types can provide an iterator, and their values can thus conveniently be iterated over in a `foreach` block.

To use them, you simply put the instance on DS prior to the `foreach`; `foreach` automatically retrieves the iterator, and starts the loop (if the iterator is not empty, eg an empty List). You access the item with the keyword `i`; in each case `i` is defined by the object, and is typically the natural value of the container. 

You can nest `foreach` loops. Each level accesses its items via `i`. You can access the items above your level via the `j` keyword (next level up) and `k` (second level up.) Nested `foreach` loops use `i`, then `j`, then `k` for the current item at each nesting level. 

The `i/j/k` accessors are only defined within the static lexical nesting of `foreach` loops. They are only defined within the word that has declared the loop; you cannot call another Final word, and use them in that context. You can however freely use them in subwords, provided they exist at the time the subword accesses them; i.e. you are calling the subword from within a `foreach` where `i/j/k` (as applicable) exist.

The following types support iteration.

## `range` and `rangeby`

As described in [Ranges](ranges.md).

```forth
> // ( end:i -- )
: countTo
  1 swap range
  foreach
    i $ sp
  /foreach
  nl
;
5 countTo

1 2 3 4 5
```

## Lists: `List`, `StringList`, all Numeric Lists, `StringKeyValueList`

* Most Lists: iterate over the values; `i` is the list item; the list is iterated in 0-based order
* StringKeyValueList: `i` is a StringKeyValue object (Key and Value are strings); this List is principally used for Web query parameters and headers; thus, unlike Maps, keys can be repeated.

```forth
> // ( -- )
: showList
  ( "alpha" "beta" "gamma" )
  foreach
    i $ sp
  /foreach
  nl
;
showList

alpha beta gamma
```

StringKeyValueList — print each pair (`.add^` takes value then key):

```forth
> // ( -- )
: showPairs
  StringKeyValueList.new^ >ts
  "1" "a" t:.add^
  "2" "b" t:.add^
  ts>
  foreach
    i >ys
    y:.key@ $ "->" $ y:.value@ $ sp
    y:drop
  /foreach
  nl
;
showPairs

a->1 b->2
```

## Map family: `Map`, `StringMap`, `DynamicObject`, `StaticObject`, `TypedObject`

* Map, StringMap: iterates over all entries in an indeterminate order; `i` is a KeyValue object
* DynamicObject, StaticObject, TypedObject: iterates deterministically in order of field definition; `i` is a KeyValue object

```forth
> // ( -- )
: showDO
  DynamicObject.new^ >ts
  10 t:.x!
  20 t:.y!
  ts>
  foreach
    i >ys
    y:.key@ $ "=" $ y:.value@ $ sp
    y:drop
  /foreach
  nl
;
showDO

x=10 y=20
```

Creation order is preserved for DynamicObject / TypedObject; Map / StringMap order is not defined — do not rely on it.

## Sets: `Set`, `StringSet`

`foreach` yields each **value** (not a KeyValue). Order is not defined.

```forth
( 1 2 3 set)
foreach
  i $ sp
/foreach
nl
```

## `VecNf` and `Mat2*`

`foreach` yields each value. A vector walks `0 … len-1`. A matrix walks **flat** row-major cells. For row/column coordinates see [Vectors and Matrices](../la/vec-mat.md).

Mats provide a convenient `getRowIter` and `getColIter` that are basically ranges set to the proper values for the row and column size respectively.

## Pipes

Pipes provide a special purpose iterator that enables you to block waiting for the next value to be received, get that value via `i`, then resume checking for the next value. When the pipe is closed by the producer, the `foreach` exits. see [Threading](../threading/index.md).
