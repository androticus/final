# UARTs

UARTs provide asynchronous serial I/O over TX/RX. Final’s type is `UartPeriph`. Prefer **`uart`** (alias of `uart0`). Pi 5 also has `uart2` / `uart3`. Optional RS-232 hardware flow control uses CTS/RTS when `.flowHW` is `true`. See the [Pin Map](pin-map.md).

Set transmission properties **before** enabling. Changing them while active is not allowed.

## Properties (set before enable)

Property | Values / default | Description
--- | --- | ---
baud | > 0, def 115200 | bit rate
dataBits | 5..8, def 8 | data bits
stopBits | 1 or 2, def 1 | stop bits
parity | `UartPeriph.parityNone` (def), `parityOdd`, `parityEven` | parity
flowHW | bool, def false | if true, claims CTS/RTS pins
enabled | bool | open the port and start RX background read
path | readonly | device path (Pi 5: `/dev/ttyAMA0`; Zero 2: `/dev/ttyS0`; override with `FINAL_UART0`)

Final picks the header node from the board profile (Zero 2 mini UART vs Pi 5 `uart0-pi5`). Do **not** use `/dev/serial0` on Pi 5 — that is the debug JST (`ttyAMA10`), not the 40-pin header. `FINAL_UART0` is only for unusual overlays (e.g. PL011 freed onto the header).

**Baud on Zero 2:** mini UART silently ignores rates below **1200** (termios appears to accept them; the wire stays at the prior rate). Final rejects `baud!` &lt; 1200 on `/dev/ttyS*`. Pi 5 PL011 accepts custom rates such as 250.

```final
  115200 uart.baud!
  8 uart.dataBits!
  1 uart.stopBits!
  UartPeriph.parityNone uart.parity!
  false uart.flowHW!
  true uart.enabled!
```

## Write

Two TX paths (timeout on both = `nBytes × frameBits / baud × 1.15`):

| Method | Stack / completion | On-wire drain |
| --- | --- | --- |
| `.write` | `( data -- true \| err:s false )` after OS accept | no |
| `.writeNotify` | `( data -- )`; Event via `setWriteNotify` | yes (`tcdrain` before notify) |

Both accept an `Int` (one byte) or a `ByteList`. Immediate problems (UART not enabled, bad argument, `writeNotify` without a handler) **panic**. TX timeout / OS errors: modal `false` on `.write`; Event `params[0]=false` (+ `params[1]` err string) on `.writeNotify`. Prefer `.writeNotify` (or a pip) when the EventLoop task must not wait on serial.

```final
  0x41 uart.write^ if … then          // 'A'; check modal ok
  myBytes uart.write^ nif "tx fail" $ then

  // TX finished on the wire:
  myHandler~ 0 uart.setWriteNotify^
  myBytes uart.writeNotify^
  // handler: ( e:Event -- )  0 e.params@() = ok; optional 1 e.params@() = err
```

## Read when ready (async)

Prime-then-notify: arm with `.readWhenReady^` after `.setReadReady^`. When the RX queue has data (already buffered or later), Final gets an `onReadReady` Event once, then disarms. In the handler, use `.newReader^` / `foreach` as usual; re-arm in the handler tail if you want another batch. A second `.readWhenReady^` while armed is a nop (modal `true`). Disable / `reset` clears armed state and both notify handlers.

```final
  myOnReady~ 0 uart.setReadReady^
  uart.readWhenReady^ nif "arm failed" error then
  // handler: ( e:Event -- ) then uart.newReader^ …
```

## Read with UartReader

Reading needs a byte budget, idle timeout, and optional delimiter. Use `UartReader` from `.newReader^`, set its properties, then `foreach`. The loop ends on timeout, max bytes, delimiter, or error. Check the readonly status fields afterward. Prefer a pip or `readWhenReady` so the EventLoop task does not block in `foreach`.

Property | R/W | Description
--- | --- | ---
maxN | R/W | max bytes this session; `0` = unlimited
timeout | R/W | idle ms between bytes (def 100)
untilByte | R/W | stop after this byte (0..255), or `-1` for none
n | R | bytes delivered
timedOut | R | true if idle timeout ended the loop
hitDelim | R | true if `untilByte` was received
ok | R | false on hard error
err | R | error string when not ok

## Example

Simple write + read-until-newline on the default UART:

```final
  115200 uart.baud!
  true uart.enabled!

  ByteList.new^ >xs
  "hello\n" foreach i x:.append^ /foreach
  xs> uart.write^ nif "write failed" error then

  uart.newReader^ >ts
  1000 t:.timeout!
  0x0A t:.untilByte!          // '\n'
  ByteList.new^ >xs
  t:peek foreach
    i x:.append^
  /foreach
  t:.hitDelim@
  t:.timedOut@
  tdrop
  x:drop
```

Pi 5 HAL tests also loopback `uart` ↔ `uart2` (requires `uart2-pi5`). There is no UART demo under `examples/pi/` yet.
