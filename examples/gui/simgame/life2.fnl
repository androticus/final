forget: reset-life-sim
requiresFyne

// install and load the life module, which declares LifeCore
"github.com/androticus/final/modules/simgame/life"
{
  "modules/simgame/life" "path"
}  
install

"" "github.com/androticus/final/modules/simgame/life" import
`
Conway's Game of Life simulation using Fyne.

This application uses the SimGame framework.
It also demonstrates loading a local module, "simgame/life".

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
    LifeCore "life" // the engine
    Int "threads" // 0=none
    Int "rows"
    Int "cols"
    
    Int "cellPx" // defCellPx
    Palette4 "pal4" // set from palettes list
    Select "selPalette"
    Slider "slider"
    Float "speed" // 0.1 to 10.0
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
  o:.pal4@ o:.life.matAge@ 0 0 o:.base@.pixmap.blit4^
;

: LifeSim_stop
  o:.base.stop^  
;

: LifeSim_pause
  o:.base.pause^  
;

: LifeSim_run
  LifeSim_ensureUiParams
  o:.base.start^ // nop if started 
;

//( idx:i -- ) o( inst )
: LifeSim_new
  LifeSim_pause // note: rendering doesn't happen in stop state
  o:.life.initModel^
  o:.base.doRender^ 
;

: LifeSim#handleWindowClose
  nop
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
  o:.life.nextGen^
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
     
  // LifeCore
  LifeCore.new^ o:.life!
  o:.rows@ o:.life.rows!
  o:.cols@ o:.life.cols!
  o:.threads o:.life.threads!
  true o:.life.useAge!
  o:.life.init^
  o:.life.initModel^
      
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
