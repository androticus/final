forget: reset-life-sim

requiresFyne
// `examples/gui/simgame/life-sim.fnl` compileFromFile
`
Conway's Game of Life simulation using Fyne.

This application uses the SimGame framework.

Try either of:

N = Pips, 0=none (no multi-threading)
    hard capped at 32 (more than 12 rarely helps, usually hurts)

N life-sim-pips

OR

life-sim // sets N to min of your core count or 12 (max effective)

optionally choose a palette, then press Run

Slow = 1 fps
Fast = limit of your machine/Pip count
`
const: life-sim-blurb


/*
LifeSim TypedObject for the app and event handlers.
Composes with SimGame via required field base (wired by
SimGame.newWithDelegate^). SimGame chrome / pixmap live on o:.base@.

We use Paren operators to conveniently create the entire spec.
*/
{      
  "LifeSim" "name"
  SimGame "baseType"
   
  {
 // SimGame "base" // added automatically
    Int "threads" // 0=none
    Int "rows"
    Int "cols"
    Int "cellPx" // defCellPx
    Mat2i1 "mat0" // curr
    Mat2i1 "mat1" // next -- we just swap after rendering
    Mat2i4 "matAge" // 0-7 dead ages, 8-15 alive
    Palette4 "pal4" // set from palettes list
    Select "selPalette"
    Slider "slider"
    Float "speed" // 0.1 to 10.0
    List "kickPipes" // N threads; nil until lazy init
    Pipe "donePipe"
  } "fields"

  {
    0 "btnRun"
    1 "btnPause"
    2 "btnNew"
    
    // sz dims / cellPx should be a power of 16
    144 "defRows" // def=144
    192 "defCols" // def=192
    4 "defCellPx" // def=4, 4 or 8 works well, should be power of 2
    15 "density" // 0 no cells, 99 all cells, 15 works well
    
    0 "palBW"
    1 "palRedGreen"
    2 "palFadeAway"
    3 "palWebColors"
    ( "B & W" "Red/Green" "Fade Away" "Web Colors" s) "paletteNames"
    ( ) "palettes" // built in init
  } "constants"
}
TypedObject.makeType^

/*
Shared inner called by regular and threaded
( rowFrom:i rowTo:i mat0 mat1 matAge -- )
*/
: LifeSim_nextGen_inner

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
  We return the live/dead 1/0
  ( -- new:i )
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
    dup j i y:![] // ( new:t )
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
  ( new:i -- )
  */
  :s updateAge
    // still dead or alive?
    // ( new:t old:t )
    j i x:@[]
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
  
  Counter.new^ >ts // neighbors
  
  >zs // matAge
  >ys // mat1
  >xs // mat0
 
  range
  foreach
    x:.getColIter^
    foreach
      updateLife
      updateAge
    /foreach
  /foreach
  xdrop ydrop zdrop tdrop
;

/*
NOTE: Mat2i1 is NOT threadsafe. This algorithm loosely assumes
that all the worker threads will be updating regions distinct from
each other, and that overlap is very unlikely. It can only occur
at the seam between the tail cells of one tranche and the head
tranche of another, and actually doesn't occur using our dims. 
( -- ) o( inst )
*/
: LifeSim_nextGen_threaded
  // dispatch to the n workers
  o:.threads@ l.n!
  
  // kick off the workers
  o:.kickPipes@
  foreach
    true i >pipe
  /foreach
  
  // get a done sentinel from each worker
  0 l.n@ 1 - range
  foreach
    o:.donePipe@ pipe> drop
  /foreach
;

/*
Determine the next generation.
We treat the grid as a torus, and wrap the coords around.
( -- ) o( inst )
*/
: LifeSim_nextGen_default
  0 o:.mat0@.rows 1 -
  o:.mat0@
  o:.mat1@
  o:.matAge@
  
  LifeSim_nextGen_inner
  ;

/*
Determine the next generation.
We treat the grid as a torus, and wrap the coords around.
( -- ) o( inst )
*/
: LifeSim_nextGen
  o:.threads@
  if 
    LifeSim_nextGen_threaded
  else
    LifeSim_nextGen_default
  then
    
  //swap them
  o:.mat0@ o:.mat1@ o:.mat0! o:.mat1!
;

/*
Initializes the model mats: mat0 mat1 matAge
( -- ) o( inst )
*/
: LifeSim_initModel
  o:.mat0@ >xs
  o:.mat1@ >ys
  o:.matAge@ >zs
  0 x:.len 1 - range
  // we get a random number 0-99 and set a cell if < density
  foreach
    100 Rand.int^ LifeSim.density < if 1 else 0 then
    dup i dup2 x:!() y:!() // set cur and last the same
    if 8 else 0 then
    i z:!()
  /foreach 
  zdrop ydrop xdrop
;


/*
PipParam fields:
mat0~ mat1~ matAge rowFrom rowTo kickPipe donePipe
( pp:AtomicDynamicObject -- )
*/
: LifeSim_nextGen_worker
  >ts
    // wait for start signal
    // loop while our kick pipe is open
    // it gets closed in the window close event
    t:.kickPipe@
    foreach
      t:.rowFrom@
      t:.rowTo@
      // note: mat0/1 are PropertyRefs, @ must be separate
      t:.mat0 @
      t:.mat1 @
      t:.matAge@
 
      LifeSim_nextGen_inner
  
      true t:.donePipe@ >pipe
    /foreach
  tdrop
;

/*

/*
( rowFrom:i rowTo:i kickPipe:pipe donePipe:pipe -- ) o( inst )
*/
: LifeSim_nextGen_initWorker
  Pip.newPipParams^ l.pp!
  l.pp.donePipe!
  l.pp.kickPipe!
  l.pp.rowTo!
  l.pp.rowFrom!
  // we use PropertyRefs for mat0/1 due to swaps
  o:.mat0~ l.pp.mat0!
  o:.mat1~ l.pp.mat1!
  o:.matAge@ l.pp.matAge!
  l.pp@ spawn: LifeSim_nextGen_worker
;

/*
This initializes our worker Pips (for threads>0)
Not called if not threaded.
( -- ) o( inst )
*/
: LifeSim_initWorkers
    // dispatch to n workers
  o:.threads@ l.n!
  
  // create our done pipe and kick pipe list
  l.n@ pipe::new o:.donePipe!
  List.new^ o:.kickPipes!
  
  Counter.new^ l.runts!
  // we get the modulus and remainder, 
  // each thread does mod + ?1 until rem exhausted
  o:.mat0@.rows l.n@ modrem // ( mod:i rem:i )
  l.runts.add^ l.each!
  0 l.start!
   
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
    1 pipe::new dup o:.kickPipes.append^ o:.donePipe@
    LifeSim_nextGen_initWorker
  /foreach
;

// if threaded, make sure they are created
: LifeSim_ensureWorkers
  o:.threads@ if
    o:.kickPipes@ Null? if
      LifeSim_initWorkers
    then
  then
;

: LifeSim_ensureUiParams
  o:.selPalette.selectedIndex 0 max
  LifeSim.palettes@() o:.pal4!
  // values <=1 are linear, log for 1-2
  o:.slider.value@ dup 1.0 > if
    5.0 *
  then 
  dup o:.speed!
  o:.base.clock.simSpeed!
;


/*
Render the current model, mat0
( -- ) o( inst )
*/
: LifeSim_renderModel
  LifeSim_ensureUiParams
  o:.pal4@ o:.matAge@ 0 0 o:.base@.pixmap.blit4^
;

: LifeSim_stop
  o:.base.stop^  
;

: LifeSim_pause
  o:.base.pause^  
;

: LifeSim_run
  LifeSim_ensureWorkers // if threaded, nop otherwise
  LifeSim_ensureUiParams
  o:.base.start^ // nop if started 
;

//( idx:i -- ) o( inst )
: LifeSim_new
  LifeSim_stop
  LifeSim_initModel
  o:.base.doRender^ 
;

: LifeSim#handleWindowClose
  o:.threads@ if
    o:.kickPipes@ dup Null? if
      drop
    else
      foreach
        i pipe::close
      /foreach
    then
  then
;

/*
 Button event handler.
( idx:i -- ) o( inst )
*/
: LifeSim#handleButtonTap
  switch
    a LifeSim.btnRun == case
      LifeSim_run
    /case
    
    a LifeSim.btnPause == case
      LifeSim_pause
    /case
    
    a LifeSim.btnNew == case
      LifeSim_new
    /case    
  /switch
;

: LifeSim#handleRender
  LifeSim_renderModel
;

/*
We use AutoRender dispatch in SimGame for rendering.
So it renders and refreshes the view after every update.
( -- ) o( inst )
*/
: LifeSim#handleUpdate
  LifeSim_nextGen
;


/*
We use this to output average update/render rates.
Note: it updates in ticks, not seconds, so if you
set the speed to more than 1 it gets called more frequently.
( index:i -- ) o( inst )
*/
: LifeSim#handleTimerEvent
  drop
  true // use stderr
  o:.base.printAvgUpdateRenderRate^
;


/* Required if options.bottomBar set
Called by SimGame during #makeUI. Return a Fyne CanvasObject,
e.g. an HBox, Border or Widget that is the content below the view.
( -- lay:Layout | widg:Widget ) o( inst:MyGame )
*/
: LifeSim#makeBottomBar
  // we make a Border layout, so the Slider expands to fill space
  o:.base@ >ys
  null null // top bottom
  ( // buttons
    LifeSim.btnRun "Run" Icon.MediaPlay y:.makeButtonWithIcon^
    LifeSim.btnPause "Pause" Icon.MediaPause y:.makeButtonWithIcon^
    LifeSim.btnNew "New" y:.makeButton^
      
    "Palette:" Label.new^     
    null 0 LifeSim.paletteNames@ Select.new^
    >ts 0 t:.selectedIndex!
    ts> dup o:.selPalette!
 
    "Update Speed:" Label.new^ // right    
  ) Box.newH^ // left
  null // right 
  0.0 2.0 0.4 Slider.new^ dup o:.slider! // center
  1.0 o:.slider.value!
  Border.new^
  ydrop
;



/*
Make the 16 entry Palette for age.
0 - 7 are dead
8 - 15 are alive
( -- Palette4 )
*/
: LifeSim_makeFadeAwayPalette
  16 List.newWithCapacity^ >ts
  // dead progress, bright blue, then fading to black
  256 8 / l.per!
  0 7 range
  foreach
    i l.per@ * 255 swap - >ts
    0 0 ts> Color.rgb^ t:.append^
  /foreach
    
  // life progress, white, then fade to light yellow
  128 8 / l.per!
  0 7 range
  foreach
    i l.per@ * 255 swap - >ts
    255 255 ts> Color.rgb^ t:.append^
  /foreach
  
  ts>
  Palette4.newFromList^
;

: LifeSim_makeFadeInPalette
  16 List.newWithCapacity^ >ts
  // dead progress, red, then fading to grey
  128 8 / l.per!
  0 7 range
  foreach
    i l.per@ * 255 swap -
    i l.per@ * dup Color.rgb^ t:.append^
  /foreach
    
  // life progress, purple, then grow to lime
  Color.purple t:.append^
  1 6 range
  foreach
    // red
    128 6 / i * 128 swap - 0 min
    // green
    255 6 / i * 255 max
    0 Color.rgb^ t:.append^
  /foreach
  Color.lime t:.append^
  
  ts>
  Palette4.newFromList^
;

: LifeSim_makeBWPalette
  16 List.newWithCapacity^ >ts
  // dead progress, black, then lighten to dark gray
  
  128 8 / l.per! 
  0 7 range
  foreach
    i l.per@ *
    dup dup Color.rgb^ t:.append^
  /foreach
    
  // life progress, white, then fade to light gray
  64 8 / l.per!
  0 7 range
  foreach
    i l.per@ * 255 swap - 
    dup dup Color.rgb^ t:.append^
  /foreach
   
  ts>
  Palette4.newFromList^
;

: LifeSim_makePalette4s
  LifeSim.palettes@ >ts
  LifeSim_makeBWPalette t:.append^
  LifeSim_makeFadeInPalette t:.append^
  LifeSim_makeFadeAwayPalette t:.append^
  255 Palette4.newWebPalette^ t:.append^ // A=255 
  
  t:drop
;


/*
Initialize a new LifeSim core.
This does not create the UI.
Called from SimGame#init via del before SimGame builds the Pixmap.
( -- ) o( inst:LifeSim )
*/
: LifeSim#preInit
  o:.base@ >ys 
  // LifeSim options and setup
  LifeSim.defCellPx o:.cellPx!
  LifeSim.defRows o:.rows!
  LifeSim.defCols o:.cols!
    
  // Pixmap size
  o:.cols@ o:.cellPx@ *
  o:.rows@ o:.cellPx@ *
  Size.new^ y:.pixSize!

  // SimGame options
  "Conway's Game of Life Simulation" y:.title!
  // Full-frame blit every paint — skip library background fill
  Color.transparent y:.backgroundColor!
  // we use auto render
  100 y:.clock.updateRate! // 10/s
  y:.clock.setAutoRenderOn^
  true y:.options.hasWindowCloseHandler! // to close kickPipes
     
  // Mats
  o:.rows@ o:.cols@
  dup2 Mat2i1.new^ o:.mat0!
  dup2 Mat2i1.new^ o:.mat1!
       Mat2i4.new^ o:.matAge!
  LifeSim_initModel
      
  // Timer/ time-related
  1.0 y:.clock.simSpeed!
  0 1000 y:.timerEventEvery^ // for outputing update rate
  
  // Palettes
  LifeSim_makePalette4s
  0 LifeSim.palettes@() o:.pal4!
  ydrop
;

: LifeSim#postInit
  o:.cellPx@ o:.base@.pixmap.params.cellPx!
;

: LifeSim_ensureThreads
  dup
  <0 if
    drop 0
  else
    dup 32 > if
      drop 32
    then
  then
;

// ( n:i -- )
: (life-sim-pips)
  LifeSim.new^ l.ls!
  l.ls.base l.base!
( peek ) "Threads: N=%d\n" fmt $$
  l.ls.threads! 
  
  l.base.init^ // sandwiches #preInit and #postInit
  l.base.makeUi^
  l.base.startPaused^
  l.base.doRender^
  
  l.base.win@.show^
;

// ( n:i -- )
: life-sim-pips
  LifeSim_ensureThreads
  (life-sim-pips)
;

: life-sim sys.coreCount 12 min (life-sim-pips) ;

life-sim-blurb $ nl
