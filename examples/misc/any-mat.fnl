forget: reset-any-mat
// `examples/misc/mat.fnl` loadFromFile
`
Example of how to use the general purpose Mat type.

try:

any-mat
`
const: any-mat-blurb

: any-mat
  // 2 rows, 3 cols
  2 3 Mat.new^
  // now fill it, use parens op, row-major order
  [ -1 2.2 3u
    "four" "five" "six"
  ] 
  >ts
  t:.getRowIter^
  foreach
    t:.getColIter^
    foreach
      ( j i t:@[] )
      "%4v " fmt $
    /foreach
    nl
  /foreach
  
  tdrop
;

any-mat-blurb $ nl
