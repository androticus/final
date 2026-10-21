// Menu and MenuItem

/*
A named menu of MenuItem values, used in a MenuBar.
*/
:doc Menu

/*
Create a Menu from a List of MenuItem.
( menuItems:List label:s -- menu:Menu )
*/
:doc Menu#new

/*
One entry in a Menu: an action, a checkable item, or a separator.
*/
:doc MenuItem

/*
True if the item can be chosen.
*/
:doc MenuItem#enabled

/*
True if a checkable item is marked.
*/
:doc MenuItem#checked

/*
Set the keyboard shortcut: Command+key on macOS, Control+key elsewhere.
key is a Keys name, such as "L" or Keys.f1.
( key:s -- )
*/
:doc MenuItem#setDefaultShortcut

/*
Create a MenuItem. The handler is ( ev:Event -- ) with empty params.
Pass nil as handler to skip the event.
( handler:MethodRef index:i label:s -- menuItem:MenuItem )
*/
:doc MenuItem#new

/*
Create a separator MenuItem.
( -- menuItem:MenuItem )
*/
:doc MenuItem#newSeparator
