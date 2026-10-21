# range and rangeby

Final provides two useful words for setting up counted iterated loops. `range` is for a definite counted loop with an increment of 1. `rangeby` is more flexible. They are described below. For detailed descriptions and stack effects, use `help`.

## `range`

`range` takes an inclusive Int start and end value, and can then provide an Iterator for   `foreach`. `i` is the value of the iteration value for that step, and starts with beg and ends with end. When using `range` to iterate over a counted zero-based range, you set beg to 0 and the end to be count-1. If the count is 0, no iterations are performed. The resulting `i` will be an Int.

```forth
> // ( -- )
: count1to5
  1 5 range
  foreach
    i $ sp
  /foreach
  nl
;
count1to5

1 2 3 4 5
```

Zero-based index style (here: three slots, indices `0` `1` `2`):

```forth
> // ( count:i -- )
: indices
  0 swap 1 - range
  foreach
    i $ sp
  /foreach
  nl
;
3 indices

0 1 2
```

An empty range is legal and runs zero iterations — useful when a computed end would otherwise need a special “nothing to do” case.

## `rangeby`

`rangeby` takes a beg, end, and step value. Step can be negative if beg>=end. The values can be of any numeric type, but they must all be the same type. This enables fractional Float step values. The iteration stops when the value goes beyond the end value. As with range, an empty range that has no first value is legal, and results in no iterations. `i` is the same type as the inputs.

```forth
> // ( -- )
: byTens
  0 40 10 rangeby
  foreach
    i $ sp
  /foreach
  nl
;
byTens

0 10 20 30 40
```

Float step:

```forth
> // ( -- )
: halves
  0.0 2.0 0.5 rangeby
  foreach
    i $ sp
  /foreach
  nl
;
halves

0 0.5 1 1.5 2
```

Counting down (negative step; `beg >= end`):

```forth
> // ( -- )
: countdown
  5 1 -1 rangeby
  foreach
    i $ sp
  /foreach
  nl
;
countdown

5 4 3 2 1
```
