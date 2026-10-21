# Threading Patterns

This section outlines some common threading patterns, and how to express them in Final using Pips and Pipes.

## Simple Delegation

This pattern uses a Pip in one of the simplest scenarios: delegating a computation to a single Pip, which will end when it finishes its task.

You will typically create a Pipe for the result, and set that, along with the parameters for the task, in the PipParams before spawning the Pip. Then you enter a `foreach` loop on the Pipe, in expectation of the result or results. The Pip closes the pipe when it has finished, which drops you out of the `foreach` loop. The Fibonnaci example uses this pattern; its code repeated below.

`examples/threading/fibonacci.fnl`

```
// ( pp:AtomicDynamicObject -- ) // .n and .pipe
: fibonacci_
  >ts t:.pipe@
  >zs // pipe
    0 >xs 1 >ys // the current two values 
    0 t:.n@ 1 - range
    foreach
      x:peek z:peek >pipe
      ys> dup xs> + >ys >xs // x, y <- y, x+y
    /foreach 
    x:drop y:drop 
    z:peek pipe::close
  z:drop
  t:drop
;

/*
Print out the first N fibonnaci numbers.
Delegates to a Pip
( n:i -- )
*/
: fibonacci
  nl
  dup pipe::new dup
  >ts
    Pip.newPipParams^
    >ts
      t:.pipe! t:.n!
    ts>
    spawn: fibonacci_
  ts> // pipe
  foreach // until closed, value is nth fib
    i $ nl
  /foreach
;
```

## Transient Workers

This pattern involves dispatching a computationally expensive, but "embarrassingly parallel", task onto a set of threads. You divide the task into chunks that each Pip will process, and pass the chunking parameter in the PipParams to each Pip's `spawn:`.

This pattern uses an "on-demand" approach, which adds more overhead per invocation, but can be easier to set up, and there is no "teardown" step when the program closes.


## Persistent Workers

This pattern is used when you know you will regularly invoking the background threads, and don't want to incure the setup/teardown cost each time.

This pattern is used in `examples/gui/simgame/life-sim.fnl`. It requires a bit more setup and teardown overhead that on-demand, but benefits from nearly instant task startup time.

To use this pattern, you will need a set of N "kicker" pipes, one for each Pip. These provide the command to the Pips to start a run of their task. They process in a `forever` loop, and you close the pipe when you are tearing down your program; for a gui program, typically in the Window.onClose handler.

The Pips can return a result or sentinel to a `done` or `results` Pipe you supply. It simply needs to count N replies in a N-counted `foreach` ranged loop.

Look at the following routines mentioned below, in life-sim.fnl:

```
// this kicks off the workers for a dispatch turn
: LifeSim_nextGen_threaded
  // dispatch to the n workers
  o:.threads@ l.n!
  
  // kick off the workers
  o:.kickPipes@
  foreach
    true i >pipe
  /foreach
  
  // get a done sentinel from each worker
  0 l.n@ 1 - range
  foreach
    o:.donePipe@ pipe> drop
  /foreach
;

// this is the setup code for the workers
// most of the parameters are elided,
// to focus on the pipe usage
/*
PipParam fields:
...more fields... kickPipe donePipe
( pp:AtomicDynamicObject -- )
*/
: LifeSim_nextGen_worker
  >ts
    // wait for start signal
    // loop while our kick pipe is open
    // it gets closed in the window close event
    t:.kickPipe@
    foreach
      // pull our parameters from pp, pass to worker
      ... 
      LifeSim_nextGen_inner

      // signal done to main Pip
      true t:.donePipe@ >pipe
    /foreach
  tdrop
;

/*
Initialize one worker Pip
( ...task params... kickPipe:pipe donePipe:pipe -- ) o( inst )
*/
: LifeSim_nextGen_initWorker
  Pip.newPipParams^ l.pp!
  l.pp.donePipe!
  l.pp.kickPipe!
  ... task params

  // spawn the Pip, it will wait for work
  l.pp@ spawn: LifeSim_nextGen_worker
;

/*
This initializes our worker N Pips
Not called if not threaded.
( n:i -- ) o( inst )
*/
: LifeSim_initWorkers
  // dispatch to n workers
  l.n!
  
  // create our done pipe and kick pipe list
  l.n@ pipe::new o:.donePipe!
  List.new^ o:.kickPipes!
  
  // task params to be chunked here
   
  0 l.n@ 1 - range
  foreach
    ... put the chunk task params on DS
    // create the Pip's kick pip, add to our list
    1 pipe::new dup o:.kickPipes.append^ o:.donePipe@
    LifeSim_nextGen_initWorker
  /foreach
;

```

## Producer/Consumer Chain

This pattern is more complex, and involves a chain of threads that are specialized for one activity, and pass their intermediate result to another Pip for further processing.

A good example would be a web-scraping task in which there are a small pool of web dialers that retrieve content, then pass the content along to a pool of workers. This enables the downloaders to have delays, while the processing Pips consume the downloads when available.

Final doesn't have an example of this pattern yet, but you build it up using Pipes as your building blocks. 

