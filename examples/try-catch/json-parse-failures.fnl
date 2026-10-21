forget: reset-json-parse-failures
`
Demonstrates a variety of json> parse failures, and that each throws
(catchable) rather than panicking. Individual fail* words only show the
malformed input and call json> — no try/catch there. The outer demo wraps
each dispatch.

Try: 

json-parse-failures
` const: json-parse-failures-blurb

// --- one dispatch per throw message category (no catch) ---

/*
Unexpected end of input (truncated value).
( -- ) throws:
*/
: failUnexpectedEof
  `
[
  1,
  2
` dup $ nl
  json>
;

/*
Unexpected token at value start.
( -- ) throws:
*/
: failUnexpectedToken
`
@not-json
`
  dup $ nl
  json>
;

/*
Invalid literal (started true/false/null, then diverged).
( -- ) throws:
*/
: failInvalidLiteral
`
tx
`
  dup $ nl
  json>
;

/*
Invalid JSON number (token collected, native reject).
( -- ) throws:
*/
: failInvalidNumber
`
1e
`
  dup $ nl
  json>
;

/*
Bad string escape (not a JSON escape letter).
( -- ) throws:
*/
: failBadEscape
`
"bad \q escape"
` 
  dup $ nl
  json>
;

/*
Bad \u escape (non-hex digits).
( -- ) throws:
*/
: failBadUnicodeEscape
`
"\uGGGG"
`
  dup $ nl
  json>
;

/*
Unterminated string.
( -- ) throws:
*/
: failUnterminatedString
`
"hello
`
  dup $ nl
  json>
;

/*
Expected character mismatch (object key not followed by ':').
( -- ) throws:
*/
: failExpectedColon
`
{
  "a"
  1
}
`
  dup $ nl
  json>
;

/*
Expected ',' or ']' in array.
( -- ) throws:
*/
: failExpectedCommaOrBracket
`
[
  1
  2
]
`
  dup $ nl
  json>
;

/*
Expected ',' or '}' in object.
( -- ) throws:
*/
: failExpectedCommaOrBrace
`
{
  "a": 1
  "b": 2
}
`
  dup $ nl
  json>
;

/*
Object body closed with ']' instead of '}'.
( -- ) throws:
*/
: failObjectClosedWithBracket
`
{
  "a": 1,
  "b": 2
]
`
  dup $ nl
  json>
;

/*
Array body closed with '}' instead of ']'.
( -- ) throws:
*/
: failArrayClosedWithBrace
`
[
  1,
  2
}
`
  dup $ nl
  json>
;

/*
Trailing garbage after a complete value.
( -- ) throws:
*/
: failTrailingGarbage
  `
{
  "ok": true
}
still here
`
  dup $ nl
  json>
;

// --- outer demo: try/catch each category ---

/*
Run all fail* cases; print each caught message.
( -- )
*/
: json-parse-failures
  "--- unexpected eof ---" $ nl
  try
    failUnexpectedEof
  catch
    "CAUGHT: " swap str::+ $ nl
  /try

  "--- unexpected token ---" $ nl
  try
    failUnexpectedToken
  catch
    "CAUGHT: " swap str::+ $ nl
  /try

  "--- invalid literal ---" $ nl
  try
    failInvalidLiteral
  catch
    "CAUGHT: " swap str::+ $ nl
  /try

  "--- invalid number ---" $ nl
  try
    failInvalidNumber
  catch
    "CAUGHT: " swap str::+ $ nl
  /try

  "--- bad escape ---" $ nl
  try
    failBadEscape
  catch
    "CAUGHT: " swap str::+ $ nl
  /try

  "--- bad unicode escape ---" $ nl
  try
    failBadUnicodeEscape
  catch
    "CAUGHT: " swap str::+ $ nl
  /try

  "--- unterminated string ---" $ nl
  try
    failUnterminatedString
  catch
    "CAUGHT: " swap str::+ $ nl
  /try

  "--- expected colon ---" $ nl
  try
    failExpectedColon
  catch
    "CAUGHT: " swap str::+ $ nl
  /try

  "--- expected comma or ] ---" $ nl
  try
    failExpectedCommaOrBracket
  catch
    "CAUGHT: " swap str::+ $ nl
  /try

  "--- expected comma or } ---" $ nl
  try
    failExpectedCommaOrBrace
  catch
    "CAUGHT: " swap str::+ $ nl
  /try

  "--- object closed with ] ---" $ nl
  try
    failObjectClosedWithBracket
  catch
    "CAUGHT: " swap str::+ $ nl
  /try

  "--- array closed with } ---" $ nl
  try
    failArrayClosedWithBrace
  catch
    "CAUGHT: " swap str::+ $ nl
  /try

  "--- trailing garbage ---" $ nl
  try
    failTrailingGarbage
  catch
    "CAUGHT: " swap str::+ $ nl
  /try

  "--- done ---" $ nl
;

json-parse-failures-blurb $ nl
