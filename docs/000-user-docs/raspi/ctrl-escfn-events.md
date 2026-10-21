# Esc and Fn Key Events

Final provides `EventEscFn` so you can receive an event when the user presses Esc or a Fn key. Handy for development, demos, and a clean shutdown path. Event model: [Events](../intro/events.md). 

Several of the Pi examples use the Esc key to exit the program. Here is the typical handler, taken from `examples/pi/led-btn-ctrl.fnl`. The handler receives the key code directly: Esc is 0, Fn keys are 1–12.

A typical EventEscFn handler code:

```final
/*
Handle an Esc or Fn key press.
We look specifically for Esc.
( code:i -- )  o( inst:LedBtnCtrl )
*/
: LedBtnCtrl#handleEscFn
  // code: 0=Esc, else Fn#
  ==0 if
    LedBtnCtrl_stop
  then
;
```

Here is the initialization code. Note: you have to call the `.start` method after creating it.

```final
    o:.handleEscFn~ EventEscFn.new^ o:.escFn!
    o:.escFn.start^
```

When your program stops, you should close the EventEscFn object, so it doesn't continue to send events to zombie code.

```final
: LedBtnCtrl_stop
  // stop receiving EscFn events
  o:.escFn.close^

  // ... more code
```
