forget: reset-uart-tx-led
`
This demo confirms the uart peripheral is enabled and working, by
pulse-width modulating an LED with its data bits.
TX is not-polarity, so High is Idle.
Wire an LED from 3.3V through a 220R or 330R resistor to GPIO14 (TX).

Baud 1200, 8N1: bytes 0..119 takes ~1s on the wire.
(Note: Zero 2 mini UART rejects baud < 1200).

You should see the LED brighten as more 0-bits are transmitted, then
idle low (off) when the port closes.

Try:

uart-tx-led
`
const: uart-tx-led-blurb


: uart-tx-led
  GpioCtrl.inst.reset^

  uart >ts
  1200 t:.baud!
  8 t:.dataBits!
  1 t:.stopBits!
  0 t:.parity!
  false t:.flowHW!
  true t:.enabled!
  0xFF // all 1s puts TX low for the data bits
  0 7 range foreach
    0 119 range foreach // about 1s step in brightness
      dup  t:.write^
      nif "A write error occurred: " $ $ nl break then
    /foreach
    1 << // shift another 0-bit into the data
  /foreach
  drop
  tdrop

  // OS accepts the writes all at once
  // so we wait, blocking
  10000 delayMs
  "Uart demo finished.\n" $
  GpioCtrl.inst.reset^
;

uart-tx-led-blurb $ nl
