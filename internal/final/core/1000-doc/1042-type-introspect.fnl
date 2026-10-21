/*
Code to enumerate type members for user help type listing.
*/

false var: _skipSystemHelpMembers

/*
Return true if we should skip the member, typically a system member.
( member:Property|MethodMetadata -- skip:b )
*/
: skipMemberDoc?
  _skipSystemHelpMembers @ if
    >ts
      t:.name@ "_" str::startsWith
    tdrop
  else
    drop
    false 
  then
;

/*
Write the doc comments, if any.
We split the doc into lines for pretty printing.
( mm:Prop|Items|Meth*Metadata -- ) z( sb:StringBuilder -- sb )
*/
: writeMemberDoc
  >ts
    t:.doc@ dup str::len 0 >
    if
      "\n" str::split
      foreach
        i z:.writeStringPretty^ "\n" z:.writeString^ 
      /foreach
    else
      drop
    then
  tdrop
;

/*
Return a descriptor string: get/set or readonly.
( pm:PropertyMetadata -- s )
*/
: (rwPropString)
  >ts
  t:.hasSet if "get/set" else "readonly" then
  tdrop
;

/*
The suffix to append to the member for rendered mixin
on the base type.
( mixin:Type -- suf:s )
*/
: mixinMemberSuffix
  >ts
  ( t:.name@ )
  " (From %s)" fmt
  tdrop
;

/*
Write the properties to sb, respecting pretty.
Note: pmm may only be interated, not introspected.
Props that come from a Mixin are not printed unless they
have an overriden .doc, and we add (from Type).
( pmm:PropertyMetadataMap -- ) z( sb:StringBuilder -- sb )
*/
: ((writeProps))
  foreach
    i .value@ nip dup
    skipMemberDoc? if
      drop continue
    then
    >ts
    // mixin handling
    t:.tryGetMixinType^ // ( -- mixin:Type true | false )
    if
      >xs
      t:.doc@ str::isEmpty // don't print if so
      if xdrop tdrop continue then
      xs> mixinMemberSuffix
    else
      ""
    then
    >xs // optional mixin suffix; under prop string

    t:peek dup writeMemberDoc
    (rwPropString) >xs

    ( t:.name t:.type.name xs> xs> )
    ".%s (%s) %s%s\n\n" fmt z:.writeStringPretty^
    tdrop
  /foreach
;

/*
Write "from <typename>:" and props when this type contributes metadata.
( t:Type -- ) y( propAccessor:Word -- pa ) z( sb:StringBuilder -- sb )
*/
: (writeProps)
  dup >ts
  y:.exec^ // ( map n | 0 )
  if
    >xs
    ( t:.name )
    "from %s:\n" fmt z:.writeStringPretty^
    z:.nest^
      xs> ((writeProps))
    z:.unnest^
  then
  tdrop
;

/*
Recusively write all the properties, up the parent chain.
Write: Type, Mixins, then Parent chain.
( t:Type -- ) y( propAccessor:Word -- pa ) z( sb:StringBuilder -- sb )
*/
: writeProps
  dup 
  >ts 
    (writeProps)
    t:.tryGetMixinTypes^ if // ( mixins:List -- )
      foreach
        i (writeProps)
      /foreach
    then
    t:.getParentType^ if recurse then
  tdrop
;

// Visitor words, to get either class or instance metadata

// ( t:Type -- map n | 0 )
: getPropertyMetadataMap
  >ts
    t:.tryGetPropertyMetadata^
  tdrop
;

// ( t:Type -- map n | 0 )
: getClassPropertyMetadataMap
  >ts
    t:.tryGetClassPropertyMetadata^
  tdrop
;


// Compounds

/*
Write the compounds to sb, respecting pretty.
There will always be at least one, since metadata exists
( im:ItemsMetadata -- ) z( sb:StringBuilder -- sb )
*/
: ((writeCompounds))
  dup writeMemberDoc
  >ts
  "" dup z:.writeStringPretty^ // for indent
  t:.hasGetByOrdinal if "@() " str::+ then
  t:.hasSetByOrdinal if "!() " str::+ then
  t:.hasGetByKey if "@{} " str::+ then
  t:.hasSetByKey if "!{} " str::+ then
  t:.hasByKey if "?{} " str::+ then
  t:.hasGetByIndices if "@[] " str::+ then
  t:.hasSetByIndices if "![]" str::+ then
  "\n\n" str::+
  z:.writeString^
  tdrop
;

/*
Write "from <typename>:" and compounds when this type contributes metadata.
( t:Type -- ) z( sb:StringBuilder -- sb )
*/
: (writeCompounds)
  >ts
  t:.tryGetItemsMetadata^
  if
    >xs
    ( t:.name )
    "from %s:\n" fmt z:.writeStringPretty^
    z:.nest^
      xs> ((writeCompounds))
    z:.unnest^
  then
  tdrop
;

/*
Recusively write all the compounds, up the parent chain
( t:Type -- ) z( sb:StringBuilder -- sb )
*/
: writeCompounds
  dup 
  >ts 
    (writeCompounds)
    t:.tryGetMixinTypes^ if // ( mixins:List -- )
      foreach
        i (writeCompounds)
      /foreach
    then
    t:.getParentType^ if recurse then
    t:.getParentType^ if recurse then
  tdrop
;


// Methods

/*
Return a descriptor string: none or (Final).
( mm:MethodMetadata -- s )
*/
: (methodTypeString)
  >ts
    t:.isFinalMethod
    if " (Final)" else "" then
  tdrop
;

/*
Write the methods to sb, respecting pretty.
Note: pmm may only be interated, not introspected.
( pmm:MethodMetadataMap -- ) z( sb:StringBuilder -- sb )
*/
: ((writeMethods))
  foreach
    i .value@ nip dup
    skipMemberDoc? if
      drop continue
    then
    >ts
    // mixin handling
    t:.tryGetMixinType^ // ( -- mixin:Type true | false )
    if
      >xs
      t:.doc@ str::isEmpty // don't print if so
      if xdrop tdrop continue then
      xs> mixinMemberSuffix
    else
      ""
    then
    >xs // optional mixin suffix

    t:peek dup writeMemberDoc
    (methodTypeString) >xs
    ( t:.name xs> xs> )
    ".%s%s%s\n\n" fmt z:.writeStringPretty^
    tdrop
  /foreach
;

/*
Write "from <typename>:" and methods when this type contributes metadata.
( t:Type -- ) y( visitor:Word -- map n | 0 ) z( sb:StringBuilder -- sb )
*/
: (writeMethods)
  dup >ts
  y:.exec^ // ( -- map n | 0 )
  if
    >xs
    ( t:.name )
    "from %s:\n" fmt z:.writeStringPretty^
    z:.nest^
      xs> ((writeMethods))
    z:.unnest^
  then
  tdrop
;

/*
Recusively write all the properties, up the parent chain
( t:Type -- ) z( sb:StringBuilder -- sb )
*/
: writeMethods
  dup 
  >ts 
    (writeMethods)
    t:.tryGetMixinTypes^ if // ( mixins:List -- )
      foreach
        i (writeMethods)
      /foreach
    then
    t:.getParentType^ if recurse then
  tdrop
;

// Visitor words, to get either class or instance metadata

// ( t:Type -- map n | 0 )
: getMethodMetadataMap
  >ts
    t:.tryGetMethodMetadata^
  tdrop
;

// ( t:Type -- map n | 0 )
: getClassMethodMetadataMap
  >ts
    t:.tryGetClassMethodMetadata^
  tdrop
;

// Type

/*
Write all the members for a type, including parent members.
( t:Type -- )  z( sb:StringBuilder -- sb )
*/
: writeMembers
  >ts
    "Properties:\n" z:.writeStringPretty^
    t:peek
    z:.nest^
      "getPropertyMetadataMap" find >ys
      writeProps
      ydrop
    z:.unnest^
    
    "Methods:\n" z:.writeStringPretty^
    t:peek
    z:.nest^
      "getMethodMetadataMap" find >ys
      writeMethods
      ydrop
    z:.unnest^
    
    "Compound Accessors:\n" z:.writeStringPretty^
    t:peek
    z:.nest^
      writeCompounds
    z:.unnest^
    
    "Class Properties:\n" z:.writeStringPretty^
    t:peek
    z:.nest^
      "getClassPropertyMetadataMap" find >ys
      writeProps
      ydrop
    z:.unnest^
    
    "Class Methods:\n" z:.writeStringPretty^
    t:peek
    z:.nest^
      "getClassMethodMetadataMap" find >ys
      writeMethods
      ydrop
    z:.unnest^

  tdrop
;

/*
Print all the members for a type, including parent members.
( t:Type -- )
*/
: members$
  "  " StringBuilder.newWithPretty^
  >zs
    writeMembers
    z:.toString^ $
  z:drop
;


