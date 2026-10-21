// TableRef and TableView

/*
A Table holder for TableView.
Change table, then call TableView.refresh so the view updates.
*/
:doc TableRef

/*
The Table shown by the view.
*/
:doc TableRef#table

/*
Create a TableRef wrapping a Table.
( table:Table -- ref:TableRef )
*/
:doc TableRef#new

/*
A grid of cells from a TableRef.
Header row uses column names when headers is true.
Also has Canvas.visible and Canvas.refresh.
*/
:doc TableView

/*
Call handler when a cell is selected.
The handler is ( ev:Event -- ); Event.params is row then column, both Int.
( handler:MethodRef index:i -- )
*/
:doc TableView#setOnSelected

/*
Call handler when a cell is unselected.
The handler is ( ev:Event -- ); Event.params is row then column, both Int.
( handler:MethodRef index:i -- )
*/
:doc TableView#setOnUnselected

/*
Set the width of one column in device-independent pixels.
( col:i width:i -- )
*/
:doc TableView#setColumnWidth

/*
Set column widths from an IntList, index matching column.
( widths:IntList -- )
*/
:doc TableView#setColumnWidths

/*
Create a TableView from a TableRef.
( ref:TableRef headers:b -- w:TableView )
*/
:doc TableView#new
