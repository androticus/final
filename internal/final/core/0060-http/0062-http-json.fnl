/*
Final Methods for getting/setting body as JSON. 
A method depending on the correctness of incoming data
does not panic, it returns a modal `val true | false' result.
*/

/*
Set the HttpReq.body as a JSON String.
Convenience method that also sets 
  Content-Type: application/json
( json:s -- ) o( inst:HttpReq )
*/
: HttpReq#setBodyAsJsonString
  "application/json" "Content-Type" o:.headers!{}
  o:.setBodyAsString^
;

/*
Set the HttpOutRes body as JSON string.
Convenience method that also sets: 
  Content-Type: application/json
( json:s -- ) o( inst:HttpOutRes )
*/
: HttpOutRes#setBodyAsJsonString
  "application/json" "Content-Type" o:.headers!{}
  o:.setBodyAsString^
;

