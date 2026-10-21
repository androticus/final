# Factor Words

Factor repeated code, or merely factored ("clean") code into plain Final words (not methods); those words will still see the instance on the O stack. Naming convention in the examples: `TypeName_helper`.

```final
/*
Both buttons get the same init.
( index:i gpio -- ) o( inst:LedBtnCtrl )
*/
: LedBtnCtrl_initBtn
  >ts
  GpioPin.pullUp t:.pull!
  GpioPin.in t:.direction!
  GpioPin.debounceButton t:.debounceMs!
  o:.handleGpioLow~ swap t:.setOnLow^
  true t:.enabled!
  tdrop
;
```

Local stanzas with `:s name ... ;s` are useful inside `#init` to split “wire the Pi” from “wire self/events” without polluting the dictionary — see `LedBtnCtrl#init` in `examples/pi/led-btn-ctrl.fnl`.
