forget: reset-key-fyne

requiresFyne
`
Shows how to set the Window.canvas key down/up handlers.
Note, this only works when you don't have any focusable
widgets on the screen.
It is intended for games and simulations.

Try:

key-fyne

Then press keys.
`
const: key-fyne-blurb


// Use Paren operators to create the spec.

{
   {
     Window "win"
     Label "lbl"
   } "fields"
     
  "KeyEventExample" "name"
}
TypedObject.makeType^

/*
Window event handler.
Attached to both the KeyDown and KeyUp events
( ev:Event -- ) o( inst )
*/
: KeyEventExample#onKey
  Keys.decode^ l.ke!

  ( l.ke@.name l.ke@.down ) 
  "Key: %s Down: %t" fmt
  o:.lbl@.setText^
;

: key-fyne
  KeyEventExample.new^ >zs
  
  App.app "Key Event Handler Demo"
  Window.new^ dup >ts z:.win!
  
  z:.onKey~ dup 0 over 1 false
  t:.setOnKeyDownUp^ 
  
  "Key:     "
  Label.new^ dup
  z:.lbl!
  t:.setContent^
  
  z:drop
  t:.show^

  t:drop
;


key-fyne-blurb $ nl
