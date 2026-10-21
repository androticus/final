// Form

/*
A labeled list of fields with optional Submit and Cancel.
Also has Canvas.visible, Canvas.refresh, and Disableable.disabled.
*/
:doc Form

/*
Add a labeled field. The widget is any canvas control.
( label:s widget -- )
*/
:doc Form#append

/*
Show Cancel and call handler when it is pressed.
The handler is ( ev:Event -- ) with empty params.
( handler:MethodRef index:i -- )
*/
:doc Form#setOnCancel

/*
Show Submit and call handler when it is pressed.
The handler is ( ev:Event -- ) with empty params.
( handler:MethodRef index:i -- )
*/
:doc Form#setOnSubmit

/*
Show this Form as a modal popup on parent and wait until it is dismissed.
Submit → Dialog.btnOk (1). Cancel / window-close → Dialog.btnCancel (0).
Caller harvests field widgets after Ok. Non-modal setOnSubmit/setOnCancel
are restored when the wait ends.
( parent:Window title:s -- closer:i )
*/
:doc Form#showModal

/*
Create an empty Form.
( -- form:Form )
*/
:doc Form#new
