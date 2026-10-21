/*
finaltui package Final (listed from package.fnlm).

Tui_ is minted by finaltui.Init (Tui.instance^ var: Tui_) before this file runs.
Dialog / help words ok to land heres.
*/

// Dialog

/*
Encapapsutes posting a confirmation or field dialog in the TUI.
*/
:doc Dialog

/*
Return button value when user presses Ok.
*/
:doc Dialog#btnOk

/*
Return button value when user presses Cancel.
*/
:doc Dialog#btnCancel

/*
Post a simple confirmation Ok/Cancel dialog with a prompt.
( prompt:s -- btn:i )
*/
:doc Dialog#showDefaultPrompt

/*
Return a new FieldDialogParams object, with default values set.
User must add fields to the .fields param of the object.
See the user manual, intro/tui-dialogs/field.md for properties.
( -- fdp:DynamicObject )
*/
:doc Dialog#newFieldDialogParams

/*
Post the Field Dialog.
If btn==btnOk, the new field values will be in their .field value
params in the fdp Object.
( fdp:DynamicObject -- fields:DO true | false )
*/
:doc Dialog#showFieldDialog

// Keep in sync with internal/final/fyneui/1000-doc/fyne/dialog.fnl (Dialog##ask).
/*
Prompts the user to provide from 1 to 7 String-valued responses.
The prompts are provided in a List, typically using the ( and )
Paren operators. The results are also in a List, in order, 
ready for a fmt output statement.
Returns results+true for Ok, false for Cancel
( prompts:List -- results:List true | false )
*/
: Dialog##ask
  Dialog.newFieldDialogParams^ >ts
  foreach
    "" i t:.fields!{}
  /foreach
  ts>
  Dialog.showFieldDialog^
  if
    List.new^ >ts
    foreach
      i >xs x:.value xdrop
      t:.append^
    /foreach
    ts>
    true
  else
    false
  then
;

