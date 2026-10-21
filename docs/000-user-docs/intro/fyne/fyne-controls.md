# Fyne Controls

This chapter is a succinct summary of the Layouts and Widgets available to use in Final. Most of these components are included in the example program `examples/gui/fyne/widgets-fyne.fnl`, which you are encouraged to peruse to see how to use each one. The SimGame framework examples also make extensive use of Fyne Events.

Most layouts and widgets also support `Canvas.visible` and `Canvas.refresh`. Interactive fields such as Button, Entry, Check, Select, RadioGroup, Slider, and Form also support `Disableable.disabled`.

## Layouts

A layout arranges the widgets it contains so you do not have to place them by hand. Box, Border, Grid, and FormLayout are factories that return a `Container`. Accordion, AppTabs, Scroll, and Split are layout types you create and keep as instances.

| Layout | Role |
| ------ | ---- |
| Accordion | Stack of titled sections that open to show detail |
| AppTabs | Row of titled tabs, each with its own content |
| Border | Five slots: top, bottom, left, right, and center |
| Box | Horizontal row or vertical stack |
| Container | Layout object returned by Box, Border, Grid, and FormLayout |
| FormLayout | Two-column label/widget pairs (no Submit/Cancel) |
| Grid | Equal cells in a fixed number of columns or rows |
| Scroll | Scrolling viewport around one child |
| Split | Two panes with a movable divider |

### Accordion

A stack of titled sections. Create an empty Accordion, then `append` each title and its detail content.

```
Accordion.new^ >ts
"Item 1" "Detail one" Label.new^ t:.append^
"Item 2" "Detail two" Label.new^ t:.append^
ts>
```

### AppTabs

A row of tabs. Create empty, then `append` each title and content. Read or set `selectedIndex` (0-based).

```
AppTabs.new^ >ts
"Basic"  o:.makeBasicTab^  t:.append^
"Border" o:.makeBorderTab^ t:.append^
ts>
```

### Border

Places up to five children: top, bottom, left, right, and center. Pass `null` for any unused slot. The center receives leftover space, which is useful for a Split or canvas that should grow with the window.

```
"TOP" Label.new^
"BOTTOM" Label.new^
"LEFT" Label.new^
"RIGHT" Label.new^
"CENTER" Label.new^
Border.new^
```

### Box

`Box.newV` stacks children vertically; `Box.newH` lays them out in a row. Both take a `List` of canvas controls and return a `Container`.

```
(
  "Hello" Label.new^
  "World" Label.new^
) Box.newV^
```

### Container

The layout object returned by Box, Border, Grid, and FormLayout. Place it as Window content or nest it inside another layout. Use `Canvas.visible` and `Canvas.refresh` when you need to hide or redraw it.

### FormLayout

A two-column form layout: label, then widget, repeated. The list must have an even length. Unlike the Form *widget*, FormLayout has no Submit or Cancel buttons.

```
(
  "Name" Entry.new^
  "Age"  Entry.new^
) FormLayout.new^
```

### Grid

Equal-sized cells. `Grid.newColumns` takes a column count and a list of children; `Grid.newRows` takes a row count.

```
2 (
  "A" Label.new^
  "B" Label.new^
  "C" Label.new^
  "D" Label.new^
) Grid.newColumns^
```

### Scroll

Wraps one child in a scrolling viewport. Use `setMinSize` when the viewport should stay a fixed size inside a larger layout.

```
(
  "Line 1" Label.new^
  "Line 2" Label.new^
  "Line 3" Label.new^
) Box.newV^ Scroll.new^ >ts
200 120 t:.setMinSize^
ts>
```

### Split

Two panes with a movable divider. `Split.newH` places leading and trailing side by side; `Split.newV` stacks top and bottom. Use `setOffset` with a value from 0 to 1 to set the divider position.

```
"Left" Label.new^
"Right" Label.new^
Split.newH^ >ts
0.4 t:.setOffset^
ts>
```

## Widgets

Widgets are the interactive or display controls you place inside layouts (or directly as Window content). Event handlers are TypedObject methods; pass a MethodRef (`o:.handlerName~`) and an integer index to the factory or setter that accepts a handler. When the event fires, your instance is on the O stack.

| Widget | Role |
| ------ | ---- |
| Activity | Animated busy indicator |
| Button | Clickable labeled control |
| CanvasImage | Image on the canvas, with optional mouse handlers |
| Check | Labeled checkbox |
| Entry | Single-line or multi-line text field |
| Form | Labeled fields with optional Submit and Cancel |
| Label | Static text |
| ListView | Scrolling list of strings (via ListRef) |
| RadioGroup | Exclusive radio choices |
| RichText | Plain or Markdown formatted text |
| Select | Drop-down list of string options |
| Separator | Thin dividing line |
| Slider | Numeric value between min and max |
| TableView | Grid of cells from a Table (via TableRef) |
| TreeView | Collapsible tree (via TreeRef / FyneTree) |

### Activity

An animated busy indicator. Call `start` to run it and `stop` to halt it.

```
Activity.new^ >ts
t:.start^
// …
t:.stop^
ts>
```

### Button

A clickable labeled button. The factory takes a MethodRef, an index, and the label text. Use `setText` to change the label later. `Button.newWithIcon` adds a theme icon from the `Icon` class (for example `Icon.Confirm`).

```
// handler: ( ev:Event -- ) o( inst )
: MyApp#onClick
  drop
  "Pressed!" o:.btn@.setText^
;

o:.onClick~ 0 "Click me" Button.new^
```

### CanvasImage

Shows a `Pixmap` or `Image` in the window. Call `refresh` after you mutate the underlying Pixmap. Optional mouse handlers cover down/up, hover, and scroll; see `examples/gui/fyne/mouse-fyne.fnl`.

```
"photo.png" Image.newFromFile^ CanvasImage.new^ >ts
// after drawing into a Pixmap shared with this image:
t:.refresh^
ts>
```

### Check

A labeled checkbox. Read or write `checked`. Attach `setOnChanged` to learn when the user toggles it; `Event.params[0]` is the new Bool.

```
"Enable" Check.new^ >ts
o:.onCheckChanged~ 0 t:.setOnChanged^
ts>
```

### Entry

A text field. `Entry.new` is single-line; `Entry.newMultiLine` wraps. Read or write `text`. Use `setOnChanged` while typing and `setOnSubmitted` when the user presses Enter on a single-line Entry; both deliver the String in `Event.params[0]`.

```
Entry.new^ >ts
o:.onEntryChanged~ 0 t:.setOnChanged^
o:.onEntrySubmitted~ 0 t:.setOnSubmitted^
ts>
```

### Form

A labeled list of fields with optional Submit and Cancel. `append` each label and widget, then `setOnSubmit` / `setOnCancel`. You can also `showModal` on a parent Window and wait for Ok or Cancel (see Dialog constants `Dialog.btnOk` / `Dialog.btnCancel`).

```
Form.new^ >ts
"Name" Entry.new^ t:.append^
"Subscribe" "Yes please" Check.new^ t:.append^
o:.onFormSubmit~ 0 t:.setOnSubmit^
o:.onFormCancel~ 0 t:.setOnCancel^
ts>
```

### Label

Static text. Create with `Label.new`, then update with `setText`.

```
"Hello from Fyne!" Label.new^ >ts
"Updated text" t:.setText^
ts>
```

### ListView

A scrolling list of strings. Wrap a `StringList` in a `ListRef`, then create the `ListView`. After you change `ListRef.list`, call `ListView.refresh`. `setOnSelected` delivers the row index in `Event.params[0]`.

```
( "Alpha" "Beta" "Gamma" s) ListRef.new^
ListView.new^ >ts
o:.onListSelected~ 0 t:.setOnSelected^
ts>
```

### RadioGroup

Exclusive radio choices from a `StringList`. The factory takes an optional MethodRef (pass `null` to skip the event), an index, and the options. Read `selected` for the chosen String (or `""` if none).

```
o:.onRadioChanged~ 0
( "One" "Two" "Three" s)
RadioGroup.new^
```

### RichText

Formatted text. `RichText.newWithText` takes plain text; `RichText.newFromMarkdown` parses Markdown. Update later with `setText` or `parseMarkdown`.

```
`
# Title

Some **bold** text.
` RichText.newFromMarkdown^
```

### Select

A drop-down of string options. Same factory shape as RadioGroup: MethodRef (or `null`), index, and a `StringList`. Read `selectedIndex` (0-based). The change handler receives the selected String in `Event.params[0]`.

```
o:.onSelectChanged~ 0
( "Red" "Green" "Blue" s)
Select.new^
```

### Separator

A thin dividing line between other controls.

```
(
  "Above" Label.new^
  Separator.new^
  "Below" Label.new^
) Box.newV^
```

### Slider

A numeric slider. Create with min, max, and step; read or write `value`. `setOnChanged` fires while dragging; `setOnChangeEnded` fires when the drag ends. Both deliver a Float in `Event.params[0]`.

```
0 100 1 Slider.new^ >ts
o:.onSliderChanged~ 0 t:.setOnChanged^
50 t:.value!
ts>
```

### TableView

A grid of cells from a Final `Table`. Wrap the table in a `TableRef`, then `TableView.new` with a Bool for whether to show column-name headers. Use `setColumnWidth` or `setColumnWidths` for sizing. After changing the table data, call `refresh`. Selection handlers receive row then column in `Event.params`.

```
myTable TableRef.new^
true TableView.new^ >ts
( 120 60 i) t:.setColumnWidths^
o:.onTableSelected~ 0 t:.setOnSelected^
ts>
```

### TreeView

A collapsible tree. Build a `FyneTree` (root id 0), add branch and leaf nodes, wrap it in a `TreeRef`, then create the `TreeView`. Selection delivers the node id in `Event.params[0]`. Refresh after changing the tree data.

```
FyneTree.new^ >ts
t:.root@ >xs
x:peek "Animals" t:.newBranchNode^ >ys
y:peek "Dog" t:.newLeafNode^ drop
y:drop x:drop
ts> TreeRef.new^
TreeView.new^ >ts
o:.onTreeSelected~ 0 t:.setOnSelected^
ts>
```
