forget: reset-led-btn-pwm
requiresPi
`
Simple Pi hardware example, wired up as follows:
gpio27: 220R to green led to ground
gpio26: button, to ground
btnHigher: button, to ground
gpio12: 220R to red LED to ground

Shows how to create a simple "legacy-style" control 
program using looping and delays.
We use the PwmPeriph pulse-width-modulation peripheral to 
control the brightness of an LED via PwmPeriph0.

Functionality:

ledBtnPressed green LED, On when up or down is pressed
gpio12 red LED, interval 1 kHz
button5: lower brightness
button6: higher brightness

Try:

led-btn-pwm

(See led-btn-pwm2-pi-ctrl.fnl for a PiController approach.)
` const: led-btn-pwm2-blurb

gpio27 const: ledBtnPressed
gpio26 const: btnLower
gpio6 const: btnHigher
pwm0 const: pwmLed
pwm1 const: pwmServo

/* 
Both buttons get the same init.
( gpio -- )
*/
: led-btn-pwm_init_btn
  >ts
  GpioPin.pullUp t:.pull!
  GpioPin.in t:.direction!
  true t:.enabled!
  tdrop
;

: led-btn-pwm_init
  // start fresh each time
  GpioCtrl.inst.reset^
  
  // the green led on ledBtnPressed
  GpioPin.pullOff ledBtnPressed.pull!
  GpioPin.out ledBtnPressed.direction!
  true ledBtnPressed.enabled!
  
  // buttons on btnLower/6, same setup
  btnLower led-btn-pwm_init_btn
  btnHigher led-btn-pwm_init_btn
  
  // Init the pwmLed peripheral (1 kHz LED)
  pwmLed >ts
  PwmMode.linear t:.mode!
  1000 t:.linear.freq!
  0.5 t:.linear.duty!
  true t:.enabled!
  tdrop
  
  // Init the pwmServo peripheral (50 Hz, servo protocol)
  pwmServo >ts
  PwmMode.servo t:.mode!
  // hobby/Parallax servo defaults
  // these are the code defaults
  // commented out to test defaults; can be made explicit here for safety
  50 t:.servo.freq!
  //1000 2000 t:.servo.setMinMaxUs^
  0.0 t:.servo.asRange! // -1..1, center
  true t:.enabled!
  tdrop
;

/*
We use the pwm peripheral, which we run at 
1 kHz for fine and fast controls.
Duty is on a scale of 0..1
*/
: led-btn-pwm
  :s init
    Counter.new^ l.cnt! // diagnostics
    0.5 l.duty! // 0..1 brightness
    // 0: down; 1: up
    2 Int32List.newWithLen^ l.wasLow!
  ;s
  
  // ( new:f -- )
  :s setClampDuty
    switch
      a <0 case 0.0 /case
      a 1.0 > case 1.0 /case
      default a /default
    /switch
    l.duty!
  ;s

  :s refreshDuty
    l.duty dup pwmLed.linear.duty!
    0.5 - 2.0 * pwmServo.servo.asRange!
  ;s
    
  /*
  handle a button
  ( index:i gpio -- )
  */
  :s handleButton
    >ts
    t:.raw
    if
      // btn not pressed
      0 swap l.wasLow!()
    else
      true l.btnPressed!
      // if wasn't pressed, apply change
      dup l.wasLow@() if
        drop
      else
        dup 1 swap l.wasLow!()
        if 0.05 else -0.05 then dup
        l.duty@ + setClampDuty
        refreshDuty  
        ( peek l.duty ) "Changed: %f; Duty: %f\n" fmt $$ drop
      then
    then
    tdrop
  ;s
  
  led-btn-pwm_init
  init
  refreshDuty
  forever
    // we light green LED while a btn is pressed    
    false l.btnPressed!
    0 btnLower handleButton
    1 btnHigher handleButton
    l.btnPressed ledBtnPressed.value!
    20 delayMs // fast enough for UI
    l.cnt.inc^ // aprox 50 hz
  /forever
;

led-btn-pwm2-blurb $ nl
