/*
GUI EventLoop root. EventLoop#run must be invoked from Final (run^), not from Go.
WarmBoot re-enters this word. Do not call from handlers.
( -- )
*/
: guiRoot
  EventLoop.inst.run^
;
