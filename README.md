# The Final Programming Language

Welcome to Final!

Final is a new, modern, stack-based programming language designed by [Brad Aisa](https://github.com/androticus) in the tradition of Forth. “But this ain’t yer granddaddy’s Forth!” Final has all the features a modern programming language needs, such as:

- rich basic primitive types, with polymorphic operators
- objects and types, including user-defined TypedObject, and DynamicObject
- collection classes such as List, Map, Set, Mat, and Vec;
  - includes a full set of strongly-typed numeric List and Mat types
- iterators, pattern matching, streams, and Regex;
- threading, JSON support, web and crypto support
  - web fetching and serving both supported;
- multiple stacks, and local objects, for ease of programming;
- nested subwords in colon words;
- an Event-driven runtime architecture;
- both a TUI and GUI interface
- Fyne library of GUI widgets, to compose GUI applications;
- rich set of drawing/blitting/image operations
  - lines, rects, circles, text, etc.
  - 1, 4, 8 bit paletted blitting
  - 32 bit rgba full-color blitting/drawing;
- configuration and logging support;
- command line execution of Final scripts, including dynamic compilation;
- support for hardware control of Raspberry Pi 5 and Pi Zero 2W (32 or 64 bit Pi OS);
  - "legacy" poll-loop or event-driven approach both supported
  - support for: all GPIO pins, and I2C, SPI, PWM, UART peripherals
  - optional pin debouncing, and level-change events;
- Final can be embedded as a scripting language in Go programs
  - via a simple Send/Reply type of mechanism, including Tool turns for AI use
  - via an approach letting you expose your Go types in Final.

In addition, Final has a framework for making Simulations and throwback Games called SimGame, that features:

- an event-driven simulation/game controller;
- a MotionHit engine, that animates your entities, and detects collisions;
- several animated blitting effects, including RadialAnimation (for explosions etc.)

## Install It

Go to the [Final Website](https://final.androticus.ai/downloads/) to get the binary for your machine.

## Run It

After installing, cd into the Final installation folder, and launch `final` or `final-gui` (`final.exe`/`final-gui.exe` on Windows).

```bash
cd ~/Final
./final
# or
./final-gui
```

## Run one of the Sims/Games!

Launch `final-gui`, then File/Load any of:
`examples/gui/simgame/life.fnl` `shape-zombies.fnl` or `invaders.fnl`

## Explore

From the TUI or GUI, try some interactive stack-based operations, eg:  

(`$` prints the value on top of the stack)

```final
> "Hello World from Final!\n" $

Hello World from Final!

> 2 3 * 10 2 - * $

48
```

## Create New Words

Creating your own words in Final is simple:

```final
> : hello "Hello World from Final!\n" $ ;
> hello

Hello World from Final!
```

Start a new definition with `:` ("colon"), then the name of the word, the operations, then a `;` ("semi-colon").

## Stack Effects

We express the values on the datastack, from deepest to topmost, using this notation; the part before the -- is the values on the stack on entry to the word; the part after is the result values (if any).

```final
// ( nice:b -- )
: nice
  if "Nice!\n" else "Naughty!\n" then 
  $
;

> true nice
Nice!
> false nice
Naughty!
```

## Expressive Dynamic Object/Collection Creation

```final
{
  1 "one"
  2 "two"
  ( 3 4 ) "list"
}
>jsonp $

{
  "one": 1,
  "two": 2,
  "list": [
    3,
    4
  ]
}
```

## You Want Stacks???

Final has the main DataStack, and defines 4 auxiliary stacks with equal capability, for ease of parameter/result parking without needing to enage in brain-melting stack gyrations.

The following example distributes some values to the auxiliary stacks, then prints their values using Final's concise parameter-gathering/formatting construct:

```final
1 >ts 2 >xs "three" >ys "four" >zs
( ts> xs> ys> zs> )
"%d %d buckle my shoe %s %s there's the door\n" fmt $

1 2 buckle my shoe three four there's the door
```

## Context Stack Notation

All the common stack operators `dup` `swap` `drop` etc. are available on all stacks, by using the stack prefix, eg `t:dup` `x:swap` `y:drop`, etc.

## DynamicObject and Local Variables

A DynamicObject is an object you create, and then assign/retrieve named field values:

```final
DynamicObject.new^ >ts
0 t:.from! 9 t:.to!
t:.from@ t:.to@ range
foreach
  i $ sp $
/foreach

0 1 2 3 4 5 6 7 8 9
```

Local variables in a word are a DynamicObject accessed using the `l` accessor. You can assign as many fields as you want; locals can help prevent stack juggling and confusion.

```
: sum-me
  0 l.sum!
  0 9 range
  foreach
    i l.sum@ + l.sum!
  /foreach
  l.sum@ $
;
> sum-me

45
```

## Help Is Always at Hand

Help is always at hand. Start with:

```final
list
```

This prints out a list of lists, categorized by topic. Then print a list for the words relevant to the topic:

```final
list-ops // prints all the basic operators, `+` `-` `&` etc.
```

Then, `help <word>` for details on the word or Type. Type help includes all the members (properties and methods) with an explanation, and for methods, stack effects.

```final
help DynamicObject
```

## Dive into the User Guide

The Final User Guide has a introductory section, to get you oriented, a more detailed section covering all the basic language constructs, then a section on every feature/subsystem that the language offers. It is available online, and is also in the installed image as a pdf with table of contents/index.

[Final User Guide](https://final.androticus.ai/docs/intro/)

## Build from Source

See [How To Build](/how-to-build.md) for more details on building `final`, `final-gui`, and the embedding examples. Point your AI at it, and ask it to help install the prerquisites needed to build. Note that `final` (TUI version) is a straight go build, and is easiest to build; `final-gui` requires `cgo` and additional prerequisites, which vary by platform. (Final is easiest to build on Mac and Linux.)
