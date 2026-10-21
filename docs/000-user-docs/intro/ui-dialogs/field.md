# Field Dialogs

A **field dialog** asks the user to fill in a small form and returns the edited values (or Cancel). `Dialog.newFieldDialogParams^` mints a params [DynamicObject](../compounds/dynamicobject.md) with defaults; you customize title, body, and the nested `fields` object, then call `Dialog.showFieldDialog^`.

## API

| Member | Kind | Effect | Description |
| ------ | ---- | ------ | ----------- |
| `Dialog.newFieldDialogParams` | class method | `( -- params:DO )` | mint outer params with defaults and an empty `fields` object |
| `Dialog.showFieldDialog` | class method | `( params:DO -- fields:DO true \| false )` | validate, show modal; Ok → edited fields + true; Cancel → false only |
| `Dialog.btnOk` / `Dialog.btnCancel` | class const | `( -- i )` | same button constants as prompts; used for `defaultButton` |

### Outer params (after `newFieldDialogParams`)

| Field | Type | Default | Role |
| ----- | ---- | ------- | ---- |
| `title` | String | `"Information Request"` | dialog title; must be non-empty if you set it |
| `bodyText` | String | `""` | instruction line above the fields; empty hides it |
| `defaultButton` | Int | `Dialog.btnOk` | which button is the default focus (`btnOk` or `btnCancel`) |
| `fields` | DynamicObject | empty object | the form rows; fill this before show |

You may set `title`, `bodyText`, and `defaultButton` on the params DO, and add rows to `fields`, before calling `showFieldDialog`. The factory already created the nested `fields` object; park it (`t:.fields@ >xs`) and store your defaults there.

### Fields object

* **1 to 8** fields. An empty object panics at show time.
* **Creation order** is display order.
* Each **field name** is the label shown in the UI.
* The **default value’s type** chooses the control:

| Default type | Control | On Ok |
| ------------ | ------- | ----- |
| String | full-width text line | edited String |
| Bool | checkbox | Bool |
| Int | narrow text | parsed Int (invalid → error, dialog stays open) |
| Float | narrow text | parsed Float (invalid → error, dialog stays open) |

Any other default type panics when you show the dialog. User field names live only under `fields`, so they never clash with outer keys such as `title`.

On **Ok**, `showFieldDialog` pushes the **edited fields object** (not the outer params) and `true`. On **Cancel** / Esc / Break, it pushes only `false` — the modal [result shape](../clean-final-coding.md) you already use elsewhere.

## Example

Collect a small account form, then print what the user entered:

```forth
> Dialog.newFieldDialogParams^ >ts
"Account" t:.title!
"Enter name, age, and admin flag." t:.bodyText!
t:.fields@ >xs
"" x:.name!
0 x:.age!
false x:.admin!
x:drop
ts> Dialog.showFieldDialog^ not if
  "cancelled" $ nl
else
  >ts
  "name=" $ t:.name@ $ nl
  "age=" $ t:.age@ $ nl
  "admin=" $ t:.admin@ $ nl
  t:drop
then

name=Ada
age=36
admin=true
```

(Again, the printed lines assume the user filled those values and pressed **Ok**.)

A shorter pattern when defaults are enough and you only customize the title:

```forth
> Dialog.newFieldDialogParams^ >ts
"Connect" t:.title!
t:.fields@ >xs
"localhost" x:.host!
8080 x:.port!
x:drop
ts> Dialog.showFieldDialog^ not if
  "cancelled" $ nl
else
  >ts
  t:.host@ $ ":" $ t:.port@ $ nl
  t:drop
then

localhost:8080
```

## Notes

1. Mint params with `Dialog.newFieldDialogParams^`, fill `title` / `bodyText` / `defaultButton` and the nested `fields` object, then `Dialog.showFieldDialog^`.
2. Int/Float rows stay open with a red error if the text does not parse; fix the field and press Ok again.
3. Same rules as prompts: one dialog at a time; Break cancels.
4. For a plain yes/no with no fields, prefer [Simple Prompt Dialogs](prompt.md).
