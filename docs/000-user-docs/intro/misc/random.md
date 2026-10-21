# Random Values and Noise

## Rand

`Rand` provides methods for creating pseudo-random values as follows. NOTE: do not use Rand for crytopgraphic contexts; use [SecureRandom](../crypto/random.md) instead.


| Method      | Parameters   | Purpose                                                |
| ----------- | ------------ | ------------------------------------------------------ |
| `int`       | n:i -- val:i | Random integer from 0 to N-1                           |
| `float`     | ( -- val:f ) | Random float in 0.0<= val <1.0                         |
| `normFloat` | ( -- val:f ) | Random float from normal distribution, mean=0 stddev=1 |


`Rand` has no instances — call class methods such as `Rand.int^`.

## Noise

`Noise` generates Perlin noise: adjacent samples are random but related. Use it for terrain maps and other fields where nearby points should vary smoothly rather than independently (as with `Rand`).


| Method        | Parameters             | Purpose                                             |
| ------------- | ---------------------- | --------------------------------------------------- |
| `new`         | ( -- n:Noise )         | Create a generator with a random seed               |
| `newWithSeed` | ( seed:i -- n:Noise )  | Create a generator with a fixed seed (reproducible) |
| `at1`         | ( x:f -- v:f )         | Sample at one coordinate                            |
| `at2`         | ( x:f y:f -- v:f )     | Sample at two coordinates (maps / terrain)          |
| `at3`         | ( x:f y:f z:f -- v:f ) | Sample at three coordinates (e.g. treat z as time)  |


Sample values are typically near -1.0 … 1.0.

```forth
Noise.new^ >ts
0.1 0.2 t:.at2^
// or reproducible:
42 Noise.newWithSeed^ >ts
0.1 0.2 t:.at2^
```

See also `examples/misc/noise-demo.fnl`.