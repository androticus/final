# Linear Algebra: Vectors, Matrices, Stats, Model Fitting

Final includes a small **dense float64** linear-algebra API for data you already hold as numbers: a vector type, a rank-2 matrix type, stack operators on those values, and ordinary least-squares (OLS) fit / compare / predict.

This is **not** a growable List. **`VecNf`** and **`Mat2f`** have a length (and matrix shape) fixed at construction. Use **[FloatList](../compounds/lists/numeric.md)** when you need append, insert, or a general numeric sequence; convert at the boundary when you are ready to compute.

| Page | Contents |
| ---- | -------- |
| [Vectors and Matrices](vec-mat.md) | `VecNf` / `Mat2f`: create, fill, index, iterate |
| [`vec::` and `mat::` Operators](operators.md) | First-wave arithmetic, product, solve, row/col, copy |
| [OLS Fit, Compare, Predict](ols.md) | `OLSModel` / `ModelCompareResult` — no factory; materialized by `fit` |

Type `help VecNf`, `help Mat2f`, `help vec::dot`, `help mat::*`, or `help OLSModel` in the TUI for the same stack effects inline.

A worked data story that uses these types is the [Auto-stats](../sample-apps/auto-stats.md) sample (used-car asking price, PRE/F model comparison).

### Pattern

```forth
3 VecNf.new^ [ 1. 2. 3. ]
2 2 Mat2f.new^ [ 1. 0.  0. 1. ]
mat::matvec          // ( A:Mat2f v:VecNf -- w:VecNf )
```
