forget: reset-dialogs-fyne
requiresFyne
`
This demo lets you see an example of each of Fyne's
modal dialogs.

Try:

dialogs-demo
` const: dialogs-fyne-blurb

// TypedObject for button callbacks
{
  "DialogDemo" "name"
  {
  } "fields"
  
  { 
    0 "btnInfo"
    1 "btnConfirm"
    2 "btnColor"
    3 "btnFileOpen"
    4 "btnFileSave"
    5 "btnFolderSelect"
    6 "btnModalForm"
    7 "btnDialog"
  } "constants"
}
TypedObject.makeType^

DialogDemo.new^ const: dialogDemo

null var: dialogWindow

/*
Handles buttons events from Parent window.
( ev:Event -- ) o( inst:SimGame )
*/
: DialogDemo#onButtonTap
  >ts t:.index@
  dialogWindow @ >xs
  switch
    a DialogDemo.btnInfo == case
      xs> "Information" "This is an InformationDialog.\nClick Ok to dismiss." InformationDialog.show^
    /case
    a DialogDemo.btnConfirm == case
      xs> "Confirm" "Proceed with this action?" ConfirmationDialog.show^
      Dialog.btnOk == if
        "ConfirmationDialog: Ok" $ nl
      else
        "ConfirmationDialog: Cancel" $ nl
      then
    /case
    a DialogDemo.btnColor == case
      xs> "Color" "Pick a color" ColorDialog.show^
      not if
        "ColorDialog: Cancel" $ nl
      else
        >ys
        "ColorDialog: Ok rgb=(" $
        y:.r@ $ "," $ y:.g@ $ "," $ y:.b@ $
        ")" $ nl
        y:drop
      then
    /case
    a DialogDemo.btnFileOpen == case
      xs> FileDialog.showOpenFile^
      ?dup if
        "FileDialog.open: " $ $ nl
      else
        "FileDialog.open: Cancel" $ nl
      then
    /case
    a DialogDemo.btnFileSave == case
      xs> FileDialog.showSaveText^
      not if
        "FileDialog.saveText: Cancel" $ nl
      else
        >ts
        "The rain in Spain falls\nmainly in the airplane.\n" t:.writeString^
        t:.close^
        "FileDialog.saveText: wrote to: " $ $ nl
        t:drop
      then
    /case
    a DialogDemo.btnFolderSelect == case
      xs> FileDialog.showOpenFolder^
      ?dup if
        "FileDialog.folder: " $ $ nl
      else
        "FileDialog.folder: Cancel" $ nl
      then
    /case
    a DialogDemo.btnModalForm == case
      Entry.new^ >ys
      Form.new^ >zs
      "Name" y:peek z:.append^
      xs> "Modal Form" z:.showModal^
      Dialog.btnOk == if
        "Form.showModal: " $ y:.text@ $ nl
      else
        "Form.showModal: Cancel" $ nl
      then
      y:drop z:drop
    /case
    a DialogDemo.btnDialog == case
      xdrop
      "Continue with Dialog.showDefaultPrompt?" Dialog.showDefaultPrompt^
      Dialog.btnOk == if
        "Dialog.prompt: Ok" $ nl
        Dialog.newFieldDialogParams^ >ys
        "Demo" y:.title!
        "Enter a name." y:.bodyText!
        y:.fields@ >zs
        "" z:.name!
        z:drop
        ys> Dialog.showFieldDialog^
        not if
          "Dialog.fields: Cancel" $ nl
        else
          >ys
          "Dialog.fields name=" $ y:.name@ $ nl
          y:drop
        then
      else
        "Dialog.prompt: Cancel" $ nl
      then
    /case
    default // shouldn't happen
      xdrop
    /default
  /switch
  tdrop
;

/*
Parent window for the stock dialog demos.
Creates and shows a small window each call.
( -- win:Window )
*/
: dialogs-parent
  dialogDemo l.dd!
  App.app "Fyne Dialogs" Window.new^ dup >ts
  dialogWindow !
  (
    "Select a Dialog Demo to run:" Label.new^ 
    l.dd.onButtonTap~ DialogDemo.btnInfo "Information Dialog" Button.new^
    l.dd.onButtonTap~ DialogDemo.btnConfirm "Confirmation Dialog" Button.new^
    l.dd.onButtonTap~ DialogDemo.btnColor "Color Picker Dialog" Button.new^
    l.dd.onButtonTap~ DialogDemo.btnFileOpen "File Open Dialog" Button.new^
    l.dd.onButtonTap~ DialogDemo.btnFileSave "File Save Dialog -- WARNING: truncates an existing file!" Button.new^
    l.dd.onButtonTap~ DialogDemo.btnFolderSelect "Folder Select Dialog" Button.new^
    l.dd.onButtonTap~ DialogDemo.btnModalForm "Modally show a Form" Button.new^
    l.dd.onButtonTap~ DialogDemo.btnDialog "Modally show the convenience Dialog Type" Button.new^
  ) Box.newV^
  t:.setContent^
  // Modal popups clamp to parent canvas; size for color + file dialogs.
  720 560 t:.resize^
  t:.show^
  tdrop
;


: dialogs-demo
  dialogs-parent
;

dialogs-fyne-blurb $ nl
