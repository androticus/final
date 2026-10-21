# Pins and Periphs

Pi hardware is expressed via numbered GPIO pins, and numbered peripheral types. Final provides support for using all GPIO pins except 0 and 1, which are reserved for the HAT EEPROM protocol, and are not a great choice to try to use yourself. 

Pin functionality is exposed via the Final Type GpioPin. Peripheral functionality is exposed via the named types I2cPeriph, SpiPeriph, UartPeriph, and PwmPeriph.

You do not create instances of these types yourself; every available Pin and Periph is a global (e.g. `gpio2`, `uart`, `i2c`, `spi`). Prefer the unnumbered defaults; numbered alts are Pi 5 only.

The following page contains a map of all the GPIO pins, and the pin mapping of the available peripherals.

Note that Final does not expose the Pulse-width modulation audio. It is a strange and not too useful throwback approach to digital sound, and is typically not used. There are many alternative choices for getting sound from a Pi, such as a USB dongle, or I2S digital hats.

Final also does not drive I2S audio. Reset skips OS pinmux-claimed header pins (`GpioCtrl.osPinMap`, pin/periph `osMapped`). The `i2s` stub still offers `.reserved` as an escape hatch — see [I2S](periph-i2s.md).
