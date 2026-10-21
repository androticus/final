# Counter / Average / MovingAverage

## Counter

`Counter` is a lighter-weight, non-threadsafe alternative to AtomicInt. It contains an Int value, with convenience methods such as `inc` `dec` etc. that do not return the new value. It is simpler to use than trying to juggle a value on a stack, fetch/increment/put back. It works especially well in a a local variable, eg 

```final
: countme
  Counter.new^ l.cnt!
  0 9 range
  foreach
    l.cnt.inc^
  /foreach
  "Count: " $ l.cnt@.value $ nl
;
```

## Average

`Average` maintains a running average, variance, and N to-date without needing to retain the samples. It is mathematically correct, in the sense that the result is the same as accumulating the samples in a list then iterating an average/variance. Use `reset` to clear.

`Average` is threadsafe, and has `getMeanAndN` and `getMeanStdDevAndN` methods for getting the values atomically.

## MovingAverage

`MovingAverage` is an exponential moving average (EMA) for signals that drift — live costs, rates, telemetry. Create a default with a roughly 20 item span with `MovingAverage.new^`. To create with a longer or shorter span, use `MovingAverage.newWithSpan^`; `.span` is fixed afterward (effective window in samples). Internally `α = 2/(span+1)`. Same `.mean` / `.n` / `add` / `reset` / `getMeanAndN` API as `Average`, but no variance.

Prefer `Average` when you need an exact cumulative mean and variance; prefer `MovingAverage` when recent samples should outweigh old ones (e.g. SimClock update/render costs).
