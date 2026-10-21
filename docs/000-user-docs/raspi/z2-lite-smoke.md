# Pi Zero 2 Lite smoke checklist

Use this on a **Raspberry Pi Zero 2 W** running **Pi OS 64-bit Lite** after installing the Pi build of Final (`final-pi`). Pi 5–only alt buses are expected to refuse; defaults should work once overlays and groups are set.

## Groups

```bash
sudo usermod -aG gpio,i2c,spi,dialout "$USER"
# log out / in, then:
groups   # confirm
```

## Overlays (`/boot/firmware/config.txt`)

Defaults only (no `*-pi5` alt overlays):

```
dtparam=i2c_arm=on
dtparam=spi=on
dtoverlay=pwm-2chan,pin=12,pin2=13,func=4,func2=4
enable_uart=1
```

Final’s Zero 2 profile defaults `uart` to `/dev/ttyS0` (header mini UART). Optional overrides: `FINAL_GPIOCHIP`, `FINAL_PWMCHIP`, `FINAL_UART0`, `FINAL_PI_SUPPORT=pi2z` (force profile in CI).

## Profile

Confirm you are on Zero 2 (device-tree model / Imager image). Alt buses are not supported here — `uart2` / `uart3` / `i2c2` / `i2c3` / `spi1` / `spi5` should error if used.

## Smoke

1. **GPIO** — load `examples/pi/hello-led.fnl` (or blink `gpio26` out).
2. **I2C** — with a sensor on the default bus: `true i2c.enabled!` (see `examples/pi/i2c-9250-9-dof.fnl`).
3. **SPI** — with panel wiring: `examples/pi/spi-jd79661-eink.fnl` uses `spi`.
4. **PWM** — enable `pwm0` after `pwm-2chan` (raw period/duty, then `.enabled!`).
5. **UART** — `true uart.enabled!` (path should be `/dev/ttyS0`; use baud ≥ 1200).

If anything fails, check `dmesg`, `/dev/i2c-*` / `/dev/spidev*` / `/sys/class/pwm/`, and group membership before chasing Final.
