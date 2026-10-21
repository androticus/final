forget: reset-ticker-timer-demo
`
Simple demo showing how to use the EventTimer and EventTicker 
event generating in both final-tui and final-gui.

It also includes the code needed to run from a commandline
--headless or a legacy final REPL binary.

Creates a ticker event that fires every two seconds, and a
timer event that fires on a random interval.
Prints which event fired. Runs for ten seconds.

Try:

ticker-timer-demo
` const: ticker-timer-demo-blurb

// Create a TypedObject, to control the demo, and handle the events.

{
  "TickerTimerDemo" "name"
  
  {
    EventTicker "ticker"
    EventTimer  "timer"
    Counter "tickersFired" // how many have been done so far?
    Bool "headless" // so we call close on the EventLoop, to exit
  } "fields"
  
  {
    2 "tickerS"
    1500 "timerMaxMs" // after ticker fires
    5 "tickerCount"
  } "constants"
}
TypedObject.makeType^

// The type now exists, add the handlers.

/*
You can factor Final methods with non-method plain words.
The stacks, including OS, carry over.
( -- ) o( inst:TickerTimerDemo )
*/
: TimerTickerDemo_primeTimer
  TickerTimerDemo.timerMaxMs Rand.int^
  Interval.millisecond dt::* o:.timer.eventAfter^
;

/*
Handles the events from the Ticker
( -- ) o( inst:TickerTimerDemo )
*/
: TickerTimerDemo#handleTickerEvent
  o:.tickersFired.inc^
  o:.tickersFired.value TickerTimerDemo.tickerCount >= 
  "Ticker event handled\n" $
  if
    o:.ticker.close^
    o:.timer.close^
    "\nticker-timer-demo done!\n" $
    o:.headless if
      EventLoop.inst.close^
    then
  else
    TimerTickerDemo_primeTimer
  then
;

/*
Handles the events from the Timer
( -- ) o( inst:TickerTimerDemo )
*/
: TickerTimerDemo#handleTimerEvent
  "Timer event handled\n" $
;

/*
Initializes the TypedObject instance
( -- ) o( inst:TickerTimerDemo )
*/
: TickerTimerDemo_init
  o:.handleTickerEvent~ EventTicker.new^ o:.ticker!
  o:.handleTimerEvent~ EventTimer.new^ o:.timer!
  Counter.new^ o:.tickersFired!
;

/*
Starts the demo.
Entry point of EventLoop.runWithStart for headless.
( e:Event -- ) o( inst:TickerTimerDemo )
*/
: TickerTimerDemo#start
  drop
  TickerTimerDemo_init
  TickerTimerDemo.tickerS Interval.second dt::* o:.ticker.eventEvery^
  TimerTickerDemo_primeTimer
;

// We create the instance in the interpreter, so it is available
// when loading headless.

TickerTimerDemo.new^ const: tt-demo

: ticker-timer-demo
  tt-demo >ts
  null t:.start^ // dummy null event, dropped
  tdrop
;

: setup-ticker-timer-demo
// Now, we modalize depending on whether EventLoop is running
  EventLoop.running if
    ticker-timer-demo-blurb $ nl
  else
    tt-demo >ts
    true t:.headless!
    t:.start~ 0 EventLoop.inst.runWithStart^
  then
;
setup-ticker-timer-demo
