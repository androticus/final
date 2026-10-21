forget: reset-mouse-fyne

requiresFyne
// `examples/gui/fyne/mouse-fyne.fnl` compileFromFile
`
CanvasImage mouse events, cursor shapes, and scribble.

Select a cursor, move and click on the tiled logo, and watch
the event log. Drag with the mouse button down to paint.

Try:

mouse-fyne
`
const: mouse-fyne-blurb


{
   {
     Size "sz"
     Painter "p"
     Pixmap "pix"
     CanvasImage "view"
     Select "sel"
     Entry "logEntry"
     LineBufferTextWriter "log"
     Window "win"
     Bool "drawing"
   } "fields"
   
   {
     432 "pixH"
     576 "pixW"
     4 "brushR"
     80 "logMaxLines"
     260 "logMinW"
     (
       "Default"
       "Text"
       "Crosshair"
       "Pointer"
       "HResize"
       "VResize"
       "NESWResize"
       "NWSEResize"
       "Hidden"
     s) "cursorNames"
     (
       CanvasImage.cursorDefault
       CanvasImage.cursorText
       CanvasImage.cursorCrosshair
       CanvasImage.cursorPointer
       CanvasImage.cursorHResize
       CanvasImage.cursorVResize
       CanvasImage.cursorNESWResize
       CanvasImage.cursorNWSEResize
       CanvasImage.cursorHidden
     ) "cursorVals"
   } "constants"
     
  "MouseFyneApp" "name"
}
TypedObject.makeType^


/*
Format a MouseEvent as one log line.
( ev:MouseEvent kind:s -- line:s )
*/
: MouseFyneApp_fmtMouse
  >ts
  >xs
  ( t:peek x:.position.xi@ x:.position.yi@ x:.button@ x:.modifier@ )
  "%s (%d,%d) btn=%d mod=%d" fmt
  x:drop
  t:drop
;

/*
Append a line to the capped log and blit the Entry.
( line:s -- ) o( inst )
*/
: MouseFyneApp#appendLog
  o:.log@.writeLine^
  o:.log@.toString^
  o:.logEntry@.text!
;

/*
Paint a scribble blob at the mouse position if in bounds.
( ev:MouseEvent -- ) o( inst )
*/
: MouseFyneApp#paintAt
  >ts
  t:.position.xi@ l.x!
  t:.position.yi@ l.y!
  l.x@ 0 <
  l.x@ o:.pix@.cols@ >= ||
  l.y@ 0 < ||
  l.y@ o:.pix@.rows@ >= || if
    t:drop
  else
    t:.position@
    MouseFyneApp.brushR o:.p@.circle^
    o:.view@.refresh^
    t:drop
  then
;


// --- event handlers ( ev:Event -- ) o( inst ) ---

: MouseFyneApp#onSelectChanged
  drop
  o:.sel@.selectedIndex@
  dup 0 < if
    drop
  else
    MouseFyneApp.cursorVals >ts t:@() t:drop
    o:.view@.cursorShape!
  then
;

: MouseFyneApp#onMouseDown
  >ts
  0 t:.params@ >xs x:@() x:drop
  dup "down" MouseFyneApp_fmtMouse
  o:.appendLog^
  true o:.drawing!
  o:.paintAt^
  t:drop
;

: MouseFyneApp#onMouseUp
  >ts
  0 t:.params@ >xs x:@() x:drop
  "up" MouseFyneApp_fmtMouse
  o:.appendLog^
  false o:.drawing!
  t:drop
;

: MouseFyneApp#onMouseIn
  >ts
  0 t:.params@ >xs x:@() x:drop
  "in" MouseFyneApp_fmtMouse
  o:.appendLog^
  t:drop
;

: MouseFyneApp#onMouseMoved
  >ts
  0 t:.params@ >xs x:@() x:drop
  dup "move" MouseFyneApp_fmtMouse
  o:.appendLog^
  o:.drawing@ if
    o:.paintAt^
  else
    drop
  then
  t:drop
;

: MouseFyneApp#onMouseOut
  drop
  "out" o:.appendLog^
  false o:.drawing!
;

: MouseFyneApp#onMouseScrolled
  >ts
  0 t:.params@ >xs x:@() x:drop
  >xs
  ( x:.position.xi@ x:.position.yi@ x:.scrolled.xi@ x:.scrolled.yi@ )
  "scroll (%d,%d) d=(%d,%d)" fmt
  o:.appendLog^
  x:drop
  t:drop
;


/*
Fill the pixmap with the tiled logo, then set navy fill for scribble.
( -- ) o( inst )
*/
: MouseFyneApp#tileBackground
  o:.p@ >ts
  t:.fill@ >xs
  FillStyle.tile x:.style!
  "examples/gui/fyne/favicon-32-f-circle.png" Image.newFromFile^
  x:.tileImage!
  2 x:.tileZoom!
  LineStyle.none t:.stroke.lineStyle!
  o:.sz.w 2 / o:.sz.h 2 / Point.new^
  o:.sz@
  t:.rect^
  FillStyle.solid x:.style!
  Color.navy x:.color!
  LineStyle.solid t:.stroke.lineStyle!
  Color.navy t:.stroke.color!
  x:drop
  t:drop
;

/*
Wire CanvasImage mouse handlers.
( -- ) o( inst )
*/
: MouseFyneApp#wireMouse
  o:.view@ >ts
  o:.onMouseDown~ 0
  o:.onMouseUp~ 0
  t:.setOnMouseDownUp^
  o:.onMouseIn~ 0
  o:.onMouseMoved~ 0
  o:.onMouseOut~ 0
  t:.setOnMouseInMovedOut^
  o:.onMouseScrolled~ 0
  t:.setOnMouseScrolled^
  t:drop
;

/*
Create the UI
( -- ) o( inst )
*/
: MouseFyneApp#makeUi
  o:.pix@.image CanvasImage.new^ dup o:.view!
  o:.wireMouse^

  o:.onSelectChanged~ 0 MouseFyneApp.cursorNames Select.new^
  dup o:.sel!
  0 o:.sel@.selectedIndex!

  (
    "Select a Cursor:" Label.new^
    o:.sel@
  ) Box.newH^
  l.top!

  Entry.newMultiLine^ dup o:.logEntry!
  true o:.logEntry@.disabled!
  Scroll.new^ >ts
  MouseFyneApp.logMinW 200 t:.setMinSize^
  ts> l.logScroll!

  (
    "Mouse Events" Label.new^
    l.logScroll@
  ) Box.newV^
  l.right!

  l.top@
  null
  null
  l.right@
  o:.view@
  Border.new^

  App.app "CanvasImage mouse demo"
  Window.new^ dup >ts o:.win!
  t:.setContent^
  o:.sz@.w MouseFyneApp.logMinW +
  o:.sz@.h 48 +
  t:.resize^
  true t:.setFixedSize^
  t:drop
;

/*
Initialize a new MouseFyneApp.
This does not create the UI.
( -- ) o( inst )
*/
: MouseFyneApp#init
  MouseFyneApp.pixW
  MouseFyneApp.pixH Size.new^
  dup o:.sz!
  Painter.new^ o:.p!
  o:.p@.pixmap@ o:.pix!
  MouseFyneApp.logMaxLines LineBufferTextWriter.new^ o:.log!
  false o:.drawing!
  o:.tileBackground^
;

: mouse-fyne
  MouseFyneApp.new^ >zs
  z:.init^
  z:.makeUi^
  z:.win@.show^
  z:drop
;


mouse-fyne-blurb $ nl
