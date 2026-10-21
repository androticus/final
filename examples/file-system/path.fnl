forget: reset

/*
Path string algebra — class-only Path helpers.

Demonstrates Path.join / split / clean / abs / basename / dirname / extname.
No Path instances: operands are strings on the data stack; Path is the type receiver.

Run (from repo or Final/ install root):
  "examples/file-system/path.fnl" compileFromFile
*/


/*
Join two segments with the host separator (no doubled or missing separators).
*/
: join
  "join" $ nl
  "  params: \"out\" \"report.md\"" $ nl
  "out" "report.md" Path.join^
  "  result: " $ $ nl
  nl
;

/*
Split one path into components (StringList). Absolute paths keep the root as the first part.
*/
: split
  "split" $ nl
  "  params: \"/var/log/syslog\"" $ nl
  "/var/log/syslog" Path.split^
  "  result:" $
  foreach
    " " $ i $
  /foreach
  nl
  nl
;

/*
Clean redundant . and .. segments.
*/
: clean
  "clean" $ nl
  "  params: \"a/./b/../c\"" $ nl
  "a/./b/../c" Path.clean^
  "  result: " $ $ nl
  nl
;

/*
Resolve to an absolute path (uses process cwd).
*/
: abs
  "abs" $ nl
  "  params: \".\"" $ nl
  "." Path.abs^
  "  result: " $ $ nl
  nl
;

/*
Leaf name without extension (\"out/reports/q1.md\" → \"q1\").
*/
: basename
  "basename" $ nl
  "  params: \"out/reports/q1.md\"" $ nl
  "out/reports/q1.md" Path.basename^
  "  result: " $ $ nl
  nl
;

/*
Parent directory of a path string.
*/
: dirname
  "dirname" $ nl
  "  params: \"out/reports/q1.md\"" $ nl
  "out/reports/q1.md" Path.dirname^
  "  result: " $ $ nl
  nl
;

/*
Extension including the leading dot, or empty if none.
*/
: extname
  "extname" $ nl
  "  params: \"out/reports/q1.md\"" $ nl
  "out/reports/q1.md" Path.extname^
  "  result: " $ $ nl
  nl
;

join
split
clean
abs
basename
dirname
extname
