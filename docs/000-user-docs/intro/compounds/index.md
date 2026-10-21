# Compound Types

Final's compound types hold related values together: ordered sequences (**Lists**), keyed tables (**Maps**), **Sets**, named field objects (**DynamicObject** / **StaticObject** / **TypedObject**), **Vecs** and **Mats**, and columnar grids (**Table**).

The [Quick Tour](../quick/compound-types.md) already introduced what each family is for. This section focuses on the **APIs** you use day to day, with short examples. For the path syntax behind `.append^`, `@()`, `!{}`, and friends, see [Objects](../objects/index.md).

| Family | Best for | Access |
| ------ | -------- | ------ |
| Lists | ordered sequences, stacks/queues | `@()` / `!()`, methods |
| Maps | dictionaries by key | `@{}` / `!{}` / `?{}` |
| Sets | list/map hybrid, of unique keys | `?{}` |
| DynamicObject | small ad-hoc records, Json objects | `.field` and `@{}` / `@()`; `{ … }` builders; `.clone^` |
| StaticObject | fixed-schema rows from a DO prototype | same API as DO for schema fields |
| TypedObject | reusable record types with a schema (and optional Final methods) | `.field` and `@{}` / `@()` |
| Vecs | 1-dimensional containers of fixed size set at construction |  index `@[]` / `!()` |
| Mats | 2-dimensional tabular containers of fixed size set at construction | flatIndex `@()` / `!()`; row col `@[]` / `![]` |
| Table | columnar grids with stable multi-key sort | `append` / `sortBy*` / `foreach` |

Use `help TypeName` in the TUI for the full member list of any type.

Note: `Mat2f` has additional uses and dedicated mat:: operations; it is documented along with `VecNf` vectors and OLS under [Linear Algebra](../la/index.md).
