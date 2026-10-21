package core

const bootstrapInitScript1 = `
// Variables and Constants

/*
Creates a new variable that can hold any value.
The initial value is on the stack.
The name is read from the input stream.
Interpreter only — not valid while compiling a colon.
At runtime <name> pushes a DTSlotRef; use @ / ! .
( init:a -- )
*/
: var:
  readIdentifier
  mintVarWord
; immediate 0x4000 setWordFlags // +InterpretOnly

/*
Creates a new constant that can hold any value.
The value is on ds.
The name is read from the input stream.
Interpreter only — not valid while compiling a colon.
At runtime <name> pushes the frozen value (OpDoLit).
( value:a -- )
*/
: const:
  readIdentifier
  mintConstWord
; immediate 0x4000 setWordFlags // +InterpretOnly

// Output Helpers

/*
Output a space.
( -- )
*/
: sp
  " " $
;

/*
Output a newline \n .
( -- )
*/
: nl
  "\n" $
;

/*
Output a tab.
( -- )
*/
: tab
  "\t" $
;

/*
Enqueue all the items on ds since ( into the list.
Supports any type of List, provided values can be enqueued.
( l:*List -- l )
*/
: enqueueFromMark
  >ys
  x:getContextMark
  depth x:.ds - 1 - 0 swap range
  foreach
    y:.enqueue^
  /foreach
  ys>
  x:drop
;

:  (})_List
  List.new^ >ys
  x:getContextMark
  depth x:.ds - 1 - 0 swap range
  x:drop
  // get items in reverse order, then we pop later
  foreach
    y:.append^
  /foreach
;

// y( List -- ) z( inst:DynamicObject | StaticObject | StringMap)
: (})_FillObject
  y:.len 2 / 1 - 0 swap range 
  foreach
    y:.pop^ // value
    y:.pop^ // name
    z:!{}
  /foreach
  y:drop
;

// ( inst:DynamicObject | StringMap )
: (})
  >zs
  (})_List
  (})_FillObject
  zs>
;

/*
Create a DynamicObject from the items on ds since {.
The items are pairs of value and name.
The name must be a string, the value can be any type.
eg: (  -- do:DynamicObject ) { 1 "one" 2 "two" }
*/
: }
  DynamicObject.new^
  (})
;

/*
Create a StaticObject from the supplied DynamicObject
Prototype, and set the values from the items on ds since {.
The items are pairs of value and name.
The name must be a string, the value can be any type.
NOTE: the fields MUST match be on the prototype.
You do not need to provide all the fields.
eg: ( do:DynamicObject -- so:StaticObject ) do { 1 "one" 2 "two" so}
*/
: so}
  (})_List // y( List -- ); prototype remains on DS
  >ts t:.newStatic^ >zs t:drop
  (})_FillObject // y( List -- ) z( inst:StaticObject )
  zs>
;

/*
Create a StringMap from the items on ds since {.
The items are pairs of value and name.
The name must be a string, the value can be any type.
eg: { 1 "one" 2 "two" -- sm:StringMap }
*/
: sm}
  StringMap.new^
  (})
;

/*
Gather up all the items on ds since ( and put them in a List.
list[0] contains the most deeply nested item, list[n-1] has ds.tos.
The items can be any primitive or compound type.
The resulting list is useful as input to fmt.
( [ds.len - cm.ds item:a] -- ) cs( -- l:list ) r( ... cm:ContextMark ... -- ... )
*/
: ) // ( -- list:AnyList )
  List.new^ 
  enqueueFromMark
;

/*
Gather up all the items on ds since ( and put them in a StringList.
list[0] contains the most deeply nested item, list[n-1] has ds.tos.
The items MUST be strings.
The resulting list is useful as input e.g. to str::join.
( [ds.len - cm.ds item:a] -- list:StringList ) cs( -- l:list ) r( ... cm:ContextMark ... -- ... )
*/
: s)
  StringList.new^ 
  enqueueFromMark
;

/*
Gather up all the items on ds since ( and put them in an IntList.
list[0] contains the most deeply nested item, list[n-1] has ds.tos.
The items MUST be Int.
( [ds.len - cm.ds item:a] -- list:StringList ) cs( -- l:list ) r( ... cm:ContextMark ... -- ... )
*/
: i)
  IntList.new^ 
  enqueueFromMark
;

/*
Gather up all the items on ds since ( and put them in an ByteList.
list[0] contains the most deeply nested item, list[n-1] has ds.tos.
The items MUST be Int, and SHOULD be in the range 0-255.
( [ds.len - cm.ds item:a] -- list:ByteList ) cs( -- l:list ) r( ... cm:ContextMark ... -- ... )
*/
: b)
  ByteList.new^ 
  enqueueFromMark
;

// ( list:List | list:StringList -- set:Set | set:StringSet ) y( set:Set | set:StringSet -- )
: fillSetFromMark
  foreach
    i y:.put^
  /foreach
  ys>
;

/*
Gather up all the items on ds since ( and put them in a Set.
( [ds.len - cm.ds item:a] -- set:Set ) r( ... cm:ContextMark ... -- ... )
*/
: set)
  )
  Set.new^ >ys
  fillSetFromMark
;

/*
Gather up all the items on ds since ( and put them in a StringSet.
( [ds.len - cm.ds item:a] -- set:Set ) r( ... cm:ContextMark ... -- ... )
*/
: sset)
  s)
  StringSet.new^ >ys
  fillSetFromMark
;

/*
Gather up all the items on ds since [ and fill a Mat.
The items MUST be the same type as the Mat holds.
There must be exactly as many items as the Mat has elements.
( mat:Mat* [ds.len - cm.ds item:a] -- mat ) r( ... cm:ContextMark ... -- ... )
*/
: ]
  ) >xs >ys // x: List y: Mat*
  x:.len dup y:.len <>
  if
    drop
    "]: list length does not match mat length" error
  then
  1 - 0 swap range
  foreach
    i x:@()
    i y:!()
  /foreach
  x:drop
  ys>
;

// Loop exit gates (kinds: Foreach=12, Begin=5, Forever=31)
: break_
  ( 12 5 31 ) "'break' requires an enclosing loop" panicRequireOneOfKind ;
// Bury the break forward-mark under any nested if/else >mark cells so that
// then/>resolve still patches those conditionals; leave the break mark just
// above the loop's 0 sentinel for loopEndBreakStep / breakResolve.
// ( ... 0 [ifMarks...] br -- ... 0 br [ifMarks...] )
: (rollBreakUnderIfs)
  over 0 == if
  else
    swap >xs
    recurse
    xs>
  then
;
: break  break_  [[ branch ]] >mark (rollBreakUnderIfs) ; immediate 0x0020 setWordFlags // +CompileOnly

: continue_
  ( 12 31 ) "'continue' requires 'foreach' or 'forever'" panicRequireOneOfKind ;
: continue  continue_  [[ branch ]] continueMark <resolve ; immediate 0x0020 setWordFlags // +CompileOnly

// Bootstrap Word Documentation Helpers

/*
Set the doc string for the existing word from the last comment.
Used to document axioms, native-go words, and bootstrap words.
Document Type members with syntax: TypeName#memberName.
Use "__items" as member name for itemsMetadata.
Does not panic if the word, type, or member is not found;
it just logs a warning.
<name> ( -- )
*/
: :doc
  false >zs // true if any succeeded
  // first, try to find it as a whole word, even if in member syntax
  readIdentifier dup >ts
  find? if
    .setDocFromLastComment^
    drop // word is a receiver, not elided, so drop it
    z:drop true >zs
  then
  
  // if it has a # then break it into TypeName memberName
  t:peek "#" str::indexOf 0 >= if
    t:peek "#" str::split >xs // note: always has 2 (if name has 1 #)
    0 x:@() dup >ys find? if
      >zs z:.isType if
        // get the underlying type object
        z:.exec^ // ( -- t:Type )
        z:drop >zs
        1 x:@() z:.getMemberForDoc^ if 
          >ts
            z:.getClearLastComment^
            t:.doc!
          t:drop
          z:drop
          z:drop true >zs
        else
          ( y:peek 1 x:@() )
          "WARNING: :doc type:%s member:%s not found.\n" fmt $
          z:drop
        then
      else
        "WARNING: :doc is not a type name: " y:peek str::+ $ nl
        z:drop
      then
      
    else
      "WARNING: :doc typename not found: " y:peek str::+ $ nl
    then
    y:drop
    
    x:drop
  then

  zs> nif
    "WARNING :doc word not found: " t:peek str::+ $ nl
  then   
  t:drop
;

// example of how to use :doc for a word
/*
Interactive List builder
( ) lets you build a list from items on ds.
( marks stack depth.
Then add items to ds.
) or s) collects all the items down to the depth mark
into a List, returning the List on ds.
You can use these in the interpreter or compiled
r( -- cm:ContextMark )
*/
:doc (

/*
Interactive Object builder
{ } or sm} lets you build a DynamicObject or StringMap from items on ds.
{ marks stack depth.
Then add items to ds.
} or sm} collects all the items down to the depth mark
into a DynamicObject or StringMap, returning the object on ds.
You can use these in the interpreter or compiled.
eg. { 1 "one" 2 "two" -- do:DynamicObject }
r( -- cm:ContextMark )
*/
:doc {

/*
Interactive Mat/VecNf filler. (NOT builder!)
NOTE: unlike ( or {, your Mat2f or VecNf must exist on DS before [.
[ ] fills values: [ marks stack depth (axiom OpBracket);
add items in flat / row-major order; ] gathers them and
writes via !(). Length must match .len or ] panics.
eg. 2 3 Mat2f.new^ [ 0 1 2 3 4 5 ]
eg. 3 VecNf.new^ [ 1 2 3 ]
( mat:Mat*|VecNf -- same ) r( -- cm:ContextMark )
*/
:doc [

/*
Close Mat/VecNf fill begun with [. See :doc [ .
( mat:Mat*|VecNf -- same )
*/
:doc ]

/*
Interpret-in-compile island (was Forth-style [ ]).
Clears Compiling so the body runs while defining a colon word.
Pair with imm]. Distinct from Mat filler [ ] and emit [[ ]].
r( -- ) c( -- )
*/
:doc [imm

/*
Close interpret-in-compile island begun with [imm.
*/
:doc imm]

/*
Creates a reload mark word in the current DictionaryContext.
Put the mark alone on line 1 of a reloadable script:
forget: reset
compileFromFile runs that mark first when it exists, then
checkpoints and loads (recreating the mark). Invoking the mark
by name also forgets the mark and everything defined after it.
dict, CT, DT, and the local type catalog roll back to the stamped
high-waters. The mark word is interpret-only.
Refuses (no mutation) if a surviving type still has a Final method
whose body is past the CT keep.
<name> ( -- )
*/
: forget:
  (forget:)
; immediate

// Defer words and 'is'

/*
Creates a word with a slot that holds the XT of the target word.
Initially unset; running the defer panics until bound with 'is'.
Interpreter only — not valid while compiling a colon.
( -- )
*/
: defer:
  readIdentifier
  mintDeferWord
; immediate 0x4000 setWordFlags // +InterpretOnly

/*
Sets the runtime behavior of a deferred word from an XT on DS.
Use it like:
&of <target> is <defer: word>
Interpret: bind now. Compile (immediate): emit store of TOS XT into the defer slot.
( xt -- ) | c( -- +lit dtIdx +_DTRef +! )
*/
: is
  readIdentifier dup >ts
  find?
  if
    >ys t:drop
  else
    drop
    "'is': requires a defer word; word not found: " ts> str::+ error
  then

  // y( w:Word -- )  DS still has xt (interpret) or empty under compile-after-&of
  y:.flags@ WordFlags.Defer hasFlag nif
    y:.name@ >ts y:drop
    getCompileStateFlags 0x1 & nif drop then
    "'is': the word was not created by defer: " ts> str::+ error
  then

  getCompileStateFlags 0x1 & if
    y:.dtIdx@
    "literal," find? if
      .xt nip exec
    else
      "'is': literal, missing" error
    then
    [[ _DTRef ! ]]
    y:drop
  else
    y:.dtIdx@ _DTRef !
    y:drop
  then
; immediate

/*
State-smart XT of next word (product is VirtXT).
Interpret: read next identifier, find it, push its VirtXT.
Compile (immediate): same lookup, then literal, (emit XT lit); resume after the word.
( -- xt ) | c( -- +lit xt )
*/
: &of
  readIdentifier dup >ts
  find? if
    .xt nip
    t:drop
    getCompileStateFlags 0x1 & if
      "literal," find? if
        .xt nip exec
      else
        "&of: literal, missing" error
      then
    then
  else
    "&of: word not found: " ts> str::+ error
  then
; immediate

/*
Compile-only immediate: same as compile path of &of (explicit comma form).
c( -- +lit xt )
*/
: &of,
  readIdentifier dup >ts
  find? if
    .xt nip
    t:drop
    "literal," find? if
      .xt nip exec
    else
      "&of,: literal, missing" error
    then
  else
    "&of,: word not found: " ts> str::+ error
  then
; 0x032 setWordFlags // CompileOnly | Immediate | Colon

/*
Compile-time gate for spawn: — Word must be a plain colon entry (not axiom/defined/immediate/hidden/compile-only).
( w:Word -- w:Word )
*/
: assertSpawnSuitable
  // Property @ leaves the receiver — do not dup before each check.
  .isColon@ not if
    .name@ "spawn: not a colon word: " swap str::+ error
  then
  .isDefined@ if
    .name@ "spawn: defined instance not spawnable: " swap str::+ error
  then
  .flags@ WordFlags.Immediate hasFlag if
    .name@ "spawn: immediate word not spawnable: " swap str::+ error
  then
  .flags@ WordFlags.Hidden hasFlag if
    .name@ "spawn: hidden word not spawnable: " swap str::+ error
  then
  .flags@ WordFlags.CompileOnly hasFlag if
    .name@ "spawn: compile-only word not spawnable: " swap str::+ error
  then
;

/*
Compile-only immediate: read next ident, check spawnable, compile XT lit + (spawnPip).
Host runtime: ( pp:AtomicDynamicObject -- )
eg: Pip.newPipParams^ spawn: helloPip
*/
: spawn:
  readIdentifier dup >ts
  find? if
    assertSpawnSuitable
    .xt nip
    t:drop
    "literal," find? if
      .xt nip exec
    else
      "spawn: literal, missing" error
    then
    [[ (spawnPip) ]]
  else
    "spawn: word not found: " ts> str::+ error
  then
; 0x032 setWordFlags // CompileOnly | Immediate | Colon; SR_Threads via viral from (spawnPip)

// Arithmetic helpers

/*
Checks if a number is within a range.
( val:n min:n max:n -- flag:b )
*/
: within
  >ts >ts 
  dup ts> >= swap ts> <= &&
;

/*
Returns the smaller of two numbers.
( a:n b:n -- n )  
*/
: min
  dup2 <= if drop else nip then
;

/*
Returns the larger of two numbers.
( a:n b:n -- n )
*/
: max
  dup2 >= if drop else nip then
;

/*
Returns the square of a number.
The result is the same type as the operand.
( a:n -- n )
*/
: sqr
  dup *
;

/*
See: Float.isClose
( value:n target:n epsilon:n -- b )
*/
: isClose
  Float.isClose^
;

/*
Returns x raised to the power of y. 
( x:n y:n -- f )
*/
: pow
  Float._pow^
;

// Convenience words for DynamicObject

/*
Create a new DynamicObject.
( -- do:DynamicObject )
*/
: newdo
  DynamicObject.new^
;

// Fyne support
false var: _hasFyne

/*
Returns true if Fyne support is available.
( -- flag:b )
*/
: hasFyne
  _hasFyne @
;

/*
Panics if Fyne support is not available.
( -- ) errors
*/
: requiresFyne
  _hasFyne @ nif
    "This app or code requires Fyne support. Run it using final-gui." error
  then
;

// Pi support — profile ints match finalpi PiSupport.None / .PiZ2 / .Pi5
0 var: _piSupport

/*
Returns true if running on a Raspberry Pi.
Note: use _piSupport / PiSupport.* for the actual platform.
( -- flag:b )
*/
: hasPi
  _piSupport @ >0
;

/*
Panics if not running on a Raspberry Pi.
Note: use _piSupport / PiSupport.* for the actual platform.
( -- ) errors
*/
: requiresPi
  hasPi nif
    "This app or code only runs on a Raspberry Pi." error
  then
;


`
