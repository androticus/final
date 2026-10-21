/*
Final extensions for ProtoVec
*/

/*
Vecs render JSON as an array, like Lists
NOTE: there is no corresponding json> operation.
o( inst )
*/
: ProtoVec#_renderJson
  JsonArray__renderJson
;

// note: rows are compact "[a, b, c]"; pretty only between rows
// z( sb -- sb ) o( inst:Mat )
: Mat__renderJson_rows
  false >ts
  o:.getRowIter^
  foreach
    t:peek
    if
      "," z:.writeString^
    else
      t:drop true >ts
    then
    renderJson_writeItemSepPretty
    false >xs
    "[" z:.writeString^
    o:.getColIter^
    foreach
      x:peek
      if
        ", " z:.writeString^
      else
        x:drop true >xs
      then
      j i o:@[] ._toString^ nip z:.writeString^
    /foreach
    x:drop
    "]" z:.writeString^
  /foreach
  t:drop
;

/*
Mats render JSON as an array of arrays, for easy reading.
NOTE: there is no corresponding json> operation.
z( sb:StringBuilder ) o( inst:Mat )
*/
: Mat__renderJson
  "[" renderJson_beginPretty
    Mat__renderJson_rows
  "]" renderJson_endPretty
;

: ProtoMat#_renderJson
  Mat__renderJson 
;