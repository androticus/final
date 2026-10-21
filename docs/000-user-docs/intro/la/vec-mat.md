# Vectors and Matrices

**`VecNf`** is a dense vector of **Float** (IEEE-754 float64) of length **N**, chosen when you call `.new^`. **`Mat2f`** is a dense rank-2 matrix of the same cell type, with shape **rows × cols** fixed at construction. Both start **zero-filled**.

They are **not** Lists: there is no `append`. Index stores require a valid slot and a **Float** — Final does not coerce Int into these cells.

Filling from a stack of values uses the `[` … `]` filler (the closer does not create the instance). See [Collection Builders](../compounds/collection-builders.md) for nesting rules.

Related: [vec:: and mat:: Operators](operators.md), [OLS](ols.md).

---

## VecNf

| Kind | Name | Effect | Description |
| ---- | ---- | ------ | ----------- |
| class | `new` | `( n:i -- v:VecNf )` | zero-filled vector; `n` must be &gt; 0 |
| prop | `len` | `( -- len:i )` | number of cells |

Indexed access is the List-style ordinal pair:

```forth
cs:@()   // ( index:i -- f:Float )
cs:!()   // ( val:Float index:i -- )
```

```forth
3 VecNf.new^ >ts
[ 1. 2. 3. ]
0 t:@() $ nl          // 1
9. 1 t:!()            // replace index 1
t:drop
```

`foreach` walks cells in order `0 … len-1`. The loop value `i` is each **Float**.

---

## Mat2f

| Kind | Name | Effect | Description |
| ---- | ---- | ------ | ----------- |
| class | `new` | `( rows:i cols:i -- m:Mat2f )` | zero-filled; both dimensions &gt; 0 |
| prop | `rows` | `( -- rows:i )` | row count |
| prop | `cols` | `( -- cols:i )` | column count |
| prop | `len` | `( -- len:i )` | `rows * cols` |
| method | `getRowIter` | `( -- rows:Range )` | row indexes `0 … rows−1` |
| method | `getColIter` | `( -- cols:Range )` | column indexes `0 … cols−1` |
| method | `wrap` | `( row:i col:i -- row:i col:i )` | torus wrap of coordinates |


Two index styles:

| Affix | Stack | Meaning |
| ----- | ----- | ------- |
| `@()` / `!()` | `( index:i -- f )` / `( f index:i -- )` | **flat** row-major index `0 … len-1` |
| `@[]` / `![]` | `( r:i c:i -- f )` / `( f r:i c:i -- )` | **coordinates** `(row, col)` |

Teaching line: `@()` = flat; `@[]` = coordinates.

```forth
2 3 Mat2f.new^ >ts
[
  0. 1. 2.
  3. 4. 5.
]
1 2 t:@[] $ nl        // 5  (row 1, col 2)
9. 0 0 t:![]          // put 9 at (0,0)
t:drop
```

`foreach` on a Mat2f walks **flat** cells in row-major order; `i` is each **Float**. Nested `range` loops plus `@[]` walk by coordinates (`i` is the inner index, `j` the outer):

```forth
: showMat
  2 2 Mat2f.new^ [ 1. 2.  3. 4. ] >ts
  0 t:.rows@ 1 - range
  foreach
    0 t:.cols@ 1 - range
    foreach
      j i t:@[] $ sp
    /foreach
    nl
  /foreach
  t:drop
;
```

`.getRowIter^` / `.getColIter^` return a `Range` over indexes if you do not want to write `0 .rows@ 1- range` by hand. Pull a whole row or column as a new `VecNf` with `mat::row` / `mat::col` — [operators](operators.md).

---

## Notes

* Every `[` … `]` fill must supply **exactly** `.len` values, each already a Float.
* Prefer `VecNf` / `Mat2f` when every cell is a Float and the shape is known (design matrix **X**, outcome **y**, coefficients). Prefer numeric Lists when the sequence still needs to grow.
* `help VecNf` / `help Mat2f` lists members and accessors.
