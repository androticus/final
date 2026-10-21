# Mats

A Mat (Matrix) is a multidimensional container whose shape is set at construction time. Final currently supports 2-dimensional mats that can hold a variety of strongly typed or Any-typed values. The two dimensions of Final's 2d Mats are its rows and columns. The cells are stored in row-major order, meaning the column cells are adjacent in the underlying array holding the values. A Mat cannot be resized once created.

The factory for all Mats has the following form, using Mat2f as an example:
```
2 3 Mat2f.new^ // creates a mat with 2 rows and 3 columns
```

The value in a cell is accessed either by the row/col matrix accessors `@[]`/`![]` or using a flat index, where cell index = row*cols + col using the ordinal accessors `@()`/`!()`.

Mat is iterable, and returns the items of the Mat in flat ordinal order. There are two convenience iterator accessors, which are basically shortcuts for a range operator:
* `getRowIter` // provides the indexes 0..rows-1
* `getColIter` // provides the indexes 0..cols-1

Typically used like this:
```
  myMat l.mat!

  l.mat.getRowIter^
  foreach
    l.mat.getColIter^
    foreach
      j i l.mat@[] // retrieves item at row j col i
      ...
      newVal j i l.mat![]
    /foreach
  /foreach
```

The following Mats are provided. Some serve special purposes, as noted

| Mat | Cell Type | Stack Type | Notes |
--- | --- | --- | ----
Mat2u8 | uint8 | Int | 1
Mat2u16 | uint16 | Int | 1
Mat2u32 | uint32 | Int | 1
Mat2u | uint | Uint | 2
Mat2i1 | 0/1 | Int | 3, bit-packed storage, can be used for true/false via Truthy semantics
Mat2i4 | 0-15 | Int | 3, stored in a Mat2i8 at present, may be packed in future
Mat2i8 | int8, 0-255 | Int | 3
Mat2i16 | int16 | Int |
Mat2i32 | int32 | Int |
Mat2i | Int | Int |
Mat2f32 | float32 | Float | 4
Mat2f | Float | Float | 5
Mat2c | Color | Color | 3
Mat | Any | Any | 

Notes:
1. unsigned value set/returned in an Int
2. unsigned value set/returned in a Uint
3. special use in Blitting operations
4. value set/returned in a Float
5. special use in linear algebra, regression/model testing, and mat:: operators
