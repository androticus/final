# Operating System Command Execution (exec)

Final offers two ways to run external programs:
* the `sys.exec` method, that runs the command synchronously;
* the EventSysExec type, that runs the command in the background, and raises an event on completion or timeout.

The apis and result sets of both are the same. 

Each form can deliver 3 types of result shapes:
* an error occurred trying to launch the command itself;
* a timeout occurred;
* the command ran to completion, possibly returning an program-specific error code, and stdout/err output

Launching an external program synchronously, with optional timeout:
```final
// ( cmd:s args:StringList timeout:i -- out:s code:i true | out:s err:s false )
  "echo" ( "sync-hello" s) 5 sys.exec?^
  if
    drop  // code
    "sys.exec? out: " $ $ nl
  else
    nip // out
    "sys.exec? failed: " $ $ nl
  then
```

See `examples/events/sys-exec-demo.fnl` for the EventSysExec code.
