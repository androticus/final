// ListRef and ListView

/*
A StringList holder for ListView.
Change list, then call ListView.refresh so the view updates.
*/
:doc ListRef

/*
The StringList shown by the view.
*/
:doc ListRef#list

/*
Create a ListRef wrapping a StringList.
( list:StringList -- ref:ListRef )
*/
:doc ListRef#new

/*
A scrolling list of strings, backed by a ListRef.
Also has Canvas.visible and Canvas.refresh.
*/
:doc ListView

/*
Call handler when a row is selected.
The handler is ( ev:Event -- ); Event.params[0] is the row Int.
( handler:MethodRef index:i -- )
*/
:doc ListView#setOnSelected

/*
Create a ListView from a ListRef.
( ref:ListRef -- w:ListView )
*/
:doc ListView#new
