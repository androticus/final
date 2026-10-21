// JSON encode — renderJson / _renderJson family.

/*
The worker for >json and >jsonp.
It uses a combination of Final Native methods, registered
typically on ProtoType objects, along with a contract thus:
* Go-native methods: z( sb:StringBuilder -- sb ), ths.Obj the instance
* Final-native: z( same ) os is the instance
*/

// ( root:a pretty:b -- json:s )
: renderJson
  if
    "  " StringBuilder.newWithPretty^
  else
    StringBuilder.new^
  then
  >zs // _render contract
  >ys
  y:._renderJson^
  z:.toString^
  y:drop
  z:drop
;

// ( root:a -- json:s )
: >json
  false renderJson
;

// ( root:a -- json:s )
: >jsonp
  true renderJson
;

: renderJson_writeItemSepPretty
  z:.pretty if // pretty: nl compact: nothing
    "\n" z:.writeString^
    "" z:.writeStringPretty^
  then
;

// ( beg:s -- )
: renderJson_beginPretty
  z:.writeString^
  z:.nest^
;

// ( end:s -- )
: renderJson_endPretty
  z:.unnest^
  z:.pretty if
    "\n" z:.writeString^
  then
  z:.writeStringPretty^
; 
// --- List-semantic ordinal compound objects, includes all strongly typed lists ---

// z( sb -- sb ) o( inst -- inst )
: List__renderJson_items
  false >ts
  o:peek
  foreach
    t:peek
    if
      "," z:.writeString^
    else
      t:drop true >ts
    then
    renderJson_writeItemSepPretty
    i ._renderJson^
    drop
  /foreach
  t:drop
;

: JsonArray__renderJson
  "[" renderJson_beginPretty
    List__renderJson_items
  "]" renderJson_endPretty
;

/*
Used by the renderer for all list-like objects that
are to become a JSON array.
z( sb:StringBuilder -- sb ) o( inst:List-like )
*/
: ProtoList#_renderJson
  JsonArray__renderJson
;

/*
Used to render Sets as a JSON array.
Note: roundtrip creates a List
z( sb:StringBuilder -- sb ) o( inst:Set-like )
*/
: ProtoSet#_renderJson
  JsonArray__renderJson
;



// --- Map, and all map-semantic objects vis ProtoMap ---

// d( item:KeyValue -- item )
: Map__renderJson_KV
  renderJson_writeItemSepPretty
  "\"" z:.writeString^
  .key z:.writeString^
  "\":" z:.writeString^ z:.writeSpacePretty^
  .value ._renderJson^ drop
;

: Map__renderJson_items
  false >ts // true for item 2+
  o:peek
  foreach
    t:peek
    if
      "," z:.writeString^
    else
      t:drop true >ts
    then
    i Map__renderJson_KV
    drop
  /foreach
  t:drop // flag
;

: JsonObject__renderJson
  "{" renderJson_beginPretty
    Map__renderJson_items
  "}" renderJson_endPretty
;

/*
Used by the renderer for all map-like objects
that are to become a JSON object.
Includes DynamicObject, StaticObject and TypedObject
z( sb:StringBuilder -- sb ) o( inst:Map-like )
*/
: ProtoMap#_renderJson
  JsonObject__renderJson
;


// Note: DynamicObject and TypedObject inherit Map
// rendering from ProtoMap
