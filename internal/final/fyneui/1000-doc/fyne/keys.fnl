// Keys and KeyEvent

/*
Named key strings for Window.setOnKeyDownUp and MenuItem.setDefaultShortcut.
Letter keys are "A"–"Z" and digit keys are "0"–"9".
*/
:doc Keys

/*
Key name for Escape.
*/
:doc Keys#escape

/*
Key name for Return.
*/
:doc Keys#return

/*
Key name for Tab.
*/
:doc Keys#tab

/*
Key name for Backspace.
*/
:doc Keys#backspace

/*
Key name for Insert.
*/
:doc Keys#insert

/*
Key name for Delete.
*/
:doc Keys#delete

/*
Key name for Right.
*/
:doc Keys#right

/*
Key name for Left.
*/
:doc Keys#left

/*
Key name for Down.
*/
:doc Keys#down

/*
Key name for Up.
*/
:doc Keys#up

/*
Key name for Page Up (value "Prior").
*/
:doc Keys#pageUp

/*
Key name for Page Down (value "Next").
*/
:doc Keys#pageDown

/*
Key name for Home.
*/
:doc Keys#home

/*
Key name for End.
*/
:doc Keys#end

/*
Key name for F1.
*/
:doc Keys#f1

/*
Key name for F2.
*/
:doc Keys#f2

/*
Key name for F3.
*/
:doc Keys#f3

/*
Key name for F4.
*/
:doc Keys#f4

/*
Key name for F5.
*/
:doc Keys#f5

/*
Key name for F6.
*/
:doc Keys#f6

/*
Key name for F7.
*/
:doc Keys#f7

/*
Key name for F8.
*/
:doc Keys#f8

/*
Key name for F9.
*/
:doc Keys#f9

/*
Key name for F10.
*/
:doc Keys#f10

/*
Key name for F11.
*/
:doc Keys#f11

/*
Key name for F12.
*/
:doc Keys#f12

/*
Key name for keypad Enter.
*/
:doc Keys#enter

/*
Key name for Space.
*/
:doc Keys#space

/*
Key name for apostrophe.
*/
:doc Keys#apostrophe

/*
Key name for comma.
*/
:doc Keys#comma

/*
Key name for minus.
*/
:doc Keys#minus

/*
Key name for period.
*/
:doc Keys#period

/*
Key name for slash.
*/
:doc Keys#slash

/*
Key name for backslash.
*/
:doc Keys#backslash

/*
Key name for left bracket.
*/
:doc Keys#leftBracket

/*
Key name for right bracket.
*/
:doc Keys#rightBracket

/*
Key name for semicolon.
*/
:doc Keys#semicolon

/*
Key name for equal.
*/
:doc Keys#equal

/*
Key name for asterisk.
*/
:doc Keys#asterisk

/*
Key name for plus.
*/
:doc Keys#plus

/*
Key name for backtick.
*/
:doc Keys#backtick

/*
Key name for left Shift.
*/
:doc Keys#leftShift

/*
Key name for right Shift.
*/
:doc Keys#rightShift

/*
Key name for left Control.
*/
:doc Keys#leftControl

/*
Key name for right Control.
*/
:doc Keys#rightControl

/*
Key name for left Alt.
*/
:doc Keys#leftAlt

/*
Key name for right Alt.
*/
:doc Keys#rightAlt

/*
Key name for left Super / Command.
*/
:doc Keys#leftSuper

/*
Key name for right Super / Command.
*/
:doc Keys#rightSuper

/*
Key name for Menu.
*/
:doc Keys#menu

/*
Key name for Print Screen.
*/
:doc Keys#printScreen

/*
Key name for Caps Lock.
*/
:doc Keys#capsLock

/*
Decode a key Event into a KeyEvent.
Event.params must be name String then down Bool.
( ev:Event -- k:KeyEvent )
*/
:doc Keys#decode

/*
A decoded keyboard event from Window.setOnKeyDownUp.
*/
:doc KeyEvent

/*
Key name string, such as "A" or Keys.escape.
*/
:doc KeyEvent#name

/*
True on key down, false on key up.
*/
:doc KeyEvent#down

/*
True if this is the Left arrow key.
*/
:doc KeyEvent#isLeft

/*
True if this is the Right arrow key.
*/
:doc KeyEvent#isRight

/*
True if this is the Up arrow key.
*/
:doc KeyEvent#isUp

/*
True if this is the Down arrow key.
*/
:doc KeyEvent#isDown

/*
True if this is the Space key.
*/
:doc KeyEvent#isSpace

/*
True if this is the Escape key.
*/
:doc KeyEvent#isEscape
