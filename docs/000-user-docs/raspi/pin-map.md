# Pin Map

This table lists all the GpioPins that Final provides, and the Periph instances it provides. Note that if you enable a numbered Periph, you can't use the correspoding GpioPins or another peripheral that it overlaps with. If you enable an optional feature, such as Uart CTS/RTS then the same restriction applies. "Def(ault) Periph" doesn't mean it is enabled by default; ALL periphs and pins must be enabled to be used; it just means this is the conventional or highlighted use, eg `i2c` and `uart`.

The `pinMap` command is available in Final, to print the current mapping.
```final
pinMap
```

Note: only the default peripherals are available on Pi 2 Zero.

| GPIO | Def Periph | Alt Periph 1 | Alt Periph 2 |
| --- | --- | --- | --- |
| 2 | I2C SDA |  |  |
| 3 | I2C SCL |  |  |
| 4 |  | I2C2 SDA | UART2 TX |
| 5 |  | I2C2 SCL | UART2 RX |
| 6 |  | I2C3 SDA | UART2 CTS^ |
| 7 | SPI CE1^ | I2C3 SCL  | UART2 RTS^ |
| 8 | SPI CE0 |  | UART3 TX |
| 9 | SPI MISO |  | UART3 RX |
| 10 | SPI MOSI  |  | UART3 CTS^  |
| 11 | SPI SCLK  |  | UART3 RTS^ |
| 12 | PWM0 |  | SPI5 CE0 |
| 13 | PWM1 |  | SPI5 MISO |
| 14 | UART TXD |  | SPI5 MOSI |
| 15 | UART RXD |  | SPI5 SCLK |
| 16 | UART CTS^ | SPI1 CE2^ |  |
| 17 | UART RTS^ | SPI1 CE1^ |  |
| 18 |  | SPI1 CE0 | I2S (`i2s`) @ |
| 19 |  | SPI1 MISO | I2S (`i2s`) @ |
| 20 |  | SPI1 MOSI | I2S (`i2s`) @ |
| 21 |  | SPI1 SCLK | I2S (`i2s`) @ |
| 22 |  |  |  |
| 23 |  |  |  |
| 24 |  |  |  |
| 25 |  |  |  |
| 26 |  | | SPI5 CE1^ |
| 27 |  |  |  |

Notes
^ - indicates an optional feature; if not enabled, GPIO pin is available. See the chapter for the peripheral type for details<br>
@ I2S audio (ALSA); Final does not drive it. Reset skips OS-claimed pinmux; `.reserved` is an escape hatch — see [I2S](periph-i2s.md).
