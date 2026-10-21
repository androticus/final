/*
Docs and extension methods for SimGame and options.
*/

/*
Playfield side index: left, top, right, bottom (0..3).
*/
:doc Boundary

/*
Left playfield side (0).
*/
:doc Boundary#left

/*
Top playfield side (1).
*/
:doc Boundary#top

/*
Right playfield side (2).
*/
:doc Boundary#right

/*
Bottom playfield side (3).
*/
:doc Boundary#bottom

/*
Outward normals for the four sides, indexed by Boundary.left … .bottom.
Points (-1,0) (0,-1) (1,0) (0,1). Shared List.
*/
:doc Boundary#normals


/*
Packed playfield-side policy bits (Int class constants). OR CrossAction, HitAction, and event bits.
Cross tests at (point-like). Hit tests HitArea vs an engine kindBorder slab.
*/
:doc BoundaryMode

/*
CrossAction: do not clamp or change velocity. Combine with crossEvents for a leave report.
*/
:doc BoundaryMode#crossActionNone

/*
CrossAction: clamp at to the inner face; set entity.stopped. Leaves vx/vy.
*/
:doc BoundaryMode#crossActionStop

/*
CrossAction: reverse that axis of vx/vy and clamp at inside. Does not set stopped.
*/
:doc BoundaryMode#crossActionBounce

/*
If set, emit a Crossing when this step first puts at past this side (none, stop, or bounce).
Requires #handleCrossings. Distinguish stop vs none via entity.stopped.
*/
:doc BoundaryMode#crossEvents

/*
HitAction: no engine slab on this side.
*/
:doc BoundaryMode#hitActionNone

/*
HitAction: engine kindBorder HitRect; freeze-for-game (set stopped). Combine with hitEvents to notify.
*/
:doc BoundaryMode#hitActionStop

/*
HitAction: engine kindBorder HitRect; in-engine reflect. Combine with hitEvents to notify.
*/
:doc BoundaryMode#hitActionBounce

/*
If set, enqueue a Hit (kindBorder) in addition to HitAction. Requires #handleHits when used.
*/
:doc BoundaryMode#hitEvents


/*
Options to a SimGame. Set these in your handlePreInit.
*/
:doc SimGameOptions

/*
Set this to add a Menu Bar to the top, above the view.
You must supply a #makeMenuBar callback method.
*/
:doc SimGameOptions#hasMenuBar

/*
Set this to add a bottom bar, such as for buttons or options.
You must supply a #makeBottomBar callback method.
*/
:doc SimGameOptions#hasBottomBar

/*
Left playfield side packed BoundaryMode bits (Int).
*/
:doc SimGameOptions#leftBoundaryMode

/*
Top playfield side packed BoundaryMode bits (Int).
*/
:doc SimGameOptions#topBoundaryMode

/*
Right playfield side packed BoundaryMode bits (Int).
*/
:doc SimGameOptions#rightBoundaryMode

/*
Bottom playfield side packed BoundaryMode bits (Int).
*/
:doc SimGameOptions#bottomBoundaryMode

/*
Set all four sides to the same packed BoundaryMode bits.
( mode:i -- )
*/
:doc SimGameOptions#setAllBoundaryModes

/*
Add a pair of HitArea actions/event.
hak* refer to the values you set in HitArea.new factories.
Use HitArea.kindBorder to refer to the borders.
Use HitArea.kindAll as a wildcard for one of the items.
You only have to make one call per pair; modes are transitive
i.e. they apply to both items, regardless the order specified.
( hak1:i hak2:i modebits:i -- )
*/
:doc SimGameOptions#addHitMatrixEntry

/*
True if any side has CrossEvents.
You must supply a #handleCrossings event handler.
*/
:doc SimGameOptions#useCrossings

/*
Set to enable HitArea events between gameplay entities.
Provide a HitArea for each participating entity.
You must supply a #handleHits event handler.
Also required (even if false) if any side has HitEvents.
*/
:doc SimGameOptions#useHits

/*
Author switch: run the MotionHit engine. Not derived from sides.
May be true with all boundary bytes 0 (integrate only).
Requires clock.mheRate > 0.
*/
:doc SimGameOptions#useMotionEngine

/*
Set this to receive KeyEvents (down and up.)
You must supply #handleKeyDown and #handleKeyUp event handlers.
*/
:doc SimGameOptions#hasKeyHandler

/*
Independent of hasKeyHandler; does not enable key handlers.
Hold starts repeating after 400 ms; a quick same-key tap then hold after 200 ms.
*/
:doc SimGameOptions#useKeyAutoRepeat

/*
Set this to receive a Window onClose event.
Normally not needed; use for resources that need
to be closed or released.
You must supply a #handleWindowClose event handler.
*/
:doc SimGameOptions#hasWindowCloseHandler

/*
Set this to get MouseDown and MouseUp events.
You must supply #handleMouseDown and #handleMouseUp
event handlers.
*/
:doc SimGameOptions#hasMouseDownUpHandler

/*
Set this to get MouseDown and MouseUp events.
You must supply #handleMouseIn, #handleMouseMoved and
#handleMouseOut event handlers.
*/
:doc SimGameOptions#hasMouseInMovedOutHandler

/*
Set this to get MouseScroll events (scroll wheel.)
You must supply #handleMouseScrolled event handler.
*/
:doc SimGameOptions#hasMouseScrolledHandler


/*
MotionHit engine (sg.mhe). Background fused step while Running; ceiling is clock.mheRate.
*/
:doc SimMHEngine

/*
If true, paint after #handleUpdate and after #onMotion. Independent of renderRate (0=no periodic paint). Not mouse-move.
*/
:doc SimClock#autoRender

/*
Author target ms between SimMHEngine fused steps. 0=off, 1=asap. Copied to mhe.actualRate (may coarsen under load). Not updateRate.
*/
:doc SimClock#mheRate

/*
If true, SimMHEngine skips motion integrate this cycle. Set by CrossAction.stop and freeze-for-game; authors may set/clear. Does not zero vx/vy.
*/
:doc SimEntity#stopped

/*
Move by dx, dy, keeping the entity in the playfield.
at is a hard stop inside (0,0)…(w-1,h-1). A HitArea further
constrains the move so the volume stays on the pixmap when it fits.
( dx:n dy:n -- )
*/
:doc SimEntity#moveInBounds

/*
Playfield clear fill used by the library paint walk (before DrawerSets).
Default Color.black. Set in #preInit (e.g. Invaders.backgroundColor sg.backgroundColor!).
If .a is 0 (Color.transparent), the library skips the fill — use when del paints every pixel (Life).
Do not full-frame fill again in del#handleRender after drawers — that would wipe them.
*/
:doc SimGame#backgroundColor

/*
The SimClock object.
*/
:doc SimGame#clock

/*
Request a one-shot timer: fire after `after` sim-ms from now.
del must implement #handleTimerEvent. after must be > 0.
( index:i after:i -- )
*/
:doc SimGame#timerEventAfter

/*
Request a one-shot timer as soon as clock.ticks reaches `tick`.
A past tick fires on the next metronome wake.
( index:i tick:i -- )
*/
:doc SimGame#timerEventAt

/*
Request a periodic timer. First fire at now+ticks, then every ticks sim-ms.
Catch-up coalesces to one dispatch per wake. ticks must be > 0.
del must implement #handleTimerEvent.
( index:i ticks:i -- )
*/
:doc SimGame#timerEventEvery

/*
Cancel all timer registrations with this index.
An Event already queued on the EventLoop may still run.
( index:i -- )
*/
:doc SimGame#cancelTimerEvent

/*
The user's TypedObject passed into .new .
*/
:doc SimGame#del

/*
The current list of enabled entities.
Use .addEntity to add to this list.
*/
:doc SimGame#entities

/*
The MotionHit engine. Bound in #init .
*/
:doc SimGame#mhe

/*
A Painter object on which you can draw to the Pixmap.
Note: do not create your own Painter, use this one.
*/
:doc SimGame#painter

/*
The Pixmap on which everything is blitted/drawn.
*/
:doc SimGame#pixmap

/*
For convenience: clock.ticks .
*/
:doc SimGame#ticks

/*
The Title shown in the main Window.
*/
:doc SimGame#title

/*
The Fyne CanvasImage widget that contains the Pixmap image.
*/
:doc SimGame#view

/*
The Fyne Window object.
*/
:doc SimGame#win

/*
Create a one-time event that is triggered after the 
requested delay, in ticks. Handle in #handleTimerEvent
( index:i -- ) // #handleTimerEvent REQUIRED
( index:i delay:i -- )
*/
:doc SimGame#timerEventAfter

/*
Create a one-time event that is triggered on or asap after 
the requested clock.tick value. Handle in #handleTimerEvent
( index:i -- ) // #handleTimerEvent REQUIRED
( index:i tick:i -- )
*/
:doc SimGame#timerEventAt

/*
Create a periodic event that is triggered every requested 
tick interval. Handle in #handleTimerEvent
( index:i -- ) // #handleTimerEvent REQUIRED
( index:i ticks:i -- )
*/
:doc SimGame#timerEventEvery

/*
Cancel all timer events with the requested index.
NOTE: a pending event may still get sent.
( index:i -- )
*/
:doc SimGame#cancelTimerEvent

/*
*System Use Only*
Mint/rebuild engine-owned kindBorder HitRect slabs for sides with HitAction ≠ none.
Not on the gameplay entity list. Called from #init after SimGame_init.
( -- )
*/
:doc SimGame#_ensureHitBoundaries

/*
*System Use Only*
Copy user clock.simSpeed onto clock.engineSpeed after render costs are warm.
No-op while cold. Called from SimGame_doRenderInner.
( -- )
*/
:doc SimGame#_publishEngineSpeed
  
/*
Working fused-step period in ms (may be coarser than clock.mheRate under load).
*/
:doc SimMHEngine#actualRate

/*
Dispatcher finished a fused MotionHit Event. Token is that event's asOfTick.
( asOfTick:i -- )
*/
:doc SimMHEngine#ack

/*
One fused integrate / bounds / hits cycle. No-op until physics is filled in.
( -- )
*/
:doc SimMHEngine#step

/*
Attach SimGame#onMotion as the EventLoop dispatcher.
( handler:MethodRef -- )
*/
:doc SimMHEngine#bindOnMotion

/*
One playfield boundary crossing occurrence.
Crossing is based on entity `at` location.
*/
:doc SimCrossing

/*
The entity whose `at` reached a boundary.
*/
:doc SimCrossing#entity

/*
The boundary crossed. One of:
Boundary.left/top/right/bottom constants.
*/
:doc SimCrossing#boundary

/*
The point in world coordinates at which the entity's
`at` crossed the border.
*/
:doc SimCrossing#at

/*
This step's crossing list. Authors: foreach (and .len). Not a general collection.
*/
:doc SimCrossings

/*
Event payload for #handleCrossings. Iterate .crossings, not .hits.
*/
:doc SimCrossingEvent

/*
Clock ticks when this engine cycle ran.
*/
:doc SimCrossingEvent#asOfTick

/*
Crossings from this cycle (foreach).
*/
:doc SimCrossingEvent#crossings

/*
One pairwise HitArea contact (freeze-for-game or info).
Entity1 Entity2 are sorted by Kind/Subkind ascending order,
so you only need to switch on the diagonal of the hit matrix.
Boundaries are assigned Kind=-1, so will be the first entity
if you keep your Kind values >= 0.
*/
:doc SimHit

/*
One of the entities involved, Kind/Subkind order <= #2
*/
:doc SimHit#entity1

/*
The other entity involved, Kind/Subkind order >= #1
*/
:doc SimHit#entity2

/*
The point in world coordinates at which the entities touched.
Note: often not needed in sim/game logic.
*/
:doc SimHit#at

/*
The normal vector in world coordinates at which the entities touched.
Note: often not needed in sim/game logic.
*/
:doc SimHit#normal

/*
This step's hit list. Authors: foreach (and .len). Not a general collection.
*/
:doc SimHits

/*
Event payload for #handleHits. Iterate .hits.
*/
:doc SimHitEvent

/*
Clock ticks when this engine cycle ran.
*/
:doc SimHitEvent#asOfTick

/*
Hits from this cycle (foreach).
*/
:doc SimHitEvent#hits


/*
SimGame is always composed with a user TypedObject delegate (del).

Preferred (based TypedObject — see 183.2): TypeDef sets SimGame "baseType";
MyType.new^ auto-mints SimGame via _newWithDelegate and stores it in .base.
O: is the extension; engine access is o:.base.…

Compat (examples until migrated): declare SimGame "sg"; call newWithDelegate^
which wires del.sg. User hooks live on del; SimGame Final methods call them
via o:.del@.
*/

/*
( del:TypedObject -- game:SimGame )
Base factory for based TypedObject mint. Requires a delegate with the
required Final #handle* / #do* / #make* methods. helpx only.
*/
:doc SimGame#_newWithDelegate

/*
Print the average update and render rates to stdout or stderr.
( useStderr:b -- ) o( inst )
*/
: SimGame#printAvgUpdateRenderRate
  >ts
  // second part of line is built first
  ( o:.clock.avgRenderCostNs.mean 1000000.0 / )
  "Avg render rate ms: %7.3f\n" fmt
  ( o:.clock.avgUpdateCostNs.mean 1000000.0 / )
  "Avg update rate ms: %7.3f  " fmt
  ts> if
    $$ $$
  else
    $ $
  then
;

/*
MouseDown and Up combined handler; we demultiplex to dispatch.
Only registered if option selected.
( ev:Event -- ) o( inst:SimGame )
*/
: SimGame#onMouseDownUp
  >ts
  0 t:.params@() // MouseEvent
  t:.index tdrop
  switch
    a 0 == case o:.del.handleMouseDown^ /case
    a 1 == case o:.del.handleMouseUp^ /case
    default drop /default // note: not expected, defensive
  /switch
;

/*
MouseIn / Moved / Out combined handler; we demultiplex to dispatch.
Only registered if option selected.
Does not autoRender — mouse-move is too frequent; periodic renderRate covers tracking.
( ev:Event -- ) o( inst:SimGame )
*/
: SimGame#onMouseInMovedOut
  >ts
  t:.index
  switch
    a 0 == case 0 t:.params@() tdrop o:.del.handleMouseIn^ /case
    a 1 == case 0 t:.params@() tdrop o:.del.handleMouseMoved^ /case
    a 2 == case tdrop o:.del.handleMouseOut^ /case 
    default tdrop /default // note: not expected, defensive
  /switch  
;

/*
MouseScrolled handler.
Only registered if option selected.
( ev:Event -- ) o( inst:SimGame )
*/
: SimGame#onMouseScrolled
  >ts
  0 t:.params@() // ScrollEvent
  tdrop
  o:.del.handleMouseScrolled^
;

/*
Window closing handler.
Stops EventTimer + SimClock metronome via #close; optional del#handleWindowClose.
( ev:Event -- ) o( inst:SimGame )
*/
: SimGame#onWindowClose
  drop
  o:.close^
  o:.options.hasWindowCloseHandler@ if
    o:.del@.handleWindowClose^
  then
;


/*
Handles a fused MotionHit Event from SimMHEngine.
params[0] = SimCrossingEvent or null; params[1] = SimHitEvent or null.
Demux to del#handleCrossings / #handleHits, ack once, then autoRender if set.
( ev:Event -- ) o( inst:SimGame )
*/
: SimGame#onMotion
  >ts t:.params@ >ts
  0 t:@() >xs
  1 t:@() >ys
  t:drop t:drop
  x:peek Null? not if
    x:peek o:.del@.handleCrossings^
  then
  y:peek Null? not if
    y:peek o:.del@.handleHits^
  then
  x:peek Null? not if
    x:.asOfTick@
  else
    y:.asOfTick@
  then
  o:.mhe@.ack^
  x:drop y:drop
  o:.clock.autoRender if
    o:.doRender^
  then
;


/*
Handles any buttons events from #makeButton registrations.
( ev:Event -- ) o( inst:SimGame )
*/
: SimGame#onButtonTap
  >ts t:.index@ t:drop
  // ( idx:i -- )
  o:.del@.handleButtonTap^
;

/*
Library EventLoop demux for clock timer events.
Unpacks ev.index and calls del#handleTimerEvent.
( ev:Event -- ) o( inst:SimGame )
*/
: SimGame#onTimerEvent
  >ts t:.index@ t:drop
  o:.del@.handleTimerEvent^
;

/*
Makes a button, with event handlers.
( index:i text:s -- btn:Button ) o( inst:SimGame )
*/
: SimGame#makeButton
  >ts >ts
  o:.onButtonTap~ ts> ts>
  Button.new^
;

/*
Makes a button, with event handlers and icon
( index:i text:s icon:fyne.Resource -- btn:Button ) o( inst:SimGame )
*/
: SimGame#makeButtonWithIcon
  >ts >ts >ts
  o:.onButtonTap~ ts> ts> ts>
  Button.newWithIcon^
;

/*
Library Drawer walk: advance each entity's cycler, then DrawerSet#invokeRender
(dispatches set.render MethodRef — auto-bound to drawer#handleRender by .new).
( -- ) o( sg:inst )
*/
: SimGame_drawDrawerSets
  o:.clock.ticks@ l.ticks!
  o:.entitiesWithDrawerSets^
  foreach
    i >ts
      t:.drawerSet@ >xs
        l.ticks@ x:.cycler@.advance^
        t:peek
        x:.cycler@
        x:.invokeRender^
      x:drop
    t:drop
  /foreach
;

/*
Factored Refresh routine -- already in a running/suspend sandwich.
Called by user #doUpdate, or from #onTimer.
Fill backgroundColor (skipped when A==0 / Color.transparent) → DrawerSets → del#handleRender → refresh.
( -- ) o( sg:inst )
*/
: SimGame_doRenderInner
  o:.clock.renderStarted^
    o:.backgroundColor.a if
      o:.backgroundColor@ o:.pixmap.fill^
    then
    SimGame_drawDrawerSets
    o:.del@.handleRender^
  o:.clock.renderEnded^
  o:.refreshView^
  o:._publishEngineSpeed^
;

/*
Factored Refresh routine, callable by user or after StartPaused for the load screen.
Allowed when Running or Paused (not Stopped). Updates stay Running-only.
( -- ) o( sg:inst )
*/
: SimGame#doRender
  o:.clock.state@ SimClock.stateStopped <> if
    true o:.setSuspended^
      SimGame_doRenderInner
    false o:.setSuspended^
  then
;

/*
Factored Update routine -- already in a running/suspend sandwich.
Called by user #doUpdate or Timer#onTimer
Also calls doRefreshInner if AutoRefresh is on.
( -- ) o( sg:inst )
*/
: SimGame_doUpdateInner
  o:.clock.updateStarted^
    o:.del@.handleUpdate^
  o:.clock.updateEnded^
  
  // auto refresh on?
  o:.clock.autoRender if
    SimGame_doRenderInner
  then
;

/*
Update routine, callable by user.
Also calls Refresh if AutoRefresh is on.
( -- ) o( sg:inst )
*/
: SimGame#doUpdate
  o:.clock.state@ SimClock.stateRunning == if
    true o:.setSuspended^
      SimGame_doUpdateInner
    false o:.setSuspended^
  then
;

/*
Timer handler: demux update/render, suspend sandwich, reprime.

When clock.updateRate > 0, del must implement #handleUpdate.
When clock.renderRate > 0, del must implement #handleRender.
Periodic lanes are independent. clock.autoRender additionally paints after update / #onMotion (not mouse-move).
( -- ) o( inst:SimGame )
*/
: SimGame#onTimer
  o:.clock.state@ SimClock.stateRunning == if
    true o:.setSuspended^
      o:.clock.ticks@ l.ticks!
      // update before render when both due
      // update also paints if autoRender
      l.ticks@ o:.clock.shouldUpdate^ if
        SimGame_doUpdateInner
      then
      l.ticks@ o:.clock.shouldRender^ if
        SimGame_doRenderInner
      then
      false o:.setSuspended^
    o:.reprimeTimer^
  then
;

/*
Keyboard event handler.
Only attached if options.hasKeyHandler.
Attached to both the KeyDown and KeyUp events.
We demultiplex via .index to dispatch.
( ev:Event -- ) o( inst )
*/
: SimGame#onKey
  dup >ts
  Keys.decode^ // KeyEvent
  t:.index tdrop
  switch
    a 0 == case o:.del@.handleKeyDown^ /case
    a 1 == case o:.del@.handleKeyUp^    /case
    default drop /default // defensive, shouldn't happen
 /switch
;

/*
Factored worker
( -- ) o( inst:SimGame )
*/
: SimGame_prePreInit
  // create the timer
  o:.onTimer~ EventTimer.new^ o:._timer!

  // MotionHit EventLoop dispatcher (Fyne CreateEventFn pattern)
  o:.onMotion~ o:.mhe@.bindOnMotion^
;

/*
Factored worker
( -- ) o( inst:SimGame )
*/
: SimGame_init
  // create the Pixmap
  o:.pixSize.h@
  o:.pixSize.w@
  Pixmap.new^ o:._pixmap!
  
  // After pixmap exists; before postInit so authors can spawn assuming walls.
  o:._ensureHitBoundaries^
;

/*
Calls:
* very early prePreInit needed for Final objects
* the required method #preInit on del,
* internal initialization,
* the required method #postInit on del.
( -- ) o( inst:SimGame )
*/
: SimGame#init
  SimGame_prePreInit
  o:.del@.preInit^
  SimGame_init
  o:.del@.postInit^
;

/*
Create the UI
( -- ) o( inst:SimGame )
*/
: SimGame#makeUi
  // create the window with close handler
  App.app o:.title@
  Window.new^ dup >ts o:._win!
  o:.onWindowClose~ 0 t:.setOnClosed^
  
  // key handlers, only if enabled
  o:.options.hasKeyHandler@ if
    o:.onKey~ 0 over 1
    o:.options.useKeyAutoRepeat@ // demultiplexes via .index to dispatch
    t:.setOnKeyDownUp^ 
  then
  tdrop

  // make our overall Border control
  // top: menuBar or null
  o:.options.hasMenuBar if
    o:.del@.makeMenuBar^
  else
    null
  then
  
  // bottom: bottomBar or null
  o:.options.hasBottomBar if
    o:.del@.makeBottomBar^
  else
    null
  then
  
  // left, right
  null null
  
  // center, our canvas view of the pixmap
  o:.pixmap@.image CanvasImage.new^ dup dup >ts o:._view!
  Border.new^ // carries over the mouse code below
    
  // mouse event handlers, now that view exists
  o:.options.hasMouseDownUpHandler@ if
    o:.onMouseDownUp~ 0 over 1
    t:.setOnMouseDownUp^
  then
  o:.options.hasMouseInMovedOutHandler@ if
    o:.onMouseInMovedOut~ 0 over 1 over 2
    t:.setOnMouseInMovedOut^
  then
  o:.options.hasMouseScrolledHandler@ if
    o:.onMouseScrolled~ 0
    t:.setOnMouseScrolled^
  then
  tdrop // canvasimage

  // set the window content, and freeze its size
  o:._setContentFixedSize^
  
;
