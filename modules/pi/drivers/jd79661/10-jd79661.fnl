forget: reset-jd79661
// NOTE modules generally should not have a forget TODO: remove this
requiresPi
/*
Adafruit JD79661 4-color e-ink (2.13" 250x122 breakout) over SpiPeriph.

Oracle: Adafruit_CircuitPython_EPD adafruit_epd/jd79661.py
Native panel RAM is 128x250, 2 bits/pixel (4 pixels/byte), MSB-first:
  0 black, 1 white, 2 yellow, 3 red
  fill bytes: 0x00 / 0x55 / 0xAA / 0xFF

BUSY is idle HIGH (wait while LOW), matching CircuitPython displayio busy_state=False.

Adj pins are ordinary GpioPins (not on SpiPeriph). ECS → spi CE0; SRCS unused
when buffering in host RAM.
*/

{
  "Jd79661" "name"
  {
    SpiPeriph "spi"
    GpioPin "dc"
    GpioPin "rst"
    GpioPin "busy"
  } "fields"
  {
    128 "width"
    250 "height"
    0 "black"
    1 "white"
    2 "yellow"
    3 "red"
    0x10 "cmdData"
    0x12 "cmdRefresh"
    0x04 "cmdPowerOn"
    0x02 "cmdPowerOff"
    0x07 "cmdDeepSleep"
  } "constants"
}
TypedObject.makeType^

/*
Mint bound to SPI + adj pins. Stack: spi, dc, rst, busy (busy TOS).
( spi:SpiPeriph dc:GpioPin rst:GpioPin busy:GpioPin -- panel:Jd79661 ) o( type:Jd79661 )
*/
: Jd79661##connect
  o:.new^ >ts
  t:.busy!
  t:.rst!
  t:.dc!
  t:.spi!
  ts>
;

/*
( c:i -- ) o( panel:Jd79661 )
*/
: Jd79661#writeCmd
  false o:.dc.value!
  ByteList.new^ >xs
  x:.append^
  xs> o:.spi.write^
;

/*
( data:ByteList -- ) o( panel:Jd79661 )
*/
: Jd79661#writeData
  true o:.dc.value!
  o:.spi.write^
;

/*
( v:i -- ) o( panel:Jd79661 )
*/
: Jd79661#writeData1
  ByteList.new^ >xs
  x:.append^
  xs> o:.writeData^
;

/*
Hardware reset pulse.
( -- ) o( panel:Jd79661 )
*/
: Jd79661#hwReset
  false o:.rst.value!
  10 delayMs
  true o:.rst.value!
  10 delayMs
;

/*
Wait until BUSY is HIGH (idle), or ~30s timeout.
( -- ) o( panel:Jd79661 )
*/
: Jd79661#busyWait
  0 >ys
  begin
    10 delayMs
    ys> 1 + >ys
    o:.busy.value@ if
      true
    else
      y:peek 3000 >
    then
  until
  ys> 3000 > if
    "Jd79661: busy timeout (check BUSY wiring/polarity)\n" $
  then
;

/*
Power-up + init sequence (Adafruit_JD79661.power_up).
( -- ) o( panel:Jd79661 )
*/
: Jd79661#powerUp
  o:.hwReset^
  o:.busyWait^
  10 delayMs

  0x4D o:.writeCmd^  0x78 o:.writeData1^
  0x00 o:.writeCmd^  0x8F o:.writeData1^  0x29 o:.writeData1^
  0x01 o:.writeCmd^  0x07 o:.writeData1^  0x00 o:.writeData1^
  0x03 o:.writeCmd^  0x10 o:.writeData1^  0x54 o:.writeData1^  0x44 o:.writeData1^
  0x06 o:.writeCmd^
    0x05 o:.writeData1^  0x00 o:.writeData1^  0x3F o:.writeData1^
    0x0A o:.writeData1^  0x25 o:.writeData1^  0x12 o:.writeData1^  0x1A o:.writeData1^
  0x50 o:.writeCmd^  0x37 o:.writeData1^
  0x60 o:.writeCmd^  0x02 o:.writeData1^  0x02 o:.writeData1^
  0x61 o:.writeCmd^  0 o:.writeData1^  128 o:.writeData1^  0 o:.writeData1^  250 o:.writeData1^
  0xE7 o:.writeCmd^  0x1C o:.writeData1^
  0xE3 o:.writeCmd^  0x22 o:.writeData1^
  0xB4 o:.writeCmd^  0xD0 o:.writeData1^
  0xB5 o:.writeCmd^  0x03 o:.writeData1^
  0xE9 o:.writeCmd^  0x01 o:.writeData1^
  0x30 o:.writeCmd^  0x08 o:.writeData1^
  Jd79661.cmdPowerOn o:.writeCmd^
  o:.busyWait^
;

/*
Power off + deep sleep (image remains on panel).
( -- ) o( panel:Jd79661 )
*/
: Jd79661#powerDown
  Jd79661.cmdPowerOff o:.writeCmd^  0 o:.writeData1^
  o:.busyWait^
  Jd79661.cmdDeepSleep o:.writeCmd^  0xA5 o:.writeData1^
  100 delayMs
;

/*
Append n copies of byte v onto xs (parked ByteList).
( v:i n:i -- ) x( bl:ByteList )
*/
: Jd79661_appendN
  >ys                      // n
  >zs                      // v
  0 y:peek 1 - range
  foreach
    z:peek x:.append^
  /foreach
  z:drop
  y:drop
;

/*
Upload one full-panel Mat2i4 and refresh.

The mat is the drawing surface: rows = Jd79661.width (128), cols =
Jd79661.height (250). Minipaint (x,y) is (col,row): 250 wide × 128 tall,
Adafruit rotation 3. Native RAM is 128×250, 4 pixels/byte, MSB first:
  nativeX = row, nativeY = 249 − col.

Indexes are panel colors: 0 black, 1 white, 2 yellow, 3 red.

( mat:Mat2i4 -- ) o( panel:Jd79661 )
*/
: Jd79661#showMat
  >ts
  o:.powerUp^
  Jd79661.cmdData o:.writeCmd^
  ByteList.new^ >xs
  0 Jd79661.height 1 - range
  foreach
    i >ys
    0 31 range
    foreach
      i 4 * >zs
      0
      0 3 range
      foreach
        2 <<
        z:peek i +
        249 y:peek -
        t:@[] 3 &
        |
      /foreach
      x:.append^
      z:drop
    /foreach
    y:drop
  /foreach
  xs> o:.writeData^
  Jd79661.cmdRefresh o:.writeCmd^
  0 o:.writeData1^
  o:.busyWait^
  o:.powerDown^
  tdrop
;

/*
Full-frame vertical stripes: black | white | yellow | red (32px each @ 128 wide).
( -- ) o( panel:Jd79661 )
*/
: Jd79661#showVerticalStripes
  o:.powerUp^
  Jd79661.cmdData o:.writeCmd^
  ByteList.new^ >xs
  0 Jd79661.height 1 - range
  foreach
    0x00 8 Jd79661_appendN
    0x55 8 Jd79661_appendN
    0xAA 8 Jd79661_appendN
    0xFF 8 Jd79661_appendN
  /foreach
  xs> o:.writeData^
  Jd79661.cmdRefresh o:.writeCmd^
  0 o:.writeData1^
  o:.busyWait^
  o:.powerDown^
;
