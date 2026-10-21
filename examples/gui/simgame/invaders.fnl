forget: reset-invaders

requiresFyne
// `examples/gui/simgame/invaders.fnl` compileFromFile
`
Alien Invaders (with apologies to Space Invaders).

This application uses the SimGame framework.
It highlights most of the SG features:
* organized SimGame object, with many options for update rates,
  rendering rates, use of keyboard and mouse, and lots more;
* The MotionHit engine, that support automatic boundary detection,
  Hit detection between game objects, programmable HitMatrix,
  definable loop speed, automatic animation of game entities;
* Rendering stack, from simple manual drawing, to Blits, Animated
  Blits, Radial Blits (eg explosions), z-order, an more.

Try:

invaders
`
const: invaders-blurb


// Game constants and data prototypes

/*
Small red phaser pulse. Idle cycler.
#handleRender ( entity:SimEntity cycler:SimDrawerCycler -- )
*/
{
  "InvadersShotDrawer" "name"
  { } "fields"
}
TypedObject.makeType^

// Create the SimGame delegate object: required field base:SimGame

{      
  "Invaders" "name"
   SimGame "baseType"
   
  {
 // SimGame "base" // added automatically
    Int "radiusScale"
    
    Palette4 "starPalette"
    Int32List "starAtFlat" // flat index
    ByteList "starColor" // 0=white, 1=red, 2=blue, 3=yellow
    EffectDrawer "starDrawer" // cycler recipe + MethodRef → #paintStars
    SimEntity "starfield" // z=-1
    SimEntity "cannon"
    BlitDrawer "cannonDrawer" // assets/cannon.png
    AnimBlitDrawer "alienDrawer" // shared 2-frame look
    List "aliens"
    SimEntity "shot" // at most one; null when none
    InvadersShotDrawer "shotDrawer"
    Mat "gooMats" // shared 4×4 Mat2c crops from alien-a
    Int "alienW"
    Int "alienH"

    Int "score"
    Int "status" // playing / waveClear / waveGround / gameWon
    Int "levelsCleared" // successful wave clears this session
    Int "waveN"
    Int "kills"
    Int "spawned"
    Float "roundVy"
    Int "spawnInterval" // sim ms between spawns
    Int "nextSpawnTick"
    Float "hitYSum"
    Int "hitYCount"
    Float "lastAvgHitY" // 0 = none yet

    Sound "bgMusic"
    Sound "sfxLaser"
    Sound "sfxExplosion"
    Sound "sfxSlime"
    Sound "sfxLevelUp"
    Sound "sfxGameWin"
  } "fields"
  
  {
    0 "btnRun"
    1 "btnPause"
    2 "btnNew"

    0 "statusPlaying"
    1 "statusWaveClear"
    2 "statusWaveGround"
    3 "statusGameWon"

    0 "subLive"
    1 "subGoo"

    SimGame.PixSize768x576 "defSize"
    Color.black "backgroundColor"
    10 "starDensity" // out of 10000; 10 is good; more is costly!
    4 "starPhaseCount" // EffectDrawer twinkle phases
    120 "starPhaseRate" // ticks per phase
    
    12 "cannonKeyMoveScale" // for arrow keys
    "examples/gui/simgame/assets/cannon.png" "cannonPng"
    "examples/gui/simgame/assets/alien-a.png" "alienAPng"
    "examples/gui/simgame/assets/alien-b.png" "alienBPng"
    "examples/gui/simgame/assets/world-of-automatons.ogg" "bgMusicOgg"
    "examples/gui/simgame/assets/laser-retro.ogg" "laserOgg"
    "examples/gui/simgame/assets/explosion.ogg" "explosionOgg"
    "examples/gui/simgame/assets/slime.ogg" "slimeOgg"
    "examples/gui/simgame/assets/nes-level-up.ogg" "levelUpOgg"
    "examples/gui/simgame/assets/nes-shooter-win.ogg" "gameWinOgg"

    // PhaserShot
    1 "HitKindShot"
    2 "HitKindAlien"
    4 "shotRadius" // Int for Painter#circle
    -480.0 "shotVy"

    // Goo RadialAnimBlitDrawer
    4 "gooGridRows"
    4 "gooGridCols"
    80 "gooPhaseRate"
    10 "gooPhases"
    10 "pointsPerKill"

    // Wave / drip spawn
    8 "nMin"
    14 "nMax"
    6 "maxLive"
    250 "alienPhaseRate"
    28.0 "vy0"
    18.0 "vyMin"
    90.0 "vyMax"
    1800 "spawnInterval0" // ms
    600 "spawnIntervalMin"
    0.9 "spawnTighten"
    10 "levelsToWin" // session win after this many clears

    // Alien AnimBlit (legacy pitch unused; drip spawn)
    3 "alienRows"
    8 "alienCols"
    48 "alienPitchX"
    40 "alienPitchY"
    64 "alienOriginX"
    32 "alienOriginY"
  } "constants"
}
TypedObject.makeType^

: InvadersShotDrawer#handleRender
  drop // idle cycler
  >ts
  t:.sg.painter@ >xs
    Color.red x:.fill.color!
    FillStyle.solid x:.fill.style!
    Color.white x:.stroke.color!
    t:.at
    Invaders.shotRadius
    x:.circle^
  x:drop
  t:drop
;

/*
Load loop + SFX once (survive New hard reset).
( -- ) o( inst )
*/
: Invaders_ensureSounds
  o:.bgMusic@ Null? if
    Invaders.bgMusicOgg Sound.new^ o:.bgMusic!
    Invaders.laserOgg Sound.new^ o:.sfxLaser!
    Invaders.explosionOgg Sound.new^ o:.sfxExplosion!
    Invaders.slimeOgg Sound.new^ o:.sfxSlime!
    Invaders.levelUpOgg Sound.new^ o:.sfxLevelUp!
    Invaders.gameWinOgg Sound.new^ o:.sfxGameWin!
  then
;

/*
One-shot SFX via the app SoundPlayer.
( sound:Sound -- )
*/
: Invaders_playSfx
  SoundPlayer.inst.play^
;

/*
BG: start or resume loop (call after base.start when entering Running).
( -- ) o( inst )
*/
: Invaders_bgMusicOnRunning
  o:.bgMusic.looping if
    o:.bgMusic.resume^
  else
    o:.bgMusic@ SoundPlayer.inst.playLooped^
  then
;

/*
BG: freeze loop (call after base.pause).
( -- ) o( inst )
*/
: Invaders_bgMusicOnPaused
  o:.bgMusic.pause^
;

/*
BG: end loop (window close / hard stop).
( -- ) o( inst )
*/
: Invaders_bgMusicOnStopped
  o:.bgMusic.stop^
;

/*
0-3 White, 3-7 Red, 8-B Blue, C-F Yellow
Level 0 - dim
Level 1 - normal
Level 2 - brighter
Level 3 - flash
( -- pal:Palette4
*/
: Invaders_initModel_createStarPalette
  (
    0x5C616C Color.hex^
    0x9AA3B0 Color.hex^
    0xD4DCE6 Color.hex^
    0xF7FBFF Color.hex^
    0x5C2A24 Color.hex^
    0x9A4A3C Color.hex^
    0xE07860 Color.hex^
    0xFFC4A8 Color.hex^
    0x2A3C5C Color.hex^
    0x4A70A0 Color.hex^
    0x7AA8E0 Color.hex^
    0xC8E0FF Color.hex^
    0x5C4C20 Color.hex^
    0x9A8038 Color.hex^
    0xE0C050 Color.hex^
    0xFFF0B8 Color.hex^
  ) Palette4.newFromList^
  o:.starPalette!
;
/*
Create a list of random points with a 16 value index to star Palette.
We just save the flat index for speed.
There is no correlation between ordinal index and point.
We use StaticObjects to minimize memory usage.
( -- )
*/
: Invaders_initModel_createStars
  o:.base.pixmap.rows  o:.base.pixmap.cols * dup l.pts!
  10000 / Invaders.starDensity * dup dup l.cnt! 
  Int32List.newWithCapacity^ dup o:.starAtFlat! l.sf!
  ByteList.newWithCapacity^ dup o:.starColor! l.sc!
  
  0 l.cnt 1 - range
  foreach
    // get flat index of a random point in the field
    l.pts Rand.int^ l.sf.append^
    4 Rand.int^ l.sc.append^ 
  /foreach
;

/*
Background Entity at z=-1; EffectDrawer (phaseCount=8) + MethodRef → #paintStars.
( -- )
*/
: Invaders_initModel_createStarfield
  SimEntity.new^ >ts
  -1 t:.zorder!
  Invaders.starPhaseCount
  Invaders.starPhaseRate
  0 0 // phaseOffset cyclesToRun (loop)
  o:.paintStars~
  EffectDrawer.newWithRender^
  dup o:.starDrawer!
  .newCycler^
  SimDrawerSet.new^
  t:.drawerSet!
  ts> dup o:.base.addEntity^
  o:.starfield!
;



/*
Load shared BlitDrawer for the cannon (single PNG).
( -- )
*/
: Invaders_initModel_loadCannonDrawer
  Invaders.cannonPng BlitDrawer.newFromPngFile^
  o:.cannonDrawer!
;

/*
Cannon at mid-x, bottom. at is blit top-left (mat size from drawer).
( -- )
*/
: Invaders_initModel_createCannon
  o:.base.pixSize@ l.sz!
  o:.cannonDrawer@ l.drw!
  SimEntity.new^ >ts
  l.sz.w 2 / l.drw.cols 2 / -
  l.sz.h l.drw.rows - 4 -
  Point.new^ t:.at!
  l.drw@
  l.drw@.newCycler^
  SimDrawerSet.new^
  t:.drawerSet!
  ts> dup o:.base.addEntity^
  o:.cannon!
;

/*
Load shared 2-frame alien AnimBlitDrawer (visual recipe; may be shared).
( -- )
*/
: Invaders_initModel_loadAlienDrawer
  (
    Invaders.alienAPng
    Invaders.alienBPng
  ) Invaders.alienPhaseRate AnimBlitDrawer.newFromPngFiles^
  o:.alienDrawer!
;

/*
4×4 Mat of Mat2c crops from alien-a (shared across explosions).
Also records alienW/alienH for per-spawn RectHitArea geometry.
( -- )
*/
: Invaders_initModel_loadGooMats
  Invaders.alienAPng Mat2c.newFromFile^ l.src!
  Invaders.gooGridRows Invaders.gooGridCols Mat.new^ dup o:.gooMats! l.grid!
  l.src.rows Invaders.gooGridRows / l.th!
  l.src.cols Invaders.gooGridCols / l.tw!
  0 Invaders.gooGridRows 1 - range
  foreach
    i l.r!
    0 Invaders.gooGridCols 1 - range
    foreach
      i l.c!
      l.r l.th *
      l.c l.tw *
      l.th l.tw
      l.src@.crop^
      l.r l.c l.grid![]
    /foreach
  /foreach
  // HitArea geometry for spawn (each alien mints its own RectHitArea)
  l.src.cols o:.alienW!
  l.src.rows o:.alienH!
;

/*
Count active live (non-goo) aliens.
( -- n:i ) o( inst )
*/
: Invaders_countLive
  0 l.n!
  o:.aliens@
  foreach
    i >ts
    t:.active if
      t:.subkind Invaders.subLive == if
        l.n 1 + l.n!
      then
    then
    tdrop
  /foreach
  l.n
;

/*
Mint one alien at random X, just above the top, with roundVy.
( -- ) o( inst )
*/
: Invaders_spawnOneAlien
  o:.base.pixSize.wi o:.alienW@ - 1 max Rand.int^ l.x!
  // top of playfield (avoid top border HitRect overlap)
  0 l.y!
  SimEntity.new^ >ts
  Invaders.HitKindAlien t:.kind!
  Invaders.subLive t:.subkind!
  l.x l.y Point.new^ t:.at!
  o:.roundVy@ t:.vy!
  // distinct HitArea per entity (MotionHit contacts key by HitArea identity)
  Point.zero
  o:.alienW@ o:.alienH@ Size.new^
  Rect.new^ Invaders.HitKindAlien RectHitArea.new^
  t:.hitArea!
  o:.alienDrawer@
  o:.alienDrawer@.newCycler^
  SimDrawerSet.new^
  t:.drawerSet!
  ts> dup o:.base.addEntity^
  o:.aliens.append^
;

/*
Try drip-spawn if under wave N and maxLive.
( -- ) o( inst )
*/
: Invaders_trySpawnAlien
  o:.spawned@ o:.waveN@ >= if
    // wave quota minted
  else
    Invaders_countLive Invaders.maxLive >= if
      // at cap — retry next update
    else
      Invaders_spawnOneAlien
      o:.spawned@ 1 + o:.spawned!
      o:.base.clock.ticks@ o:.spawnInterval@ + o:.nextSpawnTick!
    then
  then
;

/*
Drip spawn while playing and clock running.
( -- ) o( inst )
*/
: Invaders_pollSpawn
  o:.status@ Invaders.statusPlaying == if
    o:.base.clock.state SimClock.stateRunning == if
      o:.base.clock.ticks@ o:.nextSpawnTick@ >= if
        Invaders_trySpawnAlien
      then
    then
  then
;

/*
Pick N and reset kill/spawn/hit-Y counters. Keeps score and roundVy/interval.
Does not change status (caller sets playing or leaves won until Run).
( -- ) o( inst )
*/
: Invaders_rollWave
  0 o:.kills!
  0 o:.spawned!
  0.0 o:.hitYSum!
  0 o:.hitYCount!
  Invaders.nMax Invaders.nMin - 1 + Rand.int^
  Invaders.nMin +
  o:.waveN!
  o:.base.clock.ticks@ o:.spawnInterval@ + o:.nextSpawnTick!
;

/*
Start a wave as playing (init / hard New).
( -- ) o( inst )
*/
: Invaders_beginRound
  Invaders.statusPlaying o:.status!
  Invaders_rollWave
;

/*
Rebuild stars/cannon/aliens list after entity clear. Assets already loaded.
( -- ) o( inst )
*/
: Invaders_resetPlayfield
  o:.base.entities.clear^
  null o:.shot!
  List.new^ o:.aliens!
  Invaders_initModel_createStars
  Invaders_initModel_createStarfield
  Invaders_initModel_createCannon
;

/*
After a win: raise vy toward a 20% higher kill band; tighten spawn interval.
( -- ) o( inst )
*/
: Invaders_escalateFromWin
  o:.lastAvgHitY@ 0.0 == if
    // no samples — keep vy
  else
    o:.alienH@ -1 * >Float l.spawnY!
    o:.lastAvgHitY@ 0.8 * l.targetY!
    o:.lastAvgHitY@ l.spawnY@ - l.distOld!
    l.targetY@ l.spawnY@ - 1.0 max l.distNew!
    o:.roundVy@ l.distOld@ * l.distNew@ /
    Invaders.vyMax min
    Invaders.vyMin max
    o:.roundVy!
  then
  o:.spawnInterval@ >Float Invaders.spawnTighten * >Int
  Invaders.spawnIntervalMin max
  o:.spawnInterval!
;

/*
Baseline difficulty (first round / after loss).
( -- ) o( inst )
*/
: Invaders_resetDifficulty
  Invaders.vy0 o:.roundVy!
  Invaders.spawnInterval0 o:.spawnInterval!
  0.0 o:.lastAvgHitY!
;

/*
Shared shot drawer (visual; may be shared). HitArea minted per shot spawn.
( -- )
*/
: Invaders_initModel_loadShotAssets
  InvadersShotDrawer.new^ o:.shotDrawer!
;

: Invaders_initModel
  o:.base.entities.clear^
  null o:.shot!
  0 o:.score!
  0 o:.levelsCleared!
  Invaders_resetDifficulty
  List.new^ o:.aliens!

  Invaders_initModel_createStarPalette
  Invaders_initModel_createStars
  Invaders_initModel_createStarfield

  Invaders_initModel_loadAlienDrawer
  Invaders_initModel_loadGooMats

  Invaders_initModel_loadCannonDrawer
  Invaders_initModel_createCannon

  Invaders_initModel_loadShotAssets
  Invaders_beginRound
;

: Invaders#initModel
  Invaders_initModel
;

/*
Render stars with a twinkling effect.
Palette: 4 hue tranches × 4 brightness (0 dim, 1 base, 2 brighter, 3 flash).
Brightness = (phase xor flat xor i) & 3 — per-star desync from position+index,
no Rand, no shared offset (shared offset reads as a field-wide pulse).
( phase:i -- )
*/
: Invaders_renderStars
  l.phase!
  o:.base.pixmap >xs
  o:.starPalette@ l.pal!
  o:.starAtFlat@ l.sf!
  o:.starColor@ l.sc!
  
  0 o:.starAtFlat.len 1 - range
  foreach
    i l.sf@() l.flat!
    // palette idx = hue*4 + brightness
    i l.sc@() 4 *
    l.phase l.flat xor i xor 3 &
    +
    l.pal@()
    l.flat
    x:!()
  /foreach
  xdrop
;

/*
Starfield EffectDrawer paint. Library advances cycler; pass currentPhase for twinkle.
( entity:SimEntity cycler:SimDrawerCycler -- ) o( inst:Invaders )
*/
: Invaders#paintStars
  >ts drop t:.currentPhase tdrop
  Invaders_renderStars
;




//( -- ) o( inst )
: Invaders_renderModel
  
/*  o:.base.entities
  foreach
    i Invaders_renderModel_entity
  /foreach*/
;

/*
#update is only used for periodic game status updates.
Motion/Collision handles most every dynamic aspect.
( -- ) o( inst )
*/
: Invaders_updateStatus
  // score etc.
  
/*// log latest mhe mean
( o:.base.mhe.avgStepNs.mean 1000000.0 / )
"MHEngine avg step, ms: %8.3f" fmt $$
( o:.base.clock.avgRenderCostNs.mean 1000000.0 / )
"  Render avg, ms: %8.3f\n" fmt $$*/
;

: Invaders_pause
  o:.base.pause^
  Invaders_bgMusicOnPaused
;

/*
Resume clock after wave clear / ground. Game-won stays paused until New.
( -- ) o( inst )
*/
: Invaders_run
  o:.status@ Invaders.statusGameWon == if
    // session complete — New to play again
  else
    o:.status@ Invaders.statusWaveClear == if
      Invaders.statusPlaying o:.status!
      o:.base.clock.ticks@ o:.spawnInterval@ + o:.nextSpawnTick!
    else
      o:.status@ Invaders.statusWaveGround == if
        Invaders.statusPlaying o:.status!
        o:.base.clock.ticks@ o:.spawnInterval@ + o:.nextSpawnTick!
      then
    then
    o:.base.start^
    Invaders_bgMusicOnRunning
  then
;

/*
Hard reset (Life / Zombies style): score, difficulty, playfield.
( idx:i -- ) o( inst )
*/
: Invaders_new
  Invaders_pause
  Invaders_initModel
  o:.base.doRender^
;


/*
Motion Hit handler.
Shot↔border → remove shot. Shot↔alien → goo explode + score.
( he:HitEvent -- ) o( inst )
*/
: Invaders_removeShotEnt
  // ( shot:SimEntity -- )
  >ts
  t:peek o:.base.entities.remove^
  null o:.shot!
  tdrop
;

/*
Random Point velocities for goo radiate (δ pixels per phase).
( -- vels:Mat )
*/
: Invaders_newGooVels
  Invaders.gooGridRows Invaders.gooGridCols Mat.new^ >ts
  0 Invaders.gooGridRows 1 - range
  foreach
    i l.r!
    0 Invaders.gooGridCols 1 - range
    foreach
      i l.c!
      // away from grid center (~1.5,1.5) + small jitter
      l.c 1 - 4 *
      5 Rand.int^ 2 - +
      l.r 1 - 4 *
      5 Rand.int^ 2 - +
      Point.new^
      l.r l.c t:![]
    /foreach
  /foreach
  ts>
;

/*
Swap alien to RadialAnimBlitDrawer goo; stop motion; score + kill; maybe win.
( alien:SimEntity -- ) o( inst )
*/
/*
Wave cleared (kills == N). Escalate + next wave, or session win after levelsToWin.
( -- ) o( inst )
*/
: Invaders_roundWon
  o:.status@ Invaders.statusPlaying == if
    o:.hitYCount@ 0 > if
      o:.hitYSum@ o:.hitYCount@ / o:.lastAvgHitY!
    then
    o:.levelsCleared@ 1 + o:.levelsCleared!
    Invaders_pause
    o:.levelsCleared@ Invaders.levelsToWin >= if
      Invaders.statusGameWon o:.status!
      o:.sfxGameWin@ Invaders_playSfx
      Invaders_resetPlayfield
    else
      Invaders.statusWaveClear o:.status!
      o:.sfxLevelUp@ Invaders_playSfx
      Invaders_escalateFromWin
      Invaders_resetPlayfield
      Invaders_rollWave
    then
    o:.base.doRender^
  then
;

: Invaders_explodeAlien
  >ts
  t:.at.y o:.hitYSum@ + o:.hitYSum!
  o:.hitYCount@ 1 + o:.hitYCount!
  o:.kills@ 1 + o:.kills!
  Invaders.subGoo t:.subkind!
  true t:.stopped!
  true t:.anchored!
  // drop collider so goo debris cannot freeze a shot (hitActionStop)
  null t:.hitArea!
  0.0 t:.vx!
  0.0 t:.vy!
  o:.gooMats@
  Invaders_newGooVels
  Invaders.gooPhaseRate RadialAnimBlitDrawer.new^ >xs
    x:.newCycler^ >ys
      Invaders.gooPhases y:.phaseCount!
      1 y:.cyclesToRun!
    x:peek y:peek
    SimDrawerSet.new^
    t:.drawerSet!
    ydrop
  xdrop
  o:.score@ Invaders.pointsPerKill + o:.score!
  o:.sfxExplosion@ Invaders_playSfx
  tdrop
  o:.kills@ o:.waveN@ == if
    Invaders_roundWon
  then
;

: Invaders_shotHitOther
  // ( shot:SimEntity other:SimEntity -- ) o( inst )
  l.other! l.shot!
  l.other.kind SimEntity.kindBorder == if
    l.shot@ Invaders_removeShotEnt
  else
    l.other.kind Invaders.HitKindAlien == if
      l.other.subkind Invaders.subLive == if
        l.shot@ Invaders_removeShotEnt
        l.other@ Invaders_explodeAlien
      then
    then
  then
;

/*
Alien reached the ground: end round only (not session loss).
Penalty = half remaining kill value; next wave ready; press Run.
( -- ) o( inst )
*/
: Invaders_roundGround
  o:.status@ Invaders.statusPlaying == if
    Invaders.statusWaveGround o:.status!
    o:.waveN@ o:.kills@ -
    Invaders.pointsPerKill *
    2 /
    o:.score@ swap - 0 max o:.score!
    o:.sfxSlime@ Invaders_playSfx
    Invaders_pause
    // keep difficulty; roll a fresh wave for Run
    Invaders_resetPlayfield
    Invaders_rollWave
    o:.base.doRender^
  then
;

: Invaders_alienHitOther
  // ( alien:SimEntity other:SimEntity -- ) o( inst )
  l.other! l.alien!
  l.alien.subkind Invaders.subLive == if
    l.other.kind SimEntity.kindBorder == if
      // bottom only (ignore top/side); feet near playfield bottom
      l.alien.at.y o:.alienH@ +
      o:.base.pixSize.h 16 -
      >= if
        Invaders_roundGround
      then
    then
  then
;

: Invaders#handleHits
  >ts
  t:.hits
  foreach
    i >xs
      x:.entity1 l.e1!
      x:.entity2 l.e2!
      l.e1.kind Invaders.HitKindShot == if
        l.e1@ l.e2@ Invaders_shotHitOther
      else
        l.e2.kind Invaders.HitKindShot == if
          l.e2@ l.e1@ Invaders_shotHitOther
        else
          l.e1.kind Invaders.HitKindAlien == if
            l.e1@ l.e2@ Invaders_alienHitOther
          else
            l.e2.kind Invaders.HitKindAlien == if
              l.e2@ l.e1@ Invaders_alienHitOther
            then
          then
        then
      then
    xdrop
  /foreach
  tdrop
;

/*
Remove goo aliens whose one-shot cycler finished.
( -- ) o( inst )
*/
: Invaders_pollGooDone
  o:.aliens@
  foreach
    i >ts
    t:.active if
      t:.drawerSet.cycler.cyclesDone if
        t:peek o:.base.entities.remove^
      then
    then
    tdrop
  /foreach
;

/*
 Button event handler.
( idx:i -- ) o( inst )
*/
: Invaders#handleButtonTap
  switch
    a Invaders.btnRun == case
      Invaders_run
    /case
    
    a Invaders.btnPause == case
      Invaders_pause
    /case
    
    a Invaders.btnNew == case
      Invaders_new
    /case    
  /switch
;

: Invaders_renderHud
  o:.base.painter@ >xs
    0xFFFFFFB4 Color.hexa^ x:.font.color!
    "go" 18.0 x:.font.setFamilyAndSize^
    TextAlign.leftBaseline x:.font.align!
    o:.status@
    switch
      a Invaders.statusWaveClear == case "CLEAR — press Run" l.st! /case
      a Invaders.statusWaveGround == case "GROUND — press Run" l.st! /case
      a Invaders.statusGameWon == case "YOU WIN — New to replay" l.st! /case
      default "playing" l.st! /default
    /switch
    (
      o:.score@
      o:.levelsCleared@
      Invaders.levelsToWin
      o:.kills@
      o:.waveN@
      l.st@
    ) "Score %d   Lv %d/%d   %d/%d   %s" fmt
    12 22 Point.new^
    x:.text^
  x:drop
;

: Invaders#handleRender
  Invaders_renderModel
  Invaders_renderHud
;

/*
( -- ) o( inst )
*/
: Invaders#handleUpdate
  Invaders_updateStatus
  Invaders_pollSpawn
  Invaders_pollGooDone
;

/*
We use this to output average update/render rates.
Note: it updates in ticks, not seconds, so if you
set the speed to more than 1 it gets called more frequently.
( index:i -- ) o( inst )
*/
: Invaders#handleTimerEvent
  drop
  true // use stderr
  o:.base.printAvgUpdateRenderRate^
;

// ( dx:i dy:i cannon:Entity -- ) o( inst )
: Invaders_moveCannon
  l.cannon!
  l.cannon.moveInBounds^
  o:.base.doRender^
;

/*
Set cannon world X from pointer X (center); keep Y; clamp. No render.
( x:n -- ) o( inst )
*/
: Invaders_setCannonX
  o:.cannonDrawer.cols 2 / - // pointer → blit top-left
  o:.cannon >ts
  t:.at.y Point.new^
  t:.at!
  0 0 t:.moveInBounds^
  t:drop
;

/*
True when no live shot is in flight.
( -- flag:b ) o( inst )
*/
: Invaders_canFire
  o:.shot@ Null? if
    true
  else
    o:.shot.active not
  then
;

/*
Spawn one PhaserShot above cannon mid-x; vy up. At most one live.
( -- ) o( inst )
*/
: Invaders_spawnShot
  o:.cannon >ts
    t:.at.x o:.cannonDrawer.cols 2 / +
    t:.at.y Invaders.shotRadius - 2 -
    Point.new^
  tdrop
  l.at!
  SimEntity.new^ >ts
  Invaders.HitKindShot t:.kind!
  l.at t:.at!
  Invaders.shotVy t:.vy!
  Point.zero Invaders.shotRadius Invaders.HitKindShot CircleHitArea.new^
  t:.hitArea!
  o:.shotDrawer@
  SimDrawerCycler.new^
  SimDrawerSet.new^
  t:.drawerSet!
  ts> dup o:.base.addEntity^
  o:.shot!
;

/*
One fire word — space or mouse primary. At most one live shot.
( -- ) o( inst )
*/
: Invaders_fire
  o:.status@ Invaders.statusPlaying == if
    Invaders_canFire if
      Invaders_spawnShot
      o:.sfxLaser@ Invaders_playSfx
    then
  then
;

// ( ke:KeyEvent -- ) o( inst )
: Invaders_onKeyDown
  l.ke!
  o:.cannon l.can!
  Invaders.cannonKeyMoveScale l.mov!
  l.ke.name
  switch
    a Keys.left == case
      l.mov -1 * 0 l.can Invaders_moveCannon
    /case
    a Keys.right == case
      l.mov      0 l.can Invaders_moveCannon
    /case
    a Keys.space == case
      Invaders_fire
    /case
  /switch
;

/*
MouseDown — primary button fires. No render.
( me:MouseEvent -- ) o( inst )
*/
: Invaders#handleMouseDown
  >ts
  t:.button MouseEvent.btnPrimary == if
    Invaders_fire
  then
  t:drop
;

/*
MouseUp — unused.
( me:MouseEvent -- )
*/
: Invaders#handleMouseUp
  drop
;

/*
MouseIn — unused.
( me:MouseEvent -- )
*/
: Invaders#handleMouseIn
  drop
;

/*
MouseMoved — track cannon X to pointer. No render (periodic renderRate).
( me:MouseEvent -- ) o( inst )
*/
: Invaders#handleMouseMoved
  >ts
  t:.position.x
  t:drop
  Invaders_setCannonX
;

/*
MouseOut — unused.
( -- )
*/
: Invaders#handleMouseOut
;

/*
KeyDown event handler.
Attached to Window.KeyDown events
( ke:KeyEvent -- ) o( inst )
*/
: Invaders#handleKeyDown
  Invaders_onKeyDown
;

/*
KeyUp event handler.
Attached to Window.KeyUp events
( ke:KeyEvent -- ) o( inst )
*/
: Invaders#handleKeyUp
  drop
  // not used
;

/*
Window close — stop BG loop.
( -- ) o( inst )
*/
: Invaders#handleWindowClose
  Invaders_bgMusicOnStopped
;

/* Required
Called by SimGame before it does its own internal init logic.
Use this to set up all your non-Fyne objects, and set
the SimGame options you will be using.
Use this to:
* load all your assets, images, sounds, etc
* enable optional SimGame options, like MotionHit engine
( -- ) o( inst:Invaders )
*/
: Invaders#preInit
  o:.base@ >ys

  Invaders_ensureSounds

  // SimGame options
  "Alien Invaders - clear them before they reach the ground"
  y:.title!
  Invaders.backgroundColor y:.backgroundColor!
  // we use keyboard for cannon and fire
  true y:.options.hasKeyHandler!
  true y:.options.useKeyAutoRepeat!
  // user can use mouse for cannon and fire
  true y:.options.hasMouseDownUpHandler!
  true y:.options.hasMouseInMovedOutHandler!
  true y:.options.hasWindowCloseHandler!
  // update is only for things like spawning
  // all updates are done via events, and motion is automatic
  // we just enable for logging of rates
  200 y:.clock.updateRate! // 5/s
  // we periodically render for smooth motion
  // and set .autoRender on for auto rendering on most events
  50 y:.clock.renderRate! // 20/s
  y:.clock.setAutoRenderOn^
  
  // MotionHitEngine — borders + shot↔alien
  true y:.options.useMotionEngine!
  true y:.options.useHits!
  SimGameOptions.hitActionStop SimGameOptions.hitEvents | 
  y:.options.setAllBoundaryModes^
  Invaders.HitKindShot Invaders.HitKindAlien
  SimGameOptions.hitActionStop SimGameOptions.hitEvents |
  y:.options.addHitMatrixEntry^
  
  50 y:.clock.mheRate! // 20/s
  
  // Invaders options and setup
  //Invaders.defRadiusScale o:.radiusScale!
   
  // Timer/ time-related
  1.0 y:.clock.simSpeed!
  0 1000 y:.timerEventEvery^ // for outputing update rate
  
  ydrop
;

/* Required
Called by SimGame after it does its own internal init logic.
Use this to apply additional options or setup, now that the
features of SimGame have been initialized.
The main Pixmap has been created, and is accessible via o:.base@.pixmap
This is a good place to build your sim/game model.
( -- ) o( inst:Invaders )
*/
: Invaders#postInit
  Invaders_initModel
;


/* Required if options.hasBottomBar is set (default is true)
Called by SimGame during #makeUi. Return a Fyne layout object
or Widget, typically an HBox or Border that is the content
below the view.
( -- lay:Layout | widg:Widget ) o( inst:Invaders )
*/
: Invaders#makeBottomBar
  o:.base@ >ys
  ( // buttons
    Invaders.btnRun   "Run" Icon.MediaPlay y:.makeButtonWithIcon^
    Invaders.btnPause "Pause" Icon.MediaPause y:.makeButtonWithIcon^
    Invaders.btnNew   "New" y:.makeButton^
  ) Box.newH^
  ydrop
;


/*
Typical launch:
*/
: invaders
  Invaders.new^ >zs
  z:.base >ys // SimGame inst
  y:.init^
  y:.makeUi^
  y:.startPaused^
  y:.doRender^
  y:.win@.show^
  ydrop zdrop
;


invaders-blurb $ nl
