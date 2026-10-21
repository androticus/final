# Writing Methods

Controller methods use the usual Final form `TypeName#methodName`. The instance is on the O stack (`o:`).

Handlers for the EventLoop take the event and drop it (or pull fields first):

```final
/*
Handle the pin going low.
( ev:Event -- ) o( inst:LedBtnCtrl )
*/
: LedBtnCtrl#handleGpioLow
  .index@ nip                 // which button (from setOnLow index)
  LedBtnCtrl.idxEnable ==
  o:.blink!
;
```

Non-handler methods are fine for shared work (`#init`, `#read`, and so on). Prefer short methods that do one job; put reusable bits in factor words (next page).

Stack comments matter — keep `( ev:Event -- ) o( inst:YourCtrl )` accurate so `help` stays useful.

See [Objects and methods](../intro/objects/methods.md) in the main guide.
