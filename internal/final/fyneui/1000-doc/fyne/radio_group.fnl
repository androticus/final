// RadioGroup

/*
A group of exclusive radio choices.
Also has Canvas.visible, Canvas.refresh, and Disableable.disabled.
*/
:doc RadioGroup

/*
The selected option String, or "" if none.
*/
:doc RadioGroup#selected

/*
Create a RadioGroup from a StringList of options.
The handler is ( ev:Event -- ); Event.params[0] is the selected String.
Pass nil as handler to skip the event.
( handler:MethodRef index:i options:StringList -- w:RadioGroup )
*/
:doc RadioGroup#new
