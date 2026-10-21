forget: reset-i2c-ssd1306-demo
requiresPi
`
I2C demo: UCTRONICS 128x64 yellow/blue OLED (SSD1306).

Wiring (default i2c / I2C1):
  VCC 3.3V or 5V, GND, SDA GPIO2, SCL GPIO3
  7-bit address 0x3C (usually)

Host: enable the bus first, e.g.
  dtparam=i2c_arm=on
  or dtoverlay=i2c1-pi5 (Pi 5)

Load:
  "examples/pi/i2c-ssd1306-oled.fnl" compileFromFile

Then:
  i2c-ssd1306-demo

Minipaint scene: "Minipaint!" in a rounded rect (top),
then triangle / rectangle / circle across the lower band.
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
  } "fields"
}
TypedObject.makeType^

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
Vertical stripe pattern (prove data path).
( -- ) o( oled:Ssd1306 )
*/
: Ssd1306#showChecker
  o:.setFullWindow^
  ByteList.new^ >xs
  Ssd1306.ctrlData x:.append^
  0 7 range
  foreach
    0 127 range
    foreach
      i 8 / 2 % <>0 if
        0xFF
      else
        0x00
      then
      x:.append^
    /foreach
  /foreach
  o:.bus@ >ts
  o:.addr@ xs> t:.write^
  tdrop
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
Paint the Minipaint demo scene onto a 64×128 1-bit mat and push to the panel.
( -- ) o( inst:I2cSsd1306Demo )
*/
: I2cSsd1306Demo_paint
  64 128 Mat2i1.new^ >ts
  Color.black Color.white Palette1.newWithColors^ t:peek BlitSrc.new1^
  Minipaint.new^ >xs
  0 x:.ink!
  x:.fillScreen^
  1 x:.ink!
  0 x:.bg!
  30 1 68 16 4 x:.drawRoundRect^
  34 5 x:.setCursor^
  "Minipaint!" x:.print^
  6 22 38 22 22 54 x:.fillTriangle^
  50 24 28 28 x:.fillRect^
  104 40 16 x:.fillCircle^
  t:peek o:.oled.showMat^
  x:drop
  tdrop
;

/*
( -- ) o( inst:I2cSsd1306Demo )
*/
: I2cSsd1306Demo#init
  :s initPi
    true i2c.enabled!
    i2c Ssd1306.onBus^ o:.oled!
    o:.oled.begin^
    I2cSsd1306Demo_paint
  ;s

  GpioCtrl.inst.reset^
  initPi
  "SSD1306 Minipaint scene on i2c @ 0x3C\n" $
;

null var: I2cSsd1306Demo_inst

: i2c-ssd1306-demo
  I2cSsd1306Demo.new^ dup >ts
  I2cSsd1306Demo_inst !
  t:.init^
  tdrop
;

i2c-ssd1306-blurb $ nl
