# Word Doc Guidelines

The [Quick Tour](quick/control-flow.md) introduced **context effects**: a short comment that describes what a word expects on the stacks and what it leaves. This page is the full convention for writing those docs — the format the compiler stores, how `help` shows them, and how to document **Type members** (including your own [TypedObject](compounds/typedobject.md) types).

Everyday API pages already point you at `help` for the same text (`help +`, `help Regex#find`, `help DateTime`, …). Good docs make that useful.

---

## Recommended comment form

Prefer a block comment with **`/*` and `*/` each on their own line**, and the body **flush left** (no indent inside the block):

```forth
/*
Add two numbers, and return the sum.
( a:n b:n -- c:n )
*/
: + ...
```

That is the form used throughout Final’s own core docs (for example `+` in the basic-ops doc file). Single-line `// …` comments also work for a one-liner, but the block form is clearer once you have more than a sentence.

Place the comment **immediately before** the word it documents (the colon definition, or a `:doc` line — see below). The compiler takes that **last comment** and stores it on the dictionary entry’s **`.doc`** property. Typing `help wordname` in the UI prints that text.

---

## What to write

Keep three layers, in order:

1. **Opening sentence** — one complete sentence that states the purpose of the word. Lead with the verb or outcome (“Add…”, “True if…”, “Create a new empty…”). Don't repeat the word name.
2. **Additional detail** — more complete sentences only when needed (panics, modal results, related words, nesting rules). Prefer short lines; skip anything the effect line already says.
3. **Last line** — the **context / stack effect**. Put nothing after it inside the comment.

Be succinct. Help text is read in the UI; long essays belong in these intro pages, not in `.doc`.

### Good pattern

```forth
/*
Divide a by b and return the result.
Panics on divide by zero for integral types.
Returns Float.Inf for Float divide by zero.
( a:n b:n -- c:n )
*/
```

### Modal results

When a word uses the success/failure shapes from [Writing Clean Final Code](clean-final-coding.md), spell both arms on the effect line:

```forth
/*
First match substring in s, or false if none. Modal.
A successful empty match is "" true (not false).
( s -- found:s true | false )
*/
```

Follow the same pattern if the word accepts a modal result.

### Auxilliary stacks and the Object Stack

Bare `( … -- … )` is the **data stack**. When the word also uses or changes **ts / xs / ys / zs** , write those aspects with a letter prefix. Since the OS cannot be manipulated by user code, don't include before/after, only the one value that persists for the method:

```forth
/*
Render newline-separated ops onto the parked StringBuilder.
( text:s -- ) z( sb:StringBuilder -- sb )
*/

/*
Final method body: receiver is on OS for the duration of the call.
( -- s ) o( inst:Person )
*/
```

Only document aspects that change; omit stacks that are untouched, unless you want to emphasize they have a value that must flow through to code called in the word. In that case, you can use the flowthru convention.
```
/*
Helper word for parsing.
( -- ) z( sb:StringBuilder )
*/
```

---

## Stack-effect notation

Inside `( … -- … )`:

* Left of `--` — inputs (deepest first; **TOS is rightmost**).
* Right of `--` — outputs after the word runs.
* Optional `name:type` labels (e.g. `a:n`, `list:List`).

### Common type codes

| Code | Meaning |
| ---- | ------- |
| `b` or `t` | Bool / Truthy (most Bool contexts take Truthy; Json needs Bool) |
| `i` | Int |
| `u` | Uint |
| `iu` | Integral (Int or Uint) |
| `f` | Float |
| `n` | Numeric (Int, Uint, or Float) |
| `s` | String |
| `a` | Any |
| `DO` | DynamicObject (common shortcut) |

For other types, write the full type name (`List`, `Regex`, `HttpReq`, …).

---

## Attaching docs to your TypedObject with `:doc`

Final provides the `:doc` helper to attach documentation to an existing word or Type. You cannot attach such docs to sealed core or module words, but you can document your own TypedObjects this way.

Syntax:

```forth
/*
…doc text…
*/
:doc wordName // e.g. your new TypeDef

/*
…doc text…
*/
:doc TypeName#memberName // for your instance and class properties, if you want to document them
```

`:doc` reads the next identifier and attaches the **last comment** to that target’s `.doc`. For members, use a single `#` between the type name and the member name — the same spelling you use in `help`:

Then, to refresh your memory on how your type works, you just need to type:
```forth
help MyType
```

### Documenting type members

Core Final documents members this way (excerpt from the maps/objects doc file):

```forth
/*
True if the Map contains an entry for the key.
( key:a -- has:b )
*/
:doc Map#hasKey

/*
Create a new empty DynamicObject.
( -- o:DynamicObject )
*/
:doc DynamicObject#new
```

### Your TypedObject types

If you define a TypedObject (see [TypedObject](compounds/typedobject.md)), document its fields and Final methods the same way. After the type exists in the catalog:

```forth
/*
Person record: display name and age in years.
*/
:doc Person

/*
Display name.
*/
:doc Person#name

/*
Age in years.
*/
:doc Person#age

/*
One-line "Name (age)" description.
( -- s ) o( person -- person )
*/
:doc Person#describe
```

Prefer documenting members when you register them, so `help Person` and `help Person#describe` stay useful for anyone using your type — including yourself later.

You may still put a `/* … */` block immediately above `: Person#describe` so the dictionary **Word** gets `.doc` automatically; use `:doc Person#describe` when you want the text on the **type member** that `help Type#member` and the type’s member listing show.

---

## Quick checklist

* `/*` and `*/` on their own lines; body flush left.
* First line: one complete purpose sentence.
* Extra sentences only for what the effect line cannot say.
* Last line: context/stack effect; nothing after it in the comment.
* Use `:doc name` or `:doc Type#member` when the target is not the word you just defined.
* Document TypedObject fields and methods so `help` works for your types.
