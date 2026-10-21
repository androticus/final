forget: reset-led-btn
requiresPi
`
Simple Pi hardware example, wired up as follows:
gpio27: 220R to green led to ground
gpio26: button, to ground
gpio6: button, to ground

Shows how to create a controller-style 
program using the event loop and event handlers.

Functionality:

gpio27 LED, flashes at .5 Hz
button26: enable blinking
button6: disable blinking
60s inactivity timeout with 10s warning countdown;
timeout resets the hardware and exits.

Switch to Stdout/Stderr View, then 
Try:

led-btn

(See led-btn-no-ctrl.fnl for a legacy approach.)
` const: led-btn-blurb

{
  "LedBtnCtrl" "name"
  
  // fields
  {
    GpioPin "led"
    GpioPin "btnEnable"
    GpioPin "btnDisable"
    EventTicker "ticker"
    EventEscFn "escFn" // to cancel demo
    Bool "blink" // enabled or not
    Counter "timeout"
  } "fields"

  // constants
  {
    0 "idxEnable"
    1 "idxDisable"
    2 "ticksPerS" // integral, min 1
    60 "timeoutS"
  } "constants"
}
TypedObject.makeType^
// LedBtnCtrl Type now exists

// need a var to keep us alive, cleared on stop
null var: LedBtnCtrl_inst

/*
Handle a button down event.
Enable or Disable blinking
( ev:Event -- )  o( inst:LedBtnPwmCtrl )
*/
: LedBtnCtrl#handleGpioLow
  .index@ nip
  LedBtnCtrl.idxEnable ==
  o:.blink!
;

/* 
Stop all activity, and release instance.
*/
: LedBtnCtrl_stop
  "Stopping demo...\n" $
  // stop all ticker events, and close gracefully
  o:.ticker.close^

  // stop receiving EscFn events
  o:.escFn.close^
  
  // put all pins/periphs to default input
  GpioCtrl.inst.reset^

  // so the instance can be GC'ed
  null LedBtnCtrl_inst !
;

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

/*
Handle the ticker event, 1/s.
Warn user of inactivity, and shutdown.
( -- )  o( inst:LedBtnCtrl )
*/
: LedBtnCtrl#handleTicker
  o:.blink if
    o:.led.value not
  else
    false
  then
  o:.led.value!
;

/* 
Both buttons get the same init.
( index:i gpio -- ) o( inst:LedBtnPwmCtrl )
*/
: LedBtnCtrl_initBtn
  >ts
  GpioPin.pullUp t:.pull!
  GpioPin.in t:.direction!
  GpioPin.debounceButton t:.debounceMs!
  // we only need onLow event
  o:.handleGpioLow~ swap t:.setOnLow^
  true t:.enabled!
  tdrop
;

/* 
Factored init code
( gpio -- ) o( inst:LedBtnCtrl )
*/
: LedBtnCtrl#init
  :s initPi    
    // assign our named pins/periphs
    gpio27 o:.led!
    gpio26 o:.btnEnable!
    gpio6  o:.btnDisable!
    
    // initialize the gpio/periphs
    // the green led
    GpioPin.pullOff o:.led.pull!
    GpioPin.out o:.led.direction!
    true o:.led.enabled!
    
    LedBtnCtrl.idxEnable o:.btnEnable LedBtnCtrl_initBtn
    LedBtnCtrl.idxDisable o:.btnDisable LedBtnCtrl_initBtn
  ;s
  
  :s initSelf
    true o:.blink!
    o:.handleEscFn~ EventEscFn.new^ o:.escFn!
    o:.escFn.start^
    o:.handleTicker~ EventTicker.new^ o:.ticker!
    1000 LedBtnCtrl.ticksPerS / // tick interval ms
    Interval.millisecond dt::* o:.ticker.eventEvery^
  ;s
  
  // start fresh each time
  GpioCtrl.inst.reset^
  initPi
  initSelf
  //true o:.led.value! // start on
  "Starting demo. Press Esc to stop.\n" $
;


: led-btn
    
  LedBtnCtrl.new^ dup
  LedBtnCtrl_inst ! // needed to keep us alive from GC
  >ts
  t:.init^
  tdrop
;

led-btn-blurb $ nl
