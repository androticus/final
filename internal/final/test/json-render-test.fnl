forget: reset-json-render-test
`
JSON encode integration tests — primitives, List, DynamicObject, Map, nesting.

Manual (from repo root):

Try:
json-render-tests
` const: json_render_test-blurb

: jsonPrimitives
  nl
  null >json $ sp
  true >json $ sp
  -49 >json $ sp
  0xFFu >json $ sp
  3.14526 >json $ sp
  "the rain\nfalls in the plane\n" >json $
  nl 
;

: makeList_primitives
  ( null true -49 0xFFu 3.14 "the rain\nfalls in the plane\n" )
;

: makeMaplike_primitives
    null "Null"
    true "Bool"
    -49 "Int"
    0xFFu "Uint"
    3.14159256 "Float"
    "the rain\nfalls in the plane\n" "String"
;

: makeDO_primitives
  {
    makeMaplike_primitives
  }
;

: makeMap_primitives
  { 
    makeMaplike_primitives
  sm}
;

: makeList_nesting
  (
    true
    -49
    makeList_primitives
    makeDO_primitives
    makeMap_primitives
  )
;

: makeMap_nesting
  {
    makeDO_primitives "do_primitives"
    makeList_primitives "list_primitives"
    makeMap_primitives "map_primitives"
    null "Null"
    -49 "Int"
  sm}
;

: testJsonPrimitives
  jsonPrimitives
;

: testJsonListCompact
  makeList_primitives >json $
;

: testJsonListPretty
  makeList_primitives >jsonp $
;

: testDOCompact
  makeDO_primitives >json $
;

: testDOPretty
  makeDO_primitives >jsonp $
;

: testListNestingCompact
  makeList_nesting >json $
;

: testListNestingPretty
  makeList_nesting >jsonp $
;

: testMapNestingCompact
  makeMap_nesting >json $
;

: testMapNestingPretty
  makeMap_nesting >jsonp $
;

: json-render-tests
  nl testJsonPrimitives
  nl testJsonListCompact
  nl testJsonListPretty
  nl testDOCompact
  nl testDOPretty
  nl testListNestingCompact
  nl testListNestingPretty
  nl testMapNestingCompact
  nl testMapNestingPretty
;

json_render_test-blurb $ nl
