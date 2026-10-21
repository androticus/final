forget: reset-i2c-ssd1306-demo
requiresPi

// install and load the life module, which declares LifeCore
"github.com/androticus/final/modules/simgame/life"
{
  "modules/simgame/life" "path"
}  
install

"" "github.com/androticus/final/modules/simgame/life" import

`
I2C demo: UCTRONICS 128x64 yellow/blue OLED (SSD1306).

Wiring (default i2c / I2C1):
  VCC 3.3V or 5V, GND, SDA GPIO2, SCL GPIO3
  7-bit address 0x3C (usually)

Host: enable the bus first, e.g.
  dtparam=i2c_arm=on
  or dtoverlay=i2c1-pi5 (Pi 5)

Load:
  "examples/pi/i2c-ssd1306-life.fnl" compileFromFile

Then:
  i2c-ssd1306-demo

Shows Conway's Game of Life, animated.
Uses EventAgain (no-delay re-arm) + EventEscFn (Esc to stop).
` const: i2c-ssd1306-blurb

{
  "Ssd1306" "name"
  {
    I2cPeriph "bus"
    Int "addr"
  } "fields"
  {
    0x3C "defaultAddr"
    128 "width"
    64 "height"
    0x00 "ctrlCmd"
    0x40 "ctrlData"
  } "constants"
}
TypedObject.makeType^

{
  "I2cSsd1306Demo" "name"
  {
    Ssd1306 "oled"
    LifeCore "life"
    EventAgain "again"
    EventEscFn "escFn"
    Bool "stop"
  } "fields"
}
TypedObject.makeType^

// instance to prevent GC
null var: I2cSsd1306Demo_inst

/*
Write one command byte (control 0x00).
( c:i -- ) o( oled:Ssd1306 )
*/
: Ssd1306#writeCmd
  ByteList.new^ >xs
  Ssd1306.ctrlCmd x:.append^
  x:.append^
  o:.bus@ >ts
  o:.addr@ xs> t:.write^
  tdrop
;

/*
Address the full 128x64 GDDRAM window.
( -- ) o( oled:Ssd1306 )
*/
: Ssd1306#setFullWindow
  0x21 o:.writeCmd^
  0 o:.writeCmd^
  127 o:.writeCmd^
  0x22 o:.writeCmd^
  0 o:.writeCmd^
  7 o:.writeCmd^
;

/*
Standard 128x64 SWITCHCAP init (Adafruit/Arducam sequence).
( -- ) o( oled:Ssd1306 )
*/
: Ssd1306#begin
  0xAE o:.writeCmd^                 // display off
  0xD5 o:.writeCmd^ 0x80 o:.writeCmd^
  0xA8 o:.writeCmd^ 0x3F o:.writeCmd^
  0xD3 o:.writeCmd^ 0x00 o:.writeCmd^
  0x40 o:.writeCmd^
  0x8D o:.writeCmd^ 0x14 o:.writeCmd^ // charge pump on
  0x20 o:.writeCmd^ 0x00 o:.writeCmd^ // horizontal addressing
  0xA1 o:.writeCmd^                 // seg remap
  0xC8 o:.writeCmd^                 // com scan dec
  0xDA o:.writeCmd^ 0x12 o:.writeCmd^
  0x81 o:.writeCmd^ 0xCF o:.writeCmd^
  0xD9 o:.writeCmd^ 0xF1 o:.writeCmd^
  0xDB o:.writeCmd^ 0x40 o:.writeCmd^
  0xA4 o:.writeCmd^
  0xA6 o:.writeCmd^                 // normal (not inverted)
  0xAF o:.writeCmd^                 // display on
;

/*
Fill GDDRAM with one repeating byte (0x00 clear, 0xFF all on).
( v:i -- ) o( oled:Ssd1306 )
*/
: Ssd1306#fill
  >ys
  o:.setFullWindow^
  ByteList.new^ >xs
  Ssd1306.ctrlData x:.append^
  0 1023 range
  foreach
    y:peek x:.append^
  /foreach
  o:.bus@ >ts
  o:.addr@ xs> t:.write^
  tdrop
  y:drop
;

/*
Pack a 64×128 Mat2i1 (row=y, col=x) into SSD1306 page GDDRAM and write.
( mat:Mat2i1 -- ) o( oled:Ssd1306 )
*/
: Ssd1306#showMat
  >ts
  o:.setFullWindow^
  ByteList.new^ >xs
  Ssd1306.ctrlData x:.append^
  0 7 range
  foreach
    i >ys                 // page
    0 127 range
    foreach
      i >zs               // x
      0
      0 7 range
      foreach
        y:peek 8 * i + z:peek t:@[] <>0 if
          1 i << |
        then
      /foreach
      x:.append^
      z:drop
    /foreach
    y:drop
  /foreach
  o:.bus@ >ts
  o:.addr@ xs> t:.write^
  tdrop
  tdrop
;

/*
Mint bound to a bus (default 0x3C).
( bus:I2cPeriph -- oled:Ssd1306 ) o( type:Ssd1306 )
*/
: Ssd1306##onBus
  o:.new^ >ts
  t:.bus!
  Ssd1306.defaultAddr t:.addr!
  ts>
;

/*
Paint current Life mat0 onto the panel.
( -- ) o( inst:I2cSsd1306Demo )
*/
: I2cSsd1306Demo_paint
  o:.life.mat0
  o:.oled.showMat^
;

/*
Stop EventAgain / EscFn and release hardware.
( -- ) o( inst:I2cSsd1306Demo )
*/
: I2cSsd1306Demo_stop
  "Stopping SSD1306 Life demo...\n" $
  o:.again.close^
  o:.escFn.close^
  GpioCtrl.inst.reset^
  null I2cSsd1306Demo_inst !
;

/*
Esc stops the demo (Fn keys ignored).
( code:i -- ) o( inst:I2cSsd1306Demo )
*/
: I2cSsd1306Demo#handleEscFn
  ==0 if
    true o:.stop!
  then
;

/*
One Life generation + paint; re-arm EventAgain.
( -- ) o( inst:I2cSsd1306Demo )
*/
: I2cSsd1306Demo#handleFrame
  o:.stop if
    I2cSsd1306Demo_stop
  else
    o:.life.nextGen^
    I2cSsd1306Demo_paint
    o:.again.again^
  then
;

/*
( -- ) o( inst:I2cSsd1306Demo )
*/
: I2cSsd1306Demo#init
  :s initPi
    true i2c.enabled!
    i2c Ssd1306.onBus^ o:.oled!
    LifeCore.new^ o:.life!
    64  o:.life.rows!
    128 o:.life.cols!
    4 o:.life.threads!
    o:.life.init^
    o:.oled.begin^
  ;s

  :s initSelf
    o:.handleFrame~ EventAgain.new^ o:.again!
    o:.handleEscFn~ EventEscFn.new^ o:.escFn!
    o:.escFn.start^
    o:.again.again^ // kick first frame
  ;s

  GpioCtrl.inst.reset^
  initPi
  initSelf
  "SSD1306 Life on i2c @ 0x3C — Esc to stop.\n" $
;

: i2c-ssd1306-demo
  I2cSsd1306Demo.new^ dup >ts
  I2cSsd1306Demo_inst !
  t:.init^
  tdrop
;

i2c-ssd1306-blurb $ nl
