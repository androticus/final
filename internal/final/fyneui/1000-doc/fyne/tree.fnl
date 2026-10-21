// TreeRef and TreeView

/*
A FyneTree holder for TreeView.
Change tree, then call TreeView.refresh so the view updates.
*/
:doc TreeRef

/*
The FyneTree shown by the view.
*/
:doc TreeRef#tree

/*
Create a TreeRef wrapping a FyneTree.
( tree:FyneTree -- ref:TreeRef )
*/
:doc TreeRef#new

/*
A collapsible tree from a TreeRef.
Also has Canvas.visible and Canvas.refresh.
*/
:doc TreeView

/*
Call handler when a node is selected.
The handler is ( ev:Event -- ); Event.params[0] is the node id Int.
( handler:MethodRef index:i -- )
*/
:doc TreeView#setOnSelected

/*
Call handler when a node is unselected.
The handler is ( ev:Event -- ); Event.params[0] is the node id Int.
( handler:MethodRef index:i -- )
*/
:doc TreeView#setOnUnselected

/*
Create a TreeView from a TreeRef.
( ref:TreeRef -- w:TreeView )
*/
:doc TreeView#new
