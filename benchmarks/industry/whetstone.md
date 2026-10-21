# Whetstone — outline

**Suite name:** `whetstone`  
**Category:** `industry`  
**First cut:** single-threaded → `whetstone.fnl`  
**Later:** threaded → `whetstone-thread.fnl` (see naming in [136-benchmarks.md](../../../../docs/003-design-notes/136-benchmarks.md))

## Background / purpose

Classic **float + transcendental** synthetic benchmark (H.J. Curnow / B.A. Wichmann; netlib C “Painter” double-precision port). A weighted sequence of small modules (simple arith, array updates, integer ops, `sin`/`cos`/`atan`, `exp`/`log`/`sqrt`, procedure-style helpers) meant to resemble “typical” scientific FORTRAN of its era.

**Why here (after Sieve + n-body):** orthogonal shape — not prime marking, not O(n²) gravity/objects. Exercises Final float dispatch and Math transcendentals as an **industry** mix (not a coerce microbench; see [136-benchmarks.md](../../../../docs/003-design-notes/136-benchmarks.md)).

## Fixed parameters

- `LOOP` — major scale factor (netlib default **1000**). Module trip counts are multiples of `LOOP` (below).
- One major pass (`II = 1`) for v1; wall time for that pass → `avg_s` / optional MWIPS note.
- Constants: `T = 0.499975`, `T1 = 0.50025`, `T2 = 2.0`.
- Array `E1[0..4]` — **index 0 unused** (FORTRAN 1-based look); live cells `E1[1..4]`.

## Module weights (netlib)

```
N1  = 0·LOOP      // simple idents (empty)
N2  = 12·LOOP     // array elements
N3  = 14·LOOP     // array via PA (6 updates each call)
N4  = 345·LOOP    // conditional jumps
N6  = 210·LOOP    // integer arithmetic + E1 stores
N7  = 32·LOOP     // trig
N8  = 899·LOOP    // P3 procedure calls
N9  = 616·LOOP    // P0 array rotate
N10 = 0·LOOP      // integer arith (empty)
N11 = 93·LOOP     // std functions √(exp(log(x)/T1))
```

(N5 / N12 omitted historically.)

## Algorithm (sketch)

```
T,T1,T2 := 0.499975, 0.50025, 2.0
E1[1..4] := …; J,K,L,X,Y,Z,X1..X4 as needed

// M1 (N1×): X1..X4 simple linear combo · T     // often skipped (N1=0)
// M2 (N2×): same combo on E1[1..4]
// M3 (N3×): PA(E1) — six in-place E1 updates (·T / ·T2)
// M4 (N4×): nested ifs flipping J through {0,1,2,3}
// M6 (N6×): integer J,K,L update; E1[L-1], E1[K-1] stores
// M7 (N7×): X,Y := T·atan( T2·sin·cos / (cos(x+y)+cos(x-y)-1) )
// M8 (N8×): Z := P3(X,Y) with X1=T·(X+Y); Y1=T·(X1+Y); Z=(X1+Y1)/T2
// M9 (N9×): P0 — E1[J]:=E1[K]; E1[K]:=E1[L]; E1[L]:=E1[J]
// M11 (N11×): X := sqrt( exp( log(X) / T1 ) )   // X starts 0.75
```

Published score (informational): `MWIPS ≈ (100·LOOP·II) / time_s` with `II=1`.

## Final sketch notes

- `FloatList` length 5 for `E1`; helpers `pa_`, `p0_`, `p3_`.
- Hot path: float `+ - * /`, `sin` `cos` `atan` `exp` `log` `sqrt`; no I/O inside modules.
- Print final `X` (M11) and timing; optional MWIPS line.
- No Pips; no DynamicObject (contrast n-body).

## Sanity checks

- `LOOP = 1` (or small): finishes; `X` after M11 is finite and stable across reruns.
- Spot-check M2/M7 against a native netlib binary on the same host when comparing industry numbers.

## Out of scope (v1)

- Exact KIPS/MWIPS table matching without a paired native run.
- Continuous `-c` mode; multi-`II` wrapping.
- Single-precision variant.
