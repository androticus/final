# Events and the EventLoop

Under the hood, both `final` and `final-gui` use what is called an Event-driven paradigm. When you enter text in the prompt, or load a file, project, or module, the operation is dispatched via an Event sent to Final's core EventLoop.

An event-driven architecture provides benefits that a conventional procedure-driven architecture cannot, and can greatly simplify designs for such as simulations and games, low-level Pi 5 controller programs, and Graphical User Interfaces. All those topics are covered in later chapters, but it is a good idea to get a grounding in this foundation now.

## Events

An event is an occurence such as a key press, mouse movement, regular ticker notification, or single-shot timer firing. Event-driven programs operate quite differently from procedure-driven programs. Look at this startup code for Shape Zombies:

```final
: shape-zombies
  CircZomb.new^  >zs
  z:.base >ys // the SimGame base type
  y:.init^
  y:.makeUi^
  y:.startPaused^
  y:.doRender^
  y:.win@.show^
  ydrop zdrop
;
```
Notice that there is no call to a main processing routine--the code shows the main window of the game, then exits. The reason is because the actual program is distributed across a variety of *event handlers*, which are methods that are called by the system when their corresponding event occurs. Here is the `handleKeyDown` event; this gets called when you press a key (a separate event happens when the key is released)

```
/*
KeyDown event handler.
Attached to Window.KeyDown events
( ke:KeyEvent -- ) o( inst )
*/
: CircZomb#handleKeyDown
  CircZomb_onKeyDown
;
```

You can see that the handler routine is passed something called a KeyEvent, which has the name of the key that was pressed. The handler code in onKeyDown then looks for the arrow keys, and if so, moves the position of the user hexagon.

## EventLoop 

The Final EventLoop is written in Final; it listens on a Pipe for a new event, and when one arrives, dispatches it to the registered handler. 

Note that if you block the main thread by executing a continous loop, or call delayMs with lengthy delays, event processing will stall. For this reason, time-based occurrences are best handled with Timer Events (covered next).

The EventLoop is the most fundamental piece of code in `final` / `final-gui` -- if you do a stack dump with *$ you can see it sitting on the bottom of the O Stack, since it persists for the entire life of the program.

Events can originate from many sources, and do not have to come from the main thread. UI Events come from its UI thread, which is the application main thread. MotionHit events in SimGame originate on the MotionHitEngine background thread. Timer and Ticker events come from a dedicated thread. But they all get dispatched in the main Pip/thread of your program. (Only the main thread has an EventLoop; Pips cannot use event handlers, but they can send events.)

The Final UI itself is event-based, and uses an event to handle prompts sent from the main window.

## Event-based Programming

Note that you are not required to write your Final programs using an event-driven approach, unless you want to use features such as Ticker/Timer events, the Fyne UI, the SimGame framework, or Pi controllers.

Event-base programming is a different paradigm than a single procedural program, but it doesn't take much to get used to it. It is highly beneficial in many ways, because rather than one huge omnibus program with giant switch statements and interminable numbers of sub-words, each event handler in eg the SimGame framework is dedicated to handling just one specific kind of occurence. The code there can remain focused and simple. Look at the event. Determine what to do (if anything). Do it. End.

## Examples

### EventTicker and EventTimer Events

see: `examples/events/ticker-timer-demo.fnl`

This demo will run on both versions of Final.

It shows how to:
* create a TypedObject to handle events
* write ticker/timer handlers (empty stack — one MethodRef per source)
* create the Ticker and Timer, attaching to the handlers
* start the timers
* handle the timer events

### EventAgain (no-delay re-arm)

When a controller should run as fast as it can but still yield to the EventLoop (Esc, GPIO, UI), use `EventAgain` instead of a zero-interval ticker or a `forever` poll loop.

* `EventAgain.new^` binds a handler (empty stack `( -- )`); nothing fires until `.again^`.
* Call `.again^` once from `#init` to kick, and again at the end of the handler to re-arm.
* `.close^` stops further re-arms (typical Esc / shutdown path).

see: `examples/pi/i2c-ssd1306-life.fnl`

### EventEscFn (detect Esc or Fn key press)

Detecting when a user has pressed the Esc key is a convenient way to exit an Event-based controller. And having the user press a Fn key is a convenient way to offer a simple choice of execution path.

Use the `EventEscFn` event generator (in final TUI or GUI).

```final

// ( key:i -- ) // Esc=0, Fn=n
: MyCtrlr#handleEscFn 
  // your logic here
;

: MyCtrl#init
  o:.handleEscFn~ EventEscFn.new^ >ts
  t:start^ // your handleEscFn will now get events
;
```
## SimGame Events

This is an advanced look at handling events in the [SimGame Framework](sims-and-games.md). This program only runs in `final-gui`.

`examples/gui/simgame/invaders.fnl`

Have a look at the various events in Invaders to get a feel for the paradigm. Most event handlers are named `handle<EventName>` such as `handleRender` for drawing, or `handleUpdate` for the periodic update event you can enable, or `handleMouseMove` for mouse movements.
