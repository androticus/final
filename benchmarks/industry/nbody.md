# n-body (CLBG) — outline

**Suite name:** `nbody`  
**Category:** `industry`  
**First cut:** single-threaded → `nbody.fnl`  
**Later:** threaded → `nbody-thread.fnl` (see naming in [136-benchmarks.md](../../../../docs/003-design-notes/136-benchmarks.md))

## Goal

Computer Language Benchmarks Game n-body: fixed solar-system bodies, pairwise gravity, float-heavy nested loops. Next-simplest industry port after Sieve — same “arrays + loops” shape, but `Float` arith instead of integer flag marking. Not a contest for SIMD / fused-mul tricks on v1.

## Fixed parameters (pick once, document in `.fnl`)

- `steps` — how many times the simulation **advances** (one advance = all pairwise force updates, then all position updates). CLBG commonly uses **1000** (short / verify) or **50_000_000** (long / published timing); pick one that yields stable `avg_s` on target hosts. This is an integer count, not a fraction.
- Body set — classic 5: sun, jupiter, saturn, uranus, neptune (see [Suggested starting values](#suggested-starting-values-clbg)).
- Harness **repeat count** (`iterations`) — how many times to run the whole timed body (setup + `steps` advances + energy). Separate from `steps`: repeats are for averaging wall time; `steps` is the work inside one run. Prefer harness-driven repeats when the driver exists.

## Suggested starting values (CLBG)

Constants (compute once):

```
PI            := 3.141592653589793
SOLAR_MASS    := 4 * PI * PI
DAYS_PER_YEAR := 365.24
dt            := 0.01
```

Raw table (before scaling). Apply: `vx,vy,vz *= DAYS_PER_YEAR`; `mass *= SOLAR_MASS` (sun’s mass is already `1.0` → becomes `SOLAR_MASS`).

| body    | x                      | y                       | z                       | vx                       | vy                      | vz                        | mass                   |
| ------- | ---------------------- | ----------------------- | ----------------------- | ------------------------ | ----------------------- | ------------------------- | ---------------------- |
| sun     | 0                      | 0                       | 0                       | 0                        | 0                       | 0                         | 1.0                    |
| jupiter | 4.84143144246472090e+00 | -1.16032004402742839e+00 | -1.03622044471123109e-01 | 1.66007664274403694e-03  | 7.69901118419740425e-03 | -6.90460016972063023e-05  | 9.54791938424326609e-04 |
| saturn  | 8.34336671824457987e+00 | 4.12479856412430479e+00  | -4.03523417114321381e-01 | -2.76742510726862411e-03 | 4.99852801234917238e-03 | 2.30417297573763929e-05   | 2.85885980666130812e-04 |
| uranus  | 1.28943695621391310e+01 | -1.51111514016986312e+01 | -2.23307578892655734e-01 | 2.96460137564761618e-03  | 2.37847173959480950e-03 | -2.96589568540237556e-05  | 4.36624404335156298e-05 |
| neptune | 1.53796971148509165e+01 | -2.59193146099879641e+01 | 1.79258772950371181e-01  | 2.68067772490389322e-03  | 1.62824170038242295e-03 | -9.51592254519715870e-05  | 5.15138902046611451e-05 |

After building the list, run `offsetMomentum`: set sun velocity so total momentum is zero (`sun.v = -Σ(body.v * body.mass) / SOLAR_MASS`).

## Algorithm (single)

Three nesting levels in the hot path:

1. **`s` / `steps`** — outer simulation loop: each iteration is one time-step of size `dt`.
2. **`j`** — next outer: first body index of each unordered pair.
3. **`i`** — inner: second body index (`j < i`).

```
bodies := [sun, jupiter, saturn, uranus, neptune]   // mass, x,y,z, vx,vy,vz
offsetMomentum(bodies)   // zero total momentum onto sun (once, before timed steps)

// optional: energy0 := energy(bodies)   // checksum before

for s := 1 .. steps:                       // one advance per iteration
  // pairwise velocity update — j outer, i inner (Final foreach nest order)
  for j := 0 .. n-2:
    for i := j+1 .. n-1:                  // pairs where j < i
      dx := bj.x - bi.x;  dy := …;  dz := …
      dist := sqrt(dx*dx + dy*dy + dz*dz)
      mag := dt / (dist*dist*dist)          // dt usually 0.01
      bj.vx -= dx * bi.mass * mag
      bi.vx += dx * bj.mass * mag
      // same for vy, vz
  for each body b:                         // still inside this advance
    b.x += dt * b.vx;  b.y += …;  b.z += …

energy := kinetic + potential over bodies   // report / sanity
```

`energy` (CLBG-style):

```
e := 0
for each body b:
  e += 0.5 * b.mass * (vx²+vy²+vz²)
for j := 0 .. n-2:
  for i := j+1 .. n-1:
    dist := |bj.pos - bi.pos|
    e -= (bj.mass * bi.mass) / dist
return e
```

## Final sketch notes

- **First cut: clarity over speed.** Model each body as a `DynamicObject` (named fields: `mass`, `x`, `y`, `z`, `vx`, `vy`, `vz`) and hold the system in a compound `List` of those objects. Symbolic property access makes the pairwise math easy to read and reason about — that compound / object path is a deliberate signal, not an accident to micro-opt away in v1.
- Later (separate pass): flatten to typed lists / locals if we want a “native-fast” variant; keep the clear version as the readable reference.
- Hot path still: nested `foreach` / index loops + `Float` `* / + - sqrt`; no I/O inside `steps`.
- Print `energy` once after measured work for sanity (CLBG prints before/after).
- No Pips in `nbody.fnl`.

## Measured result

- Wall time for `steps` advances (or harness N repeats → `avg_s`).
- Optional: report final `energy` as a side field later; not required for v1 schema.
- Run under both `compile_context` = `core` and `local` once the harness exists.

## Sanity checks

- After `offsetMomentum`, total momentum ≈ 0.
- Known energy for CLBG ICs at `steps = 1000` (document expected value in a comment; tolerate small float drift).
- Tiny `steps` (e.g. 1): energy changes in the expected direction / magnitude vs reference.

## Out of scope (v1)

- Flattened / “native-fast” body representation (typed parallel lists, locals-only) — defer until after the DynamicObject version works.
- Magnets / soft-body / Barnes–Hut; stick to O(n²) pairwise CLBG.
- Matching published CLBG scoring / output formatting until we care about external comparison tables.
- Bit-exact float parity with a specific C reference (order-of-ops differences are fine if energy stays close).
