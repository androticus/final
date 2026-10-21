// EventLoop

/*
An occurance as raised by a source.
*/
:doc Event

/*
An optional source object that raised the event.
This is event-specific, and optional for events.
nil if no source is supplied by the event.
Typically not needed; use index to demultiplex.
*/
:doc Event#source

/*
The MethodRef passed during registration for the Event.
*/
:doc Event#handler

/*
The index value supplied during registration.
Use this to demultiplex multiple sources or 
multiple different registrations.
*/
:doc Event#index

/*
The event-specific parameters, if any.
See the definition of the source for documentation.
Events are often pre-processed, and the parameters
pulled out into strongly typed handler methods.
See SimGame which does this.
*/
:doc Event#params 

/*
The singleton receiver and dispatcher of Events.
There is only one instance: EventLoop.inst
User code typically doesn't deal with EventLoop.
*/
:doc EventLoop

/*
The global instance of EventLoop.
*/
:doc EventLoop#inst

/*
The source of a single Event that occurs after a
delay, or at a given wall time.
EventLoop must be running.
*/
:doc EventTimer

/*
Create an EventTimer for the given handler.
Does not schedule an Event until eventAfter is called.
Handler stack: ( -- ).
( handler:MethodRef -- timer:EventTimer )
*/
:doc EventTimer#new

/*
Schedule a single Event after the given Interval.
Calling again reschedules from now.
Does nothing if the timer is closed.
( interval:Interval -- )
*/
:doc EventTimer#eventAfter

/*
Schedule a single Event at the given Time.
If `when` is in the past, the event is dispatched
immediately.
Does nothing if the timer is closed.
( when:DateTime -- )
*/
:doc EventTimer#eventAt

/*
Cancel a pending Event.
The timer can be scheduled again with eventAfter.
( -- )
*/
:doc EventTimer#stop

/*
Permanently stop the timer and release it.
Further eventAfter calls do nothing.
Call when finished; stop only pauses.
( -- )
*/
:doc EventTimer#close

/*
The source of periodic Events that occur at a
specified interval.
EventLoop must be running.
*/
:doc EventTicker

/*
Create an EventTicker for the given handler.
Does not start firing until eventEvery is called.
Handler stack: ( -- ).
( handler:MethodRef -- ticker:EventTicker )
*/
:doc EventTicker#new

/*
Start or restart periodic Events at the given Interval.
Calling again changes the interval from now.
Does nothing if the ticker is closed.
( interval:Interval -- )
*/
:doc EventTicker#eventEvery

/*
Pause periodic Events.
The ticker can be started again with eventEvery.
( -- )
*/
:doc EventTicker#stop

/*
Permanently stop the ticker and release it.
Further eventEvery calls do nothing.
Call when finished; stop only pauses.
( -- )
*/
:doc EventTicker#close

/*
Re-armable one-shot EventLoop source with no timer delay.
Mint one Event per again; from the handler call again to
continue a CPU-bound loop while still yielding to other events.
EventLoop must be running.
*/
:doc EventAgain

/*
Async shell-out: runs the same capture path as sys.exec? off the EventLoop.
Completion is demuxed to the user MethodRef with the same modal stack
( out:s code:i true | out:s err:s false -- ); authors do not unpack Event.params.
Requires SR_System. Close cancels and suppresses the handler.
*/
:doc EventSysExec

/*
Start program with args under timeoutS (whole seconds; 0 = no deadline).
Starts immediately. Handler is invoked when the process finishes or times out.
( program:s args:StringList timeoutS:i handler:MethodRef -- e:EventSysExec )
*/
:doc EventSysExec#new

/*
Cancel the wait and do not invoke the user handler. Idempotent.
( -- )
*/
:doc EventSysExec#close

/*
*System Use Only* (leading _; hidden from help).
EventLoop entry: demuxes wire Event.params to the user MethodRef.
( e:Event -- ) o( inst:EventSysExec )
*/
:doc EventSysExec#_onComplete

/*
Create an EventAgain for the given handler.
Does not deliver until again is called.
Handler stack: ( -- ).
( handler:MethodRef -- ea:EventAgain )
*/
:doc EventAgain#new

/*
Enqueue one Event onto the EventLoop pipe.
Use once after new to kick, and again at the end of the
handler to re-arm. No-op if closed.
( -- )
*/
:doc EventAgain#again

/*
Permanently disable again.
Further again calls do nothing. Idempotent.
( -- )
*/
:doc EventAgain#close

/*
EventLoop Events for Escape / F1–F12 presses latched by headed
chrome (TUI or GUI). Not a general keyboard API — use Fyne canvas
keys in GUI for app typing; Esc/Fn is the controller cancel lane.

After forget:/type rollback, Final method slots are poisoned. deliver drops and
unlinks subscribers whose MethodRef targets a forgotten type so EventLoop does
not panic on orphaned controllers left open across reload.

Hosts only call LatchEscFn; this type + clearEscFn / escFn? live in core.
EventLoop must be running to dispatch.
*/
:doc EventEscFn

/*
Create an EventEscFn for the given handler.
Linked but not delivering until start (same idea as timer before
eventAfter). Handler stack: ( code:i -- ) with 0=Esc, 1..12=F1..F12.
( handler:MethodRef -- esc:EventEscFn )
*/
:doc EventEscFn#new

/*
Begin delivering latched Esc/Fn presses as Events.
Idempotent. No-op if closed.
( -- )
*/
:doc EventEscFn#start

/*
Ignore further presses (same as post-new).
May start again. No-op if closed.
( -- )
*/
:doc EventEscFn#stop

/*
Permanently unlink this subscriber.
Further start/stop do nothing.
( -- )
*/
:doc EventEscFn#close

/*
Discard any latched Escape or F1–F12 press.
Does not mint Events.
( -- )
*/
:doc clearEscFn
: clearEscFn
  EventEscFn.clearEscFn^
;

/*
Take the latched Escape/Fn press if any.
Returns n+true (n: 0=Esc, 1..12=F1..F12) or false. Clears the latch.
( -- n:i true | false )
*/
:doc escFn?
: escFn?
  EventEscFn.escFn?^
;

: EventLoop_run
  true EventLoop._running!
  EventLoop.inst.pipe
  // fetch and dispatch events, until/unless closed
  foreach
    o:.closed if
      false EventLoop._running! 
      break
    then 
    i .handler ^ // ( e:Event -- ) handler must drop it
  /foreach
;

/*
Start the EventLoop.
Fetches and dispatches Events from the event Pipe.
Blocks when there is no Event.
( -- ) o( inst )
*/
: EventLoop#run
  EventLoop_run
;

/*
Mint an Event onto the EventLoop pipe.
source is nil; params is empty. Safe before #run (queued until drained).
( handler:MethodRef index:i -- ) o( inst )
*/
:doc EventLoop#mintEvent

/*
Start the EventLoop at the end of a command-line program.
Must not already be running (headed TUI/GUI already own #run).
Mints start as the first Event, then runs until close/bye.
Start handler stack effect: ( e:Event -- ) — drop e like other handlers.
Form the MethodRef with ~, same as EventTicker/EventTimer:
  demo @ .start~ 0 EventLoop.inst.runWithStart^
( mr:MethodRef index:i -- ) o( inst )
*/
: EventLoop#runWithStart
  EventLoop.running if
    "EventLoop.runWithStart: the EventLoop is already running." error
  then
  o:.mintEvent^
  EventLoop_run
;
