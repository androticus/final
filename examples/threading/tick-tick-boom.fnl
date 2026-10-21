forget: reset-tick-tick-boom
`
Based on a Go example: https://go.dev/tour/concurrency/6

This example shows how you can sneak around the restriction of
not peeking a Pipe. It only works if you are the only receiver,
but you can simply check the len>0 to know. The switch below
works somewhat like the "select" in Go, but it is sequential,
not random like Go, and Go combines a definite pipe operation
(push or pull) in the switch. But this example gives the identical
output to the Go example cited above.

Try:

tick-tick-BOOM
` const: tick-tick-boom-blurb

: tick-tick-BOOM
  :s init
    Stopwatch.new^ l.stopwatch!
    100 Interval.millisecond dt::* Pipe.newTicker^ l.ticker!
    500 Interval.millisecond dt::* Pipe.newTimer^  l.timer!
    false l.break! // true to break
  ;s
  
  :s start
    l.stopwatch.start^
    100 Interval.millisecond dt::* l.ticker@ pipe::tickerStart
    500 Interval.millisecond dt::*  l.timer@ pipe::timerAfter
  ;s
  
  // ( -- rounded:Interval )
  :s elapsed
    l.stopwatch.elapsed 50 Interval.millisecond dt::* dt::round
  ;s
  
  :s selectOrDefault
    null switch
      // tick pipe
      l.ticker@ pipe::len >0 case
        l.ticker@ pipe> drop // drain a value from pipe, value ignored
        ( elapsed )
        "[%6s] tick.\n" fmt $
      /case
      
      // boom pipe
      l.timer@ pipe::len >0 case
        l.timer@ pipe> drop // drain a value from pipe, value ignored
        ( elapsed )
        "[%6s] BOOM!\n" fmt $
        true l.break!
      /case
      
      default
        ( elapsed )
        "[%6s]     .\n" fmt $
        50 delayMs
      /default
    /switch
  ;s
  
  init
  start
  forever    
    selectOrDefault
    
    l.break@ if
      // practice safe piping! stop the ticker
      l.ticker@ pipe::tickerStop
      
      break
    then
  /forever
;

tick-tick-boom-blurb $ nl
