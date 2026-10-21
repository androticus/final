// Canvas — visible / refresh on canvas controls

/*
visible and refresh for widgets and layouts that sit on the canvas.
Copied onto those types.
*/
:doc Canvas

/*
True if the control is shown. Setting false hides it; true shows it.
*/
:doc Canvas#visible

/*
Redraw after backing data or pixels changed.
Ordinary setters (Label.setText, Check.checked, …) already refresh.
Use this for CanvasImage after mutating a Pixmap, and for
ListView / TableView / TreeView after changing the Ref's data.
( -- )
*/
:doc Canvas#refresh
