forget: reset-hello-event

requiresFyne
// `examples/gui/fyne/event-fyne.fnl` compileFromFile
`
Demo of Final Event handling in Fyne widgets.

Shows how to use a TypedObject method as an event
sink for a Fyne event source, in this case a Button.

Try:

event-fyne

Then click the button.
`
const: event-fyne-blurb


// HelloEvent TypedObject to represent the app, and event handlers

// Use Paren operators to create the spec.

{
   {
     Window "win"
     Button "btn"
     Counter "clicks"
   } "fields"
     
  "HelloEvent" "name"
}
TypedObject.makeType^

/*
 Button click event handler.
( ev:Event -- ) o( inst )
*/
: HelloEvent#onButtonClick
  drop // ev
  o:.clicks@.inc^
  (
    o:.clicks@.value
  )
  "Clicked: %d" fmt
  o:.btn@.setText^
;

: event-fyne
  HelloEvent.new^ >zs
  
  Counter.new^ z:.clicks!
  
  App.app "Event Handler Demo"
  Window.new^ dup >ts z:.win!
  
  z:.onButtonClick~ 0 "Clicked: 0"
  Button.new^ dup
  z:.btn!
  t:.setContent^
  
  z:drop
  t:.show^

  t:drop
;

event-fyne-blurb $ nl
