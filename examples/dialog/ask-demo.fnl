forget: reset-ask-demo
`
Simple demo showing how to use Dialog.ask for quick
easy input from the user.

Try (both Ok and Cancel):

ask-demo
` const: ask-demo-blurb

: ask-demo
  ( "What is your name?" "What is your quest?" )
  Dialog.ask^ if
    "Nice to meet you, %s. Off you go — \"%s\" is a fine quest.\n" fmt $
  else
    "Then we shan't proceed. Nobody expects the Spanish Inquisition.\n" $
  then
;

ask-demo-blurb $ nl
