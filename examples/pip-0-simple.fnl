/* Simple manual spawn of a Pip
"examples/pip-0-simple.fnl" compileFromFile
*/

: helloPip
  drop
  "Hello from a Pip!\n" $
;

: spawn0
  "Hello from spawner..." $
  Pip.newPipParams^ spawn: helloPip
;
