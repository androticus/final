# Level Change Events (2)

Attach handlers on a `GpioPin` after configuring pull / direction / debounce, before or as you enable.

This is an example of setting a handler to trigger when a pin goes low. The method on the object is called `handleGpioLow` -- the method name is arbitrary, but event handlers customarily start with `handle` then the name of the event being handled.

```final
  // ( mr:MethodRef index:i -- )
  o:.handleGpioLow~ 0 t:.setOnLow^
```

Stack for both setters: `( handler:MethodRef index:i -- )` on the pin. Use `index` when several pins share one handler (`ev.index`).

This handler is used for both a "disable" and "enable" blink button; `index` distinguishes them. The code does a compare to the index of the enable button, then uses that bool results to set the `blink` field. Note that accessing the .index field still leaves the event on DS; the `nip` removes it.

```final
// ( ev:Event -- ) o( inst:LedBtnCtrl )
: LedBtnCtrl#handleGpioLow
  .index@ nip
  LedBtnCtrl.idxEnable ==
  o:.blink!
;
```

Be careful to never leave an event on DS, because they will accumulate, possibly leading to stack overflow, but certainly polluting your stack!

Debouncing is covered in [Debouncing](pin-debouncing.md). Event basics: [Events](../intro/events.md).
