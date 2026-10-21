forget: reset-sets
`
Demonstrates how to fill a set, and perform common set operations.
See help for additional operators:
set::difference
set::symmetricDifference

Try: brady-union lang-friends

` const: set-demo-blurb

( "Greg" "Peter" "Bobby" sset)
const: mike-kids

( "Marcia" "Jan" "Cindy" sset)
const: carol-kids

: brady-union
  "Mike kids: " $ mike-kids >json $ nl
  "Carol kids: " $ carol-kids >json $ nl
  "All Brady kids via set::union : " $
  mike-kids carol-kids set::union
  >json $ nl
;


( "Foo" "Bar" sset)
const: zig-friends

( "Bar" "Baz" sset)
const: rust-friends

: lang-friends
  "Ziggy's friends: " $ zig-friends >json $ nl
  "Rusty's friends: " $ rust-friends >json $ nl
  "Common friends via set::intersect : " $
  zig-friends rust-friends set::intersect
  >json $ nl
;


set-demo-blurb $
