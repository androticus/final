# I2S (audio coexistence)

Final does **not** drive I2S audio. When a digital I2S sound HAT uses BCM **18–21**, Linux usually owns those pins via pinmux.

## Prefer OS visibility

`GpioCtrl.reset` **skips** restoring any header pin the kernel has mux-claimed (see `osPinMap` / `osMapped`). With `dtparam=i2s=on` or a HAT overlay, BCM 18–21 are left alone automatically — you do **not** need to set `.reserved` first.

```final
GpioCtrl.inst.osPinMap @   // StringList by BCM; get re-reads pinmux
gpio18.osMapped @          // "" if free, else e.g. an i2s label
i2s.osMapped @             // "i2s" when req pins look mapped; else ""
GpioCtrl.inst.reset^
```

Write a `pinMap` dump word yourself if you want a printed table from one `osPinMap @`.

## `.reserved` (escape hatch)

If auto-detect misses your setup (debugfs missing, odd overlay), keep the stopgap:

1. Set `true i2s.reserved!` **before** `GpioCtrl.inst.reset^`.
2. Reset skips restoring BCM 18–21.
3. Validate still claims those pins as owner `i2s`, so Final cannot take them via `gpio18`–`gpio21` or overlapping periphs such as `spi1`.
4. `i2s.reserved` is **not** cleared by `reset`.

```final
true i2s.reserved!
GpioCtrl.inst.reset^
```

There is no `.enabled` and no audio datapath on `I2sPeriph`.

## Conflicts

Reserved or OS-mapped `i2s` overlaps SPI1 data/CE0 (18–21). Do not enable `spi1` (or those GPIO pins) while I2S owns them.

See also [Pins and Periphs](pins-periphs.md) and the [Pin Map](pin-map.md).
