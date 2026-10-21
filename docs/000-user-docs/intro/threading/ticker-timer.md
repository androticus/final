# Ticker and Timer Pipes

Final provides prebuilt **modal** Pipes that push `DateTime` either periodically (ticker) or once after an interval / at a clock time (timer).

Factories create an **unarmed** pipe. Lifecycle words both **initially arm** and **renew**.

## Ticker Pipe

```forth
Pipe.newTicker^ l.tick!
Interval.second l.tick@ pipe::tickerStart
// ...
l.tick@ pipe::tickerStop
```

- `pipe::tickerStart` `( iv:Interval p:Pipe -- )` — start or restart
- `pipe::tickerStop` `( p:Pipe -- )` — stop ticks; pipe stays open
- Do **not** `pipe::close` a ticker pipe (fault)
- **Do** `pipe::tickerStop` when no more ticks are needed

## Timer Pipe

```forth
Pipe.newTimer^ l.boom!
500 Interval.millisecond dt::* l.boom@ pipe::timerAfter
// or: someDateTime l.boom@ pipe::timerWhen
```

- `pipe::timerAfter` `( iv:Interval p:Pipe -- )` — arm or re-arm after interval
- `pipe::timerWhen` `( when:DateTime p:Pipe -- )` — arm or re-arm at time (past → immediate)
- Does **not** self-close after fire; idle if not rearmed is harmless
- Do **not** `pipe::close` a timer pipe (fault)

## Example

See `examples/threading/tick-tick-boom.fnl` for both types of Pipes. Note that the example only uses the Timer pipe once, but you can easily use it in a re-fire mode, by re-arming it after it signals.
