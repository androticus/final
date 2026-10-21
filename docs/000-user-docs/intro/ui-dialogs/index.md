# Input/UI Dialogs

When your Final code runs inside **`final` or `final-gui`**, you can raise modal dialogs that pause the script until the user chooses **Ok** or **Cancel**. 

These APIs live on the class-only type **`Dialog`**. The mechanisms used to build and show these dialogs are the same for both `final` and `final-gui`. (You do not need to know anything about [Fyne and final-gui](draw-gui/fyne-final-gui.md).)

There are three kinds:

| Kind | When to use | Entry |
| ---- | ----------- | ----- |
| **Prompt** | Yes/no confirmation with a short message | [Simple Prompt Dialogs](prompt.md) |
| **Ask** | Collect a handful of string responses, using a simple API | [Simple Ask Dialog](ask.md) |
| **Field** | Collect a handful of typed values (name, flag, number, …) using a more elaborated API than `ask` | [Field Dialogs](field.md) |

All three are **blocking**: the call does not return until the dialog closes. Esc (and Break / Ctrl-B) means Cancel. Only one dialog may be open at a time.

You already know the pieces these APIs lean on:

* **Class methods** with `^` — [Methods](../objects/methods.md)
* **Class constants** such as `Dialog.btnOk` — [Properties](../objects/properties.md)
* **fields** DynamicObjects for Field Dialogs — [DynamicObject](../compounds/dynamicobject.md)
* The **modal result** shape `( … -- result true \| false )` for Field Dialog Ok vs Cancel — [Writing Clean Final Code](../clean-final-coding.md)

The next two pages cover the exact stack effects, parameters, and short worked examples.
