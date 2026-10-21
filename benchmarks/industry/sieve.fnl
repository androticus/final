forget: reset-sieve
// `benchmarks/industry/sieve.fnl` compileFromFile
`
Sieve of Eratosthenes Prime Number Finder.
Official Final benchmark.
Single threaded.
Finds all prime nubers up to N.

Try:

100000 sieve
`
const: sieve-blurb

/*
Sieve, finds all prime numbers up to N.
At the end, provides a count of the primes.
( N:i -- count:i )
*/
: sieve
  // validate N
  dup 2 <
  if
    "N must be >= 2" error
  then
  
  DateTime.now^ >ts // for elapsed time
  
  dup >zs // N
  // many algos set all to true, but we use
  // 0=falsey to indicate primes so we don't have to 
  // initialize the array -- we use a 1xN Mat2i for speed
  1 swap Mat2i.new^
  >ys // List
  
  // init (0) (1) to truthy
  1 dup 0 y:!() 1 y:!()
  // 2 .. floor(sqrt(N)) inclusive (range end is inclusive — do not subtract 1)
  2 z:peek sqrt floor >Int range
  foreach
    i y:@()
    nif
      // for t := i*i .. N-1 step i  (list len N → valid indices 0..N-1)
      i dup * >ts
      begin
        // while t < N
        t:peek z:peek < while
        1 t:peek y:!() // set true
        ts> i + >ts
      repeat
      t:drop       
    then
  /foreach
  // count results only need 2 .. N but faster to iter all
  0 // accum
  ys>
  foreach
    i nif
      1 +
    then
  /foreach
  
  // calculate elapsed time
  DateTime.now^ ts> dt::- >ts
  ( d:peek zs> swap ts> ) // ( ) shenanigans...
  "N=%d count of primes: %d in %6s\n" fmt $
  drop    
;

sieve-blurb $ nl
