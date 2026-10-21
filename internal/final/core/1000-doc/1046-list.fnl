/*
List family of user helper words.
These words provide succinct lists of operators/word by category.
The user can then use `help <word>` to get detailed information.
We separate providing the information from printing it, for
future tui etc support.
*/

/*
Render the operations, provided as a newline-separated string value
on the supplied StringBuilder.
Nest and Unnest.
( text:s -- ) z( sb:StringBuilder -- sb )
*/
: renderListText_
  "\n" str::split
  z:.nest^
  foreach
    i z:.writeStringPretty^ 
    "\n" z:.writeString^
  /foreach
  z:.unnest^
;

/*
Make the StringBuilder used by the particular list-* commands.
Set up with pretty and 2 space indent.
( -- ) z( -- sb:StringBuilder )
*/
: makeListSB_
  "  " StringBuilder.newWithPretty^ >zs
;

// ( -- listOps:s )
: listItems_
`
list-types -- basic types, like Int and String, and helpers
list-ops -- arithmetic, bit, bool, comparison, and stack ops
list-string -- string operations
list-math -- transcendentals like sin etc, and related consts
list-control -- control flow words, like if..then, foreach etc.
list-lists -- List types, including strongly typed lists, Table
list-maps -- Map and Set types, and Set operators
list-objects -- DynamicObject, StaticObject, and TypedObject
list-json -- json encoding and decoding
list-datetime -- DateTime, Interval, and helpers
list-streams -- Text/Binary streams, plus File and StringBuilder
list-net -- Network and Http types, and helpers
list-threads -- Pips, Pipes, Events, and atomic helpers
list-db -- Database, and helpers (most are useful outside DB too)
list-filesys -- File, Dir, and Path, plus filesystem iteration
list-vec-mat -- Linear Algebra: vectors, matrices, and ops
list-draw -- Point, Rect, Color, Palettes, Blitting/Drawing
`
hasFyne if
`list-fyne -- Fyne-based Widgets and Layouts
list-simgames -- Simulations and Games
`
  str::cat
then
`list-misc -- catch-all for other words
list-os -- system-level types and operations
list-modules -- Projects and dynamically loadable modules
words -- list all User/Session words, in order of definition
`
str::cat
;

// ( -- ops:s )
: list-typesItems_
`
types: Null Bool Int Uint Float String
probe: Null? Bool? Int? Uint? Float? String?
repr coerce: >Bool? >Int? >Uint? >Float? >Bool >Int >Uint >Float
value coerce: >>Null? >>Bool? >>Int? >>Uint? >>Float? >>String?
    >>Null >>Bool >>Int >>Uint >>Float >>String
`
;

// ( -- ops:s )
: list-opsItems_
`
arithmetic: + - * / abs % modrem min max within
bitwise: & | xor bitnot bitset bitclear bittoggle
boolean: && || not
compare: == <> < <= > >= ==0 <>0 <0 <=0 >0 >=0
shift/rot: << >> rotl rotr
stack: dup ?dup dup2 drop drop2 swap over nip tuck pick rot
stack: depth peek
move: >ts ts> tdrop >xs xs> xdrop >ys ys> ydrop >zs zs> zdrop
      ds> >ds
`
;
  
// ( -- ops:s )
: list-stringItems_
`
str::len str::isEmpty
str::trimStart str::trimEnd str::trim
str::toLower str::toUpper
str::cat str::+ str::sub str::contains
str::startsWith str::endsWith str::indexOf
str::split str::splitLines str::join
`
;
  
  // ( -- ops:s )
: list-mathItems_
`
sin cos tan asin acos atan
sinh cosh tanh asinh acosh atanh
sqrt cbrt exp log log2 log10
ceil floor round trunc
Math
`
;

// ( -- ops:s )
: list-controlItems_
`
if nif else then
begin again until while repeat
forever /forever break continue
foreach /foreach i j k range rangeby
i j k (iter item in foreach, inner nested to outer)
switch case /case default /default /switch
a b c (probe value in switch, inner nested to outer)
try catch finally /try throw
l (word-local storage)
`
;

// ( -- ops:s )
: list-listsItems_
`
List StringList StringKeyValueList
ByteList Uint16List Uint32List UintList
Int8List Int16List Int32List IntList
Float32List FloatList
Table
`
;

// ( -- ops:s )
: list-mapsItems_
`
Map StringMap KeyValue StringKeyValueList
Set StringSet
set ops: set::union set::intersect
    set::difference set::symmetricDifference
`
;

// ( -- ops:s )
: list-objectsItems_
`
DynamicObject, clones via .clone
StaticObject, via DynamicObject.newStatic
TypedObject, add methods via:
//  o( inst ) OS has instance
: MyTypeName#myMethodName ... ;
// o( type )  OS has your Type
: MyTypeName##myMethodName ... ;
`
;

// ( -- ops:s )
: list-jsonItems_
`
>json >jsonp json> parseJson renderJson
`
;

// ( -- ops:s )
: list-datetimeItems_
`
DateTime Interval
dt::+ dt::- dt::* dt::/ dt::compare dt::round
dt::< dt::<= dt::> dt::>= dt::== dt::<>
dt::fmt
`
;

// ( -- ops:s )
: list-streamsItems_
`
StreamReadWrite (a parent interface Type)
TextWrite ByteWrite TextRead ByteRead (interface Types)
TextReader TextWriter ByteReader ByteWriter
StringBuilder File
`
;

// ( -- ops:s )
: list-netItems_
`
HttpAddress HttpReq HttpDialer HttpRes
httpDial printHttpRes
`
;

// ( -- ops:s )
: list-threadsItems_
`
Pip Pipe spawn: (spawnPip)
pipe::new pipe::cap pipe::len >pipe pipe> pipe>?
pipe::close pipe::closed?
pipe::tickerStart pipe::tickerStop
pipe::timerAfter pipe::timerWhen
Pipe.newTicker Pipe.newTimer
AtomicList AtomicStringList AtomicDynamicObject
AtomicInt Average pls
Event EventLoop EventTicker EventTimer
`
;

// ( -- ops:s )
: list-dbItems_
`
DbProvider
val? val|def iter>list
`
;

: list-filesysItems_
`
File Dir Path
File system iteration via Dir.iter
`
;

// ( -- ops:s )
: list-vec-mat_
`
VecNf
vec::dot vec::norm vec::+ vec::- vec::scale
vec::copy

Mat2f
mat::matvec mat::* mat::transpose mat::solve
mat::+ mat::- mat::scale 
mat::row mat::col mat::identity mat::copy

Mat2i1 Mat2i4 
Mat2i8 Mat2i16 Mat2i32 Mat2i (Int)
Mat2u8 Mat2u16 Mat2u32 Mat2u (Uint)
`
;

// ( -- ops:s )
: list-drawItems_
`
Point Size Rect Color
Palette1 Palette4 Palette8 Mat2c
BlitType BlitParams BlitSrc
Pixmap Painter
Font Stroke Fill
LineCap LineStyle FillStyle
PathBuilder
`
;
 
// ( -- ops:s )
: list-fyneItems_
`
App Prefs Window Icon Keys
EventLoop Event KeyEvent MouseEvent
ScrollEvent

Containers and Layouts:
AppTabs Border Box Grid Split Scroll
Accordion Separator FormLayout 

Widgets:
Button CanvasImage Check Entry Form Label
MenuItem Menu MenuBar RadioGroup RichText 
Select Slider Dialog ColorDialog FileDialog
ConfirmationDialog InformationDialog 

Aggregate widgets and helpers:
Data       Holder   Widget
----       ------   ------
StringList ListRef  ListView
Table      TableRef TableView
FyneTree   TreeRef  TreeView
`
;

: list-simgameItems_
`
EventLoop Event
Boundary SimGameOptions SimClock SimMHEngine
SimGame SimEntity CircleHitArea RectHitArea
SimHitEvent SimHit SimCrossingEvent SimCrossing
BlitDrawer AnimBlitDrawer RadialAnimBlitDrawer
`
;

// ( -- ops:s )
: list-miscItems_
`
Rand Color Range RangeBy
stop isTyped
Regex
output: fmt $ (stdout) $$ (stderr) r>s (rune to string)
&of &of, (xt of next word, interpret/compile)
`
;

: list-osItems_
`
Dialog (final-tui only)
Moniker Env
Session Logger LogProvider
Sys
`
;

: list-moduleItems_
`
install import
using using-default
loadProject  
`
;

/*
Print the list of list-* operations
( ops:s tagline:s -- )
*/
: doListItems_
  makeListSB_
  z:.writeString^ // tagline
  renderListText_ // ops
  z:.toString^
  $
  z:drop
;

/*
Print the list of list-* operations
*/
: list
  listItems_ 
`Run one of the following supported list commands, then
type 'help <word>' for detailed help on the word or type.
`
  doListItems_
;

: list-types
  list-typesItems_
  "Basic types, type probe and cast operators.\n"
  doListItems_
;

: list-ops
  list-opsItems_
  "Basic arithmetic, bitwise, bool, compare, and stack operations.\n"
  doListItems_
;

: list-string
  list-stringItems_
  "String operators.\n"
  doListItems_
;

: list-math
  list-mathItems_
  "Transcendentals and Math constants.\n"
  doListItems_
;

: list-control
  list-controlItems_
  "Control flow: conditionals, loops, foreach, switch.\n"
  doListItems_
;

: list-lists
  list-listsItems_
  "List and strongly typed list types.\n"
  doListItems_
;

: list-maps
  list-mapsItems_
  "Map and Set types and related helpers.\n"
  doListItems_
;

: list-objects
  list-objectsItems_
  "DynamicObject and TypedObject.\n"
  doListItems_
;

: list-json
  list-jsonItems_
  "JSON encoding and decoding.\n"
  doListItems_
;

: list-datetime
  list-datetimeItems_
  "DateTime, Interval, and dt::* helpers.\n"
  doListItems_
;

: list-streams
  list-streamsItems_
  "Streams, StringBuilder, and File helpers.\n"
  doListItems_
;

: list-net
  list-netItems_
  "HTTP types and helpers.\n"
  doListItems_
;

: list-threads
  list-threadsItems_
  "Pips, Pipes, Events, and atomic helpers.\n"
  doListItems_
;

: list-db
  list-dbItems_
  "Database Provider and helpers.\n"
  doListItems_
;

: list-filesys
  list-filesysItems_
  "File, Dir and Path.\n"
  doListItems_
;

: list-vec-mat
  list-vec-mat_
  "Linear Algebra: vectors, matrices, and ops.\n"
  doListItems_
;

: list-draw
  list-drawItems_
  "Blitting and Drawing.\n"
  doListItems_
;

: list-fyne
  list-fyneItems_
  "Fyne-base GUI classes.\n"
  doListItems_
;

: list-simgame
  list-simgameItems_
  "Simulation and Game support."
  doListItems_
;

: list-misc
  list-miscItems_
  "Miscellaneous words and types.\n"
  doListItems_
;
: list-os
  list-osItems_
  "OS and System words and types.\n"
  doListItems_
;

: list-modules
  list-moduleItems_
  "Projects and dynamically loadable modules.\n"
  doListItems_
;
