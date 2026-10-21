// Check

/*
A labeled checkbox.
Also has Canvas.visible, Canvas.refresh, and Disableable.disabled.
*/
:doc Check

/*
True if the box is checked.
*/
:doc Check#checked

/*
Call handler when checked changes.
The handler is ( ev:Event -- ); Event.params[0] is the new Bool.
( handler:MethodRef index:i -- )
*/
:doc Check#setOnChanged

/*
Create a Check with the given label.
( label:s -- check:Check )
*/
:doc Check#new
