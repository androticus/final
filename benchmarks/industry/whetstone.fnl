/*
Whetstone (netlib / Painter double-precision).
Official Final benchmark.
Single threaded.
Accepts LOOP as the major scale factor (netlib default 1000).
Float + transcendental mix; FloatList for E1[1..4] (index 0 unused).

"benchmarks/industry/whetstone.fnl" compileFromFile
*/

0.499975 const: T
0.50025  const: T1
2.0      const: T2

0 var: LOOP
0 var: E1 // FloatList len 5
0 var: J
0 var: K
0 var: L
0.0 var: X
0.0 var: Y
0.0 var: Z

// ( -- e1:FloatList )
: makeE1_
  5 FloatList.newWithLen^
;

/*
PA: six in-place updates of E1[1..4].
z( e1 -- )
*/
: pa_
  0 J !
  begin
    J @ 6 < while
    1 z:@() 2 z:@() + 3 z:@() + 4 z:@() - T * 1 z:!()
    1 z:@() 2 z:@() + 3 z:@() - 4 z:@() + T * 2 z:!()
    1 z:@() 2 z:@() - 3 z:@() + 4 z:@() + T * 3 z:!()
    0.0 1 z:@() - 2 z:@() + 3 z:@() + 4 z:@() + T2 / 4 z:!()
    J @ 1 + J !
  repeat
;

/*
P0: rotate E1 through J,K,L (netlib order — third assign sees updated E1[J]).
*/
: p0_
  E1 @ >zs
    K @ z:@() >ts
    t:peek J @ z:!()
    L @ z:@() K @ z:!()
    ts> L @ z:!()
  zs> drop
;

/*
P3: ( x y -- z )
  x1 = T*(x+y); y1 = T*(x1+y); z = (x1+y1)/T2
*/
: p3_
  over over + T *
  swap over + T *
  nip
  + T2 /
;

// Run modules for current LOOP / E1. ( -- xFinal:f )
: whetstoneOnce_
  // Module 1 N1=0 — omitted
  
  // Module 2: array elements (N2 = 12*LOOP)
  E1 @ >zs
    1.0 1 z:!()
    -1.0 2 z:!()
    -1.0 3 z:!()
    -1.0 4 z:!()
    1 LOOP @ 12 * range
    foreach
      1 z:@() 2 z:@() + 3 z:@() + 4 z:@() - T * 1 z:!()
      1 z:@() 2 z:@() + 3 z:@() - 4 z:@() + T * 2 z:!()
      1 z:@() 2 z:@() - 3 z:@() + 4 z:@() + T * 3 z:!()
      0.0 1 z:@() - 2 z:@() + 3 z:@() + 4 z:@() + T * 4 z:!()
    /foreach
  zs> drop
  
  // Module 3: PA (N3 = 14*LOOP)
  1 LOOP @ 14 * range
  foreach
    E1 @ >zs pa_ zs> drop
  /foreach
  
  // Module 4: conditional jumps (N4 = 345*LOOP)
  1 J !
  1 LOOP @ 345 * range
  foreach
    J @ 1 == if
      2 J !
    else
      3 J !
    then
    J @ 2 > if
      0 J !
    else
      1 J !
    then
    J @ 1 < if
      1 J !
    else
      0 J !
    then
  /foreach
  
  // Module 6: integer arithmetic (N6 = 210*LOOP)
  1 J !
  2 K !
  3 L !
  E1 @ >zs
    1 LOOP @ 210 * range
    foreach
      J @ K @ J @ - * L @ K @ - * J !
      L @ K @ * L @ J @ - K @ * - K !
      L @ K @ - K @ J @ + * L !
      J @ K @ + L @ + >Float L @ 1 - z:!()
      J @ K @ * L @ * >Float K @ 1 - z:!()
    /foreach
  zs> drop
  
  // Module 7: trig (N7 = 32*LOOP)
  0.5 X !
  0.5 Y !
  1 LOOP @ 32 * range
  foreach
    // X = T * atan( T2*sin(X)*cos(X) / (cos(X+Y)+cos(X-Y)-1) )
    X @ sin X @ cos * T2 *
    X @ Y @ + cos
    X @ Y @ - cos +
    1.0 - /
    atan T * X !
    
    Y @ sin Y @ cos * T2 *
    X @ Y @ + cos
    X @ Y @ - cos +
    1.0 - /
    atan T * Y !
  /foreach
  
  // Module 8: P3 calls (N8 = 899*LOOP)
  1.0 X !
  1.0 Y !
  1.0 Z !
  1 LOOP @ 899 * range
  foreach
    X @ Y @ p3_ Z !
  /foreach
  
  // Module 9: P0 (N9 = 616*LOOP)
  1 J !
  2 K !
  3 L !
  E1 @ >zs
    1.0 1 z:!()
    2.0 2 z:!()
    3.0 3 z:!()
  zs> drop
  1 LOOP @ 616 * range
  foreach
    p0_
  /foreach
  
  // Module 10 N10=0 — omitted
  
  // Module 11: std functions (N11 = 93*LOOP)
  0.75 X !
  1 LOOP @ 93 * range
  foreach
    X @ log T1 / exp sqrt X !
  /foreach
  
  X @
;

// ( LOOP:i -- )
: whetstone
  DateTime.now^ >ts
  dup LOOP !
  makeE1_ E1 !
  
  whetstoneOnce_ >ys // final X
  
  DateTime.now^ ts> dt::- >ts // elapsed
  (
    LOOP @
    y:peek
    t:peek
    t:peek LOOP @ dt::/
  )
  "LOOP=%d  X=%.9f  elapsed=%s  perLOOP=%s\n" fmt $
  
  t:drop y:drop
;
