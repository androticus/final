# Reload Marks (`forget:`)

Interactive development often needs **load → try → wipe → reload** without relaunching the process. **`forget:`** defines an interpret-only **mark word**. Invoking that word erases the mark **and everything defined after it** in the current DictionaryContext (inclusive cut).

## Mark-first script pattern

Put the mark **alone on line 1** of a reloadable script — no leading whitespace, no comment, nothing else on the line:

```forth
forget: reload
```

Session/core words stay outside the file (before the mark). Scratch definitions live under the mark.

**`compileFromFile`** reads that first line. If the named mark already exists, it runs it (wipe), **then** starts the compile-unit checkpoint and loads the file — `forget: reload` recreates the mark at the top. Load the same file again without typing the mark name first.

If line 1 is not exactly `forget: name`, `compileFromFile` does not auto-run anything (comments or `requiresFyne` before the mark disable the shortcut).

```forth
forget: reload

: helpers …
TypedObject.newTypeDef^ …
… makeType …
: MyType#m … ;

// load / reload:
"reloadable.fnl" compileFromFile
```

You can still type the mark name at the prompt to wipe without reloading.

Examples on disk often use `forget: reset` the same way (`examples/file-system/*.fnl`, format examples).

## Behavior

| Topic | Rule |
| ----- | ---- |
| Cut | Inclusive — mark word is removed too |
| Checkpoint | At mark create, stamps dict / CT / DT / local type-catalog high-waters |
| `compileFromFile` | Line 1 exactly `forget: name` → run mark if present, then compile-unit checkpoint |
| Invocation | Mark word is **interpret-only** (cannot compile its name into other words) |
| Scope | Current (task/session) DC — not sealed core |

A failed reload still leaves the previous load wiped: the auto-run happens before the new unit’s checkpoint.

## Restriction: Final methods on surviving types

Before truncating, forget **refuses** (no mutation) if a type that **survives** the cut still has a Final method whose body was compiled **after** the mark (CT past the keep). That prevents dangling method nests.

**Safe:** create the type **and** its `: Type#… ;` / `: Type##… ;` methods **under** the mark (they roll back together).

**Unsafe:** register/`makeType` a type **before** the mark, then add `: Type#m ;` **after** the mark — invoking the mark refuses and names the type/method.

## Live MethodRefs after the cut

Forget drops types from the catalog; it does not free Go objects. A `MethodRef` (EventLoop handler, window `onClosed`, or a ref parked in a word defined *before* the mark) still points at the old type’s method metadata.

On the cut, Final methods on types leaving the catalog are **poisoned**: invoking that MethodRef panics (`MethodRef: type Name#m forgotten`) instead of nesting into truncated or reissued CT. Close windows / drop outstanding refs before reload; the poison is the backstop if you do not.

## What leave alone

Sealed / locked DCs, mid-compile, and core modules are out of scope. Foreign references held outside the DC are not tracked. Surviving defer targets that point into wiped dict/CT also refuse the cut.
