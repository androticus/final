// JSON encode and decode

/*
Render any Json-capable value to a compact String.
Uses each type's _renderJson support.
( root:a -- json:s )
*/
:doc >json

/*
Same as >json, but with indentation and newlines.
( root:a -- json:s )
*/
:doc >jsonp

/*
Parse a Json String into a Final value.
Objects become DynamicObject; arrays become List;
primitives become Bool, Null, String, Int, or Float.
Panics if the text is not valid Json.
( json:s -- root:a )
*/
:doc json>

/*
Lower-level entry used by json>.
Same stack effect; prefer json> in application code.
( json:s -- root:a )
*/
:doc parseJson

/*
Encode with an explicit pretty flag.
true → indented; false → compact.
Prefer >json / >jsonp for everyday use.
( root:a pretty:b -- json:s )
*/
:doc renderJson
