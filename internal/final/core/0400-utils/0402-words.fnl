/*
Final Word listing util.
*/

/*
List all the words in the User/Session dictionary, 3 per line.
Words are listed in the order of creation.
Can be run even if the Dictionary is locked or sealed.
( -- )
*/
: words
  AtomicInt.new^ l.col! // 3 cols
  dc.dict.getUserWordsIter^
  foreach   
    i >ts ( t:.name@ ) t:drop
    "%-24s " fmt $
    l.col.inc^ 2 >
    if
      nl
      l.col.zero^
    then
  /foreach
  nl
;
