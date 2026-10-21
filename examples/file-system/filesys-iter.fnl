forget: reset

/*
Demonstration of File system iteration.

WARNING: make sure the root you choose has a small set of files
underneath it, or you will be swamped! Safe to iter . on the
release version of Final.
*/


/*
Print the filename of every file visited under root.
Root should be a directory, ex . for current.
( root:s -- )
*/
: filesysIter
  Dir.new^ >ts
  Dir.IterAll t:.iter^
  foreach
    i >ts
      t:.isDir if ">" else " " then
     ( peek t:.size@ t:.modTime@ t:.name@ t:.path@ )
     "%s %9d %v %s %s\n" fmt $
     drop
    t:drop
  /foreach
  t:drop
;
