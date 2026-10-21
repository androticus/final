# Threading

Final provides first-class threading support using a highly memory efficient model. Many languages map a language thread directly to an Operating System thread. But OS threads are heavy beasts that require a lot of setup and teardown time, and consume scary amounts of memory per thread.

Final uses an M:N model, in which it maps lightweight language threads (called `Pips`) to a fixed pool of expensive operating system threads. It has a scheduler which allocates ready-to-run Pips to an available OS thread. This enables you to create threads in the thousands or even hundred of thousands if it makes sense. N is chosen by the runtime to map effectively to the CPU, and is typically equal to the number of cores of the host.

The prefered method of interthread communication (ITC) in Final is via Pipes, which are conduits that can be shared between threads, and the root, to send or retrieve values.

Although Final encourages a "communicate to share data" not "share data to communicate" model, it does offer some threadsafe Atomic objects that can be safely shared between threads. The iterable Lists are particularly useful, since they use an efficient form of snapshotting of items at iteration time, which enables such lists to iterate without taking a lock.

Read on to learn about `Pips` `Pipes` and `Atomics`!