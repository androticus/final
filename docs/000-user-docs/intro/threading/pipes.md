# Pipes

How can a pip safely share information with other pips, or with the pip that spawned it?

The conventional approach at the hardware level, and in many languages that support multi-threading, has been to try to share oftentimes complex data structures or services, by way of locking constructs that enable multiple threads to access things in a way that won't involve the data corruption that occurs when multiple threads read and write multi-value data structures at the same time, clobbering each other's writes, or reading partially modified data that is not correct as a whole. For example, when updating the fields of an address, if a second thread were to read the address while the first thread had only updated the street but not yet the city, the second thread would read an invalid address.

The problem with locks on data structures is that oftentimes several data structures need to be updated for one coherent transaction. This can lead to locking complexity, and "deadlocks" wherein 2 threads manage to obtain one lock of two needed, but not the second. The threads stop. While there have been approaches to this issue that have some limited success, they are fragile, and require precise enforcement. They are also difficult to reason about. Worse, such issues are hard to create under test, and oftentimes only present by way of occasional lockups of processing under load in production.

## Pipe Introduction

While Final does provide a few thread-safe data constructs, the predominant way to safely and efficiently communicate between pips is via pipes. ("How many pipes could a pip piper pick if a pip piper could pick pipes?")

A pipe is just what it sounds like: it is a unidirectional conduit which can carry any type of data from a sender to a receiver. It is fully safe on both ends: a sending pip and receiving pip can read/write/check a pipe at the same time, without worry. Typically, one `Pip` has the role of writing to a `pipe`, and another `Pip` has the role of reading from it. For example, the classic case of a producer and consumer: one `Pip` produces some data items, and a second `Pip` consumes them. The act of adding an item to a pipe is called *pushing*; the act of removing an item is called *pulling*.

To work with Pipes, you use the `pipe::` API.

## Pipe Capacity and Length

Each `pipe` has a capacity that is set when it is created, and cannot be changed. The capacity determines how many items the pipe can contain before it is full. When a `pipe` is full, a sender trying to push a new item will *block* until at least one item is pulled by a receiver. When a `pipe` is empty, a receiving pip trying to pull an item will block until an item is pushed.

If a `pipe` has zero capacity, then both sides will block until the sender pushes an item and the receiver pulls it. This is typically not too useful, but can be used when you want to synchronize two pips during a multistep operation.

A common case is a `pipe` of length 1. This is often used in a situation to signal the end of an operation, or to communicate an error condition. A receiving pip blocks until the operation is complete, or an error is signalled.

Capacity greater than 1 is used for buffering between a producer and consumer. It can also be used with length of N, for N producer or consumer pips, to determine when all are finished, without needing an additional "done" channel.

To create a pipe, with capacity N:

```
// ( N:i -- p:Pipe )
N pipe::new
```

To query the number of items currently in the Pipe, use the `pipe::len` operator. NOTE: this value can change after reading in multi-threading uses due to pushes or pops made by other threads. The examples show a scenario where using `pipe::len` is safe and useful.

```
// ( p:Pipe -- len:i )
myPipe pipe::len
```

## Pushing/Popping Data into/out of a Pipe

To push an item into a Pipe, use `>pipe`. This operation will block if the Pipe is full.

To pull an item from a Pipe, use `pipe>`. This operation will block if the Pipe is empty.

To try to pull an item from a Pipe, without blocking, use `pipe>?`. This operation returns a value and true if available, or false if no value is available. Do not use this to create a tight polling loop, which doesn't release the thread for other uses; use it only when you can alternatively do or test something else being available to do work.

```
// ( item:a p1:Pipe -- )
true myPipe >pipe // eg, push a true boolean into the pipe

// ( p2:Pipe -- item:a )
myPipe pipe> // blocks until the item is available

// ( p2:Pipe -- item:a true | false )
myPipe pipe>? // don't block, if item not available
```

## Iterating a Pipe

A particularly convenient operation on a receiving pipe that will receive a possibly unknown number of requests from a sender, who will then close the pipe when finished, is to iterate the pipe using `foreach`. This operation will return a new value `i` from the pipe when available; the loop will exit when the sender closes the pipe.

To iterate a receving Pipe that the sender will close when done:
```
// ( p:Pipe -- )
myReceivingPipe
foreach // blocks until a value is available or pipe is closed
  i // handle the value popped from the pipe
/foreach
// the pipe has been closed, no more data will be sent
```

## Closing a Pipe

A sender can optionaly `close` a pipe. This lets the receiver know that no more values will be sent. When using a pipe that will be closed by the sender, the receiver should use a `foreach` on the pipe, which will return each value `i` when it is received, and stop iterating when the `pipe` is closed. This mode of using a pipe is optional; `close` does not release resources, and should only be called for use-cases involving receiver iteration. A receiver should NEVER close a sender's pipe, as this will cause a fault if the sender tries to push more data.

To close a Pipe:
```
// ( p:Pipe -- )
mySendingPipe pipe::close // nop if already closed
```

To check if a Pipe is closed. NOTE: this is a generally unsafe api, since the pipe can become closed by a sender between the time a receiver makes this call, and then attempts to read. Use `foreach` iteration on a receiving pipe that will be closed when no longer needed.
```
// ( p:Pipe -- closed:b )
mySendingPipe pipe::closed?
```
