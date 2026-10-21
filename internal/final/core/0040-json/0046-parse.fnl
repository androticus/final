// JSON parse — recursive descent over TextReader / JsonWire.
// docs/003-design-notes/124-json-parse-final-fsm.md
// docs/003-design-notes/119-json.md
//
// Stack roles: Y = JsonWire; Z = container under build; X = StringBuilder scratch; DS = values.
// Public entry: `{"a":1}` json>
//
// Go natives: TextReader.newFromString, String.parseJsonNumber (119 number policy)

// --- numeric token (119 number policy) ---

// forward references must use a defer: (sole cycle: object/array members → value → object/array)
defer: jsonParseValue

// ( r:i -- b )
: isDigit?
  48 57 within
;

// ( r:i -- b )  consumes r; true when r starts a JSON number token.
: isNumberChar?
  dup isDigit? if drop true
  else dup '.' == if drop true
  else dup 'e' == if drop true
  else dup 'E' == if drop true
  else dup '+' == if drop true
  else dup '-' == if drop true
  else drop false
  then then then then then then
;

// ( token:s -- n:a )
: parseJsonNumber
  dup >xs
  x:.parseJsonNumber^ if
    nip x:drop
  else
    x:drop "invalid JSON number" throw
  then
;

// ( -- )  x( sb -- sb ) StringBuilder on x-stack
// Iterative: one Nest frame total (not one per digit).
: jsonParseNumberMore
  begin
    peekRune if
      true
    else
      dup isNumberChar? if
        drop readRune drop
        x:.writeRune^
        y:.lenIdx@ incCheckLimit
        false
      else
        drop
        true
      then
    then
  until
;

// ( first:i -- n:a )
: jsonParseNumber
  jsonParse_maxNumber_limit_idx dup y:.lenIdx! zeroLimit
  StringBuilder.new^ >xs
  x:.writeRune^
  y:.lenIdx@ incCheckLimit
  jsonParseNumberMore
  x:.toString^
  x:drop
  parseJsonNumber
;

// --- string ---

// ( -- r:i )
: jsonParseStringEscape
  readRune not if
    "bad escape in string" throw
  else
    switch
      a '"' == case
        '"'
      /case
      a '\\' == case
        '\\'
      /case
      a '/' == case
        '/'
      /case
      a 'b' == case
        8
      /case
      a 'f' == case
        12
      /case
      a 'n' == case
        10
      /case
      a 'r' == case
        13
      /case
      a 't' == case
        9
      /case
      a 'u' == case
        jsonParseStringUnicode
      /case
    default
      "bad escape in string" throw
    /default
    /switch
  then
;

// ( -- )  x( sb -- sb ) StringBuilder on x-stack; leaves string on DS when done
// Iterative: one Nest frame total (not one per character).
// Closing '"': drop closer and the opening quote left under (key path / value arm pushes it).
: jsonParseStringMore
  begin
    readRune not if
      "unterminated string" throw
    else
      dup '"' == if
        drop drop x:.toString^ x:drop
        true
      else
        dup '\\' == if
          drop jsonParseStringEscape x:.writeRune^
          y:.lenIdx@ incCheckLimit
          false
        else
          x:.writeRune^
          y:.lenIdx@ incCheckLimit
          false
        then
      then
    then
  until
;

// ( idx:i -- )  StringBuilder on x-stack; idx is the string or ident length budget
: jsonParseStringBody
  dup y:.lenIdx! zeroLimit
  StringBuilder.new^ >xs
  jsonParseStringMore
;

// ( -- s )  opening " must be next in stream (or in pushback)
: jsonParseString
  readRune not if
    "unterminated string" throw
  else
    dup '"' == not if
      drop "expected string" throw
    else
      drop jsonParse_maxString_limit_idx jsonParseStringBody
    then
  then
;

// ( opening:i -- s )  opening " already read onto stack
: jsonParseStringOpen
  dup '"' == not if
    drop "expected string" throw
  else
    drop jsonParse_maxString_limit_idx jsonParseStringBody
  then
;

// --- literals ---

// ( kw:s -- )
: jsonEatKeyword
  foreach
    readRune not if
      "unexpected end of input" throw
    else
      i == not if
        "invalid literal" throw
      then
    then
  /foreach
;

// ( -- true:Bool )
: jsonParseTrue
  "rue" jsonEatKeyword
  true
;

// ( -- false:Bool )
: jsonParseFalse
  "alse" jsonEatKeyword
  false
;

// ( -- null:Null )
: jsonParseNull
  "ull" jsonEatKeyword
  null
;

// ( first:i -- v:a )
: jsonParseLiteral
  dup 't' == if
    drop jsonParseTrue
  else
    dup 'f' == if
      drop jsonParseFalse
    else
      dup 'n' == if
        drop jsonParseNull
      else
        drop "invalid literal" throw
      then
    then
  then
;

// --- compound values (Z-stack, begin/until loops) ---

// ( -- list:List )
: jsonParseArray
  jsonParse_maxNesting_limit_idx incCheckLimit
  List.new^ >zs
  skipWs
  begin
    peekRune if
      "unexpected end of input" throw
    else
      dup ']' == if
        drop readRune drop drop
        true
      else
        drop
        jsonParseValue
        z:.append^
        skipWs
        peekRune if
          "unexpected end of input" throw
        else
          dup ',' == if
            drop readRune drop drop
            false
          else
            dup ']' == if
              drop readRune drop drop
              true
            else
              drop "expected ',' or ']'" throw
            then
          then
        then
      then
    then
  until
  jsonParse_maxNesting_limit_idx decLimit
  zs>
;

// ( -- obj:DynamicObject )
: jsonParseObject
  jsonParse_maxNesting_limit_idx incCheckLimit
  DynamicObject.new^ >zs
  skipWs
  begin
    peekRune if
      "unexpected end of input" throw
    else
      dup '}' == if
        drop readRune drop drop
        true
      else
        drop
        skipWs
        readRune drop jsonParse_maxIdent_limit_idx jsonParseStringBody
        skipWs
        ':' expectChar
        jsonParseValue
        swap z:!{}
        skipWs
        peekRune if
          "unexpected end of input" throw
        else
          dup ',' == if
            drop readRune drop drop
            false
          else
            dup '}' == if
              drop readRune drop drop
              true
            else
              drop "expected ',' or '}'" throw
            then
          then
        then
      then
    then
  until
  jsonParse_maxNesting_limit_idx decLimit
  zs>
;

// --- dispatch ---

// ( -- v:a )
: (jsonParseValue)
  skipWs
  readRune not if
    "unexpected end of input" throw
  then
  jsonParse_maxNodes_limit_idx incCheckLimit
  switch
    a '"' == case
      a jsonParse_maxString_limit_idx jsonParseStringBody
    /case
    a '{' == case
      jsonParseObject
    /case
    a '[' == case
      jsonParseArray
    /case
    a 't' == case
      jsonParseTrue
    /case
    a 'f' == case
      jsonParseFalse
    /case
    a 'n' == case
      jsonParseNull
    /case
    a isNumberChar? case
      a jsonParseNumber
    /case
    default
      "unexpected token" throw
    /default
  /switch
;

&of (jsonParseValue) is jsonParseValue

// --- public codec entry (mirrors render.fnl >json) ---

// ( json:s -- root:a ) throws: on parse failure
: parseJson
  dup str::len >ts
  TextReader.newFromString^ not if
    "TextReader.newFromString failed" throw
  then
  JsonWire_new
  jsonParseLimits_new y:.limits!
  -1 y:.lenIdx!
  jsonParse_maxLength_limit_idx ts> setCheckLimit
  jsonParseValue
  skipWs
  readRune if
    drop "trailing garbage after JSON value" throw
  then
  JsonWire_drop
;

/*
Parse the json string into an object.
If the json is just a primitive, return that value.
Json objects render as a Final DynamicObject.
Json arrays render as a Final List.
( json:s -- result:a ) throws: on parse failure
*/
: json>
  parseJson
;
