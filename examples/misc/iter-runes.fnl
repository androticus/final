forget: reset-iter-runes
`
Shows how easy it is to iterate the Unicode runes in a Final string.
Just have the string on DS, and use a foreach loop.
"i" is the value of each rune in the string.

Try:
iter-runes
` const: iter-runes-blurb

/*
Iterate the runes of a string.
( -- )
*/
: iter-runes
  "Unicode string: " $
  "aé🙂日" dup $ nl nl
  foreach
    i $ " " $
  /foreach
  nl
;

iter-runes-blurb $ nl
