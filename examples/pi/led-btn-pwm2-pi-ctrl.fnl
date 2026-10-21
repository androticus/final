forget: reset-led-btn-pwm
requiresPi
`
Simple Pi hardware example, wired up as follows:
gpio4: 220R to green led to ground
gpio5: button, to ground
gpio6: button, to ground
gpio12/pwm0: 220R to red LED to ground

Shows how to create an object-based event-driven 
Pi Controller.
We use pulse-width-modulation to control the
brightness of an LED via PWM0 peripheral.

Functionality:

gpio27 green LED, On when lower or higher is pressed
gpio26 btnLower: lower brightness, in 0.05/1.0 steps
gpio6 btnhigher brightness, same steps
pwm0 red LED, interval 1 kHz
pwm1 servo, standard hobby servo control

Try:

led-btn-pwm

(See led-btn-pwm2-no-ctrl.fnl for a more manual approach.)
` const: led-btn-pwm2-blurb

/*
Create our TypeDef object for our fields and constants.
We use the convenient { } paren operators for the spec.
*/
{
  "LedBtnPwmCtrl" "name"
  
  // fields
  {
    // named aliases for our periphs
    GpioPin "ledBtnPressed"
    GpioPin "btnLower"
    GpioPin "btnHigher"
    PwmPeriph "pwmLed"
    PwmPeriph "pwmServo"
    EventEscFn "escFn" // for cancelling demo
    
    Float "duty" // 0..1.0, mapped to -1..1 for servo
  } "fields"
  
  
  // constants
  {
    // for linear/led, servo uses standard servo defaults
    1000 "defFreq"
    0.5 "defDuty"
    0 "idxLower"
    1 "idxHigher"
  } "constants"
}
TypedObject.makeType^
// LedBtnPwmCtrl type now exists

// need a var to keep us alive, cleared on stop
null var: LedBtnPwmCtrl_inst


/*
Handle a button Up event.
We turn off the green feedback LED
( ev:Event -- )  o( inst:LedBtnPwmCtrl )
*/
: LedBtnPwmCtrl#handleGpioHigh
  drop
  false o:.ledBtnPressed.value!

;

/*
Refresh the duty of both pwms, from our value.
Do this at startup, and after changing .duty
( -- ) o( inst:LedBtnPwmCtrl )
*/
: LedBtnPwmCtrl_refreshDuty
  o:.duty dup o:.pwmLed.linear.duty!
  0.5 - 2.0 * o:.pwmServo.servo.asRange!
;

/*
Handle a button down event.
Raise/Lower the brightness, and
turn on the feedback LED.
( ev:Event -- )  o( inst:LedBtnPwmCtrl )
*/
: LedBtnPwmCtrl#handleGpioLow
  // ( new:f -- )
  :s setClampDuty
    switch
      a <0 case 0.0 /case
      a 1.0 > case 1.0 /case
      default a /default
    /switch
    o:.duty!
  ;s

  true o:.ledBtnPressed.value!
  .index@ nip
  switch // on index
    a LedBtnPwmCtrl.idxLower == case -0.05 /case
    a LedBtnPwmCtrl.idxHigher == case 0.05 /case
  /switch
  
  dup // for diag print
  o:.duty@ + setClampDuty
  LedBtnPwmCtrl_refreshDuty  
  ( peek o:.duty ) "Changed: % 3.2f; Duty: %03.2f\n" fmt $$ drop
;

/* 
Stop all activity, and release instance.
*/
: LedBtnPwmCtrl_stop
  "Stopping PWM demo.\n" $
  // stop all EscFn events, and close gracefully
  o:.escFn.close^

  // put all pins/periphs to default input
  GpioCtrl.inst.reset^

  // so the instance can be GC'ed
  null LedBtnPwmCtrl_inst !
;

/*
Handle an Esc or Fn key press.
We look specifically for Esc.
( code:i -- )  o( inst:LedBtnPwmCtrl )
*/
: LedBtnPwmCtrl#handleEscFn
  // code: 0=Esc, else Fn#
  ==0 if
    LedBtnPwmCtrl_stop
  then
;

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

/* 
Factored init code
( gpio -- ) o( inst:LedBtnPwmCtrl )
*/
: LedBtnPwmCtrl_init
  :s initPi    
    // assign our named pins/periphs
    gpio27 o:.ledBtnPressed!
    gpio26 o:.btnLower!
    gpio6 o:.btnHigher!
    pwm0 o:.pwmLed!
    pwm1 o:.pwmServo!
    
    // initialize the gpio/periphs
    // the green led
    GpioPin.pullOff o:.ledBtnPressed.pull!
    GpioPin.out o:.ledBtnPressed.direction!
    true o:.ledBtnPressed.enabled!
    
    LedBtnPwmCtrl.idxLower o:.btnLower LedBtnPwmCtrl_initBtn
    LedBtnPwmCtrl.idxHigher o:.btnHigher LedBtnPwmCtrl_initBtn
    
    // Init the pwm0 peripheral (1 kHz LED)
    PwmMode.linear o:.pwmLed.mode!
    LedBtnPwmCtrl.defFreq o:.pwmLed.linear.freq!
    LedBtnPwmCtrl.defDuty o:.pwmLed.linear.duty!
    true o:.pwmLed.enabled!
    
    // Init the pwm1 peripheral, a Parallax hobby servo
    PwmMode.servo o:.pwmServo.mode!
    // hobby/Parallax servo defaults
    // these are the code defaults
    // commented out to test defaults; can be made explicit here for safety
    50 o:.pwmServo.servo.freq!
    //1000 2000 o:.pwmServo.servo.setMinMaxUs^
    0.0 o:.pwmServo.servo.asRange! // -1..1, center 0
    true o:.pwmServo.enabled!
  ;s
  
  :s initSelf
    LedBtnPwmCtrl.defDuty o:.duty!
    o:.handleEscFn~ EventEscFn.new^ o:.escFn!
    o:.escFn.start^
  ;s
  
  // start fresh each time
  GpioCtrl.inst.reset^
  initPi
  initSelf
  LedBtnPwmCtrl_refreshDuty
;

/* 
Initialize the instance. Nothing more to do after
#init — duty defaults to .5; inactivity watchdog
stops tickers and GpioCtrl.reset^ after 60s idle.
o( inst:LedBtnPwmCtrl )
*/
: LedBtnPwmCtrl#init
  "Press Esc to end PWM demo...\n" $
  LedBtnPwmCtrl_init
;

/*
We use the pwmLed peripheral, which we run at 
1 kHz for fine and fast controls.
Duty is on a scale of 0..1
*/
: led-btn-pwm
    
  LedBtnPwmCtrl.new^ dup
  LedBtnPwmCtrl_inst ! // needed to keep us alive
  >ts
  t:.init^
  tdrop
;

led-btn-pwm2-blurb $ nl
