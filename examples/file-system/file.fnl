forget: reset

/*
File class helpers — exists?, info, getBaseExt, makeName, copy, move, delete.

String ops need no disk. Mutators use _tmp/file-ex/ (created as needed).
Create parent _tmp if missing: the example calls Dir.mkdir on the demo dirs.

Run (from repo or Final/ install root):
  "examples/file-system/file.fnl" compileFromFile
*/


/*
Split a name into base + ext (ext without the dot; path prefix stripped).
*/
: getBaseExt
  "getBaseExt" $ nl
  "  params: \"archive/report.final.md\"" $ nl
  "archive/report.final.md" File.getBaseExt^
  "  result: base=" $ swap $ " ext=" $ $ nl
  nl
;

/*
Join base + ext into a file name (ext without the dot; empty ext → base alone).
*/
: makeName
  "makeName" $ nl
  "  params: \"out/report\" \"md\"" $ nl
  "out/report" "md" File.makeName^
  "  result: " $ $ nl
  nl
;

/*
Whether a path exists (missing → false).
*/
: exists?
  "exists?" $ nl
  "  params: \".\"" $ nl
  "." File.exists?^
  "  result: " $ $ nl
  nl
;

/*
Metadata snapshot for a path (modal: fi true | err false).
*/
: info
  "info" $ nl
  "  params: \".\"" $ nl
  "." File.info^
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
Copy a file into a destination directory (keeps leaf name; no clobber).
*/
: copy
  "copy" $ nl
  "_tmp/file-ex" Dir.mkdir^ if else "  mkdir file-ex: " $ $ nl then
  "_tmp/file-ex/archive" Dir.mkdir^ if else "  mkdir archive: " $ $ nl then
  "_tmp/file-ex/notes.txt" "hello from File.copy" File.writeText?^
  if else "  write notes: " $ $ nl then
  "  params: \"_tmp/file-ex/notes.txt\" \"_tmp/file-ex/archive\"" $ nl
  "_tmp/file-ex/notes.txt" "_tmp/file-ex/archive" File.copy^
  if
    "  result: true" $ nl
  else
    "  result: false err=" $ $ nl
  then
  nl
;

/*
Rename / relocate (dst is the full destination path; no clobber).
*/
: move
  "move" $ nl
  "  params: \"_tmp/file-ex/archive/notes.txt\" \"_tmp/file-ex/archive/notes-2026.txt\"" $ nl
  "_tmp/file-ex/archive/notes.txt" "_tmp/file-ex/archive/notes-2026.txt" File.move^
  if
    "  result: true" $ nl
  else
    "  result: false err=" $ $ nl
  then
  nl
;

/*
Delete a file or empty directory.
*/
: delete
  "delete" $ nl
  "  params: \"_tmp/file-ex/archive/notes-2026.txt\"" $ nl
  "_tmp/file-ex/archive/notes-2026.txt" File.delete^
  if
    "  result: true" $ nl
  else
    "  result: false err=" $ $ nl
  then
  nl
;

getBaseExt
makeName
exists?
info
copy
move
delete
