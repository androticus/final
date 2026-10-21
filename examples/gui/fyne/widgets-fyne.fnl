forget: reset-widgets-fyne

requiresFyne
`
This demo shows the Fyne widgets that Final supports.
Look at the code for examples of how to create every
type of control.

see: https://docs.fyne.io/ 
for detailed documentation of Fyne widgets

Try:

widgets-fyne

Use the tabs to browse controls. Press Show Form for a
Form in a second window. The Event line updates when
controls fire handlers.
`
const: widgets-fyne-blurb


// WidgetsApp — window state + event sinks

{
   {
     Window "win"
     Window "formWin"
     Label "eventDetail"
     Activity "activity"
   } "fields"
   
   {
     // list of strings to show for a menu event
     // in order of the indexes assigned to items
     (
       "File/New"
       "File/Open"
       "File/Save"
       "Help/About"
     s) "menuEventText"
   } "constants"
     
  "WidgetsApp" "name"
}
TypedObject.makeType^


/*
Update the Event detail label.
( msg:String -- ) o( inst )
*/
: WidgetsApp#setEventMsg
  o:.eventDetail@.setText^
;

// --- event handlers ( ev:Event -- ) o( inst ) ---
// Park the Event, then fmt from its params List.

: WidgetsApp#onMenuItemSelected
  >ts
  ( t:.index@ WidgetsApp.menuEventText >ts t:@() t:drop )
  "MenuItem selected: %s" fmt
  o:.setEventMsg^
  t:drop
;

: WidgetsApp#onEntryChanged
  >ts
  t:.params@ "Entry changed: %s" fmt
  o:.setEventMsg^
  t:drop
;

: WidgetsApp#onEntrySubmitted
  >ts
  t:.params@ "Entry submitted: %s" fmt
  o:.setEventMsg^
  t:drop
;

: WidgetsApp#onCheckChanged
  >ts
  t:.params@ "Check: %t" fmt
  o:.setEventMsg^
  t:drop
;

: WidgetsApp#onRadioChanged
  >ts
  t:.params@ "RadioGroup: %s" fmt
  o:.setEventMsg^
  t:drop
;

: WidgetsApp#onSelectChanged
  >ts
  t:.params@ "Select: %s" fmt
  o:.setEventMsg^
  t:drop
;

: WidgetsApp#onListSelected
  >ts
  t:.params@ "ListView selected index: %d" fmt
  o:.setEventMsg^
  t:drop
;

: WidgetsApp#onTableSelected
  >ts
  t:.params@ "TableView selected row=%d col=%d" fmt
  o:.setEventMsg^
  t:drop
;

: WidgetsApp#onTreeSelected
  >ts
  t:.params@ "TreeView selected id: %d" fmt
  o:.setEventMsg^
  t:drop
;

: WidgetsApp#onFormSubmit
  drop
  "Form: Submit" o:.setEventMsg^
  o:.formWin@.hide^
;

: WidgetsApp#onFormCancel
  drop
  "Form: Cancel" o:.setEventMsg^
  o:.formWin@.hide^
;

: WidgetsApp#onShowForm
  drop
  "Show Form pressed" o:.setEventMsg^
  o:.formWin@.show^
;

: WidgetsApp#onActivityStart
  drop
  o:.activity@.start^
  "Activity: start" o:.setEventMsg^
;

: WidgetsApp#onActivityStop
  drop
  o:.activity@.stop^
  "Activity: stop" o:.setEventMsg^
;


/*
Basic Widgets tab — build controls first (handlers need MethodRef on DS),
then assemble the Grid list.
( -- content ) o( inst )
*/
: WidgetsApp#makeBasicTab
  Entry.new^ >ts
  o:.onEntryChanged~ 0 t:.setOnChanged^
  o:.onEntrySubmitted~ 0 t:.setOnSubmitted^
  ts> l.entry!
  
  "Enable" Check.new^ >ts
  o:.onCheckChanged~ 0 t:.setOnChanged^
  ts> l.check!
  
  o:.onRadioChanged~ 0
  ( "One" "Two" "Three" s)
  RadioGroup.new^
  l.radio!
  
  o:.onSelectChanged~ 0
  ( "Red" "Green" "Blue" s)
  Select.new^
  l.sel!
  
  (
    "Scroll 1" Label.new^
    "Scroll 2" Label.new^
    "Scroll 3" Label.new^
    "Scroll 4" Label.new^
    "Scroll 5" Label.new^
    "Scroll 6" Label.new^
    "Scroll 7" Label.new^
    "Scroll 8" Label.new^
    "Scroll 9" Label.new^
    "Scroll 10" Label.new^
    "Scroll 11" Label.new^
    "Scroll 12" Label.new^
  ) Box.newV^ Scroll.new^ >ts
  200 120 t:.setMinSize^
  ts> l.scroll!
  
  (
    "Grid:" Label.new^
    2 (
      "Label" Label.new^
      "Hello Final" Label.new^
      "Entry" Label.new^
      l.entry@
      "Check" Label.new^
      l.check@
      "RadioGroup" Label.new^
      l.radio@
      "Select" Label.new^
      l.sel@
    ) Grid.newColumns^
    "Scroll" Label.new^
    l.scroll@
  ) Box.newV^
;

/*
Layouts tab — Split, Separator, Accordion, CanvasImage.
VSplit is Border center so it gets leftover height (VBox would
pack it to min height, so the divider cannot move).
( -- content ) o( inst )
*/
: WidgetsApp#makeLayoutsTab
  "Left" Label.new^
  "Right" Label.new^
  Split.newH^
  l.hSplit!
  
  "Top" Label.new^
  "Bottom" Label.new^
  Split.newV^
  l.vSplit!
  
  (
    "Above Separator" Label.new^
    Separator.new^
    "Below Separator" Label.new^
  ) Box.newV^
  l.sepBox!
  
  Accordion.new^ >ts
  "Item 1" "Accordion detail one" Label.new^ t:.append^
  "Item 2" "Accordion detail two" Label.new^ t:.append^
  ts> l.acc!
  
  "examples/gui/fyne/favicon-32-f-circle.png"
  Image.newFromFile^ CanvasImage.new^
  l.img!
  
  (
    "H Split" Label.new^
    l.hSplit@
    "V Split" Label.new^
  ) Box.newV^
  
  (
    "Separator" Label.new^
    l.sepBox@
    "Accordion" Label.new^
    l.acc@
    "CanvasImage" Label.new^
    l.img@
  ) Box.newV^
  
  null
  null
  l.vSplit@
  Border.new^
;

/*
Misc tab — Activity spinner.
( -- content ) o( inst )
*/
: WidgetsApp#makeMiscTab
  Activity.new^ o:.activity!
  (
    "Activity" Label.new^
    (
      o:.onActivityStart~ 0 "Start spinner" Button.new^
      o:.onActivityStop~ 0 "Stop spinner" Button.new^
      o:.activity@
    ) Box.newH^
  ) Box.newV^
;

/*
RichText tab.
( -- content ) o( inst )
*/
: WidgetsApp#makeRichTextTab
  `
# RichText

Fyne **markdown** via RichText.newFromMarkdown^.

- Lists
- *Emphasis*
- plain code-style text without tick marks

Use this tab to smoke RichText.
  ` RichText.newFromMarkdown^
;

/*
Border tab — labels in top, bottom, left, right, center.
( -- content ) o( inst )
*/
: WidgetsApp#makeBorderTab
  "TOP" Label.new^
  "BOTTOM" Label.new^
  "LEFT" Label.new^
  "RIGHT" Label.new^
  "CENTER" Label.new^
  Border.new^
;

/*
List tab.
( -- content ) o( inst )
*/
: WidgetsApp#makeListTab
  ( "Alpha" "Beta" "Gamma" "Delta" "Epsilon" s)
  ListRef.new^
  ListView.new^ >ts
  o:.onListSelected~ 0 t:.setOnSelected^
  ts>
;

/*
Table tab.
( -- content ) o( inst )
*/
: WidgetsApp#makeTableTab
  {
    "" "name"
    0 "age"
  } >ts
  t:peek Table.newFromDynamicObject^ >xs
  
  t:.newStatic^ >ys
  "Ada" y:.name!
  36 y:.age!
  ys> x:.append^
  
  t:.newStatic^ >ys
  "Grace" y:.name!
  85 y:.age!
  ys> x:.append^
  
  t:.newStatic^ >ys
  "Alan" y:.name!
  41 y:.age!
  ys> x:.append^
  
  t:drop
  xs> TableRef.new^
  true
  TableView.new^ >ts
  // widths: Fyne DIPs; name ~120, age ~60 (overrides CreateCell default)
  ( 120 60 i) t:.setColumnWidths^
  o:.onTableSelected~ 0 t:.setOnSelected^
  ts>
;

/*
Tree tab.
( -- content ) o( inst )
*/
: WidgetsApp#makeTreeTab
  FyneTree.new^ >ts
  t:.root@ >xs
  
  x:peek "Animals" t:.newBranchNode^ >ys
  y:peek "Dog" t:.newLeafNode^ drop
  y:peek "Cat" t:.newLeafNode^ drop
  y:drop
  
  x:peek "Plants" t:.newBranchNode^ >ys
  y:peek "Fern" t:.newLeafNode^ drop
  y:peek "Moss" t:.newLeafNode^ drop
  y:drop
  
  x:drop
  ts> TreeRef.new^
  TreeView.new^ >ts
  o:.onTreeSelected~ 0 t:.setOnSelected^
  ts>
;

/*
Secondary window with a Form (Entry, Check, RadioGroup, Submit/Cancel).
( -- ) o( inst )
*/
: WidgetsApp#makeFormWindow
  Form.new^ >ts
  
  "Name" Entry.new^ t:.append^
  "Subscribe" "Yes please" Check.new^ t:.append^
  
  null 0
  ( "Small" "Medium" "Large" s)
  RadioGroup.new^
  "Size" swap t:.append^
  
  o:.onFormSubmit~ 0 t:.setOnSubmit^
  o:.onFormCancel~ 0 t:.setOnCancel^
  
  App.app "Demo Form"
  Window.new^ dup >xs o:.formWin!
  ts> x:.setContent^
  360 280 x:.resize^
  x:.setHideOnClose^
  x:drop
;

/*
Bottom bar: prompt, Show Form, Event labels.
( -- bar ) o( inst )
*/
: WidgetsApp#makeBottomBar
  "—" Label.new^ o:.eventDetail!
  (
    "Press to see a (non-modal) form" Label.new^
    o:.onShowForm~ 0 "Show Form" Button.new^
    "Event:" Label.new^
    o:.eventDetail@
  ) Box.newH^
;

/*
Make the example MenuBar (in-window; not fyne MainMenu).
( -- bar:MenuBar )
*/
: WidgetsApp#makeMenuBar
  o:.onMenuItemSelected~ >ts
  ( 
    (
      t:peek 0 "New" MenuItem.new^
      t:peek 1 "Open..." MenuItem.new^
      MenuItem.newSeparator^
      t:peek 2 "Save" MenuItem.new^
    ) "File" Menu.new^
    
    (
      t:peek 3 "About..." MenuItem.new^
    ) "Help" Menu.new^
    
  ) MenuBar.new^
  t:drop
;

/*
Build main UI.
( -- ) o( inst )
*/
: WidgetsApp#makeUi
  AppTabs.new^ >ts
  
  "Basic" o:.makeBasicTab^ t:.append^
  "Layouts" o:.makeLayoutsTab^ t:.append^
  "RichText" o:.makeRichTextTab^ t:.append^
  "Border" o:.makeBorderTab^ t:.append^
  "List" o:.makeListTab^ t:.append^
  "Table" o:.makeTableTab^ t:.append^
  "Tree" o:.makeTreeTab^ t:.append^
  "Misc" o:.makeMiscTab^ t:.append^
  
  (
    ts>
    o:.makeBottomBar^
  ) Box.newV^
  >ts
  (
    o:.makeMenuBar^
    Separator.new^
    ts>
  )
  Box.newV^
  
  App.app "Final Fyne Widgets Demo"
  Window.new^ dup >xs o:.win!
  x:.setContent^
  720 560 x:.resize^
  
  o:.makeFormWindow^
  
  x:drop
;

: widgets-fyne
  WidgetsApp.new^ >zs
  z:.makeUi^
  z:.win@.show^
  z:drop
;

widgets-fyne-blurb $ nl
