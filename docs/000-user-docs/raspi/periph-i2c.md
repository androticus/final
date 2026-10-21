# I2C

I2C is a common two-wire bus for sensors and small displays. The lines are **SDA** (data) and **SCL** (clock). Final’s type is `I2cPeriph`. Prefer **`i2c`** (alias of `i2c1`). Pi 5 also has `i2c2` / `i2c3`. See the [Pin Map](pin-map.md).

`I2cPeriph` has no setup properties. Enable the bus, then call `write` / `read` / `tx` with a **7-bit device address** on each call (the address is not stored on the bus object).

Host OS: enable the bus in config (e.g. `dtparam=i2c_arm=on`, or `dtoverlay=i2c1-pi5` on Pi 5) and be in the `i2c` group.

```final
  true i2c.enabled!
```

## API

Property | R/W | Description
--- | --- | ---
enabled | R/W | set `true` after wiring is ready; claims SDA/SCL
bus | R | bus number (e.g. 1 for `i2c`)
path | R | device path, e.g. `/dev/i2c-1`

Method | Stack | Description
--- | --- | ---
write | `( addr:i data:ByteList -- )` | write bytes to device
read | `( addr:i n:i -- data:ByteList )` | read `n` bytes
tx | `( addr:i w:ByteList nRead:i -- data:ByteList )` | write then repeated-start read

## Template

Typical register access (write one register, or write register index then read):

```final
  // write reg, value
  ByteList.new^ >xs
  0x6B x:.append^   // reg
  0    x:.append^   // value
  0x68 xs> i2c.write^

  // read 1 byte from reg (repeated start)
  ByteList.new^ >xs
  0x75 x:.append^
  0x68 xs> 1 i2c.tx^ >ys
  0 y:@()            // value
  y:drop
```

## Drivers and demos

Copy/adapt these rather than inventing bus glue from scratch:

| Role | File |
| --- | --- |
| MPU-9250 + AK8963 driver | `modules/pi/drivers/mpu9250` |
| 9-DoF demo (loads the driver) | `examples/pi/i2c-9250-9-dof.fnl` |
| SSD1306 OLED (driver + demo) | `examples/pi/i2c-ssd1306-oled.fnl` |

Pattern in those files: a small TypedObject holds `bus` + `addr`, with helpers like `writeReg` / `readReg` that call `.write^` / `.tx^` on `i2c`.
