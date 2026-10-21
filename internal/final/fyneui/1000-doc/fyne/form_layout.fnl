// FormLayout

/*
Factory for a two-column form layout: label then widget, repeated.
new returns a Container. Distinct from Form (which adds Ok/Cancel).
*/
:doc FormLayout

/*
Create a form layout from a List of canvas controls.
The list is flat label/widget pairs; length must be even.
( objs:List -- c:Container )
*/
:doc FormLayout#new
