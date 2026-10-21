forget: reset-defer-demo
`
This demo shows how to user defer: to create a word
whose runtime behavior can be set to be that of
any colon word.
It is used (sparingly) in two general cases:
* recursive logic, where the word needs to be defined for
  one of the recursive callers -- an example is json parsing,
  which needs parseJsonValue to be defined, because it is
  called in parseArray, parseObject, and parseJsonValue itself
* in base/extended behavior logic, when you start with
  a word that has basic behavior, then after compiling more
  code, extend it with additional behavior
The initial definition of the word will panic if executed,
so you must either insure its value is assigned with "is"
prior to first call.

Try (in this order):

defer-demo-crash
defer-demo-base
defer-demo-more
` const: defer-demo-blurb

defer: defer-demo-word

: defer-demo-crash
  "This will crash, because defer-demo-word is empty\n" $
  defer-demo-word
;

: defer-demo-word-base
  "Base word behavior prints this message\n" $
;

: defer-demo-base
  &of defer-demo-word-base is defer-demo-word
  defer-demo-word
;

: defer-demo-word-more
  "More word behavior prints this message\n" $
;

: defer-demo-more
  &of defer-demo-word-more is defer-demo-word
  defer-demo-word
;

defer-demo-blurb $ nl
