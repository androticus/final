forget: reset-life-core
// NOTE modules generally should not have a forget TODO: remove this
/*
Conway's Game of Life simulation--engine.

Parameters
rows
cols // best if a multiple of 8
useAge // uses a 4 bit age mat; 0-7 dead ages, 8-15 alive
threads // 0 = none, 6-8 is usually optimum
*/

/*
LifeCore TypedObject.
call:
#init after setting params
#nextGen to advance one generation
#initModel to reset to a new random model


We use Paren operators to conveniently create the entire spec.
*/
{      
  "LifeCore" "name"
   
  {
    // required factory parameters
    Int "rows"
    Int "cols"
    // optional factory parameters
    Bool "useAge" // def=false
    Int "threads" // 0=none, def=0
    
    Mat2i1 "mat0" // curr
    Mat2i1 "mat1" // next -- we just swap after rendering
    Mat2i4 "matAge" // optional 0-7 dead ages, 8-15 alive
  } "fields"

  {
    15 "density" // 0 no cells, 99 all cells, 15 works well
  } "constants"
}
TypedObject.makeType^

/*
Shared inner called by regular and threaded.
matAge will be null for useAge=false option.
( rowFrom:i rowTo:i mat0 mat1 matAge -- )
*/
: LifeCore_nextGen_inner

  /*
  Conway's rules
  alive=1 dead=0
  cnt: how many of 8 neighbors are alive
  ( cnt:i alive:i -- alive:i )
  */
  :s lifeRules
    if // alive: stay alive rules
      dup 2 < if // isolation: death
        drop 0
      else 
        3 > if // overpopulation: death
          0
        else // goldilocks zone: 2,3
          1
        then
      then
    else // dead: birth rule
      3 == if 1 else 0 then
    then    
  ;s

  /*
  Update the nextgen life cell
  ( -- )
  */
  :s updateLife
    t:.zero^
    // clockwise from 12
    j -1 + i       x:.wrap^ x:@[] t:.add^
    j -1 + i  1 +  x:.wrap^ x:@[] t:.add^
    j      i  1 +  x:.wrap^ x:@[] t:.add^
    j  1 + i  1 +  x:.wrap^ x:@[] t:.add^
    j  1 + i       x:.wrap^ x:@[] t:.add^
    j  1 + i -1 +  x:.wrap^ x:@[] t:.add^
    j      i -1 +  x:.wrap^ x:@[] t:.add^
    j -1 + i -1 +  x:.wrap^ x:@[] t:.add^
    // ( cnt:i alive:i -- alive:i )
    t:.value j i x:@[] lifeRules
    j i y:![] 
  ;s

  // ( age:i -- age':i )
  :s incCapAge
    switch
      a 7 == case
        7
      /case
      a 15 == case
        15
      /case
      default
        a 1 +
      /default
    /switch
  ;s
  
  /*
  Update the age, based on alive/dead.
  Age is capped at 7/15
  ( -- )
  */
  :s updateAge
    j i y:@[]
    j i x:@[]
    // ( new:t old:t )
    dup2 == if // still dead or still alive
      drop2
      // advance age, capping at 7 or 15
      j i z:@[]
      incCapAge
    else // newly dead or alive
      // only need new
      drop
      if 8 else 0 then
    then
    j i z:![]
  ;s
  
  // are we using age?
  dup Null? not l.useAge!
  >zs // matAge, might be null, we don't care
  >ys // mat1
  >xs // mat0
 
  Counter.new^ >ts // live neighbor count, 0-8
  range
  foreach
    x:.getColIter^
    foreach
      updateLife
      l.useAge if
        updateAge
      then
    /foreach
  /foreach
  tdrop xdrop ydrop zdrop
;

/*
PipParam fields:
mat0 mat1 matAge rowFrom rowTo done:AtomicInt
( pp:AtomicDynamicObject -- )
*/
: LifeCore_nextGen_worker
  >ts
  t:.rowFrom@
  t:.rowTo@
  t:.mat0@
  t:.mat1@
  t:.matAge@ // may be null
 
  LifeCore_nextGen_inner
  
  t:.done@.inc^ drop
  tdrop
;

/*
Spawn one worker, with its alloted tranche of row.
( rowFrom:i rowTo:i done:AtomicInt -- ) o( inst )
*/
: LifeCore_nextGen_threaded_spawn_worker
  Pip.newPipParams^ l.pp!
  l.pp.done!
  l.pp.rowTo!
  l.pp.rowFrom!
  o:.mat0@ l.pp.mat0!
  o:.mat1@ l.pp.mat1!
  o:.matAge@ l.pp.matAge! // may be null
  l.pp@ spawn: LifeCore_nextGen_worker
;


/*
NOTE: Mat2i1 is NOT threadsafe. This algorithm loosely assumes
that all the worker threads will be updating regions distinct from
each other, and that overlap is very unlikely. It can only occur
at the seam between the tail cells of one tranche and the head
tranche of another; overlap is limited to 16 bits. 
( -- ) o( inst )
*/
: LifeCore_nextGen_threaded
  // dispatch to n workers
  o:.threads@ l.n!
  
  Counter.new^ l.runts!
  // we get the modulus and remainder, 
  // each thread does mod + ?1 until rem exhausted
  o:.mat0@.rows l.n@ modrem // ( mod:i rem:i )
  l.runts.set^ l.each!
  0 l.start!
   
  AtomicInt.new^ l.done!
  0 l.n@ 1 - range
  foreach
    l.start@ dup l.each@ + // add our full tranche, we -1 later...
    // add one runt row if available
    l.runts.value if l.runts.dec^ 1 else 0 then +
    // update next start
    dup l.start!
    // adjust because it is rowTo, not rowCount
    1 -
    // ( rowFrom:i rowTo:i -- )
    l.done
    LifeCore_nextGen_threaded_spawn_worker
  /foreach
  
  // fairly bogus... poll loop for all done
  // events from UI and timers etc will still get queued
  // but EventLoop dispatch won't cycle until done
  forever
    l.n@ l.done@.value == if break then
    50 delayUs
  /forever
;

/*
Determine the next generation on the main thread.
( -- ) o( inst )
*/
: LifeCore_nextGen_default
  0 o:.mat0@.rows 1 -
  o:.mat0@
  o:.mat1@
  o:.matAge@ // null if not using
  
  LifeCore_nextGen_inner
;

/*
Determine the next generation.
We treat the grid as a torus, and wrap the coords around.
( -- ) o( inst )
*/
: LifeCore_nextGen
  o:.threads@
  if 
    LifeCore_nextGen_threaded
  else
    LifeCore_nextGen_default
  then
    
  // swap them
  o:.mat0@ o:.mat1@ o:.mat0! o:.mat1!
;

/*
Initializes the model mats: mat0 mat1 matAge
( -- ) o( inst )
*/
: LifeCore_initModel
  o:.mat0@ >xs
  o:.mat1@ >ys
  o:.matAge@ >zs // might be null
  0 x:.len 1 - range
  // we get a random number 0-99 and set a cell if < density
  foreach
    100 Rand.int^ LifeCore.density < if 1 else 0 then
    dup i dup2 x:!() y:!() // set cur and last the same
    o:.useAge if
      if 8 else 0 then
      i z:!()
    else
      drop
    then
  /foreach 
  zdrop ydrop xdrop
;

/*
Advance the sim by one generation.
Note: CPU intensive, and blocks for duration, threaded or not.
( -- ) o( inst:LifeCore )
*/
: LifeCore#nextGen
  LifeCore_nextGen
;

/*
Initializes the model mats: mat0 mat1 matAge
( -- ) o( inst )
*/
: LifeCore#initModel
  LifeCore_initModel
;


/*
Initialize a new LifeCore.
The following need to have been set:
rows, cols
Optionally:
threads, useAge
( -- ) o( inst:LifeCore )
*/
: LifeCore#init     
  // Mats
  o:.rows@ o:.cols@
  dup2 Mat2i1.new^ o:.mat0!
  dup2 Mat2i1.new^ o:.mat1!
  o:.useAge if
       Mat2i4.new^ o:.matAge!
  else
    drop2
  then
  LifeCore_initModel
;

