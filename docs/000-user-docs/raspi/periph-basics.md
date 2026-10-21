# Peripheral Basics

Pi provides a number of peripherals that would be complex or impossible to create via manual toggling/reading of pins. Several instances of each type of peripheral are available to use, with the selection being primarily based on pin assignment.

No peripheral is enabled by default. To use a peripheral, you set any parameters it may need (some like I2C have none), then set `.enabled` to `true`.

A peripheral can have some optional pins it uses, based on its setup. Uart can optionally use the RTS/CTS pins. Spi peripherals can have more than 1 CS activated. If you do not activate the optional feature, then the corresponding pins remain available for GPIO use.

In all cases, the defined and enabled optional pins of a peripheral may not be used by a GpioPin or another peripheral. Many of the available peripherals do overlap others, so you need to make sure the pin mapping peripheral choices of your hardware schematic doesn't cause any overlap. Final does not support dynamic enabling/disabling of peripherals or pins; the schema of your hardware needs to be established during the initialization phase of your controller. (Note: you *can* dynamically enable pins or peripherals at any point, provided they are still available on the map.)

Final provides the following peripheral instances. Prefer the **unnumbered** defaults in prose and examples. Numbered alts are Pi 5 only; Zero 2 keeps defaults only. * = default / pinout-card highlight. (Some excluded Pi peripherals use GPIO0/1.)

| Kind | Prefer | Also (Pi 5 alts) |
| --- | --- | --- |
| I2C | `i2c` (`i2c1`) | `i2c2`, `i2c3` |
| SPI | `spi` (`spi0`) | `spi1`, `spi5` |
| UART | `uart` (`uart0`) | `uart2`, `uart3` |
| PWM | `pwm0`, `pwm1` | — |

Final does not expose the rather puzzling PWM audio channel, nor the I2S high speed digital audio peripherals. You may use those with operating system provided drivers.

`GpioCtrl.reset` restores Final’s in-process defaults but **does not** force GPIO pad restore on pins the kernel already muxed (UART, PWM, I²C, SPI, I2S overlays). Inspect with `GpioCtrl.inst.osPinMap @` (BCM-indexed `StringList`) or pin/periph `.osMapped @` (`""` = free).

The sections that follow provide the API for each type of peripheral.
