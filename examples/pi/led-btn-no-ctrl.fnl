forget: reset-led-btn
requiresPi
`
Simple Pi hardware example, wired up as follows:
gpio27: 220R to green led to ground
gpio26: button, to ground
gpio6: button, to ground

Shows how to create a simple "legacy-style" control 
program using looping and delays.

Functionality:

gpio27 LED, flashes at .5 Hz
button26: enable blinking
button6: disable blinking
60s inactivity timeout with 10s warning countdown;
timeout resets the hardware and exits.

Switch to Stdout/Stderr View, then 
Try:

led-btn

(See led-btn-ctrl.fnl for a controller/events approach.)
` const: led-btn-blurb

gpio27 const: led
gpio26 const: btnEnable
gpio6 const: btnDisable

/* 
Both buttons get the same init.
( gpio -- )
*/
: led-btn_init_btn
  >ts
  GpioPin.pullUp t:.pull!
  GpioPin.in t:.direction!
  true t:.enabled!
  tdrop
;

: led-btn
  :s resetTimeout
    600 l.timeout.set^
  ;s
  
  :s init
    // start fresh each time
    GpioCtrl.inst.reset^
    
    // the green led on gpio4
    GpioPin.pullOff led.pull!
    GpioPin.out led.direction!
    true led.enabled!
    
    // buttons on gpio5/6, same setup
    btnEnable led-btn_init_btn
    btnDisable led-btn_init_btn
  
    true l.blink!
    Counter.new^ l.cnt! // 100 ms/
    Counter.new^ l.timeout! // shuts down after 60s inactive
    resetTimeout // btn press refreshes
  ;s


  // warn user at 10s, every second
  // ( -- timedOut:b )
  :s warnTimeout?
    l.timeout.dec^
    l.timeout.value
    switch
      a <=0 case 
        true
      /case
      
      // less that 10 s, integral seconds
      a 100 <=  a 10 % not && case
        ( a 10 / ) "Exiting due to inactivity in %d seconds...\n" fmt $$
        false
      /case
      
      default false /default
  /switch
  ;s
  
  init
  forever
    // gpio5 is blinking "on", 6 is "off"
    btnEnable.value nif
      true l.blink!
      "set blink ON\n" $$
      resetTimeout
    then
    btnDisable.value nif
      false l.blink!
      "set blink OFF\n" $$
      resetTimeout
    then
    
    l.blink if
      // high on every high 500 ms
      l.cnt.value 10 % 5 < // t for 500, f for 500
    else
      false
    then
    led.value!
         
    100 delayMs
    l.cnt.inc^
    
    warnTimeout? if
      GpioCtrl.inst.reset^
      break
    then
    
  /forever
;

led-btn-blurb $ nl
