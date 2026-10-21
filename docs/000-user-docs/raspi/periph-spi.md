# SPI

SPI is a faster, clocked bus used for displays, storage, and similar devices. Final’s type is `SpiPeriph`. Prefer **`spi`** (alias of `spi0`). Pi 5 also has `spi1` / `spi5`. See the [Pin Map](pin-map.md).

Each SPI device on a shared bus needs its own chip-select (CS / CE). By default a `SpiPeriph` claims **CE0** only. Set `numCS` before enabling if you need more (e.g. `2` for CE0+CE1). Extra CE lines are optional pins; unused ones stay available as GPIO.

Host OS: enable SPI (e.g. `dtparam=spi=on`) and be in the `spi` group (or root).

Set mode / word size / speed / `numCS` **before** `enabled!`.

```final
  SpiPeriph.mode0 spi.mode!
  SpiPeriph.bitsPerWord8 spi.bitsPerWord!
  1000000 spi.maxHz!
  1 spi.numCS!
  0 spi.selCS!
  true spi.enabled!
```

## API

Property | R/W | Values / default | Description
--- | --- | --- | ---
enabled | R/W | bool | claim the bus (and CE0..numCS-1)
path | R | string | CE0 device path, e.g. `/dev/spidev0.0`
mode | R/W | `SpiPeriph.mode0`..`mode3` (def 0) | SPI mode; set before enable
bitsPerWord | R/W | `bitsPerWord8` or `bitsPerWord16` (def 8) | set before enable
maxHz | R/W | > 0 (def 1_000_000) | clock rate; set before enable
numCS | R/W | 1..max for that bus (def 1) | how many CE lines; set before enable
selCS | R/W | 0..numCS-1 (def 0) | which CE the next transfer uses

Method | Stack | Description
--- | --- | ---
write | `( data:ByteList -- )` | clock out bytes (RX discarded)
read | `( n:i -- data:ByteList )` | clock in `n` bytes (TX zeros)
transfer | `( tx:ByteList -- rx:ByteList )` | full-duplex transfer

Many panels also need plain GPIO for DC, RST, BUSY, and so on. Those are normal `GpioPin`s, not part of `SpiPeriph`.

## Template

```final
  // adj pins (example numbers — match your wiring)
  GpioPin.out gpio23.direction!   // DC
  false gpio23.value!
  true gpio23.enabled!

  GpioPin.out gpio24.direction!   // RST
  true gpio24.value!
  true gpio24.enabled!

  SpiPeriph.mode0 spi.mode!
  SpiPeriph.bitsPerWord8 spi.bitsPerWord!
  1000000 spi.maxHz!
  true spi.enabled!

  // command byte (DC low)
  false gpio23.value!
  ByteList.new^ >xs
  0x10 x:.append^
  xs> spi.write^

  // data (DC high)
  true gpio23.value!
  myBytes spi.write^
```

## Drivers and demos

| Role | File |
| --- | --- |
| JD79661 e-ink driver | `modules/pi/drivers/jd79661` |
| Title and shapes demo | `examples/pi/spi-jd79661-eink.fnl` |

The driver binds `SpiPeriph` plus DC/RST/BUSY pins (`Jd79661.connect^`), then uses `spi.write^` with DC selecting command vs data. Start from that pair when adding another SPI panel.
