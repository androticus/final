forget: reset-nbody
// `benchmarks/industry/nbody.fnl` compileFromFile

`
n-body problem (n=5).
Official Final benchmark.
Single threaded.
Accepts N as the number of steps.
Note: this benchmark makes use of List and TypedObject
to hold the many values needed for this algorithm.

Try:

100000 nbody
` const: nbody-blurb


// Useful constants and vars

// Number of iterations; passed in to algorithm
0                 var:   N 
0                 var:   bodies // stable parking spot
5                 const: n // nbodies
0.01              const: dt // integration
Math.pi dup * 4 * const: solarMass
365.24            const: daysPerYear

// Create a Vector using TypedObject
{
  "Vec3" "name" 
  {
    Float "x"
    Float "y"
    Float "z"
  } "fields"
}
TypedObject.makeType^
// Vec3 now exists

// Create Body Type
{
  "Body" "name"
  {
    Float "mass"
    Vec3  "pos"
    Vec3  "vel"
  } "fields"
}
TypedObject.makeType^

/*
Create a vector, scaling each value; scale=1.0 for none
( z:f y:f x:f scale:f -- vec3:Vec3 )
*/
: makeVec3_
  >ts
  Vec3.new^
  >zs
    t:peek * z:.x!
    t:peek * z:.y!
    ts>    * z:.z!
  zs>
;

/*
Create a body, with this shape, fields in this order:
  .mass
  .pos : .x .y .z
  .vel : .x .y .z

Stack (bottom→top): vz vy vx z y x mass
( vz:f vy:f vx:f z:f y:f x:f mass:f -- body:Body )
*/
: makeBody_
  Body.new^
  >zs
    solarMass * z:.mass!
    
    // make pos, no scale
    1.0 makeVec3_
    z:.pos!
    
    // make vel, scale by daysPerYear (CLBG)
    daysPerYear makeVec3_
    z:.vel!
  zs>
;

// ( -- bodies:List )
: makeBodies_
  List.new^
  >zs
    // Sun
    0.0 0.0 0.0 0.0 0.0 0.0 1.0
    makeBody_ z:.append^
    
    // Jupiter
    -6.90460016972063023e-05 7.69901118419740425e-03 1.66007664274403694e-03
    -1.03622044471123109e-01 -1.16032004402742839e+00 4.84143144246472090e+00
    9.54791938424326609e-04
    makeBody_ z:.append^
    
    // Saturn
    2.30417297573763929e-05 4.99852801234917238e-03 -2.76742510726862411e-03
    -4.03523417114321381e-01 4.12479856412430479e+00 8.34336671824457987e+00
    2.85885980666130812e-04
    makeBody_ z:.append^
    
    // Uranus
    -2.96589568540237556e-05 2.37847173959480950e-03 2.96460137564761618e-03
    -2.23307578892655734e-01 -1.51111514016986312e+01 1.28943695621391310e+01
    4.36624404335156298e-05
    makeBody_ z:.append^
    
    // Neptune
    -9.51592254519715870e-05 1.62824170038242295e-03 2.68067772490389322e-03
    1.79258772950371181e-01 -2.59193146099879641e+01 1.53796971148509165e+01
    5.15138902046611451e-05
    makeBody_ z:.append^      
  zs>
;

/*
Subtract vec1 - vec2, returning the result vector
( vec1:Vec3 vec2:Vec3 -- res:Vec3 )
*/
: Vec3-
  Vec3.new^
  >ts
    >ys >xs
    x:.x@ y:.x@ - t:.x!
    x:.y@ y:.y@ - t:.y!
    x:.z@ y:.z@ - t:.z!
  ts> x:drop y:drop
;

/*
The Euclidean distance between 2 vectors,
supplied as a delta vector.
( delta:Vec3 -- dist:f )
*/
: Vec3DistanceFromDelta
  >ts
    t:.x@ dup *
    t:.y@ dup * +
    t:.z@ dup * +
    sqrt
  t:drop
;

/*
Zero total momentum onto the sun (body 0).
z( bodies:List -- )
*/
: offsetMomentum_
  Vec3.new^ >xs
    
    0 n 1 - range
    foreach
      i z:@() >ys
        y:.vel@ >ts
        y:.mass@
        dup t:.x@ * x:.x@ + x:.x!
        dup t:.y@ * x:.y@ + x:.y!
            t:.z@ * x:.z@ + x:.z!
        t:drop
      y:drop
    /foreach
    
    // sun.vel = -p / solarMass
    0 z:@() >ys
      y:.vel@ >ts
      x:.x@ solarMass / 0.0 swap - t:.x!
      x:.y@ solarMass / 0.0 swap - t:.y!
      x:.z@ solarMass / 0.0 swap - t:.z!
      t:drop
    y:drop
  x:drop
;

/*
Total energy (kinetic + potential). CLBG checksum.
After offsetMomentum, expect about -0.169075164
z( bodies:List -- ) ( -- e:f )
*/
: energy
  0.0
  // kinetic: 0.5 * m * |v|^2
  0 n 1 - range
  foreach
    i z:@() >ys
      y:.vel@ >ts
        t:.x@ dup *
        t:.y@ dup * +
        t:.z@ dup * +
        y:.mass@ * 0.5 * +
      t:drop
    y:drop
  /foreach
  // potential: - Σ m_j*m_i / |p_j - p_i|
  0 n 2 - range
  foreach
    i 1 + n 1 - range
    foreach
      j z:@() >xs
      i z:@() >ys
      x:.pos@ y:.pos@ Vec3- Vec3DistanceFromDelta
      x:.mass@ y:.mass@ * swap / -
      x:drop y:drop
    /foreach
  /foreach
;

// ( N:i bodies:List -- bodies )
: nbodyInner_
  >zs // z=bodies
   
  // steps: 0 .. N-1 inclusive → N advances
  0 swap 1 - range
  foreach
    // pairwise velocity update (j outer, i inner)
    0 n 2 - range
    foreach
      i 1 + n 1 - range
      foreach
        // x=body j  y=body i  t=delta pos
        j z:@() >xs
        i z:@() >ys
        x:.pos@ y:.pos@ Vec3- dup >ts
        Vec3DistanceFromDelta
        // mag = dt / (dist^3)
        dup dup * * dt swap /

        // bj.v -= d * bi.mass * mag
        // bi.v += d * bj.mass * mag
        dup y:.mass@ * t:.x@ * x:.vel@.x@ swap - x:.vel@.x!
        dup y:.mass@ * t:.y@ * x:.vel@.y@ swap - x:.vel@.y!
        dup y:.mass@ * t:.z@ * x:.vel@.z@ swap - x:.vel@.z!

        dup x:.mass@ * t:.x@ * y:.vel@.x@ + y:.vel@.x!
        dup x:.mass@ * t:.y@ * y:.vel@.y@ + y:.vel@.y!
            x:.mass@ * t:.z@ * y:.vel@.z@ + y:.vel@.z!

        t:drop y:drop x:drop
      /foreach
    /foreach

    // position update once per step: p += dt * v
    0 n 1 - range
    foreach
      i z:@() >ts
        t:.pos@ >xs
        t:.vel@ >ys
        x:.x@ dt y:.x@ * + x:.x!
        x:.y@ dt y:.y@ * + x:.y!
        x:.z@ dt y:.z@ * + x:.z!
        y:drop x:drop
      t:drop
    /foreach
  /foreach
  zs>
;

// ( N:i -- )
: nbody
  DateTime.now^ >ts // total time
  dup N ! // for later
  
  makeBodies_ dup bodies !
  >zs
    offsetMomentum_
    // sanity: CLBG ICs after offsetMomentum ≈ -0.169075164
    ( energy )
    "energy0 = %.9f (expect ~ -0.169075164)\n" fmt $
  zs>
  
  nbodyInner_
  >zs
    ( energy )
    "energyN = %.9f\n" fmt $
  z:drop
  
  // print results
  ( 
    N @
    DateTime.now^ ts> dt::- // ( elapsed:Interval -- )
    dup N @ dt::/
  )
  "Did %d steps in %s for %s per N\n\n" fmt $
;

nbody-blurb $ nl
