# OLS Fit, Compare, Predict

**OLS** (ordinary least squares) fits `y ≈ X β` for a dense outcome vector and design matrix. There is **no factory** on `OLSModel`. The class method **`fit`** materializes a model from `y` and `X`.

`X` is **n × p** (rows = observations). You supply the intercept yourself if you want one — typically a first column of `1.`. `y` must have length **n**. Fit requires **n ≥ p**.

A compact-vs-augmented comparison (same `y`, nested predictors) follows Judd–McClelland–Ryan: **PRE** (proportional reduction in error) and an **F** statistic. See [Auto-stats](../sample-apps/auto-stats.md) for a full ladder on real listings.

Type `help OLSModel` and `help ModelCompareResult` in the TUI.

---

## OLSModel

| Kind | Name | Effect | Description |
| ---- | ---- | ------ | ----------- |
| class | `fit` | `( y:VecNf X:Mat2f -- model:OLSModel )` | least-squares fit |
| class | `compare` | `( c:OLSModel a:OLSModel -- r:ModelCompareResult )` | compact vs augmented (same n; `a.p` &gt; `c.p`) |
| method | `predict` | `( Xnew:Mat2f -- yHat:VecNf )` | apply `β` to new rows; `Xnew.cols` must equal `p` |
| prop | `beta` | VecNf | coefficients, length `p` |
| prop | `yHat` | VecNf | fitted values, length `n` |
| prop | `sse` | Float | residual sum of squares |
| prop | `sst` | Float | total sum of squares about the mean of `y` |
| prop | `r2` | Float | `1 − sse/sst` (0 if `sst` is 0) |
| prop | `rmse` | Float | `√(sse / (n − p))`; NaN if `n − p ≤ 0` |
| prop | `n` | Int | observations |
| prop | `p` | Int | columns of `X` |
| prop | `df` | Int | `n − p` |

```forth
4 VecNf.new^ [ 1. 3. 4. 8. ]
4 2 Mat2f.new^ [
  1. 0.
  1. 1.
  1. 2.
  1. 3.
]
OLSModel.fit^ >ts
t:.r2@ $ nl
t:.beta@ >ys
0 y:@() $ sp  1 y:@() $ nl
y:drop
// score a new row (intercept, x)
1 2 Mat2f.new^ [ 1. 2. ]
t:.predict^ >ys
0 y:@() $ nl
y:drop
t:drop
```

Park the model on **ts** (or another aux stack) when you will read several properties or call `.predict^`.

---

## ModelCompareResult

Build two models on the **same** `y`: a **compact** model **C** and an **augmented** model **A** that adds predictors. Then:

```forth
mC mA OLSModel.compare^ >ts
t:.pre@ $ sp  t:.f@ $ nl
t:drop
```

PRE is `(SSE_C − SSE_A) / SSE_C` — the fraction of C’s leftover error that A removes. **F** asks whether that PRE is large relative to the extra parameters and residual df.

| Prop | Type | Description |
| ---- | ---- | ----------- |
| `sseC` / `sseA` | Float | residual SS of C and A |
| `pre` | Float | proportional reduction in error |
| `f` | Float | F statistic |
| `df1` | Int | extra parameters (`pA − pC`) |
| `df2` | Int | residual df of A (`n − pA`) |
| `n` | Int | shared observation count |
| `pC` / `pA` | Int | column counts |

Compare **panics** if the models are not nested in this sense (different `n`, or `pA` not greater than `pC`). That is a modeling mistake, not a `throw`.

---

## Notes

* Put **Float** in every design-matrix cell. Int literals in `[` … `]` are not coerced.
* `fit` / `predict` / `compare` panic on nil, shape mismatch, or `n < p` — fix the matrices; do not wrap these in `try` for “maybe the data is the wrong width.”
* For extracting a SQL result into a List or Table first, see [`iter>list` and `dbIter>table`](../database.md).
