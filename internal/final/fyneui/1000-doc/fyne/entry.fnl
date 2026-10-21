// Entry

/*
A single-line or multi-line text field.
Also has Canvas.visible, Canvas.refresh, and Disableable.disabled.
*/
:doc Entry

/*
The current text.
*/
:doc Entry#text

/*
Call handler when the text changes.
The handler is ( ev:Event -- ); Event.params[0] is the new String.
( handler:MethodRef index:i -- )
*/
:doc Entry#setOnChanged

/*
Call handler when the user submits (Enter on a single-line Entry).
The handler is ( ev:Event -- ); Event.params[0] is the submitted String.
( handler:MethodRef index:i -- )
*/
:doc Entry#setOnSubmitted

/*
Create a single-line Entry.
( -- entry:Entry )
*/
:doc Entry#new

/*
Create a multi-line Entry with word wrap.
( -- entry:Entry )
*/
:doc Entry#newMultiLine
