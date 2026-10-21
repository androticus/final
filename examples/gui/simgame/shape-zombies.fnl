forget: reset-shape-zombies

requiresFyne
// `examples/gui/simgame/shape-zombies-v4.fnl` compileFromFile
`
Shape Zombies, using the MotionHit engine and HitAreas.
This application uses the SimGame framework.

Blessers are circles (CircleHitArea);
Zombies are axis-aligned squares (RectHitArea);
You are an octagon.
Arrow keys move you (hold for turbo!)
Gain/Lose color by contact...

Try:

shape-zombies
`
const: shape-zombies-blurb


// Game constants and data prototypes

// Create the SimGame delegate object: required field base:SimGame

{      
  "CircZomb" "name"
   SimGame "baseType"
 
   {
 // SimGame "base" // added automatically
    Int "radiusScale"
    
    /*{
      List "borders"
      List "bzs"
      SimEntity "user"
    } */
    DynamicObject "model"
  } "fields"
  
  // optional: add constants
  {
    0 "btnRun"
    1 "btnPause"
    2 "btnNew"

    SimGame.PixSize768x576 "defSize"
    16 "defRadiusScale" // for radius
    0xD0D0D0 Color.hex^ "backgroundColor"
    
    // game constants
    0 "CircKindBless"
    1 "CircKindZombie"
    2 "CircKindUser"
    
    // for Bless and Zombies
    0 "SubkindRedCyan"
    1 "SubkindGreenMagenta"
    2 "SubkindBlueYellow"
    
    // Hit Kinds
    0 "HitKindBZ" // they behave the same way for hits
    2 "HitKindUser"
   
    2 "nBlessSets"
    ( 32 24 20 i) "blessRadii"
    2 "nZombieSets"    
    ( 32 24 20 i) "zombieRadii"
    200.0 "bzBaseSpeed" // in pixels/second, 200 is good
    ( 0.6 1.0 1.4 ) "bzSpeedFactor"
    32 "userRadius"
    14 "userMoveScale" // for arrow keys
    0x7F "initUserColorValue"

    // set odds equal, so users can chase Ascension or Zombification 
     32 "blessColorFactor" // how much per contact? def=12
    -16 "drainColorFactor" // drains x2 colors... def=-6
    
     1 "stateAscended"
     0 "stateAlive"
    -1 "stateZombied"
    (
      Color.white
      Color.black
      Color.gray
    ) "borderColors" // indexed by Kind
    
    (
      Color.red
      Color.green
      Color.blue
    ) "blessColors"
    
    (
      Color.cyan
      Color.magenta
      Color.yellow
    ) "zombieColors"
    
  } "constants"
}
TypedObject.makeType^


// ( -- )
: CircZomb_initModel_createUser
  o:.base.pixSize@ l.sz!  
  SimEntity.new^ >ts
  CircZomb.CircKindUser t:.kind!
  l.sz.w 2 /
  l.sz.h 2 /
  Point.new^ t:.at!
  true t:.anchored! // so it doesn't bounce off BZs
  CircZomb.userRadius t:.data.radius!
  // initial rgb
  3 Int16List.newWithLen^ dup >xs t:.data.rgb!
  CircZomb.initUserColorValue dup dup
  0 x:!()  1 x:!()  2 x:!()  xdrop

  Point.zero t:.data.radius CircZomb.HitKindUser CircleHitArea.new^
  t:.hitArea!
  
  CircZomb.stateAlive t:.data.state!
  ts> dup o:.base.addEntity^
  o:.model.user!
;


// ( subkind:i -- dxy:f )
: CircZomb_randomBZ_xy
  CircZomb.bzSpeedFactor@()
  CircZomb.bzBaseSpeed // in pixels/second
  * 
  128 Rand.int^ 16 max 64 - 128.0 /
  *
;

/*
Set the BZ in motion, random split of vx/vy.
Smaller ones move faster.
( ent:SimEntity -- )
*/
: CircZomb_initModel_setBZmotion
  >ts
  t:.subkind dup
  CircZomb_randomBZ_xy
  t:.vx!
  CircZomb_randomBZ_xy
  t:.vy!
  tdrop
;

// ( radius:n at:Point subkind:i kind:i ha:HitArea -- ent:SimEntity )
: CircZomb_initModel_createBZ
  SimEntity.new^ >ts
  t:.hitArea!
  t:.kind!
  t:.subkind!
  t:.at!
  t:.data.radius!
  ts> dup o:.base.addEntity^
  dup CircZomb_initModel_setBZmotion
;

// Local RectHitArea centered on entity `at`: topLeft (-r,-r), size (2r,2r).
// ( radius:i -- rha:RectHitArea )
: CircZomb_zombieRectHitArea
  l.r!
  l.r -1 * dup Point.new^
  l.r 2 * dup Size.new^
  Rect.new^ CircZomb.HitKindBZ RectHitArea.new^
;

: CircZomb_initModel_createBZs
  List.new^ dup o:.model.bzs! l.bzs!
  o:.base.pixSize@ l.sz! // should be world, not pix, ok for now
  
  CircZomb.nZombieSets l.n!
  // split start field into N+1 y sections
  l.sz.h 2 / l.n 2 + / l.hTranche!
  0 l.n 1 - range
  foreach
    i CircZomb.zombieRadii@() l.radius!
    // share one RectHitArea per radius class; at is square center
    l.radius CircZomb_zombieRectHitArea l.rha!
    0 2 range // subkind / color
    foreach
      l.radius
      
      // x location
      l.sz.w 4 / i 1 + * // quarters
      // y location
      l.sz.h l.hTranche j 1 + * -
      Point.new^
      
      i  CircZomb.CircKindZombie
      l.rha
      CircZomb_initModel_createBZ
      l.bzs.append^
    /foreach
  /foreach    

  CircZomb.nBlessSets l.n!
  // split start field into N+1 y sections
  l.sz.h 2 / l.n 2 + / l.hTranche!
  0 l.n 1 - range
  foreach // set
    i CircZomb.blessRadii@() l.radius!
    // they can share a HitCircle
    Point.zero l.radius CircZomb.HitKindBZ CircleHitArea.new^ l.cha!
    0 2 range // subkind / color
    foreach
      l.radius
      
      // x location
      l.sz.w 4 / i 1 + * // quarters
      // y location
      l.hTranche j 1 + *
      Point.new^
      
      i CircZomb.CircKindBless
      l.cha
      CircZomb_initModel_createBZ
      l.bzs.append^
    /foreach
  /foreach  
;

: CircZomb_initModel
  o:.base.entities.clear^ // for New
  newdo o:.model!

  CircZomb_initModel_createBZs
  CircZomb_initModel_createUser
;

// ( ent:Entity -- col:Color )
: CircZomb_colorForUser
  >ts 
  0 t:.data.rgb@()  1 t:.data.rgb@()  2 t:.data.rgb@() 
  tdrop 
  Color.rgb^
;

// render a Bless, Zombie (square), or User
//( ent:Entity -- ) o( inst )
: CircZomb_renderModel_entity
  o:.base.painter@ l.p!
  >ts
    t:.kind l.kind!
    t:.subkind l.sub!

    l.kind CircZomb.borderColors@()
    l.p.stroke.color!

    l.kind switch
      a CircZomb.CircKindBless == case
        l.sub CircZomb.blessColors@()
      /case
      a CircZomb.CircKindZombie == case
        l.sub CircZomb.zombieColors@()
      /case
      a CircZomb.CircKindUser == case
        t:peek CircZomb_colorForUser
      /case
    /switch
    l.p.fill.color!
    FillStyle.solid l.p.fill.style!
    l.kind switch
      a CircZomb.CircKindZombie == case
        // square centered on at; data.radius = half width/height
        t:.at
        t:.data.radius 2 * dup Size.new^
        l.p@.rect^
      /case
      a CircZomb.CircKindBless == case
        t:.at
        t:.data.radius
        l.p@.circle^
      /case
      default // user
        8 // sides
        t:.at // center
        t:.data.radius  // radius
        l.p@.regularPolygon^
      /default
    /switch
    
    // ascended or zombified user
    l.kind CircZomb.CircKindUser == if
      t:.data.state if
        t:.data.state switch
          a CircZomb.stateAscended == case
            "A" l.ltr!
            0xFFD700 Color.hex^ l.p@.font.color!
          /case
          default // zombified
            "Z" l.ltr!
            Color.gray l.p@.font.color!
          /default
        /switch
        
        "go" 48.0 l.p.font.setFamilyAndSize^
        TextAlign.center TextAlign.middle |
        l.p@.font.align!
        l.ltr t:.at l.p.text^
      then
    then
  tdrop
;

//( -- ) o( inst )
: CircZomb_renderModel
  CircZomb.backgroundColor o:.base.pixmap.fill^
  
  o:.base.entities
  foreach
    i CircZomb_renderModel_entity
  /foreach
;

/*
#update is only used for periodic game status updates.
Motion/Collision handles most every dynamic aspect.
( -- ) o( inst )
*/
: CircZomb_updateStatus
  // score etc.
  
// log latest mhe mean
( o:.base.mhe.avgStepNs.mean 1000000.0 / )
"MHEngine avg step, ms: %8.3f" fmt $$
( o:.base.clock.avgRenderCostNs.mean 1000000.0 / )
"  Render avg, ms: %8.3f\n" fmt $$
;

: CircZomb_pause
  o:.base.pause^  
;

: CircZomb_run
  o:.base.start^ // nop if started 
;

//( idx:i -- ) o( inst )
: CircZomb_new
  CircZomb_pause
  CircZomb_initModel
  o:.base.doRender^ 
;

// add factor, clamping to 0:255
// ( x:i factor:i -- x':i ) 
: CircZomb_blessDrain_clamp
  + 0 max 255 min
;

/*
Update user state after bless/drain
( user:Ent -- )
*/
: CircZomb_blessDrain_updateState
  >ts
  // if sum is 0: zombie, 3x255 then ascend
  Counter.new^ l.sum!
  0 2 range
  foreach
    i t:.data.rgb@() l.sum.add^
  /foreach
  
  l.sum.value
  switch
    a 0 == case
      CircZomb.stateZombied t:.data.state!
    /case
    a 765 == case
      CircZomb.stateAscended t:.data.state!
    /case
  /switch
  
  tdrop
;

/*
Handle one BZ/User interaction.
The user is still active (not trancended or zombified.)
( bz:Ent user:Ent -- )
*/
: CircZomb_blessDrain_inner
  l.user! l.bz!
  ByteList.new^ l.colorTargets!

  l.bz.kind
  switch
    a CircZomb.CircKindBless == case
      CircZomb.blessColorFactor l.cf!
      l.bz.subkind l.colorTargets.append^
    /case
    
    a CircZomb.CircKindZombie == case
      CircZomb.drainColorFactor l.cf!
      // color targets for zombies are more complicated
      // because we drain 2
      l.bz.subkind
      switch
        a CircZomb.SubkindRedCyan == case
          1 2 l.colorTargets.append^ l.colorTargets.append^
        /case
        a CircZomb.SubkindGreenMagenta == case
          0 2 l.colorTargets.append^ l.colorTargets.append^
        /case
        a CircZomb.SubkindBlueYellow == case
          0 1 l.colorTargets.append^ l.colorTargets.append^
        /case       
      /switch
    /case
    
    default // bug
      "Entity was expected to be Bless or Zombie" error
    /default
  /switch

  l.colorTargets
  foreach
    i l.user.data.rgb@() l.cf
    CircZomb_blessDrain_clamp
    i l.user.data.rgb!()
( i l.cf i l.user.data.rgb@() ) "chan: %d delta: %3d new: %2X  " fmt $$  
  /foreach
"\n" $$   
  l.user CircZomb_blessDrain_updateState
;

/*
Handle one BZ/User interaction.
( bz:Ent user:Ent -- )
*/
: CircZomb_blessDrain
  // only if still active
  dup >ts t:.data.state tdrop
  if
    drop2
  else 
    CircZomb_blessDrain_inner
  then
;

/*
Motion Hit handler.
We only asked for Events for BZ-User hits.
Determine the type, and do the bless/drain processing.
( he:HitEvent -- ) o( inst:CZ )
*/
: CircZomb#handleHits
  >ts
//( t:.asOfTick ) "CircZomb#handleHits asOfTick %8d\n" fmt $$
  // entities are sorted by kind:subkind
  t:.hits // SimHit: entity1:Ent entity2:Ent at normal
  foreach // note: could be multiple if several hit in one step
    i >ts
      t:.entity1 t:.entity2
      CircZomb_blessDrain
    tdrop
  /foreach
  tdrop
  // an autorender happens here, and periodic gets pushed
  // forward to next period
;

/*
 Button event handler.
( idx:i -- ) o( inst )
*/
: CircZomb#handleButtonTap
  switch
    a CircZomb.btnRun == case
      CircZomb_run
    /case
    
    a CircZomb.btnPause == case
      CircZomb_pause
    /case
    
    a CircZomb.btnNew == case
      CircZomb_new
    /case    
  /switch
;

: CircZomb#handleRender
  CircZomb_renderModel
;

/*
( -- ) o( inst )
*/
: CircZomb#handleUpdate
  CircZomb_updateStatus
;

/*
We use this to output average update/render rates.
Note: it updates in ticks, not seconds, so if you
set the speed to more than 1 it gets called more frequently.
( index:i -- ) o( inst )
*/
: CircZomb#handleTimerEvent
  drop
  true // use stderr
  o:.base.printAvgUpdateRenderRate^
;

// ( dx:i dy:i user:Entity -- ) o( inst )
: CircZomb_moveUser
  l.user!
  l.user.moveInBounds^
  o:.base.doRender^
;

// ( ke:KeyEvent -- ) o( inst )
: CircZomb_onKeyDown
  l.ke!
  o:.model.user l.user!
  CircZomb.userMoveScale l.mov!
  l.ke.name
  // note: y down is +ve
  switch
    a Keys.left == case
      l.mov -1 * 0 l.user CircZomb_moveUser
    /case
    a Keys.right == case
      l.mov      0 l.user CircZomb_moveUser
    /case
    a Keys.up == case
      0 l.mov -1 * l.user CircZomb_moveUser
    /case
    a Keys.down == case
      0 l.mov      l.user CircZomb_moveUser
    /case  
  /switch
;

/*
KeyDown event handler.
Attached to Window.KeyDown events
( ke:KeyEvent -- ) o( inst )
*/
: CircZomb#handleKeyDown
  CircZomb_onKeyDown
;

/*
KeyUp event handler.
Attached to Window.KeyUp events
( ke:KeyEvent -- ) o( inst )
*/
: CircZomb#handleKeyUp
  drop
  // not used
;

/* Required
Called by SimGame before it does its own internal init logic.
Use this to set up all your non-Fyne objects, and set
the SimGame options you will be using.
Use this to:
* load all your assets, images, sounds, etc
* enable optional SimGame options, like MotionHit engine
( -- ) o( inst:CircZomb )
*/
: CircZomb#preInit
  o:.base@ >ys

  // SimGame options
  "Shape Zombies -  hit: circles to Ascend, squares to be Zombified... your choice..." y:.title!
  true y:.options.hasKeyHandler!
  true y:.options.useKeyAutoRepeat!
  // update is only for things like spawning
  // we don't need any independent update in the current code
  // all updates are done via events, and motion is automatic
  // we just enable for logging of rates
  1000 y:.clock.updateRate! // 1/s
  // we periodically render for smooth motion
  // and set .autoRender on for auto rendering on most events
  50 y:.clock.renderRate! // 20/s
  y:.clock.setAutoRenderOn^
  
  // MotionHitEngine
  // We use Hits to bounce off borders and each other!
  true y:.options.useMotionEngine!
  true y:.options.useHits!
  SimGameOptions.hitActionBounce 
  y:.options.setAllBoundaryModes^
  // Set up the Hit Matrix -- how entities interact
  // BZs just bounce off each other; no event
  CircZomb.HitKindBZ dup SimGameOptions.hitActionBounce
  y:.options.addHitMatrixEntry^
  // BZs also bounce off User, but we get an event, for bless/drain
  CircZomb.HitKindBZ CircZomb.HitKindUser
  SimGameOptions.hitActionBounce SimGameOptions.hitEvents |
  y:.options.addHitMatrixEntry^
  
  50 y:.clock.mheRate! // 20/s
  
  // CircZomb options and setup
  CircZomb.defRadiusScale o:.radiusScale!

  CircZomb_initModel
    
  // Timer/ time-related
  1.0 y:.clock.simSpeed!
  0 1000 y:.timerEventEvery^ // for outputing update rate
  
  ydrop
;

/* Required
Called by SimGame after it does its own internal init logic.
Use this to apply additional options or setup, now that the
features of SimGame have been initialized.
The main Pixmap has been created, and is accessible via o:.base@.pixmap
This is a good place to build your sim/game model.
( -- ) o( inst:CircZomb )
*/
: CircZomb#postInit
;


/* Required if options.hasBottomBar is set (default is true)
Called by SimGame during #makeUi. Return a Fyne layout object
or Widget, typically an HBox or Border that is the content
below the view.
( -- lay:Layout | widg:Widget ) o( inst:CircZomb )
*/
: CircZomb#makeBottomBar
  o:.base@ >ys
  ( // buttons
    CircZomb.btnRun   "Run" Icon.MediaPlay y:.makeButtonWithIcon^
    CircZomb.btnPause "Pause" Icon.MediaPause y:.makeButtonWithIcon^
    CircZomb.btnNew   "New" y:.makeButton^
  ) Box.newH^
  ydrop
;


/*
Typical launch:
*/
: shape-zombies
  CircZomb.new^ >zs
  z:.base >ys
  y:.init^
  y:.makeUi^
  y:.startPaused^
  y:.doRender^
  y:.win@.show^
  ydrop zdrop
;


shape-zombies-blurb $ nl
