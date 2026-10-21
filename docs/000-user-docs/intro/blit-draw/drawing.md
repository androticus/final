# Drawing

Final provides a set of flexible and powerful drawing operations for things like lines, shapes, and text. It provides options for stroke (lines/shape outlines) and fill (the area inside shapes). Text can be aligned vertically and horizontally to suit the context. A `push` and `pop` facility enables preserving previously set parameters when drawing a set of new values; this is useful when calling higher level drawing constructs composed from many primitives.

## Painter

Drawing is exposed on the `Painter` object, which can take an existing `Pixmap` or can create one automatically. Therefore, blitting and drawing operations can be mixed in any sequence, keeping in mind that new operations work over top of previous ones. The api is briefly summarized below. All the available operations are used in the examples/blit-draw/draw.fnl. Painting is also used in the Shape Zombies and Invaders example applications.

As with blitting, a Painter has a default set of parameters, which can be overriden. Unlike Pixmap, the mechanism of overriding is not on a per api basis, but by way of using `push`, setting the new parameters, using them, then calling `pop`, which restores the previous values.

Use `help Painter` for details on the parameter values needed for each operation. "Rot" indicates an operation that rotates the shape clockwise around its center by the specified degrees.


| API                                  | Description |
| ------------------------------------ | ----------- |
| `pixel` | write a pixel with current `stroke` value |
| `pixelC` | write a pixel with a provided Color |
| `line`                               | Draw a line            |
| `circle`                             | Draw a circle            |
| `ellipse`<br>`ellipseRot`            | Draw an ellipse            |
| `rect`<br>`rectRot`                     | Draw a rectangle            |
| `roundedRect`<br>`roundedRectRot`       | Draw a rounded rectangle            |
| `regularPolygon`<br>`regularPolygonRot` | Draw a regular polygon            |
| `text`<br>`textWrap`                    | Write text, auto-wraps \n;<br>textWrap wraps to a rectangular area            |

Text writing can be done using the following alignments:
* Vertical: Top, Middle, Baseline (line above descenders such as in "g"), Bottom
* Horizontal: Left, Center, Right

Pens stay on the Painter (`stroke`, `fill`, `font`). `push` / `pop` are a pen bracket around whole sessions, not part of the builder.

Contours that are not a stock shape use `pathBuilder`: a singleton session on the Painter.

### PathBuilder

You can use Painter to create paths and solids of arbitrary strokes, using straight lines, bezier curves, arcs, or elliptical curves. These all use a `PathBuilder`.

`Painter.pathBuilder` is created with the Painter. Note that it is **not** in `PainterState` and is not cloned on `push`. It is one of three states: **ready**, **path**, **solid**.


| Word                                                       | From          | Does                                                                 |
| ---------------------------------------------------------- | ------------- | -------------------------------------------------------------------- |
| `startPath`                                                | ready         | Open polyline (chart line). `end` strokes; does **not** `ClosePath`. |
| `startSolid`                                               | ready         | Closed shape. `end` closes, then fill then stroke.                   |
| `moveTo` / `lineTo` / `bezierTo` / `arc` / `ellipticalArc` | path or solid | Append. Point or `x y` (Int, Uint, or Float).                        |
| `end`                                                      | path or solid | Paint, clear the gg path, back to ready.                             |


Do **not** `lineTo` the start point on a solid; `end` closes.

One-shots that would wipe the contour (`line`, `circle`, `ellipse`, `ellipseRot`, `rect`, `rectRot`, `roundedRect`, `roundedRectRot`, `regularPolygon`, `regularPolygonRot`) and `push` / `pop` **panic** unless the builder is **ready**. `pixel` / `pixelC` / `text` / `clear` do not use the path and stay allowed.

Invalid combinations panic with the state in the message (`already in path`, `not started`, `call end first`).

#### Good

Park the builder, start, emit, `end`, drop. Mix one-shots only while **ready**. Reuse the same object after `end`.

```forth
// Open stroke (chart / polyline)
p.pathBuilder >ts
  t:.startPath^
  10 10 t:.moveTo^
  80 10 t:.lineTo^
  80 80 t:.lineTo^
  t:.end^
t:drop

// Closed fill + stroke — do not repeat the start point
FillStyle.solid p.fill.style!
Color.aqua p.fill.color!
p.pathBuilder >ts
  t:.startSolid^
  16 16 t:.moveTo^
  80 16 t:.lineTo^
  48 72 t:.lineTo^
  t:.end^
t:drop

// Fractional control points (flex numeric)
t:.startPath^
10 40 t:.moveTo^
40.5 8  90 40 t:.bezierTo^
t:.end^
```

#### Don'ts

 `p.line` (or `circle` / `ellipse` / `rect` / `roundedRect` / `regularPolygon` / `*Rot` / `push`) after `startPath` and before `end`. **Don't:** `end` with no `start`*. **Don't:** `startSolid` while already in **path**.

Temporary pens for a series of sessions: `push`, set `stroke` / `fill`, several `start*`…`end` (and one-shots), then `pop`.

## Reading/Writing to files or streams

Both `Mat2c` and `Pixmap` have methods for saving to PNG, which is lossless, or JPEG/GIF which are lossy. It is recommended to using PNG.

They also both have a factory method for creating an instance from a file. The newFromFile api automatically determines the file format.