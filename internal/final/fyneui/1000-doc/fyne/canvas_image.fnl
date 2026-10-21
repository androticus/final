// Docs for CanvasImage control

/*
An image on the canvas, from a Pixmap or Image.
Supports mouse, hover, scroll, and cursor shape.
Also has Canvas.visible and Canvas.refresh.
*/
:doc CanvasImage

/*
The current cursor shape.
Set it with a CanvasImage.cursor* constant.
*/
:doc CanvasImage#cursorShape

/*
Default arrow cursor.
*/
:doc CanvasImage#cursorDefault

/*
Text I-beam cursor.
*/
:doc CanvasImage#cursorText

/*
Crosshair cursor.
*/
:doc CanvasImage#cursorCrosshair

/*
Pointer (hand) cursor.
*/
:doc CanvasImage#cursorPointer

/*
Horizontal resize cursor.
*/
:doc CanvasImage#cursorHResize

/*
Vertical resize cursor.
*/
:doc CanvasImage#cursorVResize

/*
Northeast–southwest resize cursor.
*/
:doc CanvasImage#cursorNESWResize

/*
Northwest–southeast resize cursor.
*/
:doc CanvasImage#cursorNWSEResize

/*
Hidden cursor.
*/
:doc CanvasImage#cursorHidden

/*
Redraw after updating the underlying image.
Ordinary setters already refresh; use this after mutating a Pixmap.
( -- )
*/
:doc CanvasImage#refresh

/*
Set the mouse down and up handlers.
Each handler is ( ev:Event -- ); Event.params[0] is the MouseEvent.
Pass nil as a handler to skip that event.
( down:MethodRef downIndex:i up:MethodRef upIndex:i -- )
*/
:doc CanvasImage#setOnMouseDownUp

/*
Set the mouse in, move, and out handlers.
In and move: ( ev:Event -- ); Event.params[0] is the MouseEvent.
Out: ( ev:Event -- ) with empty params.
Pass nil as a handler to skip that event.
( in:MethodRef inIndex:i move:MethodRef moveIndex:i
  out:MethodRef outIndex:i -- )
*/
:doc CanvasImage#setOnMouseInMovedOut

/*
Set the mouse-wheel handler.
The handler is ( ev:Event -- ); Event.params[0] is the ScrollEvent.
Pass nil as handler to skip the event.
( scroll:MethodRef index:i -- )
*/
:doc CanvasImage#setOnMouseScrolled

/*
Create a CanvasImage from a Pixmap or Image.
( img -- img:CanvasImage )
*/
:doc CanvasImage#new
