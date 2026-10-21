"Do not load this file; copy/paste/edit it, and remove this line" error
/*
This is a template for a Simulation or Game using the
SimGame framework (composition / delegation).

It contains the required code and methods, with comments on 
how to use them.
Not all methods are used in every instance, some only get called
when you enable an option, such as the MotionHit engine.

See life-fyne-poc.fnl for a simple Sim example.
*/


// Create your TypedObject; required field sg:SimGame

{      
  "MyGame" "name"
  
  { // add your fields: Type "name"
    SimGame "sg" // required — wired by SimGame.newWithDelegate^
    Int "myIntProp"
    // more fields
  } "fields"
  
  // optional: add constants
  {
    0 "DefaultProp"
    // more constants
  } "constants"
}
TypedObject.makeType^

// Your MyGame Type now exists

// Most of your code will go here, before the events are created

// Callback methods -- required ones are marked as such
// Just leave unneeded ones empty. All Methods will provide
// your MyGame instance on the O Stack.
// Access SimGame via o:.sg@ (title, options, pix, makeButton, …).


/*
Timer Event
Required if you call timerEventAfter, timerEventAt, or
ctimerEventEvery.
It returns the index you used when you created the event.
Use different indexes if you have multiple types of events.
( index:i -- )
*/

: MyGame#handleTimerEvent
  drop
;


// Mouse Events

/*
Required if you use hasMouseDownUp option.
( ev:MouseEvent -- ) o( inst:MyGame )
*/
: MyGame#handleMouseDown
  drop
;

/*
Required if you use hasMouseDownUp option.
( ev:MouseEvent -- ) o( inst:MyGame )
*/
: MyGame#handleMouseUp
  drop
;

/*
Required if you use hasMouseInMovedOut option.
( ev:MouseEvent -- ) o( inst:MyGame )
*/
: MyGame#handleMouseIn
  drop
;
/*
Required if you use hasMouseInMovedOut option.
( ev:MouseEvent -- ) o( inst:MyGame )
*/
: MyGame#handleMouseMoved
  drop
;

/*
Required if you use hasMouseInMovedOut option.
NOTE: no event
( -- ) o( inst:MyGame )
*/
: MyGame#handleMouseOut
;

/*
Required if you use hasMouseScrolled option.
( ev:ScrollEvent -- ) o( inst:MyGame )
*/
: MyGame#handleMouseScroll
  drop
;

/*
Motion Hit handler.
Required if you use useMotionHit via:
* any boundary with hitEvents flag
* Hit Matrix with any hitEvents flag
( he:HitEvent -- ) o( inst:MyGame )
*/
: MyGame#handleHits
  drop
;

/*
Motion Crossing handler.
Required if you use useCrossEvents flag:
* on any boundary
Only called when there are any crossings.
( ce:CrossingEvent -- ) o( inst:MyGame )
*/
: MyGame#handleCrossings
  drop
;

/*
Required if you use renderRate, autoRender or manual #doRender calls
( -- ) o( inst )
*/
: MyGame#handleRender
// render (draw) your model
;

/*
Required if you use updateRate or manual #doUpdate calls
( -- ) o( inst:MyGame )
*/
: MyGame#handleUpdate
// update your model
;

/* Required
Called by SimGame before it does its own internal init logic.
Use this to set up all your non-Fyne objects, and set
the SimGame options you will be using.
Use this to:
* load all your assets, images, sounds, etc
* enable optional SimGame options, like MotionHit engine
( -- ) o( inst:MyGame )
*/
: MyGame#preInit
;

/* Required
Called by SimGame after it does its own internal init logic.
Use this to apply additional options or setup, now that the
features of SimGame have been initialized.
The main Pixmap has been created, and is accessible via o:.sg@.pix
This is a good place to build your sim/game model.
( -- ) o( inst:MyGame )
*/
: MyGame#postInit
;


/* Required if options.hasMenuBar is set (default is false)
Called by SimGame during #makeUi. Return a MenuBar object.
( -- mb:MenuBar ) o( inst:MyGame )
*/
: MyGame#makeMenuBar
// make a MenuBar, with Menu and MenuItem content
;

/* Required if options.hasBottomBar is set (default is true)
Called by SimGame during #makeUi. Return a Fyne layout object
or Widget, typically an HBox or Border that is the content
below the view.
( -- lay:Layout | widg:Widget ) o( inst:MyGame )
*/
: MyGame#makeBottomBar
// make a layout, eg buttons in an HBox
// or a Border, or a Widget
// use o:.sg@.makeButton^ / makeButtonWithIcon^ for chrome buttons
;

/* Required if any makeButton calls were made
Handle the button press for the button of the index.
( btnIndex:i -- ) o( inst:MyGame )
*/
: MyGame#handleButtonTap
// you will typically consume btnIndex in a switch
  drop
;

/* Required if options.hasWindowCloseHandler set
Called when the main window closes.
Use this to release resources, especially open Pipes.
See life.fnl for an example of closing Pipes.
SimGame already disposes of its timers.
( -- ) o( inst:MyGame )
*/
: MyGame#handleWindowClose
;

/*
Typical launch:

: my-game
  MyGame.new^ dup >zs
  SimGame.newWithDelegate^ >ys   // sets z.sg from y
  y:.init^
  y:.makeUi^
  y:.startPaused^
  y:.doRender^ // initial render
  y:.win@.show^
  ydrop zdrop
;
*/
