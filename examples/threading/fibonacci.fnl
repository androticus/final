forget: reset-fibonacci
// `examples/threading/fibonacci.fnl` compileFromFile

`
Threaded Fibonacci, using the Simple Delegation pattern.
Based on a Go example: https://go.dev/tour/concurrency/4

Try: (keep N ~< 60)

N fibonacci
` const: fibonacci-blurb

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
    zs> pipe::close
  t:drop
;

/*
Print out the first N fibonnaci numbers.
Delegates to a Pip
( n:i -- )
*/
: fibonacci
  nl
  dup l.n!
  Stopwatch.new^ l.sw! l.sw.start^
  dup pipe::new dup
  >ts
    Pip.newPipParams^
    >ts
      t:.pipe! t:.n!
    ts>
    spawn: fibonacci_
  ts> // pipe
  foreach // until closed; recurse path: one Fib(n)
    i $ nl
  /foreach
  l.sw.stop^
  ( l.n 1 - l.sw.elapsed )
  "Calculated fib(%d) in %v\n" fmt $
;

fibonacci-blurb $ nl
