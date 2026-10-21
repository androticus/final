# Simple Prompt Dialogs

A **prompt dialog** shows a short message with **Ok** and **Cancel**. Use it when you only need a confirmation, not typed input.

## API

| Member | Kind | Effect | Description |
| ------ | ---- | ------ | ----------- |
| `Dialog.showDefaultPrompt` | class method | `( body:s -- btn:i )` | show modal; return which button closed it |
| `Dialog.btnOk` | class const | `( -- 1 )` | Ok was chosen |
| `Dialog.btnCancel` | class const | `( -- 0 )` | Cancel was chosen (also Esc / Break) |

`showDefaultPrompt` always uses title `"Prompt"` and focuses **Ok** by default. You pass only the body string. The body is shown centered; keep it short (about three lines).

Compare the returned Int to `Dialog.btnOk` / `Dialog.btnCancel` (or use `==` / `<>` against those constants).

## Example

```forth
> "Overwrite notes.txt?" Dialog.showDefaultPrompt^
Dialog.btnOk == if
  "overwriting…" $ nl
else
  "left alone" $ nl
then

overwriting…
```

(In the run above, the user pressed **Ok**. Cancel or Esc would print `left alone`.)

A minimal check that only cares about Ok:

```forth
> "Continue?" Dialog.showDefaultPrompt^ Dialog.btnOk == $

true
```

## Notes

1. The call **blocks** your Final word until the dialog closes; the rest of the UI (menus, Break) still works.
2. For a simple way to collect one or a few string fields, use an [Ask Dialog](ask.md).
3. For collecting typed fields rather than a yes/no, use a [Field Dialog](field.md).
