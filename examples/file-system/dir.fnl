forget: reset

/*
Dir helpers — new, cwd, setCwd, exists?, info, mkdir, string, iter.

Mutators / iter use _tmp/dir-ex/ (created as needed). setCwd saves and
restores the process working directory so the session is left unchanged.

Run (from repo or Final/ install root):
  "examples/file-system/dir.fnl" compileFromFile
*/


/*
Build a Dir from a path string (directory root value, not an OS open handle).
*/
: new
  "new" $ nl
  "  params: \".\"" $ nl
  "." Dir.new^ >ts
  "  result: " $ t:.string@ $ nl
  t:drop
  nl
;

/*
Process current working directory as a Dir.
*/
: cwd
  "cwd" $ nl
  "  params: (none)" $ nl
  Dir.cwd^ >ts
  "  result: " $ t:.string@ $ nl
  t:drop
  nl
;

/*
Change process cwd (modal). This example restores the previous cwd afterward.
*/
: setCwd
  "setCwd" $ nl
  Dir.cwd^ .string@ >ts
  "_tmp/dir-ex" Dir.mkdir^ if else "  mkdir: " $ $ nl then
  "  params: \"_tmp/dir-ex\"" $ nl
  "_tmp/dir-ex" Dir.setCwd^
  if
    "  result: true" $ nl
    "  cwd now: " $ Dir.cwd^ .string@ $ nl
  else
    "  result: false err=" $ $ nl
  then
  t:peek Dir.setCwd^ if else "  restore cwd: " $ $ nl then
  t:drop
  nl
;

/*
Whether a path exists (missing → false).
*/
: exists?
  "exists?" $ nl
  "  params: \".\"" $ nl
  "." Dir.exists?^
  "  result: " $ $ nl
  nl
;

/*
Metadata snapshot for a path (modal: fi true | err false).
*/
: info
  "info" $ nl
  "  params: \".\"" $ nl
  "." Dir.info^
  if
    >ts
    "  result: name=" $ t:.name@ $ " isDir=" $ t:.isDir@ $ nl
    t:drop
  else
    "  err: " $ $ nl
  then
  nl
;

/*
Create a directory and any missing parents (modal; idempotent if already a dir).
*/
: mkdir
  "mkdir" $ nl
  "  params: \"_tmp/dir-ex/nested/leaf\"" $ nl
  "_tmp/dir-ex/nested/leaf" Dir.mkdir^
  if
    "  result: true" $ nl
  else
    "  result: false err=" $ $ nl
  then
  nl
;

/*
Instance string: canonical path for a Dir root.
*/
: string
  "string" $ nl
  "  params: Dir for \".\"" $ nl
  "." Dir.new^ >ts
  "  result: " $ t:.string@ $ nl
  t:drop
  nl
;

/*
Modal FileSystemIter under a Dir (IterFiles on a tiny demo tree).
*/
: iter
  "iter" $ nl
  "_tmp/dir-ex" Dir.mkdir^ if else "  mkdir: " $ $ nl then
  "_tmp/dir-ex/a.txt" "a" File.writeText?^ if else "  write a: " $ $ nl then
  "_tmp/dir-ex/b.txt" "b" File.writeText?^ if else "  write b: " $ $ nl then
  "  params: Dir \"_tmp/dir-ex\"  Dir.IterFiles" $ nl
  "_tmp/dir-ex" Dir.new^ >ts
  Dir.IterFiles t:.iter^
  "  result names:" $
  foreach
    i >ts
    " " $ t:.name@ $
    t:drop
  /foreach
  nl
  t:drop
  nl
;

new
cwd
setCwd
exists?
info
mkdir
string
iter
