forget: reset-draw-demo

requiresFyne
`
Examples showing how to use Final Painter to draw.
This covers the entire Paint API.
We use Fyne to display the result, but Fyne/final-gui is not
needed to use the Pixmap and Painter classes.

Try:

draw-demo
`
const: draw-demo-blurb

// Create a type for the app

{
   {
     Size "sz" // pixmap size
     Palette4 "pal"
     Painter "p" // the api for drawing
     Pixmap "pix" // what painter draws on (p.pixmap)
     CanvasImage "view"
     Select "sel"
     Select "lsSel"
     Window "win"
   } "fields"
   
   {
     0 "btnRun"
     
     // a pleasant screen size: not too big or small
     480 "pixH"
     640 "pixW" // maybe these should be params to the Sim

     0 "demoLine"
     1 "demoCircle"
     2 "demoRectangle"
     3 "demoPolygon"
     4 "demoText"
     5 "demoFill"
     6 "demoPathBuilder"
     7 "demoSolidBuilder"
     8 "demoPushPop"
     (
       "Line"
       "Circle"
       "Rectangle"
       "Polygon"
       "Text"
       "Fill"
       "Path Builder"
       "Solid Builder"
       "Push/Pop"
     s) "demos"
     (
       "solid"
       "none"
       "dot"
       "dash"
       "dotDash"
     s) "lineStyles"
     (
       LineStyle.solid
       LineStyle.none
       LineStyle.dot
       LineStyle.dash
       LineStyle.dotDash
     ) "lineStyleVals"
   } "constants"
   
  "DrawFyneApp" "name"
}
TypedObject.makeType^

// small size (integral bounds for Rand.int)
: DrawFyneApp_rndSize
  o:.sz@.wi 4 / Rand.int^
  o:.sz@.hi 4 / Rand.int^
  Size.new^
;

: DrawFyneApp_rndPoint
  o:.sz@.wi Rand.int^
  o:.sz@.hi Rand.int^
  Point.new^
;

: DrawFyneApp_rndRotation
  360 Rand.int^
  >Float
;

: getNonWhiteColor
  forever
    16 Rand.int^
    dup 3 == if drop else break then
  /forever
  o:.pal@ >ts t:@() t:drop
;

: getCenterPoint
  o:.sz@.w 2 /
  o:.sz@.h 2 /
  Point.new^
;

: DrawFyneApp_line
  Color.red o:.p@.stroke.color!
  // draw random red lines
  0 9 range
  foreach
    DrawFyneApp_rndPoint
    DrawFyneApp_rndPoint
    o:.p@.line^
  /foreach
  o:.view@.refresh^
;

: DrawFyneApp_circle
  // random green circles and ellipses
  Color.green o:.p@.stroke.color!
  0 5 range
  foreach
    DrawFyneApp_rndPoint
    o:.sz@.hi 4 / Rand.int^ 4 +
    o:.p@.circle^
  /foreach
  
  Color.lime o:.p@.stroke.color!
  0 5 range
  foreach
    DrawFyneApp_rndPoint
    o:.sz@.wi 8 / Rand.int^ 8 +
    o:.sz@.hi 8 / Rand.int^ 8 +
    o:.p@.ellipse^
  /foreach

  Color.olive o:.p@.stroke.color!
  0 5 range
  foreach
    DrawFyneApp_rndPoint
    o:.sz@.wi 8 / Rand.int^ 8 +
    o:.sz@.hi 8 / Rand.int^ 8 +
    DrawFyneApp_rndRotation
    o:.p@.ellipseRot^
  /foreach
  o:.view@.refresh^
;

: DrawFyneApp_rectangle
  // random blue rects, in bounds, filled (center + Size)
  Color.blue o:.p@.stroke.color!
  FillStyle.solid o:.p.fill.style!
  Color.aqua o:.p@.fill.color!
  
  0 5 range
  foreach
    DrawFyneApp_rndPoint
    DrawFyneApp_rndSize
    o:.p@.rect^
  /foreach

  Color.gray o:.p@.stroke.color!
  Color.silver o:.p@.fill.color!
  0 5 range
  foreach
    DrawFyneApp_rndPoint
    DrawFyneApp_rndSize
    DrawFyneApp_rndRotation
    o:.p@.rectRot^
  /foreach

  Color.navy o:.p@.stroke.color!
  Color.teal o:.p@.fill.color!
  0 5 range
  foreach
    DrawFyneApp_rndPoint
    DrawFyneApp_rndSize dup l.sz!
    l.sz.wi@ l.sz.hi@ min 2 /
    1 + Rand.int^ 1 +
    o:.p@.roundedRect^
  /foreach

  Color.olive o:.p@.stroke.color!
  Color.lime o:.p@.fill.color!
  0 5 range
  foreach
    DrawFyneApp_rndPoint
    DrawFyneApp_rndSize dup l.sz!
    l.sz.wi@ l.sz.hi@ min 2 /
    1 + Rand.int^ 1 +
    DrawFyneApp_rndRotation
    o:.p@.roundedRectRot^
  /foreach
  o:.view@.refresh^
  FillStyle.none o:.p.fill.style!
;

: DrawFyneApp_polygon
  // regular polygons centered (angled via regularPolygonRot)
  16 dup l.n!
  o:.sz@.h 2 / swap 1 + / l.rad!
  360.0 l.n@ / l.rot!
  getCenterPoint >ts
  0 l.n@ 1 - range
  foreach
    getNonWhiteColor o:.p@.stroke.color!
    i 3 + // sides
    t:peek // center
    i 1 + l.rad@ *  // radius (i=0 => 20; radius 0 is a point)
    i l.rot@ * // rotation (clockwise degrees)
    o:.p@.regularPolygonRot^
  /foreach
  t:drop
  o:.view@.refresh^
;

// ( s:s at:Point family:s size:n )
: DrawFyneApp_text_write
  o:.p@ >ts
  t:.font.setFamilyAndSize^
  t:.text^
  t:drop
;

: DrawFyneApp_text
  o:.p@ >ts
  FillStyle.none t:.fill.style!
  LineStyle.solid t:.stroke.lineStyle!
  1 t:.stroke.lineWidth!
  Color.silver t:.stroke.color!
  Color.black t:.font.color!

  14 l.pt!
  110 l.cw!
  46 l.ch!
  36 l.gapX!
  10 l.gapY!
  160 l.mx! // ~one column right of the old origin; room for row labels
  12 l.my!
  44 l.head! // column titles + air before row 1

  (
    Font.go
    Font.goItalic
    Font.goBold
    Font.goBoldItalic
    Font.goMono
    Font.goSmallCaps
  ) l.faces!
  (
    TextAlign.top
    TextAlign.middle
    TextAlign.baseline
    TextAlign.bottom
    TextAlign.baseline
    TextAlign.baseline
  ) l.vers!
  (
    TextAlign.left
    TextAlign.center
    TextAlign.right
  ) l.hors!
  (
    "Left"
    "Center"
    "Right"
  ) l.colLabs!
  (
    "Top"
    "Middle"
    "Baseline"
    "Bottom"
    "Word Wrap"
    "Line Wrap"
  ) l.rowLabs!

  Font.go l.pt@ t:.font.setFamilyAndSize^

  // column titles: Left / Center / Right, same H as the column, above row 1
  l.my@ 16 + l.hy!
  0 2 range
  foreach
    l.mx@ l.cw@ 2 / + i l.cw@ l.gapX@ + * + l.cx!
    i l.hors@() TextAlign.bottom | t:.font.align!
    l.cx@
    i l.hors@() TextAlign.left == if l.cw@ 2 / - then
    i l.hors@() TextAlign.right == if l.cw@ 2 / + then
    l.hy@ Point.new^
    i l.colLabs@() swap
    t:.text^
  /foreach

  0 5 range
  foreach
    0 2 range
    foreach
      j l.faces@() l.pt@ t:.font.setFamilyAndSize^

      l.mx@ l.cw@ 2 / + i l.cw@ l.gapX@ + * + l.cx!
      l.my@ l.head@ + l.ch@ 2 / +
      j l.ch@ l.gapY@ + * +
      l.cy!

      l.cx@ l.cy@ Point.new^
      l.cw@ l.ch@ Size.new^
      t:.rect^

      i 0 == if
        Font.go l.pt@ t:.font.setFamilyAndSize^
        TextAlign.right j l.vers@() | t:.font.align!
        l.mx@ 10 -
        l.cy@
        j l.vers@() TextAlign.top == if l.ch@ 2 / - then
        j l.vers@() TextAlign.bottom == if l.ch@ 2 / + then
        j l.vers@() TextAlign.baseline == if l.ch@ 4 / + then
        Point.new^
        j l.rowLabs@() swap
        t:.text^
        j l.faces@() l.pt@ t:.font.setFamilyAndSize^
      then

      i l.hors@() j l.vers@() | t:.font.align!

      l.cx@
      i l.hors@() TextAlign.left == if l.cw@ 2 / - then
      i l.hors@() TextAlign.right == if l.cw@ 2 / + then
      l.cy@
      j l.vers@() TextAlign.top == if l.ch@ 2 / - then
      j l.vers@() TextAlign.bottom == if l.ch@ 2 / + then
      j l.vers@() TextAlign.baseline == if l.ch@ 4 / + then
      Point.new^ l.at!

      j 4 == if
        "snug wrap in this box" l.at@ l.cw@ t:.textWrap^
      else
        j 5 == if
          "hello\\n\nworld" l.at@ t:.text^
        else
          j l.faces@() l.at@ t:.text^
        then
      then
    /foreach
  /foreach

  Font.go l.pt@ t:.font.setFamilyAndSize^
  TextAlign.centerTop t:.font.align!
  l.mx@ l.cw@ 3 * l.gapX@ 2 * + 2 / +
  l.my@ l.head@ + 6 l.ch@ l.gapY@ + * + l.gapY@ - 12 +
  Point.new^
  "All text in 14 pt" swap
  t:.text^

  t:drop
  o:.view@.refresh^
;

// ( s:s x:i y:i -- )
: DrawFyneApp_fill_label
  Point.new^ Font.go 14 DrawFyneApp_text_write
;

: DrawFyneApp_fill
  o:.sz.w 2 / l.qw!
  o:.sz.h 2 / l.qh!
  16 l.pad!
  22 l.lab! // top inset so labels sit above the shape

  Color.black o:.p@.font.color!
  "solid"        l.pad@         l.pad@         DrawFyneApp_fill_label
  "checkerboard" l.qw@ l.pad@ + l.pad@         DrawFyneApp_fill_label
  "gradient"     l.pad@         l.qh@ l.pad@ + DrawFyneApp_fill_label
  "tile"         l.qw@ l.pad@ + l.qh@ l.pad@ + DrawFyneApp_fill_label

  o:.p@ >ts
  t:.fill@ >xs

  Color.silver t:.stroke.color!
  0 l.qh@ Point.new^  o:.sz.w l.qh@ Point.new^  t:.line^
  l.qw@ 0 Point.new^  l.qw@ o:.sz.h Point.new^  t:.line^

  // TL — solid circle, fuchsia / purple
  FillStyle.solid x:.style!
  Color.fuchsia x:.color!
  Color.purple t:.stroke.color!
  l.qw@ 2 /  l.qh@ 2 / l.lab@ 2 / +  Point.new^
  l.qh@ 2 / l.pad@ -  12 -
  t:.circle^

  // TR — checkerboard rect, navy / aqua (center of quadrant)
  FillStyle.checkerboard x:.style!
  Color.navy x:.color!
  Color.aqua x:.color2!
  14 x:.checkerSize!
  Color.navy t:.stroke.color!
  l.qw@ l.pad@ +  l.qw@ l.pad@ 2 * - 2 / +
  l.lab@  l.qh@ l.lab@ - l.pad@ - 2 / +
  Point.new^
  l.qw@ l.pad@ 2 * -  l.qh@ l.lab@ - l.pad@ -  Size.new^
  t:.rect^

  // BL — gradient pentagon, maroon → yellow, 90° (down)
  FillStyle.gradient x:.style!
  Color.maroon x:.color!
  Color.yellow x:.color2!
  90 x:.setGradientDegrees^
  Color.olive t:.stroke.color!
  5
  l.qw@ 2 /  l.qh@ dup 2 / + l.lab@ 2 / +  Point.new^
  l.qh@ 2 / l.pad@ -  14 -
  18
  t:.regularPolygonRot^

  // BR — tile rect, compact logo (center of quadrant)
  FillStyle.tile x:.style!
  "examples/gui/fyne/favicon-32-f-circle.png" Image.newFromFile^
  x:.tileImage!
  2 x:.tileZoom!
  Color.teal t:.stroke.color!
  l.qw@ l.pad@ +  l.qw@ l.pad@ 2 * - 2 / +
  l.qh@ l.lab@ +  l.qh@ l.lab@ - l.pad@ - 2 / +
  Point.new^
  l.qw@ l.pad@ 2 * -  l.qh@ l.lab@ - l.pad@ -  Size.new^
  t:.rect^

  FillStyle.none x:.style!
  x:drop
  t:drop
  o:.view@.refresh^
;

/*
Shared initial stroke for BezierTo / Arc / EllipticalArc exemplars.
PathBuilder is on t:; pass the caller's locals (has .x0 .dx .dy .yMid).
( locals -- )
*/
: DrawFyneApp_path_stem
  >xs
  x:.x0@ x:.yMid@ t:.moveTo^
  x:.x0@ x:.dx@ +  x:.yMid@ x:.dy@ +  t:.lineTo^
  x:drop
;

/*
Place exemplar slot 0..=2 in a path band: column .x0 and stem .dy
(up, down, up). Mutates the passed locals object.
( slot:i locals -- )
*/
: DrawFyneApp_path_slot
  >xs
  x:.slot!
  o:.sz.w 3 / x:.colW!
  36 x:.slot@ x:.colW@ * + x:.x0!
  // stem angle: first & third up (−dy), second down (+dy)
  x:.slot@ 1 == if
    x:.stemDy@
  else
    x:.stemDy@ -1 *
  then
  x:.dy!
  x:drop
;

: DrawFyneApp_path
  Color.navy o:.p@.stroke.color!
  3 o:.p@.stroke.lineWidth!
  Color.black o:.p@.font.color!

  o:.sz.h 4 / l.rowH!
  14 l.pad!
  // shared stem length and |rise|; sign set per slot
  72 l.dx!
  16 l.stemDy!

  o:.p@.pathBuilder@ >ts

  // --- 0 Line Paths (zigzag: up, down, up) ---
  "Line Paths"
  l.pad@  l.pad@
  DrawFyneApp_fill_label
  l.rowH@ 2 / l.yMid!
  48 l.x0!
  t:.startPath^
  l.x0@            l.yMid@ 16 +  t:.moveTo^
  l.x0@ 140 +      l.yMid@ 16 -  t:.lineTo^
  l.x0@ 280 +      l.yMid@ 16 +  t:.lineTo^
  l.x0@ 420 +      l.yMid@ 16 -  t:.lineTo^
  t:.end^

  // --- 1 BezierTo: 3 exemplars (stem up / down / up) ---
  "BezierTo"
  l.pad@  l.rowH@ l.pad@ +
  DrawFyneApp_fill_label
  l.rowH@ l.rowH@ 2 / + l.yMid!
  0 2 range
  foreach
    i l DrawFyneApp_path_slot
    t:.startPath^
    l DrawFyneApp_path_stem
    // pull away from stem; magnitude grows slightly per slot
    24 l.slot@ 6 * + l.pull!
    l.slot@ 1 == if
      l.x0@ l.dx@ + 36 +   l.yMid@ l.dy@ + l.pull@ +
      l.x0@ l.dx@ + 100 +  l.yMid@ l.dy@ + 4 +
    else
      l.x0@ l.dx@ + 36 +   l.yMid@ l.dy@ + l.pull@ -
      l.x0@ l.dx@ + 100 +  l.yMid@ l.dy@ + 4 -
    then
    t:.bezierTo^
    t:.end^
  /foreach

  // --- 2 Arc: compact quarters so they stay in-band ---
  "Arc"
  l.pad@  l.rowH@ 2 * l.pad@ +
  DrawFyneApp_fill_label
  l.rowH@ 2 * l.rowH@ 2 / + l.yMid!
  0 2 range
  foreach
    i l DrawFyneApp_path_slot
    // r: 22, 26, 30
    22 l.slot@ 4 * + l.r!
    t:.startPath^
    l DrawFyneApp_path_stem
    // tip at left of circle (180°); bend down after up-stem, up after down-stem
    l.x0@ l.dx@ + l.r@ +
    l.yMid@ l.dy@ +
    Point.new^
    l.r@
    l.slot@ 1 == if
      180.0  270.0
    else
      180.0  90.0
    then
    t:.arc^
    t:.end^
  /foreach

  // --- 3 EllipticalArc: 3 exemplars, modest ry ---
  "EllipticalArc"
  l.pad@  l.rowH@ 3 * l.pad@ +
  DrawFyneApp_fill_label
  l.rowH@ 3 * l.rowH@ 2 / + l.yMid!
  0 2 range
  foreach
    i l DrawFyneApp_path_slot
    // rx/ry vary mildly per slot
    48 l.slot@ 6 * + l.rx!
    18 l.slot@ 3 * + l.ry!
    t:.startPath^
    l DrawFyneApp_path_stem
    l.x0@ l.dx@ + l.rx@ +
    l.yMid@ l.dy@ +
    Point.new^
    l.rx@ l.ry@
    l.slot@ 1 == if
      180.0  270.0
    else
      180.0  90.0
    then
    t:.ellipticalArc^
    t:.end^
  /foreach

  t:drop
  o:.view@.refresh^
;

: DrawFyneApp_solid
  FillStyle.solid o:.p@.fill.style!
  Color.aqua o:.p@.fill.color!
  Color.navy o:.p@.stroke.color!
  3 o:.p@.stroke.lineWidth!
  // Irregular hexagon (moveTo + 5 lineTos; end closes) — not a regularPolygon
  o:.sz.w 2 / l.cx!
  o:.sz.h 2 / l.cy!
  o:.p@.pathBuilder@ >ts
  t:.startSolid^
  l.cx@ 140 -  l.cy@ 20 -   t:.moveTo^
  l.cx@ 40 -   l.cy@ 110 -  t:.lineTo^
  l.cx@ 130 +  l.cy@ 90 -   t:.lineTo^
  l.cx@ 160 +  l.cy@ 30 +   t:.lineTo^
  l.cx@ 50 +   l.cy@ 120 +  t:.lineTo^
  l.cx@ 120 -  l.cy@ 80 +   t:.lineTo^
  t:.end^
  t:drop
  FillStyle.none o:.p@.fill.style!
  o:.view@.refresh^
;

// Same marks in every column — only Painter state changes.
// ( x0:i -- )
: DrawFyneApp_pushpop_draw
  l.x0!
  o:.sz.w 3 / l.cw!
  14 l.pad!
  36 l.top!

  o:.p@ >ts
  l.x0@ l.pad@ +  l.cw@ l.pad@ 2 * - 2 / +
  l.top@ 24 +
  Point.new^
  l.cw@ l.pad@ 2 * -  48  Size.new^
  t:.rect^

  l.x0@ l.cw@ 2 / +  l.top@  86 + Point.new^
  26
  t:.circle^

  l.x0@ l.pad@ +  l.top@ 124 + Point.new^
  l.x0@ l.cw@ + l.pad@ -  l.top@ 148 + Point.new^
  t:.line^

  6
  l.x0@ l.cw@ 2 / +  l.top@ 198 + Point.new^
  34
  t:.regularPolygon^

  "Aa"
  l.x0@ l.pad@ +  l.top@ 258 + Point.new^
  t:.text^
  t:drop
;

: DrawFyneApp_pushpop_beforeState
  o:.p@ >ts
  t:.stroke@ >xs
  t:.fill@ >ys
  Color.navy x:.color!
  3 x:.lineWidth!
  LineCap.round x:.lineCap!
  LineStyle.solid x:.lineStyle!
  FillStyle.solid y:.style!
  Color.aqua y:.color!
  Color.navy t:.font.color!
  Font.go 14 t:.font.setFamilyAndSize^
  x:drop
  y:drop
  t:drop
;

: DrawFyneApp_pushpop_pushState
  o:.p@ >ts
  t:.stroke@ >xs
  t:.fill@ >ys
  Color.red x:.color!
  8 x:.lineWidth!
  LineCap.square x:.lineCap!
  LineStyle.dash x:.lineStyle!
  FillStyle.checkerboard y:.style!
  Color.yellow y:.color!
  Color.red y:.color2!
  10 y:.checkerSize!
  Color.maroon t:.font.color!
  Font.goBold 20 t:.font.setFamilyAndSize^
  x:drop
  y:drop
  t:drop
;

: DrawFyneApp_pushpop
  o:.sz.w 3 / l.cw!

  Color.black o:.p@.font.color!
  Font.go 14 o:.p@.font.setFamilyAndSize^
  "Before"  8                 16 DrawFyneApp_fill_label
  "Push"    l.cw@ 8 +         16 DrawFyneApp_fill_label
  "Pop"     l.cw@ 2 * 8 +     16 DrawFyneApp_fill_label

  o:.p@ >ts
  Color.silver t:.stroke.color!
  1 t:.stroke.lineWidth!
  LineStyle.solid t:.stroke.lineStyle!
  l.cw@ 0 Point.new^  l.cw@ o:.sz.h Point.new^  t:.line^
  l.cw@ 2 * 0 Point.new^  l.cw@ 2 * o:.sz.h Point.new^  t:.line^
  t:drop

  DrawFyneApp_pushpop_beforeState
  0 DrawFyneApp_pushpop_draw

  o:.p@.push^
  DrawFyneApp_pushpop_pushState
  l.cw@ DrawFyneApp_pushpop_draw
  o:.p@.pop^

  l.cw@ 2 * DrawFyneApp_pushpop_draw

  FillStyle.none o:.p@.fill.style!
  o:.view@.refresh^
;

: DrawFyneApp_applyLineStyle
  o:.lsSel@.selectedIndex@
  dup 0 < if
    drop
  else
    DrawFyneApp.lineStyleVals >ts t:@() t:drop
    o:.p@.stroke.lineStyle!
  then
;

: DrawFyneApp_runAt
  o:.p.clear^
  o:.view.refresh^
  
  switch
    a DrawFyneApp.demoLine == case
      DrawFyneApp_line
    /case
    
    a DrawFyneApp.demoCircle == case
      DrawFyneApp_circle
    /case

    a DrawFyneApp.demoRectangle == case
      DrawFyneApp_rectangle
    /case

    a DrawFyneApp.demoPolygon == case
      DrawFyneApp_polygon
    /case

    a DrawFyneApp.demoText == case
      DrawFyneApp_text
    /case

    a DrawFyneApp.demoFill == case
      DrawFyneApp_fill
    /case

    a DrawFyneApp.demoPathBuilder == case
      DrawFyneApp_path
    /case

    a DrawFyneApp.demoSolidBuilder == case
      DrawFyneApp_solid
    /case

    a DrawFyneApp.demoPushPop == case
      DrawFyneApp_pushpop
    /case
  /switch
;

: DrawFyneApp_run
  o:.sel@.selectedIndex@
  DrawFyneApp_runAt
;

/*
 Button event handler.
( ev:Event -- ) o( inst )
*/
: DrawFyneApp#onButtonClick
  >ts t:.index@ t:drop
  switch
    a DrawFyneApp.btnRun == case
      DrawFyneApp_applyLineStyle
      DrawFyneApp_run
    /case
  /switch
;

/*
Demo menu: same indices as DrawFyneApp.demos / Select.
( ev:Event -- ) o( inst )
*/
: DrawFyneApp#onMenuDemo
  >ts t:.index@ t:drop
  dup o:.sel@.selectedIndex!
  DrawFyneApp_applyLineStyle
  DrawFyneApp_runAt
;

/*
( index:i text:s -- btn:Button ) o( inst )
*/
: DrawFyneApp_makeButton
  >ts >ts
  o:.onButtonClick~ ts> ts>
  Button.new^
;

/*
In-window Demo menu (Window.setMenuBar, not fyne MainMenu).
( -- bar:MenuBar )
*/
: DrawFyneApp#makeMenuBar
  o:.onMenuDemo~ >ts
  (
    (
      t:peek 0 "Line" MenuItem.new^
      t:peek 1 "Circle" MenuItem.new^
      t:peek 2 "Rectangle" MenuItem.new^
      t:peek 3 "Polygon" MenuItem.new^
      t:peek 4 "Text" MenuItem.new^
      t:peek 5 "Fill" MenuItem.new^
      t:peek 6 "Path Builder" MenuItem.new^
      t:peek 7 "Solid Builder" MenuItem.new^
      t:peek 8 "Push/Pop" MenuItem.new^
    ) "Demo" Menu.new^
  ) MenuBar.new^
  t:drop
;

/*
Create the UI
( -- ) o( inst )
*/
: DrawFyneApp#makeUi
  o:.pix@.image CanvasImage.new^ dup o:.view!

  (
    "Select a demo:" Label.new^
    null 0 DrawFyneApp.demos Select.new^ dup o:.sel!

    "Line style:" Label.new^
    null 0 DrawFyneApp.lineStyles Select.new^ dup o:.lsSel!

    DrawFyneApp.btnRun "Run" DrawFyneApp_makeButton
  ) Box.newH^

  // Border: pixmap in center (takes leftover space); controls stay at the bottom
  null swap
  null null
  o:.view@
  Border.new^

  // create the window, and set the content
  App.app "Final Painter drawing example"
  Window.new^ dup >ts o:.win!
  o:.makeMenuBar^ t:.setMenuBar^
  t:.setContent^
  
  // close handler
  //o:.onWindowClose~ 0 t:.setOnClosed^
  
  o:.sz@.w o:.sz@.h 48 + 36 + // +button bar +menu bar
  t:.resize^
  true t:.setFixedSize^
  t:drop
;

/*
Initialize a new DrawFyneApp.
This does not create the UI.
( -- ) o( inst )
*/
: DrawFyneApp#init
   
  DrawFyneApp.pixW
  DrawFyneApp.pixH Size.new^
  dup o:.sz!
  
  255 Palette4.newWebPalette^ o:.pal!
  // Painter
  Painter.new^ o:.p!
  // Pixmap
  o:.p@.pixmap@ o:.pix! 
;

: draw-demo
  DrawFyneApp.new^ >zs
  z:.init^
  z:.makeUi^
  
  z:.win@.show^

  z:drop
;


draw-demo-blurb $ nl
