# Simulations and Games

Final provides a framework called SimGame to help create Simulations and Games. There are several example programs that exercise most of the functionality of the framework, and can serve as a reference or template for new sims and games:
* life.fnl -- Conway's Game of Life -- a classic and beguiling cellular automata simulation 
* shape-zombies.fnl -- Shape Zombies - a simple but surprisingly fun game of hitting Bless circles and avoiding Zombie squares
* invaders.fnl -- Invaders -- a simple take on the classic Space Invaders, that automatically gets harder the better you perform

NOTE: the SimGame framework uses the Fyne GUI library and is only available when running final-gui.

The framework provides the following constructs and facilities:

## SimGame

`SimGame` is the top-level organizational/containing type for any simulation or game it provides:

* creation of the GUI, with hooks for adding a menu and bottom bar/area of any content
* a hybrid composition of a Final catalog type with your own TypedObject for holding your game objects and state, and writing handler methods for things like rendering, key/mouse events, etc
* support for keyboard and mouse support
* a Game clock which advances in millisecond increments, and can be used for automatically triggering render and update events
* an optional MotionHit engine, described below
* render dispatching of Drawer objects you may provide, respecting z-order

Creating a SimGame instance involves two steps:
1. create a TypedObject that will hold your game state, event handlers, and methods
2. create a SimGame object, passing it an instance of your TypedObject

The SimGame instance will be available to you in all your event handlers and methods as the .sg property on your TypedObject instance. Likewise, the SimGame has your instance, and uses it to send events to.

## Coordinates

By default, the game world and viewport are the same, with 0,0 at the upper left, x extending postiive right, y down. You can create a Sim/Game that is larger than the Pixmap used to visualize it, by setting the Pixmap.viewOrigin to a non 0,0 point in your world coordinates. The blitting and drawing routines will then accept your world coordinates.

## SimGameClock

This clock operates continuously on a background thread, and updates a SimGame.tick counter with milliseconds since Start. This value becomes a reference for all other time-based operations in the game, and serves as your reference to real time in your updating and rendering handlers. It is used:
* to trigger `update` and `render` events, if those are enabled;
* to cycle animations and effects that have time-based phases;
* to time and dispatch any time-based events you request via timerEventAfter/At/Every apis.

## Entities

An `Entity` is a visible or invisible object or effect in your game. In Shape Zombies, the Entities are the circles, squares, and the user hexagon. In Invaders, the Entities are the the star field, aliens, and the cannon. You create an instance of an `Entity` for every instance of a game object. An Entity has these key properties:

Property | Type | Description
--- | --- | ---
data | flex | Defaults to a DynamicObject, but can take any Final type -- put your custom Entity data in here
sg | SimGame | the SimGame instance
kind | Int | an ordinal you manage that reprsents a major kind of Entity<br>Zombies has 3: Bless, Zombie, and User
subkind | Int | you can have subkinds of your Kinds; Zombies used 0/1/2 for the three Bless or Zombie color sets
state | Int | an ordinal you can use for a state of the entity, if it has states
at | Point | where the entity is in the game world
hitArea | HitArea | an optional HitArea if the entity will be involved in collisions or boundary actions
drawerSet | DrawerSet | the current visual representation
vx, vy | Float | if it is moving the x and y velocity components, used by Motion engine
zorder | Int | from -ve background to positive, the order things should get drawn ("in front of"/"behind")
anchored | Bool | a fixed place Entity, or one that does not move when hit
visible | Bool | if it is visible; if not, it is not drawn
enabled | Bool | a "soft switch"; not enabled won't participate in motion/collisions
active | Bool | is live in the Entities collection; false before adding to game, and when removed

## HitAreas

A HitArea defines a spatial extent that is used by the Hit and Boundary collision processes. Every Entity that participates in that process needs a HitArea. Two types are provided:
* CircleHitArea -- whose center defaults to the `at` of its Entity, with a defined radius
* RecHitArea -- a rectangulare are whose center defaults to the `at` of its Entity, with defined width/height

A HitArea is not a visible element--it is just an aproximate (or exact) outline of the entity. It is does not have to exactly correspond to the visible extent of the Entity.

**One HitArea instance per entity.** Do not share a single HitArea object across multiple entities. MotionHit tracks contacts by HitArea identity; sharing collapses distinct bodies into one contact and causes wrong stop/enter/event behavior. Mint (or clone) a HitArea when you spawn each entity. Drawers and image mats may still be shared; HitAreas may not. You may clear `.hitArea` to `null` on a live entity (e.g. non-colliding debris); replacing a non-nil HitArea while the entity is active is not allowed.

## MotionHit Engine

The `MotionHit Engine (MHE)` is a centerpiece of SimGame. It is a background process that moves entities with non-zero vx/vy smoothly in the game space, and registers "hits" between registered Entities, or between Entities and an enabled boundary. Because it works in the background, on its own thread, it does not take CPU time away from game logic or rendering. It operates precisely, using a high-precision version of game time, and floating point positions and changes. So while you use the simple integral ms game clock at the level of game actions and events, the MHE is operating much more precisely, without "chunking" or rounding.

### Action and Event options

All the action and event options of MHE share a common vocabulary, summarized below:

Option | Description
--- | ---
none/off | the option is not used in the context
stop | the entity will stop
bounce | the entity will rebound
events | the defined contact will raise an event

### Motion

The engine can move Entities in space, updating their `at` coordinate, according to the velocity set via their vx/vy (in pixels/second). Motion does not require HitAreas. See the following sections for boundary an hit processing of moving entities.

You enable Motion by setting these properties (we show typical code contexts)
```
// o( inst:TypedObject )
true o:sg.useMotionEngine!
// entity on t:
<xvalue> t:.vx!
<yvalue> t:.vy!
// OR
<xvalue> <yvalue> t:.setVelocity^
```
That's it! As long as your entity is `enabled` (default) and not `stopped` the MHE will now move it around, by changing its `at` value. All you need to do is render it using its `at` location in your `handleRender` routine.

## Boundary Crossings or Hits

There are two kinds of boundary options:
* `Cross`, based only on the `at` centroid, which doesn't need a HitArea
* `Hit`, which is based on an entity's HitArea (required for this mode)

The SimGame object has separate action and event properties for each of: left, top, right, and bottom boundary. For example, Invaders doesn't need its left and right boundaries. There is a convenience operator `setAllBoundardModes` that you can call when the options are the same for all four boundaries. You provide both the Cross: none/stop/bounce + event on/off, and Hit: none/stop/bounce + event on/off for a boundary. You are free to "mix and match" both types, which will depend on your entity definitions, and whether you use HitAreas. You are not required to provide a HitArea for all your entities, only the ones you want to participate in the hit-based operations.

For example, just using `at` and Cross detection, an Entity can drift partially offscreen until its `at` reaches the boundary; but with Hit boundary detection, the entity stops or bounces at its border. An event can be triggerd by either type of boundary check, which creates a `bounds` event that you handle. Invaders uses this to detect when a Phaser shot leaves the screen at the top, or an alien hits the ground. In both cases it uses the `stop` option, and the event then handles the occurence, such as removing the phaser shot from the playfield, and scoring a loss for the alien hitting the ground. Shape Zombies uses the `bounce` option, so the Bless and Zombie entities bounce off the boundaries, the same way they bounce off each other.

If you don't want your entity to move when hit, you set the `anchored` property. This just means that it is either a permanent fixture like a wall or barrier, or that the entity doesn't other move when hit.
```
true t:.anchored!
```

### Hit Detection

The Hit engine can detect collisions between any two entities with a HitArea, or between entities and a boundary, when that boundary's Hit option is `stop` or `cross`. As with Motion, an event can be enabled, providing an easy way to process these occurrences, without having to manually iterate your entities, and solve the complex NxN-1 problem!

**HitRegistration**

During the init event of your simgame, you register what happens in collisions between pairs of entities, and between an entity and a boundary. The registration uses an integral `HitKind` value, which you define to group similar entities together, and differebt fior collision semantics with other entities gets its own kind. Unregistered entities simple pass through each other. Registration is transitive, so you only need to register each pair once. The HitKind gets set on the HitArea. The boundaries have a system-defined HitKindBoundary (-1).

Here is the init code for ShapeZombies showing all the calls. `y` contains the SimGame object

```forth
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
```

## Cyclers, Drawers, and DrawSets

### Cycler

A Cycler is an auxiliary object with properties related to cyclical or counted rendering processes.

Property | Type | Description 
--- | --- | ---
start | Int | The Simgame.tick when the object was activated -- used for cycle/phase calcs
*Setup props:* | | 
phaseCount | Int | Phases per cycle; 0 = not phased
phaseRate | Int | ticks per phase; 0 = not phased
phaseOffset | Int | ticks before phase 0 starts; usually 0
cyclesToRun | Int | 0 = continuous; else N cycles then done
*Dynamic r/o:* | |
currentPhase | Int | 0..phaseCount-1
currentCycle | Int | 0..cyclesToRun-1 when finite
cyclesDone | Bool | true after finite run completes; poll in update

### Drawer

You can choose to do all your rendering manually. Life and Zombies use this approach. You can also use SimGame's Drawing Engine, by attaching a DrawerSet to your entities. You can also use a hybrid approach, by letting the DrawingEngine first draw entities with DrawerSets, then your `handlerRender` event handler on either an object method you provide, or on the SimGame delegate. 

These are the types of Drawers SimGame offers:
* BlitDrawer -- draws from a BlitSrc object (Mat2Xi/PaletteX or Mat2c)
* AnimBlitDrawer -- uses a set of source mats; the animation cycle is controlled by the associated Cycler. A 2-blit animation would set phaseCount to 2; a 4-blit set to 4; etc.
* RadialAnimBlitDrawer -- used a Mat of small blits that radiate outward according to the cycler; super useful for explosion-type effects. Invaders uses this for its exploding alien effect. The blits were just a 4x4 chop of a base alien blit rect. 
* "Roll Your Own" -- this is not a specific class, but is a Drawer created by using a TypedObject with a `handleRender` method, or else a method on your SimGame delegate object. Invaders uses this approach to handle its StarField entity's drawing: it is just a `renderStars` method on the Invaders TypedObject delegate.

### DrawerSet

A Drawer Set is a compound of a Drawer, a Cycler, and a MethodRef that can be used for manual drawing. Animated Drawers have their own drawing built in; the MethodRef is for "hand-rolled" Drawers, usually either a method on the SimGame delegate, or one on a TypedObject, if the Drawer needs addiitonal state.

## Putting it all together

Start small, and iterate your sim or game in steps. Don't try to include every feature from the start. Begin with a subset of your game objects, and implement the code to render them. Then you can add some simple motion. This will motivate the need for boundaries, often using `stop`. It is quite rewarding to get to a point where things are moving. Invaders started with the cannon, and the keyboard control. It was quite satisfying to see it rendered, and moving according to the key presses. The phase shot came next, and the obvious need to have it stop at the top boundary, otherwise it would literally go to "infinity and beyond!".

Above all, have fun!
