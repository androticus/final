forget: reset-config
`
An example of using Session.config.
Use a JSON string to initialize the Session.config object.
Uses a 2-step process for robustness, and Embedding compatibility.

1. Set your JSON string in sess.configJson
2. Call sess.initConfig
3. example the sess.config variable

Try:

init-config
print-config
`
const: config-blurb

: init-config
  // 1. put your JSON on session.configJson
  `{ "paramA" : 1, "paramB" : 2 }` sess.configJson!

  // 2. Call the init
  sess.initConfig^
;

: print-config
  ( sess.config.paramA@ sess.config.paramB@ )
  "paramA: %d  paramB: %d\n" fmt $
;


config-blurb $ nl
