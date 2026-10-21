/*
Help functionality
*/

/*
Print word documentation from a dictionary Word.
( w:Word -- )
*/
: helpWord
  >xs
  x:.doc@ dup str::isEmpty
  if
    drop
    "help: no doc string for: " x:.name@ str::+ $ nl
  else
    $ nl
  then
  xdrop
;

/*
Print type member documentation for a Type dictionary word.
( w:Word -- )
*/
: helpType
  >xs
  ( x:.name )
  "%s is a Type; member documentation:\n" fmt $
  x:.exec^ // ( -- typ:Type )
  members$
  xdrop
;

: (help)
  readIdentifier dup >ts find?
  if
    dup >xs
    helpWord
    x:.isType if x:peek helpType then
    x:drop
  else
    "help: word not found: " t:peek str::+ $ nl
  then
  t:drop
  nl
;

/*
Print the doc string for the given word.
Prints a message if there is no doc string.
If the word is a Type, print the Type documentation.
Prints a message if the word is not found.
use: help <word>
<word> ( -- )
*/
: help
  true _skipSystemHelpMembers !
  (help)
;

: helpx
  false _skipSystemHelpMembers !
  (help)
;

