// Select

/*
A drop-down list of string options.
Also has Canvas.visible, Canvas.refresh, and Disableable.disabled.
*/
:doc Select

/*
Index of the selected option, starting at 0.
*/
:doc Select#selectedIndex

/*
Create a Select from a StringList of options.
The handler is ( ev:Event -- ); Event.params[0] is the selected String.
Pass nil as handler to skip the event.
( handler:MethodRef index:i options:StringList -- w:Select )
*/
:doc Select#new
