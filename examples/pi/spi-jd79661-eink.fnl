forget: reset-spi-jd79661-demo
requiresPi

// install and load the jd79661 driver module (Jd79661)
"github.com/androticus/final/modules/pi/drivers/jd79661"
{
  "modules/pi/drivers/jd79661" "path"
}
install

"" "github.com/androticus/final/modules/pi/drivers/jd79661" import

`
SPI demo: Adafruit JD79661 4-color e-ink (2.13").

Wiring (spi / spi0 + adj GPIOs — match your harness):
  Vin 3.3V, GND
  SCK  GPIO11, MOSI GPIO10, ECS → CE0 (GPIO8)
  MISO optional (SRAM/SD only; unused here)
  BUSY GPIO22, DC GPIO23, RST GPIO24

Host:
  dtparam=spi=on
  (spi group / root)

Load:
  "examples/pi/spi-jd79661-eink.fnl" compileFromFile

Then:
  spi-jd79661-demo

Paints one 128×250 Mat2i4 (the panel RAM, landscape 250×128): white
background, a black top bar with "Final Pi E-ink Demo" at text size 2,
then overlapping circle / oval / rect / rounded rect (black border,
yellow and red fills). Updates the display only after the mat is complete.
Powers the controller down (image persists), then GpioCtrl.reset^.
` const: spi-jd79661-demo-blurb

{
  "SpiJd79661Demo" "name"
  {
    Jd79661 "panel"
  } "fields"
}
TypedObject.makeType^

/*
One Mat2i4 the size of the panel; paint the whole scene into it, then show.
Palette indexes match the panel: 0 black, 1 white, 2 yellow, 3 red.
( -- ) o( inst:SpiJd79661Demo )
*/
: SpiJd79661Demo_paint
  Jd79661.width Jd79661.height Mat2i4.new^ >ts
  4 List.newWithCapacity^ >ys
  Color.black y:.append^
  Color.white y:.append^
  Color.yellow y:.append^
  Color.red y:.append^
  ys> Palette4.newFromList^ t:peek BlitSrc.new4^
  Minipaint.new^ >xs

  Jd79661.white x:.ink!
  x:.fillScreen^

  Jd79661.black x:.ink!
  0 0 250 22 x:.fillRect^
  Jd79661.white x:.ink!
  Jd79661.black x:.bg!
  2 x:.textSize!
  false x:.wrap!
  11 3 x:.setCursor^
  "Final Pi E-ink Demo" x:.print^

  // circle, yellow
  Jd79661.yellow x:.ink!
  48 74 34 x:.fillCircle^
  Jd79661.black x:.ink!
  48 74 34 x:.drawCircle^

  // oval, red
  Jd79661.red x:.ink!
  110 78 46 28 x:.fillEllipse^
  Jd79661.black x:.ink!
  110 78 46 28 x:.drawEllipse^

  // rect, yellow
  Jd79661.yellow x:.ink!
  140 36 64 52 x:.fillRect^
  Jd79661.black x:.ink!
  140 36 64 52 x:.drawRect^

  // rounded rect, red
  Jd79661.red x:.ink!
  175 56 64 52 12 x:.fillRoundRect^
  Jd79661.black x:.ink!
  175 56 64 52 12 x:.drawRoundRect^

  t:peek o:.panel.showMat^
  x:drop
  tdrop
;

/*
Enable spi + BUSY/DC/RST, paint the scene, tear down.
( -- ) o( inst:SpiJd79661Demo )
*/
: SpiJd79661Demo#init
  GpioCtrl.inst.reset^

  // adj pins before bus enable (Validate claims)
  GpioPin.in gpio22.direction!
  true gpio22.enabled!

  GpioPin.out gpio23.direction!
  false gpio23.value!
  true gpio23.enabled!

  GpioPin.out gpio24.direction!
  true gpio24.value!
  true gpio24.enabled!

  SpiPeriph.mode0 spi.mode!
  SpiPeriph.bitsPerWord8 spi.bitsPerWord!
  1000000 spi.maxHz!
  1 spi.numCS!
  0 spi.selCS!
  true spi.enabled!

  spi gpio23 gpio24 gpio22 Jd79661.connect^ o:.panel!

  "JD79661: painting scene (refresh may take several seconds)…\n" $
  SpiJd79661Demo_paint
  "JD79661: done — image should persist; resetting GpioCtrl.\n" $

  GpioCtrl.inst.reset^
;

null var: SpiJd79661Demo_inst

: spi-jd79661-demo
  SpiJd79661Demo.new^ dup >ts
  SpiJd79661Demo_inst !
  t:.init^
  tdrop
;

spi-jd79661-demo-blurb $ nl
