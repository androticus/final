/*
Application Configuration
*/

/*
Initialize Session.config from Session.configJson.
If configJson is empty, does nothing.
Otherwise, it parses the json, and overwrites the
existing Session.config with the new DynamicObject.
It then clears Session.configJson.
( -- ) throws // invalid Json, or limits exceeded
*/
: Session#initConfig
  sess.configJson@ dup
  if // non-empty
    json>
    sess.setConfig^
    "" sess.configJson!
  else 
    drop
  then
;
