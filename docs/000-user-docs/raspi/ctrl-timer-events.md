# Timer Events

Use `EventTicker` for a steady beat (blink, poll, heartbeat). Use `EventTimer` for a one-shot delay or deadline. Details and a desktop demo: [Events](../intro/events.md) and `examples/events/ticker-timer-demo.fnl`.

Create with a handler MethodRef (one handler per timer/ticker), then arm:

```final
  o:.handleTicker~ EventTicker.new^ o:.ticker!
  500 Interval.millisecond dt::* o:.ticker.eventEvery^

  o:.handleTimeout~ EventTimer.new^ o:.timer!
  60 Interval.second dt::* o:.timer.eventAfter^
```

| Type | Arm | Pause | Done |
| --- | --- | --- | --- |
| EventTicker | `.eventEvery` `( Interval -- )` | `.stop` | `.close` |
| EventTimer | `.eventAfter` or `.eventAt` | `.stop` | `.close` |

Handler sketch (empty stack — no Event / index to drop):

```final
: LedBtnCtrl#handleTicker
  o:.blink if
    o:.led.value not o:.led.value!
  then
;
```

Always `.close^` in your stop path so events do not keep firing after teardown.
