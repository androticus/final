forget: reset-to-point
`
Make a simple "PointF" TypedObject, with fields and methods (no constants).

Try:

to-point
` const: to-point-blurb

// Use Paren operators to create the entire spec.
// We don't create constants for this type.

{
   {
     Float "x"
     Float "y"
   } "fields"
   
  "PointF" "name"
}

// make the type
TypedObject.makeType^

// Doc it

/*
A point in 2D space.
Properties are Float.
*/
:doc PointF

// Add a method

/*
Calculate the absolute value, aka distance from origin.
( -- val:f ) o( inst )
*/
: PointF#abs
  o:.x@ sqr o:.y sqr + sqrt 
;

/*
Create a Point, and compute its abs.
( -- )
*/
: to-point
  PointF.new^ >ts
  3 t:.x!
  4 t:.y!
  // we use the MethodRef fetch, then the ^ operator
  // to show how you can use that pattern
  ( t:.x@ t:.y@ t:.abs~ ^ )
  "x: %f  y: %f  | pt | : %f\n" fmt $
  t:drop
;

to-point-blurb $ nl
