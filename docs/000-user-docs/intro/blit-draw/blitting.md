# Raster Blitting

Raster operations are of two basic types:

- pixel and pixel mat operations called "blitting" (Bit BLock Transfer)
- drawing, such as lines, shapes, and text

Final provides a helpful assortment of objects and operations to facilitate both types of operations.

The best way to understand these objects is to read the source code where most of them are put to use, in examples/blit-draw and examples/gui/simgame (Life, Shape Zombies, and Invaders).

## Coordinate System

The coordinate system for both blitting and drawing follows the longstanding game convention of:

- x=0 is the far right, and increases positive to the right
- y=0 is the top left, and increases positive going down

This is a bit different from mathematic coodinates, but it does correspond to things like spreadsheets.

Blitting and matrix access is always done from a reference point of top-left, down to bottom right. 

Most painting operations involving shapes use a center-0 reference, such as the center of a circle or rectangle.

`Pixmap` supports a world transformation `viewportAt` property which enables you to locate the viewport anywhere in an unconstrained drawing space; blits and draws clip automatically.

## Utility Types

There are several utility types used widely across drawing and painting. They are briefly described here. All these types are immutable value types, meaning they can't be modified after creation, and are passed by value, not by pointer. You create a new one when you need new values. Use `help` to view the members and factories provided by each type.

### Point

A 2-dimensional point, using fractional x and y.

| Member | Stack effect | Description |
| ------ | ------------ | ----------- |
| `.x` `.y` | `( -- n:f )` | exact coordinates
| `.xi` `.yi` | `( -- n:i )` | rounded integral coordinates
| `Point.new` | `( x:n y:n -- p:Point )` | create a new Point; accepts any numeric type
| `Point.zero` | `( -- p:Point )` | the 0,0 Point (prefer this to the `new` factory for known 0,0)

Note that most APIs that take a Point in Final will also accept the ( x:n y:n -- ) shape.

Also note that x and y are reversed, coordinate-wise, from the row/col parameters used in things like Mats, so be wary!

### Size

A 2-dimension size, using Float w (width) and h (height).

| Member | Stack effect | Description |
| ------ | ------------ | ----------- |
| `.w` `.h` | `( -- n:f )` | exact width and height |
| `.wi` `.hi` | `( -- n:i )` | rounded integral width and height |
| `Size.new` | `( w:n h:n -- s:Size )` | create a new Size; accepts any numeric type |

Note that most APIs that take a Size in Final will also accept the `( w:n h:n -- )` shape. Discrete sinks such as `Pixmap.new` and `Painter.new` require exact integral dimensions.

### Rect

Describes a rectangular region; comprises a Point topLeft and a Size size.

| Member | Stack effect | Description |
| ------ | ------------ | ----------- |
| `.topLeft` | `( -- p:Point )` | top-left corner |
| `.size` | `( -- s:Size )` | extent (W×H) |
| `Rect.new` | `( topLeft:Point \| x:n y:n  size:Size \| w:n h:n -- r:Rect )` | create a Rect; each slot accepts a value type or a numeric pair |

Examples of the modal factory shapes:

```
pt sz Rect.new^
x y sz Rect.new^
pt w h Rect.new^
x y w h Rect.new^
```

## Blitting

Blitting, stated simply, is moving pixels from A to B, typically from some source, to in Final a destination called a `Pixmap`. A `Pixmap` is like a canvas of pixels, upon which you sucessively build up a final image from a series of simpler primitives. `Pixmap` is used for both Bliting and Drawing, and the two types of operations can be intermixed, with some caveats. A `Pixmap` is created with a specific size, which is immutable for its lifetime; so the canvas size must be set before any drawing operation commence. 

The following sections explain the various "atoms" of the blitting capabilities

### `Color`

A color object is a 32 bit composite of 8 unsigned bits each for Red, Green, Blue, and Alpha. Alpha is how transparent or opaque is the color value, with 0 being completely transparent (and thus not really a color--the other 3 values are irrelevant) and 255 being completely opaque. When you do not specify an A value for a `Color` it is automatically set to be 255/opaque.

Alpha is not meaningful in isolation; it only makes sense when drawing upon a Pixmap that may already have had drawing operations made on it. In this case, if using the `blend` operation, Alpha will determine how much of the new color to mix with the existing values. (Of course opaque stomps over top.)

The `Color` object has constants for some common colors from the Web Palette, and also `translucent` and `mask`, which is used in masked blitting operations.

| Member | Stack effect | Description |
| ------ | ------------ | ----------- |
| `.r` `.g` `.b` `.a` | `( -- n:i )` | Channel values 0–255 (Int); read-only |
| `Color.black` … `Color.aqua` | `( -- c:Color )` | Named Web Palette colors (also `cyan`/`magenta` aliases for aqua/fuchsia) |
| `Color.transparent` | `( -- c:Color )` | Fully transparent (0,0,0,0) |
| `Color.mask` | `( -- c:Color )` | Same as transparent; skip pixel when blit mask is on |
| `Color.rgb` | `( r:i g:i b:i -- c:Color )` | From 0–255 channels; alpha 255 |
| `Color.rgba` | `( r:i g:i b:i a:i -- c:Color )` | From 0–255 channels including alpha |
| `Color.rgbf` | `( r:f g:f b:f -- c:Color )` | From 0.0–1.0 channels; alpha 1.0 |
| `Color.rgbaf` | `( r:f g:f b:f a:f -- c:Color )` | From 0.0–1.0 channels including alpha |
| `Color.hex` | `( packed:i \| hex:s -- c:Color )` | From `0xRRGGBB` Int or hex String (optional `0x`/`0X`); alpha 255 |
| `Color.hexa` | `( packed:i \| hex:s -- c:Color )` | From `0xRRGGBBAA` Int or hex String (optional `0x`/`0X`) |

### `Palette`

A `Palette` is a lookup table, that maps a narrow index value to an actual color. Final supports 2, 4, and 8-bit Palettes. This lets you maintain a source of pixels in a more compact format than full 32-bits, economizing on memory. Although memory is not as scarce as in the 70s and 80s when this technique was widely used, it is still useful in many cases. For example, the `life-sim.fnl` simulation uses 4-bits, 16 values, to encode the age of the cells, and then a palette to hold the actual colors.

The Final types are:

- `Palette1` -- 1 bit, 2 color values 
- `Palette4` -- 4 bits, 16 color values
- `Palette8` -- 8 bits, 256 color values

### Mat2x

A Mat is a tabular structure of a fixed number of rows and columns, established when created. (For details on Mats, see [Mats](../compounds/mats.md)). Four of these types are used in blitting, as shown in this table:


| Mat    | Palette  | Used for                             |
| ------ | -------- | ------------------------------------ |
| Mat2i1 | Palette1 | 2-value color indexes and lookup     |
| Mat2i4 | Palette4 | 16-value color indexes and lookup    |
| Mat2i8 | Palette8 | 256-value color indexes and lookup   |
| Mat2c  | n/a      | Full color operations, import/export |


Mats in Final have matrix compound path accessors:

- @[] ( row:i col:i -- value ) for fetch
- ![] ( value row:i col:i -- ) for store

They also have list-like flat-index accessors; storage is in row-major order.

- @() ( flat:i -- value ) for fetch
- !() ( value flat:i -- ) for store

Flat indexes are useful for setting the entire matrix to a specific value, or sometimes when assigning initial randomized values that are unrelated to row/col.

### BlitSrc

A `BlitSrc` is the source data of a blitting operation. It always holds one of the four Mat types, and for the Palettized ones (1/4/8) the accompanying `PaletteX`. (Mat2c does not use a Palette.) You provide a `BlitSrc` to the `blit` method of `Pixmap`.

### Pixmap

A `Pixmap` is a rectangular blitting and drawing surface whose size is set at creation. Its underlying representation differs from Mat*, but accessing individual pixels uses the same operators. (The underlying representation is based on what the core code for blitting and drawing are most efficient using.)

A typical session with a `Pixmap` involves creating, optionally setting a backgrouind color with `fill`, then performing a series of `blit` or draw operations. Successive operations either write over and blend their content with the existing state of the `Pixmap`. 

A `Pixmap` can be exported to a PNG/JPEG/GIF file, and can be serialized using a Stream for Http operations. It can also be directly visualized using the Fyne `CanvasImage` object (for final-gui binary.)

`Pixmap` has several blitting options, which can be set on the Pixmap, and then optionally overridden for an individual `blit` call:


| Option | Values         | Default | Description                                                                                                                  |
| ------ | -------------- | ------- | ---------------------------------------------------------------------------------------------------------------------------- |
| mask   | on or off      | off     | when on, uses the 0 palette index, or for Mat2c the `mask` color to indicate pixels that should not be written to the Pixmap |
| op     | stomp or blend | stomp   | stomp overwrites the destination pixel; blend mixes according the alpha values of the src and dst colors                      |
| cellPx | Int            | 1       | how many dst pixels to write, per source pixel see life.fnl for an example, where it uses 4 to make each cell a good size    |
