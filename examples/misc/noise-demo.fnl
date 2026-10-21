forget: reset-noise-demo

`
Short Noise demo: seeded Perlin samples (at2).

Nearby points are related; distant points diverge.
Try: noise-demo
` const: noise-demo-blurb

/*
Print a small grid of at2 samples from a fixed seed.
( )
*/
: noise-demo
  42 Noise.newWithSeed^ >ts
  0 5 range
  foreach
    i >Float 10.0 / >xs
    0 5 range
    foreach
      ( x:peek i >Float 10.0 / t:.at2^ )
      "%7.3f " fmt $
    /foreach
    nl
    x:drop
  /foreach
  t:drop
;

noise-demo-blurb $ nl
