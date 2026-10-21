package core

const bootstrapInitScript0 = `

// === Word Definition Helpers ===

: immediate 0x0010 setWordFlags ;

: [imm_ 24 pushBlock ;
: [imm [imm_ // ( -- laststate:i ) interpret-in-compile island
  getCompileStateFlags dup
  0x01 bitnot & // clear Compiling aka now Interpreting
  setCompileStateFlags
; 0x032 setWordFlags // CompileOnly | Immediate | Colon

: imm]_ "'imm]' requires a preceding '[imm'" 24 panicDropBlock ;
: imm] imm]_ // ( state:i )
  setCompileStateFlags
; 0x032 setWordFlags // CompileOnly | Immediate | Colon


: [[_ 26 pushBlock ;
: [[ [[_ // ( -- laststate:i )
  getCompileStateFlags dup
  0x05 | // set Compiling | Emitting
  setCompileStateFlags
; 0x032 setWordFlags // CompileOnly | Immediate | Colon

: ]]_  "']]' requires a preceding '[['" 26 panicDropBlock ;
: ]] ]]_ // ( state:i )
  setCompileStateFlags
; 0x1032 setWordFlags // ExecutesInEmitting | CompileOnly | Immediate | Colon

// === Compile-flag scopes (compiled into parent CT; not immediate) ===
// Snapshots on XS — do not use DS (break marks / >mark walk live there).
// System | CompileOnly — future: forbidden outside System mode.

: [i
  getCompileStateFlags >xs
  getCompileStateFlags 0x05 bitnot & setCompileStateFlags
; 0x0820 setWordFlags

: [c
  getCompileStateFlags >xs
  getCompileStateFlags 0x04 bitnot & 0x01 | setCompileStateFlags
; 0x0820 setWordFlags

: *]
  xs> setCompileStateFlags
; 0x0820 setWordFlags

// === Recursion (compile-time self-call) ===
// recurse — immediate; emits OpNest to PeekBuilding().colonCode (no dictionary find).

// Control Structure Helpers
// <name>_ handles block token push/check/pop on ts
// It panics for unexpected or illegalblock kinds.
// see compile_state.go for the block kind constants.

// Colon/Semicolon implement these in bootstrap go code:
// :_ 1 pushBlock ;
// ;_ "; (semicolon) requires all code blocks to be closed; eg if ... then" 1 panicDropBlock ;

// === Mark / Resolve Primitives ===
// >mark ( -- orig ); >resolve ( orig -- ) patches j=here-(orig+1) and drops orig (axiom).
: >mark    here 0 , ;

: <mark    here ;
// Backward j = orig - (here+1); branch base is one past operand slot (06-control-flow.md).
: <resolve here - 1 - , ;

// === Conditionals ===
: if_      2 pushBlock ;
: if if_      [[ branch? ]] >mark ; immediate 0x0020 setWordFlags // +CompileOnly
: nif if_ [[ not branch? ]] >mark ; immediate 0x0020 setWordFlags // +CompileOnly

: else_
    "'else' requires a preceding 'if'" 2 panicRequireBlock
    3 pushBlock ;
: else else_    [[ branch ]] >mark swap >resolve ; immediate 0x0020 setWordFlags // +CompileOnly

: then_ 
    3 acceptDropBlock
    "'then' requires a preceding 'if' or 'else'" 2 panicDropBlock ;
: then then_    >resolve ; immediate 0x0020 setWordFlags // +CompileOnly

// === Indefinite Loops ===
// begin pushes compile-DS sentinel 0 for break marks (like switch).
// forever … /forever is the continue-capable infinite loop (sandwich closer).
: begin_      5 pushBlock ;
: begin begin_   <mark 0 ; immediate 0x0020 setWordFlags // +CompileOnly

: forever_    31 pushBlock ;
: forever forever_   <mark dup setBlockMark 0 ; immediate 0x0020 setWordFlags // +CompileOnly

: loopEndBreakStep
  dup if
    [c breakResolve *]
    recurse
  else
    drop
  then ;

: again_ "'again' requires a preceding 'begin'" 5 panicDropBlock ;
: again again_    [[ branch ]] [i loopEndBreakStep *] <resolve ; immediate 0x0020 setWordFlags // +CompileOnly

: /forever_ "'/forever' requires a preceding 'forever'" 31 panicDropBlock ;
: /forever /forever_    [[ branch ]] [i loopEndBreakStep *] <resolve ; immediate 0x0020 setWordFlags // +CompileOnly

: until_ "'until' requires a preceding 'begin'" 5 panicDropBlock ;
: until until_    [[ branch? ]] [i loopEndBreakStep *] <resolve ; immediate 0x0020 setWordFlags // +CompileOnly

// === BEGIN / WHILE / REPEAT ===
: while_      
    "'while' requires a preceding 'begin'" 5 panicRequireBlock
    10 pushBlock ;
: while while_   [[ branch? ]] >mark ; immediate 0x0020 setWordFlags // +CompileOnly

: repeat_ 
    "'repeat' requires a preceding 'while'" 10 panicDropBlock
    // note: the while_ logic confirmed the 'begin'; following should be redundant
    "'repeat' requires a preceding 'begin'" 5 panicDropBlock ;
: repeat repeat_
  >xs
  [[ branch ]]
  [i loopEndBreakStep *]
  <resolve
  xs> >resolve
; immediate 0x0020 setWordFlags // +CompileOnly

// === Foreach (iterator loop; pre-test) ===
// Shared exit epilogue after /foreach: (dropIter) (exhaust and break).
// Iterators park on ColonFrameMark Slot.Foreach (not RS). <mark = continue/back-target; 0 = break sentinel.

// get a native go iterarator, or one from ._tryGetIter
// error if no iterator
// ( provider:a -- iter:Iterator ) error if no iterator
: (foreachPrologue)
  (tryGetIter) not if
    ._tryGetIter^ not if
      drop "foreach requires an object that is iterable" error
    then
  then
;

: foreach_      12 pushBlock ;
: foreach foreach_ [[ (foreachPrologue) (pushIter) ]] <mark dup setBlockMark 0 [[ (foreachTop) ]] [[ branch? ]] >mark ; immediate 0x0020 setWordFlags // +CompileOnly

: /foreach_ "'/foreach' requires a preceding 'foreach'" 12 panicDropBlock ;
// Stash exitOrig; drain break marks; back-branch; patch exit+breaks to shared drop-iter epilogue.
: /foreach /foreach_
  >xs
  [[ branch ]]
  [i loopEndBreakStep *]
  <resolve
  xs> >resolve
  [[ (dropIter) ]]
; immediate 0x0020 setWordFlags // +CompileOnly

: ijk_ "'i/j/k' require 1/2/3 outer 'foreach' loops" 12 panicRequireNKind ;
: i 1 ijk_ [[ (i) ]] ; immediate 0x0020 setWordFlags // +CompileOnly
: j 2 ijk_ [[ (j) ]] ; immediate 0x0020 setWordFlags // +CompileOnly
: k 3 ijk_ [[ (k) ]] ; immediate 0x0020 setWordFlags // +CompileOnly

// === Target switch (first-match-wins; Slot SwitchTargets) ===
// switch pops target from DS via (pushSwitch); /switch shared epilogue (dropSwitch).
// a/b/c fetch nested targets (like i/j/k). Default body optional between default and /default.
// Early exit mid-arm: wrap remaining body in if (no exitSwitch word).

: switch_
  (assertSwitchNest)
  14 pushBlock ;
: switch switch_ [[ (pushSwitch) ]] 0 ; immediate 0x0020 setWordFlags // +CompileOnly

: case_
    16 acceptDropBlock
    "'case' requires a preceding 'switch'" 14 panicRequireBlock
    15 pushBlock ;
: case case_    [[ branch? ]] >mark ; immediate 0x0020 setWordFlags // +CompileOnly

: /case_ "'/case' requires a preceding 'case'" 15 panicDropBlock 16 pushBlock ;
: /case /case_    [[ branch ]] >mark swap >resolve ; immediate 0x0020 setWordFlags // +CompileOnly

: default_
    16 acceptDropBlock
    "'default' requires a preceding 'switch'" 14 panicRequireBlock
    18 pushBlock ;
: default default_   nop ; immediate 0x0020 setWordFlags // +CompileOnly

: /default_ "'/default' requires a preceding 'default'" 18 panicDropBlock 19 pushBlock ;
: /default /default_   nop ; immediate 0x0020 setWordFlags // +CompileOnly

: /switchStep  dup if [c >resolve *] recurse else drop then ;

: /switch_
    19 acceptDropBlock
    16 acceptDropBlock
    "'/switch' requires a preceding 'switch'" 14 panicDropBlock ;
// Resolve exit marks to HERE, then emit shared (dropSwitch) epilogue (like /foreach + dropIter).
: /switch /switch_    [i /switchStep *] [[ (dropSwitch) ]] ; immediate 0x0020 setWordFlags // +CompileOnly

: abc_ "'a/b/c' require 1/2/3 outer 'switch' blocks" 14 panicRequireNKind ;
: a 1 abc_ [[ (a) ]] ; immediate 0x0020 setWordFlags // +CompileOnly
: b 2 abc_ [[ (b) ]] ; immediate 0x0020 setWordFlags // +CompileOnly
: c 3 abc_ [[ (c) ]] ; immediate 0x0020 setWordFlags // +CompileOnly

// === try / catch / finally / /try (161-try-catch) ===
// BlockKinds 32–35. CatchMark snapshots value stacks at try; throw restores then enters catch.
// Compile DS holds catchIP then skip-over-catch mark (same discipline as if/else).

: try_ 32 pushBlock ;
: try try_ [[ pushCatchMark ]] here 0 , ; immediate 0x0020 setWordFlags // +CompileOnly

: catch_
    "'catch' requires a preceding 'try'" 32 panicDropBlock
    33 pushBlock ;
: catch catch_
  [[ branch ]] >mark
  swap >resolve
; immediate 0x0020 setWordFlags // +CompileOnly

: finally_
    "'finally' requires a preceding 'catch'" 33 panicRequireBlock
    34 pushBlock ;
: finally finally_ >resolve ; immediate 0x0020 setWordFlags // +CompileOnly

: /try_
    34 acceptDropBlock
    "'/try' requires a preceding 'catch'" 33 panicDropBlock ;
: /try /try_
  depth 0 > if >resolve then
  [[ dropCatchMark ]]
; immediate 0x0020 setWordFlags // +CompileOnly

// === Literal compile (slice-002.2) ===
// s( value type -- ) c( -- +literal )
: compileLiteral  drop [[ literal, ]] ;

// addWord ( w:*Word -- ) — commit word to execute DC dictionary
: addWord  dc.dict.add^ ;

// === Flag helpers ===

: hasFlag // ( value:i mask:iu -- b )
  & <>0
;

// === Quit leave request (slice-002.6 partial; Go repl loop still authoritative) ===

: bye  true byeReq ! ;

: checkResetByeReq
  byeReq @
  dup if false byeReq ! then
;

: exit  stop ;

// === Type member dispatch (110-objects-and-types.md prototype) ===
// s( index:i kind:i -- … ) ths on context stack.
: typeMemberInvoke  armSubCode typeMemberDispatch ;

/*
Find the word whose name is on the stack.
Returns the word if found, or raises an error.

( name:s -- w:Word ) or error if not found
*/
: find
  dup >ts
  find? if
    t:drop
  else
    "find: word not found: " ts> str::+ error
  then 
;

`
