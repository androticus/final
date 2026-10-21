/*
Compile a MODULE's files into the current Session.
Assumes the moduleJson is a vetted MODULE file.
Assumes that all required modules are installed.
( rootPath:s moduleJson:s -- )
*/
: compileModule
  json>
  >ts
  >xs
  t:.files@
  foreach
    x:peek i Path.join^
    dup
    compileFromFile
    "Compiled " $
    $
    ".\n" $
  /foreach
  "zzCMOD_DONE_zz\n" $
  x:drop
  t:drop
;
