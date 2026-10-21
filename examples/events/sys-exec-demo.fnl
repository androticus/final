forget: reset-sys-exec-demo
`
Demo: sys.exec? (sync) and EventSysExec (async demuxed handler).

Shows the unified modal stack:
  ( out:s code:i true | out:s err:s false -- )

Switch to View:StdOut/StdErr

Try:

// succeeds
true sys-exec-demo
// fails
false sys-exec-demo
` const: sys-exec-demo-blurb

{
  "SysExecDemo" "name"

  {
    EventSysExec "exec"
    Bool "done"
  } "fields"
}
TypedObject.makeType^

// var to keep alive from GC
null var: se-demo

// either "echo" or "echobad" to test both paths
"" var: cmd

// ( -- out:s code:i true name:s | out:s err:s false name:s )
: execOutput
  // A+ first-class stack shenanigans for `if` `name` and args
  >ts >xs >ys >ys
  ( ts> ys> ys> swap )
  xs> if
    "%s code: %d out: %s\n" fmt $
  else
    "%s err: %s out: %s\n" fmt $
  then
;

/*
Sync path: echo with a short timeout bound.
( -- ) o( inst:SysExecDemo )
*/
: SysExecDemo_runSync
  cmd @ ( "sync-hello" s) 5 sys.exec?^
  // ( -- out:s code:i true | out:s err:s false )
  "sys.exec?" execOutput
;

/*
Async completion — same modal as sys.exec?.
( out:s code:i true | out:s err:s false -- ) o( inst:SysExecDemo )
*/
: SysExecDemo#handleExec
  "EventSysExec" execOutput
  o:.exec.close^
  true o:.done!
  null se-demo ! // release instance
;

/*
( -- ) o( inst:SysExecDemo )
*/
: SysExecDemo_init
  false o:.done!
  cmd @ ( "async-hello" s) 5
  o:.handleExec~ EventSysExec.new^ o:.exec!
;

: SysExecDemo#start
  SysExecDemo_runSync
  SysExecDemo_init
;

: sys-exec-demo
  if "echo" else "echobad" then cmd !
  SysExecDemo.new^ dup >ts se-demo !
  t:.start^
  tdrop
;


