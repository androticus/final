# `vec::` and `mat::` Operators

Linear-algebra operators live in two families, like `str::` and `dt::`. **Last argument is TOS.** They return **new** vectors or matrices; inputs are not mutated.

There is no element-wise (Hadamard) multiply in this wave. `vec::dot` is the scalar inner product. `mat::*` is matrix product, not cell-wise `*`.

Type `help vec::+` or `help mat::solve` in the TUI for stack effects.

---

## `vec::` (vectors only)

| Word | Effect | Description |
| ---- | ------ | ----------- |
| `vec::+` | `( u:VecNf v:VecNf -- w:VecNf )` | element-wise sum; same length |
| `vec::-` | `( u:VecNf v:VecNf -- w:VecNf )` | element-wise `u − v`; same length |
| `vec::scale` | `( v:VecNf s:f -- w:VecNf )` | multiply every cell by `s` |
| `vec::dot` | `( u:VecNf v:VecNf -- f )` | inner product → Float; same length |
| `vec::norm` | `( v:VecNf -- f )` | Euclidean (L2) norm |
| `vec::copy` | `( v:VecNf -- w:VecNf )` | deep copy |

```forth
3 VecNf.new^ [ 1. 2. 3. ]
3 VecNf.new^ [ 4. 5. 6. ]
vec::+ >ts
0 t:@() $ nl          // 5
t:drop

3 VecNf.new^ [ 1. 2. 3. ]
3 VecNf.new^ [ 4. 5. 6. ]
vec::dot $ nl         // 32
```

---

## `mat::` (at least one matrix)

| Word | Effect | Description |
| ---- | ------ | ----------- |
| `mat::+` | `( A:Mat2f B:Mat2f -- C:Mat2f )` | element-wise sum; same shape |
| `mat::-` | `( A:Mat2f B:Mat2f -- C:Mat2f )` | element-wise `A − B`; same shape |
| `mat::scale` | `( A:Mat2f s:f -- C:Mat2f )` | multiply every cell by `s` |
| `mat::*` | `( A:Mat2f B:Mat2f -- C:Mat2f )` | matrix product; `A.cols` must equal `B.rows` |
| `mat::matvec` | `( A:Mat2f v:VecNf -- w:VecNf )` | `w = A v`; `v.len` must equal `A.cols` |
| `mat::transpose` | `( A:Mat2f -- At:Mat2f )` | shape becomes `cols × rows` |
| `mat::solve` | `( A:Mat2f b:VecNf -- x:VecNf )` | solve `A x = b` (square or least-squares); `b.len` = `A.rows`; result length `A.cols` |
| `mat::row` | `( A:Mat2f r:i -- v:VecNf )` | copy of row `r` (length = cols) |
| `mat::col` | `( A:Mat2f c:i -- v:VecNf )` | copy of column `c` (length = rows) |
| `mat::identity` | `( n:i -- I:Mat2f )` | `n × n` identity; `n` &gt; 0 |
| `mat::copy` | `( A:Mat2f -- B:Mat2f )` | deep copy |

```forth
2 2 Mat2f.new^ [ 1. 2.  3. 4. ]
2 VecNf.new^ [ 1. 0. ]
mat::matvec >ts
0 t:@() $ sp  1 t:@() $ nl     // 1  3
t:drop

2 mat::identity
2 2 Mat2f.new^ [ 1. 2.  3. 4. ]
mat::*                         // same as the right-hand matrix
```

Mismatched lengths or shapes **panic** (programmer error: the sizes were known when you built the values).

For fitting `y ≈ X β` and comparing nested models, use [OLS](ols.md) rather than assembling `mat::solve` by hand — OLS also reports SSE, R², RMSE, and compact-vs-augmented PRE/F.
