// Pathing smoke — create, exists?, read back. Temp file beside this source.
// Run when changing filesystem / path APIs. Launcher deletes the tmp afterward.

: run-pathing-smoke
  "internal/final/extras/.pathing-smoke-tmp.txt" TextWriter.newFromFileTruncateOrCreate?^ not if
    "SMOKE_FAIL pathing: create writer: " $$ $ nl
    1 0 /
  then
  >ts
  "pathing-smoke-ok" t:.writeLine^
  t:.flush^
  t:.close^
  t:drop

  "internal/final/extras/.pathing-smoke-tmp.txt" File.fileExists?^ not if
    "SMOKE_FAIL pathing: fileExists? false\n" $$
    1 0 /
  then

  "internal/final/extras/.pathing-smoke-tmp.txt" TextReader.newFromFile?^ not if
    "SMOKE_FAIL pathing: open reader: " $$ $ nl
    1 0 /
  then
  >ts
  t:.readLine^ not if
    "SMOKE_FAIL pathing: unexpected EOF\n" $$
    1 0 /
  then
  "pathing-smoke-ok" <> if
    "SMOKE_FAIL pathing: content mismatch\n" $$
    1 0 /
  then
  t:.close^
  t:drop

  "SMOKE_OK pathing\n" $
;
run-pathing-smoke
