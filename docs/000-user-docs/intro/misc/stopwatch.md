# Stopwatch

Stopwatch enables you to measure the time between a start time and a later time. You can use it in two modes:
* Stop mode -- call the `stop` method when you want accumulation to stop
* Interval mode -- call the `elapsed` method at a point you want the interval since `start`

Here is a typical use pattern:

```forth
: myword
  Stopwatch.new^ l.sw!
  //...
  l.sw.start^
  // intervening code
  l.sw.elapsed // ( int:Interval ) since .start
  // more code
  l.sw.stop^
  l.sw.elapsed // ( int:Interval ) since .start
  // use elapsed
;
```

There is no reset method; call `start` to begin a new timing session.
