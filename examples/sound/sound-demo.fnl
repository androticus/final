forget: reset-sound-demo
// `examples/sound/sound-demo.fnl` compileFromFile
`
Shows how to play sounds in Final.
This example plays a looped background music,
then some random laser shots over top.

Try:

sound-demo
`
const: sound-demo-blurb

`assets/sounds/music/loops/sci-fi/world-of-automatons.ogg`
const: bg-music-file
`assets/sounds/sfx/sci-fi/laser-retro.ogg`
const: laser-file


: sound-demo
  SoundPlayer.inst l.pl!
  bg-music-file Sound.new^ l.bg!
  laser-file Sound.new^ l.laser!
  
  l.bg@ l.pl.playLooped^
  
  0 19 range
  foreach
    100 Rand.int^
    15 *
    delayMs
    l.laser@ l.pl.play^
  /foreach 
  
  l.bg.stop^
  l.pl.stop^
  "sound-demo is finished\n" $ 
;

sound-demo-blurb $ nl
