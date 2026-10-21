// FyneTree and FyneTreeNode

/*
A node in a FyneTree: a branch with children, or a leaf.
*/
:doc FyneTreeNode

/*
Display text of the node.
*/
:doc FyneTreeNode#value

/*
Numeric id of the node. The tree root is 0.
*/
:doc FyneTreeNode#id

/*
A tree of labeled nodes for TreeView.
new creates a tree whose root has id 0 and value "Root".
*/
:doc FyneTree

/*
The root node (id 0).
*/
:doc FyneTree#root

/*
Add a branch node under parent.
( parent:FyneTreeNode value:s -- node:FyneTreeNode )
*/
:doc FyneTree#newBranchNode

/*
Add a leaf node under parent.
( parent:FyneTreeNode value:s -- node:FyneTreeNode )
*/
:doc FyneTree#newLeafNode

/*
Create a FyneTree with an empty root.
( -- tree:FyneTree )
*/
:doc FyneTree#new
