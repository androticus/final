// Dialog — TUI-compatible prompt and field dialogs for Fyne

/*
Ok/Cancel prompt and typed field form for GUI apps.
Same show stacks as TUI Dialog (modal wait). Field params match the TUI
Field Dialog (title, bodyText, defaultButton, fields).
*/
:doc Dialog

/*
Ok was chosen. Value 1.
*/
:doc Dialog#btnOk

/*
Cancel was chosen. Value 0.
*/
:doc Dialog#btnCancel

/*
Show Ok/Cancel confirmation with title "Prompt" and wait until dismissed.
Returns Dialog.btnOk or Dialog.btnCancel.
( prompt:s -- btn:i )
*/
:doc Dialog#showDefaultPrompt

/*
Mint field-dialog params with defaults and an empty fields bag.
Fill title, bodyText, defaultButton, and fields before showFieldDialog.
( -- fdp:DynamicObject )
*/
:doc Dialog#newFieldDialogParams

/*
Show a form from field-dialog params and wait until dismissed.
Ok → edited fields bag and true. Cancel / window-close → false only.
Invalid Int/Float keeps the form open.
( fdp:DynamicObject -- fields:DO true | false )
*/
:doc Dialog#showFieldDialog

// Keep in sync with finaltui/final/000-tui.fnl (Dialog##ask).
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
