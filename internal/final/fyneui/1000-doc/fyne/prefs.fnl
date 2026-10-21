// Prefs — global Fyne App preferences

/*
Class-only global preferences (Fyne App preferences).
No Prefs instances — use Prefs.getBool^ etc.
Keys are strings. Values persist for this app.
*/
:doc Prefs

/*
Read a Bool, or default if the key is missing.
( key:s default:b -- b )
*/
:doc Prefs#getBool

/*
Store a Bool.
( val:b key:s -- )
*/
:doc Prefs#setBool

/*
Read an Int, or default if the key is missing.
( key:s default:i -- i )
*/
:doc Prefs#getInt

/*
Store an Int.
( val:i key:s -- )
*/
:doc Prefs#setInt

/*
Read a Float, or default if the key is missing.
( key:s default:f -- f )
*/
:doc Prefs#getFloat

/*
Store a Float.
( val:f key:s -- )
*/
:doc Prefs#setFloat

/*
Read a String, or default if the key is missing.
( key:s default:s -- s )
*/
:doc Prefs#getString

/*
Store a String.
( val:s key:s -- )
*/
:doc Prefs#setString

/*
Read an IntList. False if the key is missing.
An empty stored list is found (list + true).
( key:s -- l:IntList true | false )
*/
:doc Prefs#getIntList

/*
Store an IntList.
( val:IntList key:s -- )
*/
:doc Prefs#setIntList

/*
Read a FloatList. False if the key is missing.
An empty stored list is found (list + true).
( key:s -- l:FloatList true | false )
*/
:doc Prefs#getFloatList

/*
Store a FloatList.
( val:FloatList key:s -- )
*/
:doc Prefs#setFloatList

/*
Read a StringList. False if the key is missing.
An empty stored list is found (list + true).
( key:s -- l:StringList true | false )
*/
:doc Prefs#getStringList

/*
Store a StringList.
( val:StringList key:s -- )
*/
:doc Prefs#setStringList

/*
Delete a preference key.
( key:s -- )
*/
:doc Prefs#remove
