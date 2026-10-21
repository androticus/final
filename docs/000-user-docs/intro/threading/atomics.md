# Atomics

## Shared state (optional; prefer pipes)

The preferred method of ITC is via `pipes` (“share by communicating”).

However, thread-safe variables and collections can have their use, so Final provides a curated collection of thread-safe constructs.

## `AtomicInt`

A single thread-safe integer. Readonly `.value`; mutating methods return the **new** value except `.set^` / `.zero^`.

| Kind | Name | Effect | Description |
| ---- | ---- | ------ | ----------- |
| class | `new` | `( -- a:AtomicInt )` | starts at 0 |
| prop | `value` | Int | current value (const) |
| method | `inc` | `( -- n:i )` | add 1; return new value |
| method | `dec` | `( -- n:i )` | subtract 1; return new value |
| method | `add` | `( n:i -- n':i )` | add `n`; return new value |
| method | `sub` | `( n:i -- n':i )` | subtract `n`; return new value |
| method | `set` | `( n:i -- )` | store `n` |
| method | `zero` | `( -- )` | store 0 |

```forth
AtomicInt.new^ >ts
t:.inc^ $ nl          // 1
10 t:.add^ $ nl        // 11
t:.value@ $ nl        // 11
t:drop
```

Prefer a Pipe when you are passing work between Pips. Use `AtomicInt` when several Pips only need to bump a counter.

## `AtomicList` / `AtomicStringList`

These are mutex-guarded Lists for reading/writing an item.

Iteration via `foreach` uses a snapshot clone, so is threadsafe. NOTE: do NOT iterate AtomicLists with a range-based iterator or other counted loop; iteration is not threadsafe using that approach.

## `AtomicDynamicObject` 

A mutex-guarded DynamicObject (string keys) that is threadsafe for reading a key, writing a key value, and adding a new field. `foreach` uses a snapshot clone, similar to the AtomicLists.

## `Average` and `MovingAverage`

Average is a type used for accumulating an exactly correct mean and variance for data samples added via the `add` method, without retaining the samples. `MovingAverage` is used for calculating a recent, moving average. Both are threadsafe for `add`, and the atomic read methods.

See `help Average` and `help MovingAverage` for the apis.

See [Counter / Average / MovingAverage](../misc/count-etal.md)

Plain `var:` is **not** threadsafe for concurrent writers. Prefer pipes / atomic objects for shared mutation. Concurrent **reads** of a `var:` published once before Pips start are fine. `const:` is safe to read after definition.

## General Caveats

In all the collections above, threadsafety extends only to the operations listed. In particular, the threadsafety of any objects stored in these containers have the threadsafety of the object itself, which for most Objects is NOT threadsafe.

### Threadsafe Operations

* reading any class-level sealed readonly property; eg Math.pi
* reading properties from object instances that are not mutated over the lifetime of the reading Pips
* creating/mutating objects created by a Pip, and not shared
* reading/mutating an object created by a Pip, then sent via a Pipe to another Pip, under contract that the sending Pip then no longer mutates or uses the object; eg. preparing a DynamicObject by Pip1, then sending it to Pip2, abandoning its ownership
* calling object factories, such as opening new file streams
* writing to the shared stdout/stderr, eg using $/$$
