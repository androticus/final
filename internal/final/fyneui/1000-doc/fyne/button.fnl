// Button

/*
A clickable labeled button.
Also has Canvas.visible, Canvas.refresh, and Disableable.disabled.
*/
:doc Button

/*
Set the button label.
( text:s -- )
*/
:doc Button#setText

/*
Create a Button. The handler is ( ev:Event -- ) with empty params.
( handler:MethodRef index:i text:s -- btn:Button )
*/
:doc Button#new

/*
Create a Button with a theme icon from Icon.
The handler is ( ev:Event -- ) with empty params.
( handler:MethodRef index:i text:s icon -- btn:Button )
*/
:doc Button#newWithIcon
