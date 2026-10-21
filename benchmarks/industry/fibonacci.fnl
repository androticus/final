forget: reset-fibonacci
`
Classic: fib(n) = fib(n-1) + fib(n-2); fib(0)=0, fib(1)=1

Try:

N fibonacci

Suggest N <= 40, can take awhile...
Hard capped at 50.
` const: fibonacci-blurb

// Classic: fib(n) = fib(n-1) + fib(n-2); fib(0)=0, fib(1)=1
// poly ops; ddup/dswap avoid cs selector
// ( n:i -- f:i )
: (fibonacci)
  ddup 1 > if
    ddup 1 - recurse
    dswap 2 - recurse
    +
  then
;

// ( n:i -- )
: fibonacci
  dup 50 > if
    $ " will take too much time, exiting.\n" $
  else
    dup l.n!
    Stopwatch.new^ l.sw! l.sw.start^
    (fibonacci)
    l.sw.stop^
    $ nl
    ( l.n l.sw.elapsed )
    "Calculated fib(%d) in %v\n" fmt $
  then
;

fibonacci-blurb $ nl
