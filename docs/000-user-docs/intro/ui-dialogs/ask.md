# Simple Ask Dialogs

The Ask Dialog api provides a simple means for collecting from 1 to 7 string-typed values from the user, with a default prompt message.

The following example is in `examples/dialog/ask-demo.fnl`.

```
: ask-demo
  ( "What is your name?" "What is your quest?" )
  Dialog.ask^ if
    "Nice to meet you, %s. Off you go — \"%s\" is a fine quest.\n" fmt $
  else
    "Then we shan't proceed. Nobody expects the Spanish Inquisition.\n" $
  then
;
```
The code provides the requested field labels using the convenient Paren operators for building a List. It then asks Dialog to post the dialog via the `ask` method, which returns one of two types of result shape:
* `true` then a List of the replies, in order of request
* `false` and no List

You test for the response type using an `if` statement, and if `true` can then pass the List directly to a `fmt` statement and print it out (or use the replies however otherwise). If `false`, you can print a default reply, as above.

`Dialog` is covered in more detail in [Input/UI Dialogs](../ui-dialogs/index.md),<br>Paren operators in [Compound Builders](../compounds/compound-builders.md), and<br>`fmt` in [Output Formatting](../out-formatting.md).
