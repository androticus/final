# Event Handlers

Controllers can use Final’s EventLoop: you register one of your methods on an Event object, and then Final raises an `Event`, and your TypedObject method runs. Background on the model: [Events and the EventLoop](../intro/events.md).

For Pi work you mainly use:

| Source | Handler pattern | Covered in |
| --- | --- | --- |
| GPIO level | `#handleGpioLow` / `#handleGpioHigh` | [Level Change Events (2)](ctrl-level-events.md) |
| Periodic tick | `#handleTicker` | [Timer Events](ctrl-timer-events.md) |
| One-shot delay | `#handleTimer` (name as you like) | [Timer Events](ctrl-timer-events.md) |
| Tight re-arm loop | `#handleFrame` (name as you like) + `EventAgain` | [Events](../intro/events.md) |
| Esc / Fn | `#handleEscFn` | [Esc/Fn key Events](ctrl-escfn-events.md) |

Rules of thumb:

* Form a `MethodRef` with `~`, pass it when you create the source or call `setOnLow` / `setOnHigh`.
* Handlers receive `( ev:Event -- )` — drop `ev` or read `.index` / `.params` then drop.
* Do not block the EventLoop with `forever` / long `delayMs` loops; use ticker/timer, or `EventAgain` for no-delay re-arm.
* On stop, `.close^` `EventAgain` / tickers / timers and `EventEscFn`, then `GpioCtrl.inst.reset^`.

Worked examples: `examples/pi/led-btn-ctrl.fnl` (ticker + Esc); `examples/pi/i2c-ssd1306-life.fnl` (`EventAgain` + Esc). General ticker/timer demo: `examples/events/ticker-timer-demo.fnl`.
