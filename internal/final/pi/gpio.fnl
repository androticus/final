// internal/final/pi/gpio.fnl — Pi foundation (BCM/SoC first)
// Compiled into sealed core when finalpi is linked (bootstrap tranche4).
//
// Platforms: Pi 5 (RP1) and Pi Zero 2 (classic BCM). GpioCtrl.piSupport /
// _piSupport select the profile. Prefer unnumbered defaults: uart / i2c / spi
// (plus pwm0/pwm1). Alt buses (uart2/3, i2c2/3, spi1/spi5) are Pi 5 only
// (assertPi5 + Go needsPi5).
//
// Overlays (config.txt) — Pi 5 uses *-pi5 names; Zero 2 uses classic dtparam /
// non-pi5 overlays:
//   PWM:    dtoverlay=pwm-2chan,pin=12,pin2=13,func=4,func2=4
//           (Alt0 on 12/13; omit func → Alt5 defaults, not PWM on those pins)
//   I2C:    Pi5: dtoverlay=i2c1-pi5  |  Z2: dtparam=i2c_arm=on
//   SPI:    both: dtparam=spi=on     (spi1/spi5 overlays are Pi5-only alts)
//   UART0:  Pi5: dtoverlay=uart0-pi5 → /dev/ttyAMA0
//           Z2:  enable_uart=1; if BT owns PL011, header may be /dev/ttyS0 —
//                set FINAL_UART0=/dev/ttyS0 (or disable BT / use miniuart-bt)
//   gpiochip: auto pinctrl-rp1 (Pi5) / pinctrl-bcm2835 (Z2); override FINAL_GPIOCHIP
//
// GpioCtrl.inst exposes BCM 2–27 (0/1 HAT ID reserved).
// Pin IDs are BCM/SoC numbers only (not header/BOARD).
//
// Soft setup:
//   gpio4 >ts
//   GpioPin.out t:.direction!
//   true t:.value!
//   true t:.enabled!
//   ts>
//
// PWM (mode Raw default; pwm0→12, pwm1→13; stock pwm-2chan overlay):
//   pwm0 >ts
//   PwmMode.linear t:.mode!
//   1000 t:.linear.freq!
//   0.25 t:.linear.duty!
//   true t:.enabled!
//   ts>
//
// UART (uart1 / GPIO0–1 HAT ID not supported):
//   uart0 GPIO14/15 ttyAMA0  dtoverlay=uart0-pi5 (Pi5) / enable_uart (Z2; see above)
//   uart2 GPIO4/5   ttyAMA2  dtoverlay=uart2-pi5           flowHW → CTS/RTS 6/7   (Pi5 only)
//   uart3 GPIO8/9   ttyAMA3  dtoverlay=uart3-pi5           flowHW → CTS/RTS 10/11 (Pi5 only)
//   (uart4 not supported — GPIO12/13 clash with spi5; use uart0/2/3)
//   uart >ts
//   115200 t:.baud!
//   true t:.flowHW!   // before enabled; claims opt CTS/RTS for Validate
//   true t:.enabled!
//   72 t:.write   // 'H' as Int; or ByteList for a burst
//   ts>
//
// I2C (i2c0 HAT EEPROM not exposed):
//   i2c1 GPIO2/3  /dev/i2c-1   dtparam=i2c_arm=on (Z2) or dtoverlay=i2c1-pi5 (Pi5)
//   i2c2 GPIO4/5  /dev/i2c-2   dtoverlay=i2c2-pi5   (Pi5 only)
//   i2c3 GPIO6/7  /dev/i2c-3   dtoverlay=i2c3-pi5   (Pi5 only)
//   i2c >ts
//   true t:.enabled!
//   // ( addr:i data:ByteList -- )           .write
//   // ( addr:i n:i -- data:ByteList )       .read
//   // ( addr:i w:ByteList n:i -- r:ByteList ) .tx   // write + repeated-start read
//   ts>
//   Device chips (e.g. modules/pi/drivers/mpu9250) wrap these with register helpers.
//
// SPI:
//   spi0 MOSI10 MISO9 SCLK11 CE0=8 CE1=7  /dev/spidev0.0|.1  numCS 1..2  dtparam=spi=on
//   spi1 MOSI20 MISO19 SCLK21 CE0=18 CE1=17 CE2=16  spidev1.*  numCS 1..3  (Pi5 only)
//   spi5 MOSI14 MISO13 SCLK15 CE0=12 CE1=26  spidev5.0|.1  numCS 1..2  (Pi5 only)
//        (CE0/data clash with pwm0–3 / uart0 on 12–15; CE1=26 when numCS=2)
//   spi >ts
//   SpiPeriph.mode0 t:.mode!  SpiPeriph.bitsPerWord8 t:.bitsPerWord!
//   1000000 t:.maxHz!   // free Int; driver rounds — not an enum
//   2 t:.numCS!         // before enabled; claims CE1.. via optPins (Validate)
//   true t:.enabled!    // opens spidevB.0 .. B.(numCS-1) eagerly
//   0 t:.selCS!         // modal CE for following transfers (default 0)
//   // ByteList -- ByteList  .transfer   (uses selCS)
//   ts>
//
// I2S stub (ALSA-owned header audio; Final does not drive audio):
//   true i2s.reserved!   // before GpioCtrl.inst.reset^ if using an ALSA audio HAT (BCM 18–21)
//
// Optional: FINAL_GPIOCHIP=… / FINAL_UART0=… / FINAL_UARTn=… / FINAL_I2Cn=… /
//   FINAL_SPIn=… / FINAL_PWMCHIP=… / FINAL_PI_SUPPORT=none|pi2z|pi5
// User must be in the gpio/i2c/spi/dialout groups (or root) for devices.

GpioCtrl.piSupport _piSupport !

/*
The alternative peripherals are only supported on Pi 5.
Panic otherwise.
( -- ) error
*/
: assertPi5
  _piSupport @ PiSupport.Pi5 <>
  if
    "Not supported on this Pi version. Use the default peripherals only."
    error
  then
;

// Lazy names: first use builds GpioCtrl; park on a controller field to fetch once.
// Default peripherals (the ones on Header maps) also have an unnumbered alias, preferred.

: pwm0  GpioCtrl.inst.pwm0 ;
: pwm1  GpioCtrl.inst.pwm1 ;
: uart  GpioCtrl.inst.uart0 ;
: uart0 GpioCtrl.inst.uart0 ;
: uart2 assertPi5 GpioCtrl.inst.uart2 ;
: uart3 assertPi5 GpioCtrl.inst.uart3 ;
: i2c   GpioCtrl.inst.i2c1 ;
: i2c1  GpioCtrl.inst.i2c1 ;
: i2c2  assertPi5 GpioCtrl.inst.i2c2 ;
: i2c3  assertPi5 GpioCtrl.inst.i2c3 ;
: spi   GpioCtrl.inst.spi0 ;
: spi0  GpioCtrl.inst.spi0 ;
: spi1  assertPi5 GpioCtrl.inst.spi1 ;
: spi5  assertPi5 GpioCtrl.inst.spi5 ;
: i2s   GpioCtrl.inst.i2s ;

: gpio2  GpioCtrl.inst.gpio2 ;
: gpio3  GpioCtrl.inst.gpio3 ;
: gpio4  GpioCtrl.inst.gpio4 ;
: gpio5  GpioCtrl.inst.gpio5 ;
: gpio6  GpioCtrl.inst.gpio6 ;
: gpio7  GpioCtrl.inst.gpio7 ;
: gpio8  GpioCtrl.inst.gpio8 ;
: gpio9  GpioCtrl.inst.gpio9 ;
: gpio10 GpioCtrl.inst.gpio10 ;
: gpio11 GpioCtrl.inst.gpio11 ;
: gpio12 GpioCtrl.inst.gpio12 ;
: gpio13 GpioCtrl.inst.gpio13 ;
: gpio14 GpioCtrl.inst.gpio14 ;
: gpio15 GpioCtrl.inst.gpio15 ;
: gpio16 GpioCtrl.inst.gpio16 ;
: gpio17 GpioCtrl.inst.gpio17 ;
: gpio18 GpioCtrl.inst.gpio18 ;
: gpio19 GpioCtrl.inst.gpio19 ;
: gpio20 GpioCtrl.inst.gpio20 ;
: gpio21 GpioCtrl.inst.gpio21 ;
: gpio22 GpioCtrl.inst.gpio22 ;
: gpio23 GpioCtrl.inst.gpio23 ;
: gpio24 GpioCtrl.inst.gpio24 ;
: gpio25 GpioCtrl.inst.gpio25 ;
: gpio26 GpioCtrl.inst.gpio26 ;
: gpio27 GpioCtrl.inst.gpio27 ;
