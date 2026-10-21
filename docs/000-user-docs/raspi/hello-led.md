# Hello LED!

Well... the LED doesn't actually say, "Hello!', but it does blink!

Wire the green LED as follows:

![Hello LED schematic: GPIO 27 to resistor to LED to GND](images/hello-led.svg)

from GPIO pin 27, to a 220R or 330R resistor<br>
from the resistor to the anode of the LED<br>
the cathode (flat edge) of the LED to GND

(Note: it is suggested to always do wiring with the Pi shutdown, and disconnected from power.)

Here is the program we will write. It is also in examples/pi/hello-led.fnl

Run the example file, and enter
```
hello-led

```

You can press Esc or any Fn keep to stop the demo (which you should do, so the hardware gets reset.)

```final
: hello-led
  // set all pins/periphs to default state+inactive
  GpioCtrl.inst.reset^
  
  // set gpio to a named local for clarity/ease of change of pin
  gpio27 l.led!

  // initialize the led: no pullup/down, output
  GpioPin.pullOff l.led.pull!
  GpioPin.out l.led.direction!
  true l.led.enabled!
  
  clearEscFn // clears the Esc/Fn key latch
  "\nPress Esc or any Fn to exit...\n" $ 
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
```
This simple example has the general steps you will use in any control program, simple or complex. They are:
1. Reset the Gpio pins to inputs, and disable all peripherals;
2. (optional but highly recommended) create mnemonic names for the pins and peripherals you will use; this makes the code much clearer, and makes changing pin or peripheral numbers simpler;
3. initialize every pin and peripheral you will use; after setting the parameters needed for a pin/periph, set `.enabled` to `true`;
4. initialize your own controller program's state variables;
5. establish a way to gracefully stop the program--using the clearEscFn then escFn? is especially well suited;
6. start the action:
  * for "legacy/poll loop" style, you create a `forever` loop with a `break` to exit
  * we will see later how to stop using an event-based controller

Using the Pi with Final is this simple!

Let's move on to discuss controllers in more detail, then we will cover pin and peripheral setup and use.
