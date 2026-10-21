forget: reset-hello-led
requiresPi
`
"Hello Led!" Pi hardware example, wired up as follows:
gpio27: 220R to a green or red LED to ground
(Don't use blue or white, they need drive voltage > 3.3V.)

Functionality:

gpio27 LED, flashes at .5 Hz
Press Esc or Fn* button to exit

Switch to Stdout/Stderr View, then 
Try:

hello-led
` const: led-blurb


: hello-led
  // set all pins/periphs to default state+inactive
  GpioCtrl.inst.reset^
  
  // set gpio to a named local for clarity/ease of change of pin
  gpio27 l.led!

  // initialize the led
  GpioPin.pullOff l.led.pull! // output, no pullup
  GpioPin.out l.led.direction!
  true l.led.enabled!
  
  clearEscFn // clears the Esc/Fn key latch
  "\nPress Esc to exit...\n" $ 
  forever
    // check for Esc/Fn pressed
    escFn? if
      ==0 if // Esc=0 else Fn#
        "Stopping demo...\n" $
        // put all pins/periphs to default input
        GpioCtrl.inst.reset^
        break
      then
    then
    
    // toggle the LED and wait the delay
    l.led.value not l.led.value!
    500 delayMs
  /forever
;

led-blurb $ nl
