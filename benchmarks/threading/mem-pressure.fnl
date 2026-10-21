/*
This benchmark tests the memory pressure/usage caused by
creating N Pips.
Each spawnee puts a small amount of pressure on each aux stack
to mimic a typical highwater mark; but it doesn't create objects,
and doesn't use much CPU or additional memory.
So we are measuring a baseline of memory, CPU, for a single Pipe.
`spawnee` takes the following:
Pipe(?) .childDone -- children signal up, per below
A child Pip signals .childDone to parent Pip when
it has done it's work; root does a foreach on the .childDone
Pipe, and breaks when it gets all N reports 
Every Pip waits for endDelayMs after signaling before exiting,
to give the top level Task time to synchronize all Pips, and
make its runtime snapshot.
*/

5 1000 * const: endDelayMs
15 const: memspawnWorkN // const is inherently atomic
1024 const: childDonePipeCapMax // cap is min(N, Max)

/*
Do some work that expands stacks, to give a reasonable 
benchmark of stack high-waters of a Pip.
( -- )
*/
: memspawn_work
  memspawnWorkN 1 - >ts
  0 t:peek range
  foreach
    i dup >ts dup >xs dup >ys dup >zs // ( -- i )
  /foreach 
  0 ts> range
  foreach
    z:drop y:drop x:drop t:drop drop
  /foreach
;

// ( N:i -- params:DO ) convenient place to put N
: spawnPrep
  Pip.newPipParams^
  >ts
    dup t:.N!
    childDonePipeCap min Pipe.newWithCapacity^
    t:.childDone!
  ts>
;

// ( childParams:DO -- )
: waitForChildren
  >ts
    t:.childDone@ // cache on ds
     // sync all children
    0 t:.N@ 1 - range
    foreach
      dup pipe> drop // dummy value
    /foreach
    drop
  t:drop
;

// ( myParams:DO -- )
: signalDoneToParent
  >ts
    true t:.childDone@ >pipe // dummy value
  t:drop
;

// ( params:DynamicObject -- )
: spawnee
  memspawn_work
  signalDoneToParent
  endDelayMs delayMs
;

// ( N:i -- childParams:DO )
: spawnChildren
  dup spawnPrep
  swap 0 swap 1 - range
  foreach
    dup spawn: spawnee
  /foreach
;

// ( N:i -- )
: memPressure
  Benchmark.new^
  >zs // z( bm:Benchmark )
    dup "N=" swap >>String str::+ z:.name!
    z:.start^
    spawnChildren >ys // y( childParams:DO )
    y:peek waitForChildren
    z:.end^ >xs // x( res:BenchResults )
    x:.fmt^ $ nl
  x:drop y:drop z:drop
  // we should delay like the pips are, to more likely return
  // to an unlocked DC, hopefully the Pips clean up lively
  "Waiting for Pips to finish...\n" $
  endDelayMs delayMs
;
