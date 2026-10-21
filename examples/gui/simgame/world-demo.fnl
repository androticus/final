forget: reset-world-demo
requiresFyne
`
This is a demonstration of world/viewport mapping: a larger world
drawn in world coordinates, viewed through a fixed Pixmap via
Pixmap.viewOrigin. Patterns draw from blit-draw/draw.fnl; SimGame
is only for setup and events. Blit and paint work on any Pixmap.

Try:

world-demo

Arrow keys / drag / scroll to pan. Home recenters on the hub.
` const: world-demo-blurb


{
  "WD" "name"
   SimGame "baseType"
   
   {
 // SimGame "base" // added automatically
    Label "originLabel"
    Bool "dragging"
    Point "dragLast"
    Palette1 "pal1"
    Palette1 "pal1Blend"
    Mat2i1 "mat1"
    Palette4 "pal4"
    Palette4 "pal4Blend"
    Mat2i4 "mat4"
    Palette8 "pal8"
    Palette8 "pal8Blend"
    Mat2u8 "mat8"
    Mat2c "matC"
    Mat2c "matCBlend"
  } "fields"

  {
    Color.white "backgroundColor"
    40 "panStep"
    -520 "westX"
    900 "eastX" // room for 4×3 blit grid at cellPx 16
    -420 "northY"
    420 "southY"
    // identity viewOrigin footprint for PixSize640x480
    640 "viewW"
    480 "viewH"
    16 "blitCellPx"
    8 "blitMatN" // source mat N×N; × cellPx → 128×128 stamps
    200 "blitPitch" // spacing between stamp top-lefts
  } "constants"
}
TypedObject.makeType^


// --- pan / HUD -----------------------------------------------------

// Center world (0,0) in the current pixmap.
// ( -- ) o( inst )
: WD_homeOrigin
  o:.base@.pixmap@ >ts
  t:.cols@ 2 / -1 *
  t:.rows@ 2 / -1 *
  t:.setViewOrigin^
  tdrop
;

// ( dx:n dy:n -- ) o( inst )
: WD_panBy
  l.dy!
  l.dx!
  o:.base@.pixmap@ >ts
  t:.viewOrigin.x@ l.dx@ +
  t:.viewOrigin.y@ l.dy@ +
  t:.setViewOrigin^
  tdrop
;

// ( -- ) o( inst )
: WD_updateOriginLabel
  o:.base@.pixmap.viewOrigin@ >ts
  ( t:.x@ t:.y@ )
  "Current origin: %.0f,%.0f" fmt
  o:.originLabel.setText^
  tdrop
;

// ( dx:n dy:n -- ) o( inst )
: WD_panAndRedraw
  WD_panBy
  WD_updateOriginLabel
  o:.base.doRender^
;

// Title at a world point. Painter on ts.
// ( s:s wx:n wy:n -- ) t( p:Painter )
: WD_labelAt
  l.wy!
  l.wx!
  l.s!
  Color.black t:.font.color!
  Font.go 18 t:.font.setFamilyAndSize^
  TextAlign.centerMiddle t:.font.align!
  l.s@ l.wx@ l.wy@ t:.text^
;

// Arrow + caption. Painter on ts.
// ( s:s x0:n y0:n x1:n y1:n labX:n labY:n -- ) t( p:Painter )
: WD_arrow
  l.labY!
  l.labX!
  l.y1!
  l.x1!
  l.y0!
  l.x0!
  l.s!
  Color.maroon t:.stroke.color!
  2 t:.stroke.lineWidth!
  LineStyle.solid t:.stroke.lineStyle!
  l.x0@ l.y0@  l.x1@ l.y1@  t:.line^
  Color.maroon t:.font.color!
  Font.go 14 t:.font.setFamilyAndSize^
  TextAlign.centerMiddle t:.font.align!
  l.s@ l.labX@ l.labY@ t:.text^
;


// --- lands (world coords; Painter on ts unless noted) --------------

// Identity viewOrigin footprint: world (0,0)‥(viewW-1, viewH-1).
// ( -- ) t( p:Painter )
: WD_paintDefaultViewport
  Color.gray t:.stroke.color!
  2 t:.stroke.lineWidth!
  LineStyle.dot t:.stroke.lineStyle!
  FillStyle.none t:.fill.style!
  // center of [0, viewW)×[0, viewH)
  WD.viewW 2 /  WD.viewH 2 /
  WD.viewW WD.viewH Size.new^
  t:.rect^

  Color.gray t:.font.color!
  Font.go 14 t:.font.setFamilyAndSize^
  TextAlign.leftTop t:.font.align!
  "Default Viewport" 8 8 t:.text^

  LineStyle.solid t:.stroke.lineStyle!
;

// ( -- ) t( p:Painter )
: WD_paintHub
  WD_paintDefaultViewport

  Color.navy t:.stroke.color!
  2 t:.stroke.lineWidth!
  LineStyle.solid t:.stroke.lineStyle!
  FillStyle.none t:.fill.style!

  -80 0  80 0  t:.line^
  0 -80  0 80  t:.line^

  FillStyle.solid t:.fill.style!
  Color.yellow t:.fill.color!
  Color.olive t:.stroke.color!
  0 0 14 t:.circle^
  FillStyle.none t:.fill.style!

  "World Gallery" 0 -110 WD_labelAt
  "viewOrigin pans this viewport" 0 110 WD_labelAt
  Color.gray t:.font.color!
  Font.go 12 t:.font.setFamilyAndSize^
  TextAlign.centerMiddle t:.font.align!
  "(0,0) hub — Home to recenter" 0 132 t:.text^

  "<-- Circle Land"      -40 0   -160 0    -100 -22  WD_arrow
  "Blit Land -->"         40 0    160 0     100 -22  WD_arrow
  "^ Shape / Path Land"    0 -40    0 -150     0 -90  WD_arrow
  "v Text Land"            0  40    0  150     0  90  WD_arrow
;

// ( -- ) t( p:Painter )
: WD_paintCircleLand
  "Circle Land" WD.westX -160 WD_labelAt

  Color.green t:.stroke.color!
  2 t:.stroke.lineWidth!
  FillStyle.none t:.fill.style!
  WD.westX 0 50 t:.circle^

  Color.lime t:.stroke.color!
  FillStyle.solid t:.fill.style!
  Color.aqua t:.fill.color!
  WD.westX 80 -  -40  36 24 t:.ellipse^

  Color.olive t:.stroke.color!
  Color.silver t:.fill.color!
  WD.westX 80 +  -40  40 22 35.0 t:.ellipseRot^

  FillStyle.none t:.fill.style!
  Color.teal t:.stroke.color!
  WD.westX 30 -  70
  WD.westX 90 +  110
  t:.line^
;

// ( -- ) t( p:Painter )
: WD_paintShapeLand
  "Shape / Path Land" 0 WD.northY 140 - WD_labelAt

  Color.blue t:.stroke.color!
  FillStyle.solid t:.fill.style!
  Color.aqua t:.fill.color!
  0 WD.northY  90 55 Size.new^ t:.rect^

  Color.navy t:.stroke.color!
  Color.silver t:.fill.color!
  110 WD.northY 20 +  70 40 Size.new^ 25.0 t:.rectRot^

  Color.purple t:.stroke.color!
  Color.fuchsia t:.fill.color!
  -110 WD.northY 20 +  80 50 Size.new^ 12 t:.roundedRect^

  FillStyle.none t:.fill.style!
  Color.maroon t:.stroke.color!
  5  -90 WD.northY 70 +  28 t:.regularPolygon^
  6   90 WD.northY 70 +  28 18.0 t:.regularPolygonRot^

  Color.red t:.stroke.color!
  2 t:.stroke.lineWidth!
  t:.pathBuilder@ >xs
  x:.startPath^
  -40 WD.northY 120 +  x:.moveTo^
  0 WD.northY 160 +    x:.lineTo^
  40 WD.northY 120 +   x:.lineTo^
  x:.end^
  xdrop

  FillStyle.solid t:.fill.style!
  Color.yellow t:.fill.color!
  Color.maroon t:.stroke.color!
  t:.pathBuilder@ >xs
  x:.startSolid^
  -30 WD.northY 200 +  x:.moveTo^
  30 WD.northY 200 +   x:.lineTo^
  0 WD.northY 240 +    x:.lineTo^
  x:.end^
  xdrop
  FillStyle.none t:.fill.style!
;

// ( -- ) t( p:Painter )
: WD_paintTextLand
  "Text Land" 0 WD.southY 140 - WD_labelAt

  Color.black t:.font.color!
  Font.go 16 t:.font.setFamilyAndSize^
  TextAlign.leftMiddle t:.font.align!
  "Painter.text — left" -180 WD.southY t:.text^

  TextAlign.centerMiddle t:.font.align!
  Font.goBold 16 t:.font.setFamilyAndSize^
  "center bold" 0 WD.southY t:.text^

  TextAlign.rightMiddle t:.font.align!
  Font.goItalic 16 t:.font.setFamilyAndSize^
  "italic right —" 180 WD.southY t:.text^

  Font.go 14 t:.font.setFamilyAndSize^
  TextAlign.leftTop t:.font.align!
  Color.gray t:.font.color!
  "word wrap stays in world space;\npan to read me fully"
  -120 WD.southY 50 +
  240
  t:.textWrap^

  Color.red t:.stroke.color!
  LineStyle.solid t:.stroke.lineStyle!
  -20 WD.southY 100 + t:.pixel^
  Color.blue -18 WD.southY 100 + t:.pixelC^
;

// Blit wing — 4 kinds × 3 modes (stomp / blend / mask), cellPx-scaled.
// Grid top-left at (eastX, -blitPitch); cols = kinds, rows = modes.
// ( col:i row:i -- wx:n wy:n )  stamp top-left in world
: WD_blitCellAt
  l.row!
  l.col!
  WD.eastX l.col@ WD.blitPitch * +
  WD.blitPitch -1 * l.row@ WD.blitPitch * +
;

// Teal underlay slightly larger than the stamp (shows through mask/blend).
// ( wx:n wy:n -- ) t( p:Painter )
: WD_blitUnderlay
  l.wy!
  l.wx!
  WD.blitMatN WD.blitCellPx * l.side!
  12 l.pad!
  FillStyle.solid t:.fill.style!
  Color.teal t:.fill.color!
  LineStyle.none t:.stroke.lineStyle!
  l.wx@ l.side@ 2 / +
  l.wy@ l.side@ 2 / +
  l.side@ l.pad@ 2 * +
  l.side@ l.pad@ 2 * +
  Size.new^
  t:.rect^
  FillStyle.none t:.fill.style!
  LineStyle.solid t:.stroke.lineStyle!
;

// ( src:BlitSrc op:i mask:i wx:n wy:n -- ) y( pix:Pixmap )
: WD_blitStamp
  l.wy!
  l.wx!
  l.mask!
  l.op!
  l.src!
  BlitParams.new^ >ts
  WD.blitCellPx t:.cellPx!
  l.op@ t:.op!
  l.mask@ t:.mask!
  l.src@ ts>
  l.wx@ l.wy@
  y:.blitWithOverrides^
;

// ( -- ) o( inst )
: WD_paintBlitLand
  o:.base@.painter@ >ts
  "Blit Land — 4×3 (stomp / blend / mask), cellPx 16"
  WD.eastX 300 +
  WD.blitPitch -1 * 70 -
  WD_labelAt

  Color.gray t:.font.color!
  Font.go 12 t:.font.setFamilyAndSize^
  (
    "blit1" "blit4" "blit8" "blitC"
  ) l.colLabs!
  (
    "stomp" "blend" "mask"
  ) l.rowLabs!
  WD.blitMatN WD.blitCellPx * l.side!

  TextAlign.centerMiddle t:.font.align!
  0 3 range foreach
    i 0 WD_blitCellAt
    l.wy! l.wx!
    i l.colLabs@()
    l.wx@ l.side@ 2 / +
    l.wy@ 24 -
    t:.text^
  /foreach

  TextAlign.rightMiddle t:.font.align!
  0 2 range foreach
    0 i WD_blitCellAt
    l.wy! l.wx!
    i l.rowLabs@()
    l.wx@ 12 -
    l.wy@ l.side@ 2 / +
    t:.text^
  /foreach

  0 2 range foreach
    0 3 range foreach
      i j WD_blitCellAt
      WD_blitUnderlay
    /foreach
  /foreach
  tdrop

  o:.base@.pixmap@ >ys

  // row 0 stomp
  o:.pal1@ o:.mat1@ BlitSrc.new1^
  BlitParams.OpStomp BlitParams.MaskOff
  0 0 WD_blitCellAt WD_blitStamp

  o:.pal4@ o:.mat4@ BlitSrc.new4^
  BlitParams.OpStomp BlitParams.MaskOff
  1 0 WD_blitCellAt WD_blitStamp

  o:.pal8@ o:.mat8@ BlitSrc.new8^
  BlitParams.OpStomp BlitParams.MaskOff
  2 0 WD_blitCellAt WD_blitStamp

  o:.matC@ BlitSrc.newC^
  BlitParams.OpStomp BlitParams.MaskOff
  3 0 WD_blitCellAt WD_blitStamp

  // row 1 blend — A=128 pals / matCBlend
  o:.pal1Blend@ o:.mat1@ BlitSrc.new1^
  BlitParams.OpBlend BlitParams.MaskOff
  0 1 WD_blitCellAt WD_blitStamp

  o:.pal4Blend@ o:.mat4@ BlitSrc.new4^
  BlitParams.OpBlend BlitParams.MaskOff
  1 1 WD_blitCellAt WD_blitStamp

  o:.pal8Blend@ o:.mat8@ BlitSrc.new8^
  BlitParams.OpBlend BlitParams.MaskOff
  2 1 WD_blitCellAt WD_blitStamp

  o:.matCBlend@ BlitSrc.newC^
  BlitParams.OpBlend BlitParams.MaskOff
  3 1 WD_blitCellAt WD_blitStamp

  // row 2 mask — index0 / A==0 skipped → teal underlay
  o:.pal1@ o:.mat1@ BlitSrc.new1^
  BlitParams.OpStomp BlitParams.MaskOn
  0 2 WD_blitCellAt WD_blitStamp

  o:.pal4@ o:.mat4@ BlitSrc.new4^
  BlitParams.OpStomp BlitParams.MaskOn
  1 2 WD_blitCellAt WD_blitStamp

  o:.pal8@ o:.mat8@ BlitSrc.new8^
  BlitParams.OpStomp BlitParams.MaskOn
  2 2 WD_blitCellAt WD_blitStamp

  o:.matC@ BlitSrc.newC^
  BlitParams.OpStomp BlitParams.MaskOn
  3 2 WD_blitCellAt WD_blitStamp

  ydrop
;


// --- assets --------------------------------------------------------

// Shared N×N patterns; rendered at blitCellPx (16 → 128×128).
// ( -- ) o( inst )
: WD_buildBlitAssets
  Color.black Color.white Palette1.newWithColors^ o:.pal1!
  0 0 0 128 Color.rgba^  255 255 255 128 Color.rgba^
  Palette1.newWithColors^ o:.pal1Blend!

  WD.blitMatN dup Mat2i1.new^ >ts
  0 WD.blitMatN 1 - range foreach
    0 WD.blitMatN 1 - range foreach
      i j + 2 %
      j i t:![]
    /foreach
  /foreach
  ts> o:.mat1!

  255 Palette4.newWebPalette^ o:.pal4!
  128 Palette4.newWebPalette^ o:.pal4Blend!
  WD.blitMatN dup Mat2i4.new^ >ts
  0 WD.blitMatN 1 - range foreach
    0 WD.blitMatN 1 - range foreach
      i j + WD.blitMatN %
      j i t:![]
    /foreach
  /foreach
  ts> o:.mat4!

  255 Palette8.new332Palette^ o:.pal8!
  128 Palette8.new332Palette^ o:.pal8Blend!
  WD.blitMatN dup Mat2u8.new^ >ts
  0 WD.blitMatN 1 - range foreach
    0 WD.blitMatN 1 - range foreach
      i 32 * j 4 * +
      j i t:![]
    /foreach
  /foreach
  ts> o:.mat8!

  WD.blitMatN dup Mat2c.new^ >ts
  WD.blitMatN dup Mat2c.new^ >xs
  0 WD.blitMatN 1 - range foreach
    0 WD.blitMatN 1 - range foreach
      i 0 == j 0 == && if
        Color.transparent j i t:![]
        0xFFFFFF7F Color.hexa^ j i x:![]
      else
        256 Rand.int^ l.r!
        256 Rand.int^ l.g!
        256 Rand.int^ l.b!
        l.r@ l.g@ l.b@ 255 Color.rgba^ j i t:![]
        l.r@ l.g@ l.b@ 128 Color.rgba^ j i x:![]
      then
    /foreach
  /foreach
  ts> o:.matC!
  xs> o:.matCBlend!
;


// --- events --------------------------------------------------------

: WD#handleMouseDown
  >ts
  true o:.dragging!
  t:.position@ o:.dragLast!
  tdrop
;

: WD#handleMouseUp
  drop
  false o:.dragging!
;

: WD#handleMouseIn
  drop
;

: WD#handleMouseMoved
  o:.dragging@ not if
    drop
  else
    >ts
    o:.dragLast.x@ t:.position.x@ -
    o:.dragLast.y@ t:.position.y@ -
    t:.position@ o:.dragLast!
    tdrop
    WD_panAndRedraw
  then
;

: WD#handleMouseOut
  false o:.dragging!
;

: WD#handleMouseScrolled
  >ts
  t:.scrolled.x@ -1 *
  t:.scrolled.y@ -1 *
  tdrop
  WD_panAndRedraw
;

: WD#handleKeyDown
  >ts
    t:.name
  tdrop
  switch
    a Keys.left == case
      WD.panStep -1 *  0  WD_panAndRedraw
    /case
    a Keys.right == case
      WD.panStep  0  WD_panAndRedraw
    /case
    a Keys.up == case
      0  WD.panStep -1 *  WD_panAndRedraw
    /case
    a Keys.down == case
      0  WD.panStep  WD_panAndRedraw
    /case
    a Keys.home == case
      WD_homeOrigin
      WD_updateOriginLabel
      o:.base.doRender^
    /case
  /switch
;

: WD#handleKeyUp
  drop
;

: WD#handleRender
  o:.base@.painter@ >ts
  LineStyle.solid t:.stroke.lineStyle!
  1 t:.stroke.lineWidth!
  FillStyle.none t:.fill.style!
  Color.black t:.font.color!

  WD_paintHub
  WD_paintCircleLand
  WD_paintShapeLand
  WD_paintTextLand
  tdrop

  WD_paintBlitLand
;

: WD#handleUpdate
;


: WD#makeBottomBar
  (
    "Arrows / drag / scroll to pan · Home = hub" Label.new^
    Separator.new^
    "Current origin: 0,0" Label.new^ dup o:.originLabel!
  ) Box.newH^
;

: WD#preInit
  o:.base@ >ys

  "World Demo — scroll the gallery (viewOrigin)" y:.title!
  WD.backgroundColor y:.backgroundColor!
  true y:.options.hasKeyHandler!
  true y:.options.useKeyAutoRepeat!
  true y:.options.hasMouseDownUpHandler!
  true y:.options.hasMouseInMovedOutHandler!
  true y:.options.hasMouseScrolledHandler!

  SimGame.PixSize640x480 y:.pixSize!

  ydrop
;

: WD#postInit
  false o:.dragging!
  WD_buildBlitAssets
  WD_homeOrigin
;

: world-demo
  WD.new^ >zs
  z:.base >ys
  y:.init^
  y:.makeUi^
  y:.startPaused^
  y:.doRender^
  y:.win@.show^
  ydrop zdrop
;

world-demo-blurb $ nl
