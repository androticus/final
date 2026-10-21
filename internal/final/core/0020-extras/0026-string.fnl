/*
Final string operations
*/

/*
Does the string contain another string.
( s:String sub:String -- yes:b )
*/
: str::contains
  str::indexOf >=0
;