# Controllers

"Controller" is a conceptual term that refers to the totality of code that makes up your control program. It includes the pin and peripheral setup, and then your code that accesses the hardware to achieve your end system goal.

There are basically two approaches to writing a controller. We discuss each in turn.

## "Legacy" ("Poll Loop") Controllers

The simplest approach, which we call "legacy" or "poll loop" control, involves creating a single structured routine that executes in a loop. In the loop it checks the states of the hardware, and according to the goals of your system then responds appropriately. For example, if you have a small display screen and two scroll buttons, you frequently poll for the state of the buttons, pressed/not, track the transition from one to another, and scroll your display content accordingly on the down-press detection.

Poll loop controllers can be easy to write... at first... when there are only a couple of devices and the actions are simple. But they soon become complex when there are many devices, and time also factors in, such as checking things on intervals, and so on. And if you have to add manual switch debouncing in.. it gets very hard to reason about and maintain.

Poll loops are fine for simple smoke tests of a device or two, but Final offers a more structured approach, with many benefits...

## Event-driven Controller

An event-driven controller in Final takes advantage of the EventLoop built in to Final, and structures a controller as a TypedObject with methods for handling specific kinds of events, such as level changes, and time/interval triggers.

A controller like this has a bit more "ceremony" to set up, but it pays off by greatly simplifying some of the most common situations, which are level-change response (eg. buttons, sensors, etc.), and incorporating time elements: regular "heart beats", which Final calls `tickers`, or interval-based events, which final calls `timers`.

We will revisit this type of controller in more detail in the Writing Controllers section. First let's dive in to learn all about Pins and Periphs (Peripherals)...

