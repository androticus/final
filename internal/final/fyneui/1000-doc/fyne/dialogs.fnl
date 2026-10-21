// Standard dialogs

/*
A color picker popup.
*/
:doc ColorDialog

/*
Show a color picker on parent and wait until it is dismissed.
Ok → Color and true. Cancel / window-close → false only.
( parent:Window title:s message:s -- color:Color true | false )
*/
:doc ColorDialog#show

/*
A Yes/No confirmation popup.
*/
:doc ConfirmationDialog

/*
Show a confirmation popup on parent and wait until it is dismissed.
Returns Dialog.btnOk (1) for Yes, Dialog.btnCancel (0) for No / window-close.
( parent:Window title:s message:s -- btn:i )
*/
:doc ConfirmationDialog#show

/*
Open-file, save-file, and open-folder popups.
Open/folder wait and return a path String ("" if cancelled).
Save returns path plus an open TextWriter or ByteWriter (caller closes).
Confirming Save creates or truncates the file immediately (Fyne storage.Writer);
it does not append. Use TextWriter/ByteWriter.newFromFileAppendOrCreate for append.
*/
:doc FileDialog

/*
Show an open-file picker on parent and wait until it is dismissed.
Returns the chosen path, or "" if cancelled.
( parent:Window -- path:s )
*/
:doc FileDialog#showOpenFile

/*
Show a save-file picker on parent and wait until it is dismissed.
Ok → path, open TextWriter, and true (caller must close). Cancel → false only.
Confirm creates or truncates the file; not append.
( parent:Window -- path:s w:TextWriter true | false )
*/
:doc FileDialog#showSaveText

/*
Show a save-file picker on parent and wait until it is dismissed.
Ok → path, open ByteWriter, and true (caller must close). Cancel → false only.
Confirm creates or truncates the file; not append.
( parent:Window -- path:s w:ByteWriter true | false )
*/
:doc FileDialog#showSaveBytes

/*
Show an open-folder picker on parent and wait until it is dismissed.
Returns the chosen path, or "" if cancelled.
( parent:Window -- path:s )
*/
:doc FileDialog#showOpenFolder

/*
An information popup with an OK button.
*/
:doc InformationDialog

/*
Show an information popup on parent and wait until it is dismissed.
No result is pushed. Ok and window-close both complete the wait.
( parent:Window title:s message:s -- )
*/
:doc InformationDialog#show
