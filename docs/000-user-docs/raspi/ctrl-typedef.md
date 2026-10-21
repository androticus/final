# Creating the TypeDef

An event-driven controller is a `TypedObject`: fields for pins/periphs and event sources, plus constants, then methods for handlers and init.

Use a paren-object spec and `TypedObject.makeType^`. Full example: `examples/pi/led-btn-ctrl.fnl`.

```final
{
  "LedBtnCtrl" "name"
  {
    GpioPin "led"
    GpioPin "btnEnable"
    GpioPin "btnDisable"
    EventTicker "ticker"
    EventEscFn "escFn"
    Bool "blink"
  } "fields"
  {
    0 "idxEnable"
    1 "idxDisable"
    2 "ticksPerS"
  } "constants"
}
TypedObject.makeType^
```

Keep a global var holding the live instance so it is not garbage-collected while the EventLoop still has handlers pointing at it. Clear that var in your stop path.

```final
null var: LedBtnCtrl_inst
```

See also [TypedObject](../intro/compounds/typedobject.md) in the main guide.
