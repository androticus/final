// Window

/*
A top-level window. Create with Window.new, then setContent and show.
*/
:doc Window

/*
True if key-down auto-repeat is enabled.
*/
:doc Window#autoRepeat

/*
Install the in-window MenuBar. Call once, before setContent.
( bar:MenuBar -- )
*/
:doc Window#setMenuBar

/*
Set the window contents to a canvas control.
( widget -- )
*/
:doc Window#setContent

/*
Show the window.
( -- )
*/
:doc Window#show

/*
Hide the window without destroying it. A later show can bring it back.
( -- )
*/
:doc Window#hide

/*
Make the window chrome close button Hide instead of Close.
Call once after create so Show can reopen the same window.
( -- )
*/
:doc Window#setHideOnClose

/*
Set the window size.
( w:n h:n -- )
*/
:doc Window#resize

/*
If true, the user cannot resize the window.
( b:b -- )
*/
:doc Window#setFixedSize

/*
Call handler when the window closes.
The handler is ( ev:Event -- ) with empty params.
( handler:MethodRef index:i -- )
*/
:doc Window#setOnClosed

/*
Install key down and up handlers.
Each handler is ( ev:Event -- ); Event.params is name String then down Bool.
Use Keys.decode to turn the Event into a KeyEvent.
If autoRepeat is true, both handlers are required.
Pass nil as a handler to skip that event.
( down:MethodRef downIdx:i up:MethodRef upIdx:i autoRepeat:b -- )
*/
:doc Window#setOnKeyDownUp

/*
Create a Window with the given title.
( app:App title:s -- window:Window )
*/
:doc Window#new
