# Getting Ready

## Final Features used

Make sure you familiarize yourself with these Final features. We use them in the Pi examples:

* Stacks, and basic stack operations
* Objects
* Compound Object syntax
* DynamicObject
* Paren object constructors
* TypedObject

For more advanced controller designs, it is helpful to be familiar with these features:

* Events and the EventLoop
* EventTimer and EventTicker for creating time-based events in your controller
* EventEscFn for ease of stopping a controller, or selecting options

## Enabling Peripherals On Your Pi

Some of the breakout functions of the Pi are not enbled by default in the Pi OS. The pins themselves are available, but the following needs to be done to enable I2C, SPI, and PWM. You make these changes in `/boot/firmware/config.txt`. BE CAREFUL! The `nano` editor is intuitive for this purpose; it must be launched with `sudo` permission, as follows:

```bash
sudo nano /boot/firmware/config.txt
```

### I2C and SPI

Make sure you enable i2c and spi (uncomment.) Only uncomment i2s if you are going to use an I2S audio dongle or hat. `GpioCtrl.reset` skips pins the kernel has mux-claimed (`osPinMap` / `osMapped`); use `true i2s.reserved!` only as an escape hatch if that miss — see [I2S](periph-i2s.md).

The following enable the header defaults (`i2c` / `spi`) on both Pi 5 and Zero 2:
```
dtparam=i2c_arm=on
#dtparam=i2s=on
dtparam=spi=on
```

On **Pi 5 only**, to enable alt buses Final supports, *add* lines like these (not available on Zero 2):

```
# I2C alts — Pi 5 only; only include ones you use
dtoverlay=i2c2-pi5
dtoverlay=i2c3-pi5

# SPI alts — Pi 5 only (spi1-Ncs / spi5-1cs|2cs)
dtoverlay=spi1-1cs
dtoverlay=spi5-1cs
```

### UARTs

Prefer the unnumbered `uart` (header TX/RX). Enable only the UARTs whose pins you will use.

**Pi 5** header UART + optional loopback partner used in HAL tests:
```
dtoverlay=uart0-pi5
# optional Pi5-only alt for loopback tests:
# dtoverlay=uart2-pi5
```

**Pi Zero 2:** `enable_uart=1`. Final defaults `uart` to `/dev/ttyS0` on Zero 2 (see [UART](periph-uarts.md) and [Z2 Lite smoke](z2-lite-smoke.md)).

### PWM

Don't enable PWM overlays if you aren't using PWM. 

```
# Header PWM on BCM 12/13 (pwm0/pwm1) — both Pi 5 and Zero 2
# func=4 (Alt0) is required on 12/13; overlay defaults are Alt5 (for 18/19).
dtoverlay=pwm-2chan,pin=12,pin2=13,func=4,func2=4
```


## Pi Prototyping Hardware

You should already have a breakout cable and a T-header for a breadboard. You can get this from CanaKit or on Amazon. Several of the examples use the following:
* a red and a green LED, blue for the TX pwm demo
* 3x 220 or 330 ohm resistors for the LEDs
* 2 breadboard compatible push buttons.

We also have an example of using:
* a Parallax hobby servo
* a 9250 9-dof accelerometer/gyro/compass
* an SSD1306-class OLED display
* a JD79661 AdaFruit 4-color e-ink display

## Conventions Used

We talk exclusively about GPIO pins, not header pins. GPIO is the conceptual pin number, from 0 to 27. These are mapped in a kind of haphazard way to the 40-pin header on the Pi. Most Pi boards come with a card showing how the GPIO pins map to the header pins. To avoid any confusion, we never reference the header pins. So just make sure if you follow any of our examples, that you are using the GPIO pin, not the header pin number.