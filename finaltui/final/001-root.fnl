/*
TUI EventLoop root. EventLoop#run must be invoked from Final (run^), not from Go.
WarmBoot re-enters this word. Do not call from handlers.
Headed final owns this loop (do not call EventLoop.inst.run^ from TUI-loaded source).
( -- )
*/
: tuiRoot
  EventLoop.inst.run^
;
