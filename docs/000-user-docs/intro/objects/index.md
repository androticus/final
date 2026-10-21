# Objects

The [Quick Tour](../quick/objects-types.md) introduced Types, members, and the basics of compound path access. This Foundations section goes deeper on the **APIs** and the **path syntax** you use every day.

Everything you reach through a path—Go-native types, DynamicObject / StaticObject, TypedObject, Words, collections—uses the same API:

| Kind | Role |
| ---- | ---- |
| **Type** | Catalog entry that describes members; CapCase dictionary word (`List`, `Int`, …) |
| **Instance** | Concrete value of that Type |
| **Property** | Named readable/writable (or const) member — `@` / `!` |
| **Method** | Named operation — invoke with `^` |
| **Collection accessors** | Indexed `@()` / `!()`, keyed `@{}` / `!{}` / `?{}`, coordinate `@[]` / `![]` (`Mat2f`) |

Parameters and results for member ops still use **DS**. The **receiver** (the object the path acts on) usually lives on an auxilliary stack (`t:` / `x:` / `y:` / `z:`) or on **OS** (`o:`) inside a Final method. See [Stacks](../stacks/index.md) and [Writing Clean Final Code](../clean-final-coding.md).

Concrete collection APIs (List, Map, DynamicObject, StaticObject, TypedObject, Table) are in [Collection Types](../compounds/index.md). DynamicObject also offers `.clone^` (shallow prototype copies) and `.newStatic^` (mint a [StaticObject](../compounds/staticobject.md) from a DO prototype). Use `help TypeName` in the TUI for the full member list of any type.

## In this section

* [Types](types.md) — Type objects, class vs instance members, factories
* [Properties](properties.md) — fetch, store, const, refs
* [Methods](methods.md) — instance and class methods, `^`, Final methods on OS
* [Collection Accessors](collection-accessors.md) — `@()` `!()` `@[]` `![]` `@{}` `!{}` `?{}`
* [Compound Path Syntax](compound-path-syntax.md) — path elements, stacking rules, nested walks, examples
