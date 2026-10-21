# Sieve (Eratosthenes) — outline

**Suite name:** `sieve`  
**Category:** `industry`  
**First cut:** single-threaded → `sieve.fnl`  
**Later:** threaded → `sieve-thread.fnl` (see naming in [136-benchmarks.md](../../../../docs/003-design-notes/136-benchmarks.md))

## Goal

Classic prime sieve: integer loops + array/list marking. Simple Final port; memory-ish VM signal. Not a bit-packed micro-opt contest on v1.

## Fixed parameters (pick once, document in `.fnl`)

- `N` — upper bound (inclusive or exclusive; state which). Common: 8190 / 10000 / 1e6 — choose one that runs long enough for stable `avg_s`.
- Outer **repeat count** — harness `iterations`, or an inner loop in the body if the harness only times one `Send`. Prefer harness-driven repeats when the driver exists.

## Algorithm (single)

```
flags[0..N] := true          // or false = composite; pick one convention
flags[0] := false; flags[1] := false   // if N includes 0/1

for p := 2 .. floor(sqrt(N)):
  if flags[p]:
    for k := p*p .. N step p:
      flags[k] := false

count := number of true flags in 2..N   // optional checksum / printed result
```

## Final sketch notes

- Prefer a `List` (or typed integral list if handy) of bool/int flags over ad-hoc compounds.
- Inner clear loop: `forever` + index math, or `foreach` over a range if that stays cheap.
- Keep the hot path free of I/O; print `count` once after measured work if useful for sanity.
- No Pips in `sieve.fnl`.

## Measured result

- Wall time for one full sieve (or harness N repeats → `avg_s`).
- Optional: report `count` as a side field later; not required for v1 schema.
- Run under both `compile_context` = `core` and `local` once the harness exists.

## Sanity checks

- Small `N` (e.g. 30): primes match a known list.
- Known `count` for chosen `N` (document expected count in a comment).

## Out of scope (v1)

- Bit packing, wheel factorization, segmented sieve.
- Matching a specific published score formula until we care about external comparison tables.
