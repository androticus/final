# GpioPin Basics

Each pin constant, from `gpio2` to `gpio27` is an instance of the GpioPin Type. You can configure the following:

* internal pullup or pulldown resistors, or none
  - these are typically only used on input pins, output pins actively drive both high and low
* input or output (direction)
* for inputs: debouncing interval, or none
* level change event handlers: onLow and onHigh

We already saw an example in Hello LED! of how to set a pin for output. Here is an example of using all the pin features, from several example code files. The hardware is a button connecting the pin to ground when pressed. It uses the internal pullup resistor, to avoid needing an external, one, and uses debouncing and level change events. ([Event Handlers](ctrl-events.md) are covered later in the manual.)

Any property that requires some kind of mode or set of limited values has corresponding enums defined as constants on the Type, as you see here.

```final
/* 
Both buttons get the same init.
( index:i gpio -- ) o( inst:LedBtnPwmCtrl )
*/
: LedBtnPwmCtrl_initBtn
  >ts
  GpioPin.pullUp t:.pull!
  GpioPin.in t:.direction!
  GpioPin.debounceButton t:.debounceMs!
  true t:.enabled!
  // events
  dup o:.handleGpioHigh~ swap t:.setOnHigh^
      o:.handleGpioLow~ swap t:.setOnLow^
  tdrop
;
```

The default value of .pull is pullOff, and the default value of debounceMs is 0, so you don't need to set those if the defaults are fine (typical for an output). You should also set `pull` to `pullOff` when your input is being driven from an active digital source.
