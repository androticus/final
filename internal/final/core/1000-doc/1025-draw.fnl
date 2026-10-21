// Drawing adjuncts: Point, Size, Rect, Font; Painter

/*
Continuous (x, y) in drawing space (Y grows down). Stored as float64.
.x / .y are Float; .xi / .yi are rounded Int for pixel-grid work.
Methods that take a Point also accept two numerics (Int, Uint, or Float).
Point is an immutable value type.
*/
:doc Point

/*
Create a Point from any numeric pair.
( x:n y:n -- p:Point )
*/
:doc Point#new

/*
X coordinate (Float).
*/
:doc Point#x

/*
Y coordinate (Float).
*/
:doc Point#y

/*
Rounded integral X (Int).
*/
:doc Point#xi

/*
Rounded integral Y (Int).
*/
:doc Point#yi

/*
Continuous W×H in drawing space (columns × rows). Stored as float64.
.w / .h are Float; .wi / .hi are rounded Int for pixel-span samples.
Methods that take a Size also accept two numerics (Int, Uint, or Float).
Discrete sinks (Painter.new, Pixmap.new, pixSize) require exact integral dims.
Size is an immutable value type.
*/
:doc Size

/*
Create a Size from any numeric pair.
( w:n h:n -- s:Size )
*/
:doc Size#new

/*
Width (Float).
*/
:doc Size#w

/*
Height (Float).
*/
:doc Size#h

/*
Rounded integral width (Int).
*/
:doc Size#wi

/*
Rounded integral height (Int).
*/
:doc Size#hi

/*
Axis-aligned rectangle: top-left Point plus Size.
Factory slots are polymorphic (Point or x y; Size or w h).
Rect is an immutable value type.
*/
:doc Rect

/*
Create a Rect.
( topLeft:Point size:Size -- r:Rect )
*/
:doc Rect#new

/*
Top-left corner.
*/
:doc Rect#topLeft

/*
Extent (W×H).
*/
:doc Rect#size

/*
RGBA color value (each channel is uint8).
Final caches the pre-multiplied channel values
for blitting and drawing.
Color is an immutable value type.
*/
:doc Color

/*
Red channel (0–255).
*/
:doc Color#r

/*
Green channel (0–255).
*/
:doc Color#g

/*
Blue channel (0–255).
*/
:doc Color#b

/*
Alpha channel (0–255).
*/
:doc Color#a

/*
The all-zero sentinel color when using the Blit mask option.
Skips the blit operation for the pixel.
*/
:doc Color#mask

/*
Create a Color from red, green, blue (alpha=255).
( r:i g:i b:i -- c:Color )
*/
:doc Color#rgb

/*
Create a Color from red, green, blue, alpha.
( r:i g:i b:i a:i -- c:Color )
*/
:doc Color#rgba

/*
Create a Color from a packed 0xRRGGBB int (alpha=255) or hex string.
The value can be an Int; hex notation is the usual way to write it.
The value can be a hex string, with optional 0x/0X prefix.
( packed:i | hex:s -- c:Color )
*/
:doc Color#hex

/*
Create a Color from a packed 0xRRGGBBAA int or hex string.
The value can be an Int; hex notation is the usual way to write it.
The value can be a hex string, with optional 0x/0X prefix.
( packed:i | hex:s -- c:Color )
*/
:doc Color#hexa

/*
Create a Color from fractional red, green, blue (alpha=1.0).
0.0 <= val <= 1.0
( r:f g:f b:f -- c:Color )
*/
:doc Color#rgbf

/*
Create a Color from fractional red, green, blue, alpha.
0.0 <= val <= 1.0
( r:f g:f b:f a:f -- c:Color )
*/
:doc Color#rgbaf

/*
Two colors (1-bit / blit1). Indexes 0 and 1. @() / !().
new^ is empty (zeros); newWithColors^ takes c0 c1.
*/
:doc Palette1

/*
Empty Palette1 (2 zero Color slots). Fill with !() or newWithColors^.
( -- pal:Palette1 )
*/
:doc Palette1#new

/*
Palette1 from two Color values.
( c0:Color c1:Color -- pal:Palette1 )
*/
:doc Palette1#newWithColors

/*
16 colors (4-bit / blit4). @() / !(). new^ is empty (zeros).
newFromList^ copies a Color list (slots past the list stay zero).
newWebPalette^ fills the 16 CSS/VGA names at a given alpha.
*/
:doc Palette4

/*
Empty Palette4 (16 zero Color slots).
( -- pal:Palette4 )
*/
:doc Palette4#new

/*
Palette4 from a Color list (slots past the list stay zero).
( colors:List -- pal:Palette4 )
*/
:doc Palette4#newFromList

/*
The 16 named web colors as a Palette4; every slot uses straight alpha a.
( a:i -- pal:Palette4 )
*/
:doc Palette4#newWebPalette

/*
256 colors (8-bit / blit8). @() / !(). Pair with Mat2u8.
new^ is empty (zeros); fill slots with !().
new332Palette^ builds the classic RRRGGGBB table at a given alpha.
*/
:doc Palette8

/*
Empty Palette8 (256 zero Color slots).
( -- pal:Palette8 )
*/
:doc Palette8#new

/*
3-3-2 RGB palette (RRRGGGBB); every slot uses straight alpha a.
( a:i -- pal:Palette8 )
*/
:doc Palette8#new332Palette

/*
Raster surface (write-only). Store is Go *image.RGBA premultiplied bytes.
Author ink is Color; !() / ![] write cells. No @() / @[] readback.
image.Image At returns premul color.RGBA for codecs/Fyne.

Create:  rows cols Pixmap.new^
Shape:   .rows .cols  (Painter pixmap: rows = Size.h, cols = Size.w)
Write:   !() / ![]  — flat or ( row col ) stores
Blit:    .params (BlitParams base); blit / blitWithOverrides; blit1 / blit4 / blit8 / blitC
Wipe:    clear (zeros) / fill (color). Blit does not clear.
*/
:doc Pixmap

/*
Create an RGBA Pixmap.
( rows:i cols:i -- pm:Pixmap )
*/
:doc Pixmap#new

/*
Number of rows. Readonly.
*/
:doc Pixmap#rows

/*
Number of columns. Readonly.
*/
:doc Pixmap#cols

/*
Go image.Image for Fyne / codecs. Readonly.
*/
:doc Pixmap#image

/*
Wrap row/col indexes onto the torus.
( row:i col:i -- row:i col:i )
*/
:doc Pixmap#wrap

/*
Write-only !() / ![] (Color cells). No @() / @[] readback.
*/
:doc Pixmap#__items

/*
Set every pixel to transparent. (Note: Blit does not clear.)
( -- )
*/
:doc Pixmap#clear

/*
Set every pixel to a color.
( c:Color -- )
*/
:doc Pixmap#fill

/*
Concrete BlitParams base for this Pixmap (stomp, cellPx 1).
Readonly property; change policy via field setters (Inherit not allowed).
*/
:doc Pixmap#params

/*
Stamp BlitSrc using Pixmap.params (no override). Intersection-clips to pixmap.
( src:BlitSrc at:Point -- )
*/
:doc Pixmap#blit

/*
Stamp BlitSrc after resolving over against Pixmap.params. Intersection-clips to pixmap.
( src:BlitSrc over:BlitParams at:Point -- )
*/
:doc Pixmap#blitWithOverrides

/*
Blit Mat2i1 through Palette1. Intersection-clips to pixmap.
( pal:Palette1 mat:Mat2i1 at:Point -- )
The Point slot also accepts two Ints (x y).
*/
:doc Pixmap#blit1

/*
Blit Mat2i4 through Palette4. Intersection-clips to pixmap.
( pal:Palette4 mat:Mat2i4 at:Point -- )
The Point slot also accepts two Ints (x y).
*/
:doc Pixmap#blit4

/*
Blit Mat2u8 through Palette8. Intersection-clips to pixmap.
( pal:Palette8 mat:Mat2u8 at:Point -- )
The Point slot also accepts two Ints (x y).
*/
:doc Pixmap#blit8

/*
Blit Mat2c (full-color cells). Intersection-clips to pixmap.
( mat:Mat2c at:Point -- )
The Point slot also accepts two Ints (x y).
*/
:doc Pixmap#blitC

/*
Snapshot buffer into a new Mat2c (author/read model). Premul pixels are
un-premultiplied; A=0 cells become transparent Color (0,0,0,0).
( -- mat:Mat2c )
*/
:doc Pixmap#toMat2C

/*
Per-pixel blit composite. Class consts: blend (0, Go Over), stomp (1, Go Src).
*/
:doc BlitType

/*
Per-source blit override bag. Factory new: all ParamUseDefault (-1 inherit from Pixmap.blit).
Setters only accept concrete values (not Inherit). Class consts: ParamUseDefault, MaskOn,
MaskOff, OpStomp, OpBlend. Dest overhang is always intersection-clipped.
Resolve is internal.
*/
:doc BlitParams

/*
Validated mat+palette (or Mat2c) pairing for blit.
Factories: new1 (Palette1+Mat2i1), new4, new8, newC (Mat2c, no pal).
Readonly .mat .pal; setMat / setPal keep the same depth kind. Change depth via a new BlitSrc.
*/
:doc BlitSrc


/*
Raster drawing context. Factory new mints an RGBA Pixmap (gg-compatible)
and binds to it; newWithPixmap binds an existing Pixmap. pixmap is writable
(rebind via .pixmap!); stroke, fill, and font are readonly adjuncts;
mutate those objects, do not replace them.
pathBuilder is a singleton session for contours (startPath / startSolid / end).
Default stroke: black, lineWidth 1, cap round, join round, style solid.
Default fill: style none (outline-only until Fill.style is solid).
Default font: Font.go at 12pt, black.
*/
:doc Painter

/*
Create a Painter for a positive Size (or w h). Mints a new Pixmap.
( size:Size | w:i h:i -- p:Painter )
*/
:doc Painter#new

/*
Create a Painter bound to an existing Pixmap (no new buffer). Panics if pm is null.
( pm:Pixmap -- p:Painter )
*/
:doc Painter#newWithPixmap

/*
The Pixmap this Painter draws on. Read with .pixmap; rebind with .pixmap!
(painter.pixmap! pm). Panics if pm is null.
*/
:doc Painter#pixmap

/*
Stroke pen (readonly). Mutate .color / .lineWidth / .lineCap / .lineStyle.
*/
:doc Painter#stroke

/*
Fill pen (readonly). Mutate .color / .style.
*/
:doc Painter#fill

/*
Font on this Painter (readonly). Mutate .family / .size / .color on it.
*/
:doc Painter#font

/*
Singleton path session (readonly). Not cloned by push.
startPath for open strokes; startSolid for closed fill+stroke; end to paint.
One-shots (line, circle, ellipse, ellipseRot, rect, rectRot, roundedRect, roundedRectRot, regularPolygon, regularPolygonRot) and push/pop panic unless ready.
*/
:doc Painter#pathBuilder

/*
Wipe the pixmap to transparent zeros (same buffer as new).
Does not replace pixmap; does not change Stroke/Fill/Font.
( -- )
*/
:doc Painter#clear

/*
Save Stroke/Fill/Font and gg drawing state. Nested; pair with pop.
After push, .stroke / .fill / .font are new objects (value copies).
Do not keep a parked adjunct across push/pop.
( -- )
*/
:doc Painter#push

/*
Restore the last pushed Stroke/Fill/Font and gg drawing state.
Panics if there is no matching push.
( -- )
*/
:doc Painter#pop

/*
Write one pixel on the pixmap using stroke.color (straight Color → RGBA).
No-op if stroke.lineStyle is none. Panics if the point is out of range.
Does not change Stroke/Fill/Font. Not a gg DrawPoint.
( at:Point -- )
The Point slot also accepts two Ints (x y).
*/
:doc Painter#pixel

/*
Write one pixel with an explicit color (not stroke.color). Always writes.
Color is deepest so it can be dup'd for several pokes:
  c dup x1 y1 pixelC  dup x2 y2 pixelC  x3 y3 pixelC
Panics if the point is out of range. Does not change Stroke/Fill/Font.
( c:Color at:Point -- )
The Point slot also accepts two Ints (x y).
*/
:doc Painter#pixelC

/*
Draw a line (stroke). No-op if stroke.lineStyle is none.
( from:Point to:Point -- )
Each Point slot also accepts two Ints (x y).
*/
:doc Painter#line

/*
Draw a circle. Fill then stroke when those pens are live.
( ctr:Point radius:i -- )
*/
:doc Painter#circle

/*
Draw an axis-aligned ellipse (gg DrawEllipse).
Fill then stroke when those pens are live.
Radii accept Int, Uint, or Float (→ float64 at the gg boundary).
( ctr:Point rx:n ry:n -- )
*/
:doc Painter#ellipse

/*
Same geometry as ellipse, plus clockwise rotationDegrees about center.
Zero rotation delegates to ellipse.
( ctr:Point rx:n ry:n rotationDegrees:n -- )
*/
:doc Painter#ellipseRot

/*
Draw a rectangle centered at ctr with the given Size.
Fill then stroke when those pens are live.
( ctr:Point size:Size -- )
*/
:doc Painter#rect

/*
Same geometry as rect, plus clockwise rotationDegrees about center.
Zero rotation delegates to rect.
( ctr:Point size:Size rotationDegrees:n -- )
*/
:doc Painter#rectRot

/*
Draw a rounded rectangle centered at ctr (gg DrawRoundedRectangle; one radius for all corners).
Fill then stroke when those pens are live.
( ctr:Point size:Size radius:n -- )
*/
:doc Painter#roundedRect

/*
Same geometry as roundedRect, plus clockwise rotationDegrees about center.
Zero rotation delegates to roundedRect.
( ctr:Point size:Size radius:n rotationDegrees:n -- )
*/
:doc Painter#roundedRectRot

/*
Draw a regular polygon (no rotation; 0 = point-up odd n-gon / flat-top even n-gon).
Fill then stroke when those pens are live.
( sides:i center:Point radius:n -- )
*/
:doc Painter#regularPolygon

/*
Same geometry as regularPolygon, plus clockwise rotationDegrees.
( sides:i center:Point radius:n rotationDegrees:n -- )
*/
:doc Painter#regularPolygonRot

/*
Typeface for Painter text. Mutate the Font on Painter (family, size, color, align);
do not replace Painter.font.
Default family is Font.go; default size is 12 points; default color is black;
default align is TextAlign.centerMiddle.
Built-in gofont families only; loading a file face is a later release.
*/
:doc Font

/*
Family name (one of Font.names). Assign a class const, e.g. Font.goBold.
*/
:doc Font#family

/*
Point size (Int, Uint, or Float). Must be positive.
*/
:doc Font#size

/*
Glyph ink. Applied at text time (not on setter).
*/
:doc Font#color

/*
Placement pose (TextAlign). One horizontal flag ORed with one vertical.
Default TextAlign.centerMiddle. Incomplete or conflicting flags panic.
*/
:doc Font#align

/*
Go Regular (default).
*/
:doc Font#go

/*
Go Bold.
*/
:doc Font#goBold

/*
Go Italic.
*/
:doc Font#goItalic

/*
Go Bold Italic.
*/
:doc Font#goBoldItalic

/*
Go Mono.
*/
:doc Font#goMono

/*
Go Small Caps.
*/
:doc Font#goSmallCaps

/*
Known family names (builtins).
( -- names:StringList )
*/
:doc Font#names

/*
Draw text at a Point. Font.align names which point of the ink box sits on
the origin (default centerMiddle). A newline is a hard line break; each line
uses the same horizontal align. Uses font.color. Y down.
( s:s at:Point -- )
The Point slot also accepts two Ints (x y).
*/
:doc Painter#text

/*
Word-wrap s to width, then place like text (same Font.align).
Width must be positive (Int, Uint, or Float).
( s:s at:Point width:n -- )
The Point slot also accepts two Ints (x y).
*/
:doc Painter#textWrap

/*
Outline pen on a Painter. No Stroke.new; use painter.stroke.
*/
:doc Stroke

/*
Stroke ink. Applied immediately before Stroke() (not on setter).
*/
:doc Stroke#color

/*
Stroke width in pixels (flex number). Must be positive. Flows through to gg.
*/
:doc Stroke#lineWidth

/*
End cap: LineCap.round (default), .butt, .square. Flows through to gg.
*/
:doc Stroke#lineCap

/*
Corner join: LineJoin.round (default), .bevel. Flows through to gg.
*/
:doc Stroke#lineJoin

/*
LineStyle.solid (default), .none (skip Stroke), .dot, .dash, .dotDash.
Dash patterns scale with lineWidth.
*/
:doc Stroke#lineStyle

/*
Interior pen on a Painter. No Fill.new; use painter.fill.
Default style is none (no interior). Styles: none, solid, checkerboard, gradient, tile.
*/
:doc Fill

/*
Fill ink. Solid; checker A; gradient 0%.
*/
:doc Fill#color

/*
Second ink. Checker B; gradient 100%. Default white.
*/
:doc Fill#color2

/*
FillStyle.none (default), .solid, .checkerboard, .gradient, .tile.
*/
:doc Fill#style

/*
Checker tile size in pixels. Default 8. Must be > 0.
*/
:doc Fill#checkerSize

/*
Tile source (Image or Pixmap). Tiling origin is pixmap (0,0); the path clips.
*/
:doc Fill#tileImage

/*
Tile scale. Default 1. Values other than 1 pre-scale (nearest neighbor) then tile.
*/
:doc Fill#tileZoom

/*
Place 0% and 100% in pixmap space. Path still clips the fill.
( from:Point to:Point -- )
*/
:doc Fill#setGradientFromTo

/*
Sugar: gradient along the Rect's topLeft → opposite corner.
( r:Rect -- )
*/
:doc Fill#setGradientRect

/*
Direction. 0° = right, 90° = down. Resolved at fill from the path AABB.
( deg:n -- )
*/
:doc Fill#setGradientDegrees

/*
Direction as a vector (Point or dx dy). Length ignored. AABB at fill time.
( dir:Point | dx dy -- )
*/
:doc Fill#setGradientDir

/*
Placement flags for Painter.text. OR one horizontal with one vertical, or use a
composed name. Class consts are Int so `|` works. Assign to Font.align.
Horizontal: left, center, right. Vertical: top, middle, baseline, bottom.
*/
:doc TextAlign

/*
Left edge of the run (or each line) on the origin.
*/
:doc TextAlign#left

/*
Horizontal center of the run on the origin.
*/
:doc TextAlign#center

/*
Right edge of the run on the origin.
*/
:doc TextAlign#right

/*
Top of the em-box on the origin.
*/
:doc TextAlign#top

/*
Vertical center of the em-box on the origin.
*/
:doc TextAlign#middle

/*
First-line typographic baseline on the origin.
*/
:doc TextAlign#baseline

/*
Bottom of the em-box (descender line) on the origin.
*/
:doc TextAlign#bottom

/*
left | top
*/
:doc TextAlign#leftTop

/*
left | middle
*/
:doc TextAlign#leftMiddle

/*
left | baseline
*/
:doc TextAlign#leftBaseline

/*
left | bottom
*/
:doc TextAlign#leftBottom

/*
center | top
*/
:doc TextAlign#centerTop

/*
center | middle (Font.align default)
*/
:doc TextAlign#centerMiddle

/*
center | baseline
*/
:doc TextAlign#centerBaseline

/*
center | bottom
*/
:doc TextAlign#centerBottom

/*
right | top
*/
:doc TextAlign#rightTop

/*
right | middle
*/
:doc TextAlign#rightMiddle

/*
right | baseline
*/
:doc TextAlign#rightBaseline

/*
right | bottom
*/
:doc TextAlign#rightBottom

/*
Stroke end cap. Class consts: round, butt, square.
*/
:doc LineCap

/*
Stroke corner join. Class consts: round, bevel.
*/
:doc LineJoin

/*
Stroke style. Class consts: none, solid, dot, dash, dotDash.
dot / dash / dotDash scale with lineWidth.
*/
:doc LineStyle

/*
Fill style. Class consts: none, solid, checkerboard, gradient, tile.
*/
:doc FillStyle

/*
Path session on a Painter (painter.pathBuilder). One singleton per Painter.
startPath: open polyline, stroked on end (no ClosePath).
startSolid: closed shape; end does ClosePath then fill then stroke.
Do not lineTo the start point; startSolid's end closes.
Vertex slots accept Point or x y (Int, Uint, or Float).
*/
:doc PathBuilder

/*
Begin an open path (chart line). Panics if not ready.
( -- )
*/
:doc PathBuilder#startPath

/*
Begin a closed solid. Panics if not ready.
( -- )
*/
:doc PathBuilder#startSolid

/*
Paint the session and return to ready. Panics if not started.
path: stroke (LineStyle.none skips). solid: ClosePath + fillThenStroke.
( -- )
*/
:doc PathBuilder#end

/*
Move the current point. Panics if not in a session.
( at:Point | x y -- )
*/
:doc PathBuilder#moveTo

/*
Line to a point. Panics if not in a session.
( at:Point | x y -- )
*/
:doc PathBuilder#lineTo

/*
Quadratic bezier to end, with control. Panics if not in a session.
( control end -- ) each slot Point or x y
*/
:doc PathBuilder#bezierTo

/*
Arc about center. Angles in degrees. Continues the current point (LineTo to the
arc start if live). Panics if not in a session.
( center radius:n angle1:n angle2:n -- )
*/
:doc PathBuilder#arc

/*
Elliptical arc. Angles in degrees. Panics if not in a session.
( center rx:n ry:n angle1:n angle2:n -- )
*/
:doc PathBuilder#ellipticalArc

/*
Draw lines, shapes, and text on a small display.
Create one from a 1-, 4-, or 8-color BlitSrc.
Colors are palette numbers, held in ink.
The origin is the top-left corner, and y grows downward.
Marks that fall outside the picture are ignored.
*/
:doc Minipaint

/*
Create a Minipaint for a 1-, 4-, or 8-color BlitSrc.
( src:BlitSrc -- mp:Minipaint )
*/
:doc Minipaint#new

/*
The BlitSrc this Minipaint draws on.
*/
:doc Minipaint#blitSrc

/*
Color number used for lines, shapes, and text.
Starts at 1.
*/
:doc Minipaint#ink

/*
Color number drawn behind text.
Starts at 0.
Set this equal to ink to leave the background unchanged.
*/
:doc Minipaint#bg

/*
Column where the next character is drawn.
*/
:doc Minipaint#cursorX

/*
Row where the next character is drawn.
*/
:doc Minipaint#cursorY

/*
How many times to enlarge each character.
1 draws the built-in size, a cell 6 pixels wide and 8 pixels tall.
*/
:doc Minipaint#textSize

/*
When true, text that would pass the right edge continues on the next line.
Starts out true.
*/
:doc Minipaint#wrap

/*
Width of the picture, in pixels.
*/
:doc Minipaint#width

/*
Height of the picture, in pixels.
*/
:doc Minipaint#height

/*
Set one pixel to a palette color.
( x:i y:i c:i -- )
*/
:doc Minipaint#putPixel

/*
Set one pixel to the current ink.
( x:i y:i -- )
*/
:doc Minipaint#pixel

/*
Draw a straight line in the current ink.
( x0:i y0:i x1:i y1:i -- )
*/
:doc Minipaint#drawLine

/*
Draw a vertical line in the current ink, h pixels down from x y.
( x:i y:i h:i -- )
*/
:doc Minipaint#drawFastVLine

/*
Draw a horizontal line in the current ink, w pixels right from x y.
( x:i y:i w:i -- )
*/
:doc Minipaint#drawFastHLine

/*
Draw a rectangle outline in the current ink.
x y is the top-left corner.
( x:i y:i w:i h:i -- )
*/
:doc Minipaint#drawRect

/*
Fill a rectangle with the current ink.
x y is the top-left corner.
( x:i y:i w:i h:i -- )
*/
:doc Minipaint#fillRect

/*
Fill the whole picture with the current ink.
( -- )
*/
:doc Minipaint#fillScreen

/*
Draw a circle outline in the current ink.
x y is the center.
( x:i y:i r:i -- )
*/
:doc Minipaint#drawCircle

/*
Fill a circle with the current ink.
x y is the center.
( x:i y:i r:i -- )
*/
:doc Minipaint#fillCircle

/*
Draw an ellipse outline in the current ink.
x y is the center.
( x:i y:i rx:i ry:i -- )
*/
:doc Minipaint#drawEllipse

/*
Fill an ellipse with the current ink.
x y is the center.
( x:i y:i rx:i ry:i -- )
*/
:doc Minipaint#fillEllipse

/*
Draw a triangle outline in the current ink.
( x0:i y0:i x1:i y1:i x2:i y2:i -- )
*/
:doc Minipaint#drawTriangle

/*
Fill a triangle with the current ink.
( x0:i y0:i x1:i y1:i x2:i y2:i -- )
*/
:doc Minipaint#fillTriangle

/*
Draw a rounded rectangle outline in the current ink.
x y is the top-left corner, and r is the corner radius.
( x:i y:i w:i h:i r:i -- )
*/
:doc Minipaint#drawRoundRect

/*
Fill a rounded rectangle with the current ink.
x y is the top-left corner, and r is the corner radius.
( x:i y:i w:i h:i r:i -- )
*/
:doc Minipaint#fillRoundRect

/*
Place the text cursor at x y.
( x:i y:i -- )
*/
:doc Minipaint#setCursor

/*
Draw one character at x y, using ink, bg, and textSize.
Does not move the cursor.
( x:i y:i c:i -- )
*/
:doc Minipaint#drawChar

/*
Draw one character at the cursor, then move right one character.
A newline moves to the start of the next line.
( c:i -- )
*/
:doc Minipaint#writeChar

/*
Draw text at the cursor, and leave the cursor after the last character.
A newline moves to the start of the next line.
( s:s -- )
*/
:doc Minipaint#print

