# Command Line Invocation

Interactive work uses the TUI/GUI. For systemd, cron, or a headless box, launch with the `final` binary (not `final-gui`).

General scripting forms are in [Command line scripts](../intro/app-svcs/cmd-piping.md):

```bash
final my-controller.fnl
```

When no TUI owns the EventLoop, start it yourself with `EventLoop.inst.runWithStart^` and a start handler. Pattern from `examples/events/ticker-timer-demo.fnl`:

```final
: setup
  EventLoop.running if
    // already in TUI/GUI — just run your start word
    my-demo
  else
    myCtrl >ts
    true t:.headless!
    t:.start~ 0 EventLoop.inst.runWithStart^
  then
;
```

On clean exit from headless mode, close the EventLoop (`EventLoop.inst.close^`) after you tear down hardware, or the process will keep waiting for events.
