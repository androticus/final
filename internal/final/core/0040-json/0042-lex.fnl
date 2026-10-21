// Shared text lexer helpers for wire-format parsers (JSON, YAML, …).
// docs/003-design-notes/124-json-parse-final-fsm.md
// docs/003-design-notes/126-string-helpers.md
//
// Convention: JsonWire (DynamicObject) on y-stack via JsonWire_new / >ys.
// Fields: reader TextReader, pushback Int (-1 = none).
// parseJson also sets .limits (tracker DO) and .lenIdx (active length-budget idx).
// Parse faults use "msg" throw (catchable; uncaught → fault).

// --- JsonWire ---

// ( reader:TextReader -- )
: JsonWire_new
  DynamicObject.new^ >ys
  y:.reader!
  -1 y:.pushback!
;

// ( -- )
: JsonWire_drop
  y:drop
;

// --- pushback rune (-1 = empty) ---

// ( -- b )
: pushbackEmpty?
  "pushback" y:?{} if
    -1 y:.pushback@ ==
  else
    true
  then
;

// ( r:i -- )
: unreadRune
  y:.pushback!
;

// ( -- r:i true | false )
: readRune
  pushbackEmpty? if
    y:.reader@ >ts
    t:.getNextRune^
    t:drop
  else
    y:.pushback@ -1 y:.pushback! true
  then
;

// ( -- true ) | ( r:i false )   true = no rune (eof)
: peekRune
  readRune if
    dup unreadRune false
  else
    true
  then
;

// --- whitespace ---

// ( r:i -- b )
: isJsonWs?
  dup 32 == over 9 == || over 10 == || over 13 == || nip
;

// ( -- )
// Iterative: one Nest frame total (not one per whitespace rune — avoids RS whitespace attacks).
: skipWs
  begin
    readRune if
      dup isJsonWs? if
        drop
        false
      else
        unreadRune
        true
      then
    else
      true
    then
  until
;

: jsonSkipWs skipWs ;

// --- JSON string \uXXXX (must live outside defer: files — string/const literals) ---

// ( r:i -- n:i ok:true ) || ( -- false )
: hexDigitValue
  >ts
  t:peek 48 57 within if
    ts> 48 - true
  else
    t:peek 65 70 within if
      ts> 65 - 10 + true
    else
      t:peek 97 102 within if
        ts> 97 - 10 + true
      else
        t:drop false
      then
    then
  then
;

// ( acc:i -- acc':i )
: jsonParseHexDigit
  readRune not if
    "bad \\u escape" throw
  else
    hexDigitValue
    not if
      drop "bad \\u escape" throw
    then
    16 * +
  then
;

// ( -- r:i )
: jsonParseStringUnicode
  0 jsonParseHexDigit jsonParseHexDigit jsonParseHexDigit jsonParseHexDigit
;

// --- char expectations ---

// ( expected:i -- )
: expectChar
  readRune not if
    "unexpected end of input" throw
  else
    == not if
      "expected character mismatch" throw
    then
  then
;

// ( expected:i -- )
: expectRune
  skipWs
  readRune not if
    "unexpected end of input" throw
  else
    == not if
      "mismatched close delimiter" throw
    then
    drop
  then
;

// readWhile — backlog (YAML line bodies); needs runtime gather word.
