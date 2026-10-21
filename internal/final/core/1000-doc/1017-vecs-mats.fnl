// Linear Algebra — VecNf, Mat2f, and vec:: / mat:: ops
// See also: list-vec-mat ; design notes 160 / 160.1

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

/*
Dense float64/Float 0-init vector of length N set in new^.
Not a growable List (use FloatList for that.)
Index i is typically aligned with matrix rows,
paired vectors, or coefficients, not “slot in a bag.”

Create:  N VecNf.new^     (N > 0; zero-filled)
Fill:    N VecNf.new^ [ v0 … vN-1 ]
         (same [ ] filler as Mat2f; needs exactly .len values)
Access:  @() / !() over 0…len−1 (store is Float; no coerce)
Iterate: foreach yields each Float (VecNfIter)

Ops live under vec:: (vectors only; last arg TOS).
First wave: vec::+ vec::- vec::scale vec::dot vec::norm
vec::copy. All return new vectors/scalars — inputs unchanged.
vec::dot is the scalar inner product, not element-wise *.

Use for y, β̂, residuals, matvec results, Table feature rows
when every ordinal cell is Float. Prefer FloatList for
general sequences; convert at extract/persist boundaries.

eg: `` 3 VecNf.new^ [ 1 2 3 ] ``
See also: list-vec-mat ; Mat2f ; help vec::dot
*/
:doc VecNf

/*
Length N (fixed).
( -- len:i )
*/
:doc VecNf#len

/*
Create a zero-filled VecNf of length N. N must be > 0.
( n:i -- v:VecNf )
*/
:doc VecNf#new

/*
Indexed Float cells via @() / !().
*/
:doc VecNf#__items

/*
Flat cell iterator over a VecNf (Current → Float).
Obtained via foreach / ._tryGetIter — not usually constructed by hand.
*/
:doc VecNfIter

/*
Dense rank-2 float64/Float 0-init matrix, row-major order.
Shape rows×cols is fixed at construction (both > 0).

Create:  rows cols Mat2f.new^     (zero-filled)
Fill:    rows cols Mat2f.new^ [ … rows*cols values … ]
Access:  @() / !()  — flat index 0…len−1 (row-major)
         @[] / ![]  — ( r c -- val ) / ( val r c -- )
         Teaching line: @() = flat; @[] = coordinates
Iterate: foreach yields each Float flat
Nest:    .getRowIter^ / .getColIter^ → Range (0…n−1);
         then j i for @[]

Ops live under mat:: when ≥1 operand is a matrix
(last arg TOS). First wave: mat::matvec mat::*
mat::transpose mat::solve mat::+ mat::- mat::scale
mat::row mat::col mat::identity mat::copy.
Functional: return new mats/vecs; inputs unchanged.
mat::* is matrix product (A B), not element-wise —
Hadamard would be mat::.* when added. mat::solve is
A x = b (gonum; square or least-squares).

Primary home for design matrix X in OLS / regression.

eg: `` 2 3 Mat2f.new^ [ 0 1 2  3 4 5 ] ``
See also: list-vec-mat ; VecNf ; help mat::*
*/
:doc Mat2f

/*
Number of rows.
( -- rows:i )
*/
:doc Mat2f#rows

/*
Number of columns.
( -- cols:i )
*/
:doc Mat2f#cols

/*
Total cells (rows*cols).
( -- len:i )
*/
:doc Mat2f#len

/*
Create a zero-filled Mat2f. Both rows and cols must be > 0.
( rows:i cols:i -- m:Mat2f )
*/
:doc Mat2f#new

/*
Flat @()/!() and multi-index @[]/![].
Flat: index 0…len−1 in row-major order.
Multi: ( r c ) as for nested [r][c]; store casts to Float.
*/
:doc Mat2f#__items

/*
Row indexes as a Range (0 … rows−1). foreach; i is the row Int.
Pair with .getColIter^ on the matrix (not on the range).
( -- rows:Range )
*/
:doc Mat2f#getRowIter

/*
Column indexes as a Range (0 … cols−1). foreach; i is the col Int.
( -- cols:Range )
*/
:doc Mat2f#getColIter

/*
Wrap row/col indexes onto the torus (convolution / Life).
( row:i col:i -- row:i col:i )
*/
:doc Mat2f#wrap

/*
Go uint8 cells as Int (0..255). Same members as Mat2f. Palette8 blit source.
*/
:doc Mat2u8

/*
Go uint16 cells as Int. Same members as Mat2f.
*/
:doc Mat2u16

/*
Go uint32 cells as Int. Same members as Mat2f.
*/
:doc Mat2u32

/*
Go uint cells as Uint (wide unsigned). Same members as Mat2f.
Store is Uint; !() accepts Int or Uint.
*/
:doc Mat2u

/*
Int cells (Go int8). Same members as Mat2f.
*/
:doc Mat2i8

/*
1-bit indexes 0 or 1 as Int. Packed dense row-major bits.
Same Mat2 members as Mat2f (rows/cols/len, @()/![], getRowIter/getColIter, wrap).
Non-zero Put stores 1. Pair with Palette1 for indexed blit.
*/
:doc Mat2i1

/*
4-bit indexes 0..15 as Int (not Uint). Same Mat2 members as Mat2f (rows/cols/len,
@()/![], getRowIter/getColIter, wrap). Put panics outside 0..15.
Pair with Palette4 for indexed blit.
*/
:doc Mat2i4

/*
Int cells (Go int16). Same members as Mat2f.
*/
:doc Mat2i16

/*
Int cells (Go int32). Same members as Mat2f.
*/
:doc Mat2i32

/*
Int cells (Go int). Same members as Mat2f.
*/
:doc Mat2i

/*
Float cells (Go float32). Same members as Mat2f.
*/
:doc Mat2f32

/*
Color cells (Go Color / NRGBA). Same members as Mat2f (rows/cols/len, @()/![],
getRowIter/getColIter, wrap, foreach by value). Store is Color, not Float.
Not a mat:: operand. Full-color blit source; Pixmap remains the display surface.
Mat2c is an NRGBA image.Image (codec interop) as well as a color matrix.
*/
:doc Mat2c

// ---------------------------------------------------------------------------
// vec::  (vectors only; last arg TOS; return new values — inputs unchanged)
// ---------------------------------------------------------------------------

/*
Element-wise sum. Same length required.
( u:VecNf v:VecNf -- w:VecNf )
*/
:doc vec::+

/*
Element-wise difference u − v. Same length required.
( u:VecNf v:VecNf -- w:VecNf )
*/
:doc vec::-

/*
Scale every entry by s (s on TOS).
( v:VecNf s:f -- w:VecNf )
*/
:doc vec::scale

/*
Dot product → Float. Same length required.
This is the scalar inner product, not element-wise multiply
(that would be vec::.* when added). Not mat::*.
( u:VecNf v:VecNf -- f )
*/
:doc vec::dot

/*
Euclidean (L2) norm √(Σ xᵢ²).
( v:VecNf -- f )
*/
:doc vec::norm

/*
Deep copy of the vector.
( v:VecNf -- w:VecNf )
*/
:doc vec::copy

// ---------------------------------------------------------------------------
// mat::  (≥1 matrix; last arg TOS; return new values — inputs unchanged)
// ---------------------------------------------------------------------------

/*
Matrix × vector: w = A v. Length of v must equal A.cols.
( A:Mat2f v:VecNf -- w:VecNf )
*/
:doc mat::matvec

/*
Matrix product C = A B (inner dims must match: A.cols = B.rows).
This is true matrix multiply, not element-wise (Hadamard).
Element-wise would be mat::.* when added.
( A:Mat2f B:Mat2f -- C:Mat2f )
*/
:doc mat::*

/*
Transpose: At has shape cols×rows.
( A:Mat2f -- At:Mat2f )
*/
:doc mat::transpose

/*
Solve A x = b for x (gonum; works for square and
overdetermined least-squares paths). Length of b must
equal A.rows. Result length is A.cols.
( A:Mat2f b:VecNf -- x:VecNf )
*/
:doc mat::solve

/*
Element-wise sum. Same shape required.
( A:Mat2f B:Mat2f -- C:Mat2f )
*/
:doc mat::+

/*
Element-wise difference A − B. Same shape required.
( A:Mat2f B:Mat2f -- C:Mat2f )
*/
:doc mat::-

/*
Scale every entry by s (s on TOS).
( A:Mat2f s:f -- C:Mat2f )
*/
:doc mat::scale

/*
Copy of row r as a VecNf (length = cols).
( A:Mat2f r:i -- v:VecNf )
*/
:doc mat::row

/*
Copy of column c as a VecNf (length = rows).
( A:Mat2f c:i -- v:VecNf )
*/
:doc mat::col

/*
n×n identity matrix. n must be > 0.
( n:i -- I:Mat2f )
*/
:doc mat::identity

/*
Deep copy of the matrix.
( A:Mat2f -- B:Mat2f )
*/
:doc mat::copy
