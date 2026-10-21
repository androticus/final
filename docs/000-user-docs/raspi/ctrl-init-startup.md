# Init and Startup

Recommended order (see `examples/pi/led-btn-ctrl.fnl` and friends):

1. `GpioCtrl.inst.reset^` — all pins input, peripherals off.
2. Assign named aliases: `gpio27 o:.led!`, `pwm0 o:.pwmLed!`, etc.
3. Configure each pin/periph, then `true … .enabled!`.
4. Create event sources (`EventEscFn`, `EventTicker`, …), `.start^` / `.eventEvery^` as needed.
5. Keep the instance in a global var so it stays alive.

```final
: led-btn
  LedBtnCtrl.new^ dup
  LedBtnCtrl_inst !
  >ts
  t:.init^
  tdrop
;
```

Stop path: close event sources, `GpioCtrl.inst.reset^`, clear the instance var.

For headless / command-line launch (no TUI EventLoop yet), see [Command Line Invocation](run-cmdline.md) and the `runWithStart` pattern in `examples/events/ticker-timer-demo.fnl`.
