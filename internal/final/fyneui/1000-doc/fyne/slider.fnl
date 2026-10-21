// Slider

/*
A numeric slider between min and max.
Also has Canvas.visible, Canvas.refresh, and Disableable.disabled.
*/
:doc Slider

/*
Lower bound, set at creation.
*/
:doc Slider#min

/*
Upper bound, set at creation.
*/
:doc Slider#max

/*
Current value.
*/
:doc Slider#value

/*
Step size, set at creation.
*/
:doc Slider#step

/*
Call handler while the value is changing.
The handler is ( ev:Event -- ); Event.params[0] is the Float value.
( handler:MethodRef index:i -- )
*/
:doc Slider#setOnChanged

/*
Call handler when the drag ends.
The handler is ( ev:Event -- ); Event.params[0] is the Float value.
( handler:MethodRef index:i -- )
*/
:doc Slider#setOnChangeEnded

/*
Set the slider size.
( w:n h:n -- )
*/
:doc Slider#resize

/*
Create a Slider.
( min:n max:n step:n -- slider:Slider )
*/
:doc Slider#new
