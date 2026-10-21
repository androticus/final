// JSON parse budgets — ordinal indexes, cap table, per-parse tracker factory.
// docs/003-design-notes/200-system-hardening.md

0 const: jsonParse_maxLength_limit_idx
1 const: jsonParse_maxNesting_limit_idx
2 const: jsonParse_maxNodes_limit_idx
3 const: jsonParse_maxIdent_limit_idx
4 const: jsonParse_maxString_limit_idx
5 const: jsonParse_maxNumber_limit_idx

(
  0x800000 // 8 MB
  64
  1000000
  2048
  0x100000 // 1 MB
  64
i) const: jsonParse_limits

(
  "JSON parse limit exceeded: input length (limit %d)"
  "JSON parse limit exceeded: nesting depth (limit %d)"
  "JSON parse limit exceeded: node count (limit %d)"
  "JSON parse limit exceeded: identifier length (limit %d)"
  "JSON parse limit exceeded: string length (limit %d)"
  "JSON parse limit exceeded: number token length (limit %d)"
s) const: jsonParse_limitFmts

// ( -- limits:DynamicObject )
// Fields in idx order so idx @() is the AtomicInt for that budget.
: jsonParseLimits_new
  {
    AtomicInt.new^ "inputLen"
    AtomicInt.new^ "nesting"
    AtomicInt.new^ "nodes"
    AtomicInt.new^ "identLen"
    AtomicInt.new^ "stringLen"
    AtomicInt.new^ "numberLen"
  }
;

// Tracker lives on JsonWire (ys) as .limits.
// idx is parked on t:. Collection goes to t: as well so t:@() / t:.inc^ have a CS receiver.

// ( collection -- value )  idx on t:
: jsonParse_atIdx
  t:peek swap >ts t:@()
  t:drop
;

// ( idx:i -- ) throws: when the incremented counter exceeds jsonParse_limits[idx]
: incCheckLimit
  >ts
  y:.limits@ jsonParse_atIdx >ts
  t:.inc^
  t:drop
  jsonParse_limits jsonParse_atIdx >
  if
    ( jsonParse_limits jsonParse_atIdx )
    jsonParse_limitFmts jsonParse_atIdx
    fmt throw
  then
  t:drop
;

// ( idx:i val:i -- ) throws: when val exceeds jsonParse_limits[idx]
: setCheckLimit
  swap >ts
  dup y:.limits@ jsonParse_atIdx >ts
  t:.set^
  t:drop
  jsonParse_limits jsonParse_atIdx >
  if
    ( jsonParse_limits jsonParse_atIdx )
    jsonParse_limitFmts jsonParse_atIdx
    fmt throw
  then
  t:drop
;

// ( idx:i -- )
: decLimit
  >ts
  y:.limits@ jsonParse_atIdx >ts
  t:.dec^ drop
  t:drop
  t:drop
;

// ( idx:i -- )
: zeroLimit
  >ts
  y:.limits@ jsonParse_atIdx >ts
  t:.zero^
  t:drop
  t:drop
;
