// Threading

// Pips

/*
Pip is an object that provides support for Pips.
It only provides class members, and is never instantiated.
Use spawn: to create a new Pip.
Note: Pip panics are viral, and bring down the entire Pip tree from 
the root Task. **Test your spawn words well, including error paths.**
*/
:doc Pip

/*
Build a default PipParams AtomicDynamicObject for spawn:
._stdout / ._stderr = the parent writers (TextWrite), typically the root
  you usually don't need to change these;
._name = optional name, otherwise spawn word name used.
  in both cases, affixed with "#pipId" (for fault reports);
then, add your own payload fields as needed; 
you will typically pass parameters and Pipes in.
( -- pp:AtomicDynamicObject )
*/
:doc Pip#newPipParams

/*
Launch a child pip (lower-level axiom).
VirtXT is TOS so the caller can leave PipParams under it.
Fire-and-forget like Go `go` — returns nothing. No join; child exit only
unregisters from the parent map. ITC via pipes if needed.
( pp:AtomicDynamicObject xt:XT -- )
*/
:doc (spawnPip)

/*
Create and launch a new Pip, using the next word as the worker.
The worker Word must be a plain colon definition, not
immediate/hidden/compile-only/defined/or an axiom.
eg: `Pip.newPipParams^ spawn: doWork`
parent: ( pp:AtomicDynamicObject -- )
child:  ( pp:AtomicDynamicObject )
*/
:doc spawn:

// Pipes

/*
Pipe is an object that provides support for Pipes.
It only provides class members, and is never instantiated.
Use `pipe::new` to create a new Pipe.
*/
:doc Pipe

/*
Synonym for `pipe::new`, which is preferred.
*/
:doc Pipe#newWithCapacity

/*
Create an unarmed ticker-mode Pipe.
Arm with `pipe::tickerStart`; stop with `pipe::tickerStop`.
Do not `pipe::close` a ticker pipe (fault).
eg:
`Pipe.newTicker^ l.t!
 Interval.second l.t@ pipe::tickerStart`
create: ( -- p:Pipe )
pushes after start: ( dt:DateTime ) periodically
*/
:doc Pipe#newTicker

/*
Create an unarmed timer-mode Pipe.
Arm with `pipe::timerAfter` or `pipe::timerWhen`.
Does not self-close after fire; re-arm to fire again.
Do not `pipe::close` a timer pipe (fault).
eg:
`Pipe.newTimer^ l.t!
 Interval.second l.t@ pipe::timerAfter`
create: ( -- p:Pipe )
pushes after arm: ( dt:DateTime ) once per arm
*/
:doc Pipe#newTimer

/*
Create a new pipe with the given capacity.
Cap=0 synchronizes reader and writer.
Cap=1 useful for "done" and other such signalling.
Cap>1 useful for buffering.
Parent should pass required Pipes to Child in the 
DynamicObject passed to the child after spawn:.
( cap:i -- p:Pipe )
*/
:doc pipe::new

/*
Returns the capacity of the pipe.
Callable by any Pip.
( p:Pipe -- cap:i )
*/
:doc pipe::cap

/*
Returns the number of items currently in the pipe.
Callable by any Pip.
( p:pipe -- len:i )
*/
:doc pipe::len

/*
Push a value into the pipe, blocking if it is full.
( item:a p:Pipe -- )
*/
:doc >pipe

/*
Pull a value from the pipe, blocking if it is empty.
( p:Pipe -- item:a )
*/
:doc pipe>

/*
Pull a value from the pipe, blocking if it is empty.
Returns a value and true, or nil and false if closed.
Note: prefer iterating with foreach when using 
Closed semantics.
( p:Pipe -- item:a true | nil false )
*/
:doc pipe>?

/*
Close the pipe.
Optional: it is typically for the iterated receiver scenario.
Only the sender should close a pipe.
Closing may be called multiple times; only the first closes.
Sending on a closed channel panics.
Faults on ticker- or timer-mode pipes — use
`pipe::tickerStop` / re-arm instead.
( p:Pipe -- )
*/
:doc pipe::close

/*
Check if the pipe is closed.
Note: closing is for iterated receiver scenarios,
or similar, where a finite, possibly unknown number
of values will be pushed into the pipe by one worker.
Only the sender should close a pipe.
( p:Pipe -- b )
*/
:doc pipe::closed?

/*
Start or restart a ticker-mode Pipe at the given Interval.
Pushes DateTime values periodically until `pipe::tickerStop`.
Faults if the Pipe is not ticker-mode.
( iv:Interval p:Pipe -- )
*/
:doc pipe::tickerStart

/*
Stop a ticker-mode Pipe. The pipe stays open.
No-op if not yet started. Faults if not ticker-mode.
( p:Pipe -- )
*/
:doc pipe::tickerStop

/*
Arm or re-arm a timer-mode Pipe to push DateTime once
after Interval. Does not close the pipe.
Faults if the Pipe is not timer-mode.
( iv:Interval p:Pipe -- )
*/
:doc pipe::timerAfter

/*
Arm or re-arm a timer-mode Pipe to push DateTime once
at the given DateTime. Past times fire immediately.
Does not close the pipe. Faults if not timer-mode.
( when:DateTime p:Pipe -- )
*/
:doc pipe::timerWhen

/*
Iteration
`foreach` continues until the pipe is closed.
The iterator blocks on the GetNext operation when
there is no value available.
Once a value is read, it is available as i:Any.
`foreach` will hang if the Pipe is never closed.
*/

// Atomic* Helpers

/*
A threadsafe list that can be shared across multiple Pips.
Iteration is also safe, by way of a snapshot of items--just
a snapshot of the list of items at the beginning
of `foreach`, not a copy of the items themselves.
Iteration does NOT lock the List.
Note: Range iteration is NOT threadsafe; only use foreach to iterate.
Note: The thread safety of items themselves is based on the item type.
Primitives are always threadsafe, because they are immutable. 
*/
:doc AtomicList

/*
Create a new empty AtomicList.
( -- list:AtomicList )
*/
:doc AtomicList#new

/*
Shuffle is threadsafe, but locks the list for the entire shuffle.
*/
:doc AtomicList#shuffle

/*
A threadsafe StringList that can be shared across multiple Pips.
see AtomicList for thread details.
*/
:doc AtomicStringList

/*
Create a new empty AtomicStringList.
( -- list:AtomicStringList )
*/
:doc AtomicStringList#new

/*
Shuffle is threadsafe, but locks the list for the entire shuffle.
*/
:doc AtomicStringList#shuffle

/*
Sorting is threadsafe, but locks the list for the entire sort.
*/
:doc AtomicStringList#sort

/*
A DynamicObject that is threadsafe for reading and writing across
multiple Pips.
Iteration is also safe, by way of a snapshot of items--just
a snapshot of the list of items at the beginning
of `foreach`, not the items themselves.
Iteration does not lock the List.
Note: Thread safety of the items themselves is based on the item type.
Primitives are always threadsafe, because they are immutable. 
*/
:doc AtomicDynamicObject

/*
Create a new empty AtomicDynamicObject.
( -- o:AtomicDynamicObject )
*/
:doc AtomicDynamicObject#new

/*
A threadsafe Int with count/add convenience members.
*/
:doc AtomicInt 

/*
Create a new AtomicInt with 0 value.
( -- ai:AtomicInt )
*/
:doc AtomicInt#new

/*
Get the value.
( ai:AtomicInt -- val:i )
*/
:doc AtomicInt#value

/*
Atomically increment the value, returning the new value.
( ai:AtomicInt -- val':i )
*/
:doc AtomicInt#inc

/*
Atomically decrement the value, returning the new value.
( ai:AtomicInt -- val':i )
*/
:doc AtomicInt#dec

/*
Atomically add to the value, returning the new value.
( a:i ai:AtomicInt -- val':i )
*/
:doc AtomicInt#add

/*
Atomically subtract from the value, returning the new value.
( a:i ai:AtomicInt -- val':i )
*/
:doc AtomicInt#sub

/*
Atomically set the value.
( a:i ai:AtomicInt -- )
*/
:doc AtomicInt#set

/*
Atomically zero the value.
( a:i ai:AtomicInt -- )
*/
:doc AtomicInt#zero

/*
Return the Pip-local Storage DynamicObject instance.
( -- pls:DynamicObject )
*/
:doc pls
