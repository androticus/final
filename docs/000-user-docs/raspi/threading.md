# Multi-threading

Most Pi controllers stay on the main EventLoop thread: GPIO events, tickers, and Esc/Fn all dispatch there. Prefer that model.

If you need background work (long compute, blocking I/O that must not stall handlers), use Final **Pips** and **Pipes**. Overview: [Threading](../intro/threading/index.md) and [Patterns](../intro/threading/patterns.md).

Notes for controllers:

* Only the main thread has an EventLoop. A Pip cannot register event handlers; it can send work results back on a Pipe for the main side to handle.
* Do not call long `delayMs` inside a handler — use `EventTicker` / `EventTimer` instead ([Timer Events](ctrl-timer-events.md)).
* Hardware setup/teardown (`enabled!`, `GpioCtrl.inst.reset^`) should stay on the main controller path unless you have a clear ownership scheme.
