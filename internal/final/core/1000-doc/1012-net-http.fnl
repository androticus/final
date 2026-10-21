// HTTP networking

/*
Host address for an HttpReq: scheme, host, and optional port.
*/
:doc HttpAddress

/*
Create an empty HttpAddress.
( -- addr:HttpAddress )
*/
:doc HttpAddress#new

/*
URL scheme, typically "http" or "https".
Empty means https when dialing.
*/
:doc HttpAddress#scheme

/*
Host name or IP.
Required before send.
*/
:doc HttpAddress#host

/*
Port number. 0 means the scheme default
(80 for http, 443 for https).
*/
:doc HttpAddress#port

/*
Outgoing HTTP request.
Set hostAddress, then path/method/headers/body as needed.
*/
:doc HttpReq

/*
Create a new empty HttpReq with a fresh HttpAddress,
empty query params, and empty headers.
( -- req:HttpReq )
*/
:doc HttpReq#new

/*
Request timeout in seconds. 0 means no timeout
(or the advisory run budget when dialing).
*/
:doc HttpReq#timeoutS

/*
Target address (read-only property; mutate its fields).
*/
:doc HttpReq#hostAddress

/*
HTTP method such as "GET" or "POST".
Empty: POST when a body is set, else GET.
*/
:doc HttpReq#method

/*
URL path, e.g. "/v1/users".
Leading slash is optional.
*/
:doc HttpReq#path

/*
Ordered query parameters (StringKeyValueList).
*/
:doc HttpReq#queryParams

/*
Ordered request headers (StringKeyValueList).
*/
:doc HttpReq#headers

/*
Set the request body from a String (UTF-8 text).
Does not set Content-Type; set headers yourself, or use a
WithContentType helper when available.
( s:String -- )
*/
:doc HttpReq#setBodyFromString

/*
Set the request body from Bytes.
In Final, Bytes are always carried as a ByteList.
Does not set Content-Type (binary has no default) —
set Content-Type explicitly for the media type.
( bytes:ByteList -- )
*/
:doc HttpReq#setBodyFromBytes

/*
Set the request body from a TextRead (UTF-8 stream).
Does not set Content-Type; set headers yourself.
( r:TextRead -- )
*/
:doc HttpReq#setBodyFromTextRead

/*
Set the request body from a ByteRead stream.
Does not set Content-Type — set it for the media type.
( r:ByteRead -- )
*/
:doc HttpReq#setBodyFromByteRead

/*
Clear the request body (useful before a GET reuse).
( -- )
*/
:doc HttpReq#clearBody

/*
Append path segments from a List of String/Int/Uint.
( segments:List -- )
*/
:doc HttpReq#appendPathFromList

/*
Performs HTTP sends. Usually you do not keep one around —
prefer httpDial for a one-shot send.
*/
:doc HttpDialer

/*
Create a new HttpDialer.
( -- dialer:HttpDialer )
*/
:doc HttpDialer#new

/*
Send the request.
Transport errors return an error String and false;
HTTP status codes (4xx/5xx) still return an HttpRes and true —
check .status yourself.
( req:HttpReq -- res:HttpRes true | err:s false )
*/
:doc HttpDialer#send

/*
HTTP response: status, headers, and body.
*/
:doc HttpRes

/*
HTTP status code, e.g. 200 or 404.
*/
:doc HttpRes#status

/*
Response headers as a StringKeyValueList.
*/
:doc HttpRes#headers

/*
Body as Bytes (Final carrier: ByteList).
( -- bytes:ByteList )
*/
:doc HttpRes#getBodyAsBytes

/*
Body as a UTF-8 String.
( -- s:String )
*/
:doc HttpRes#getBodyAsString

/*
Body as a TextRead (UTF-8 stream over the response bytes).
( -- r:TextRead )
*/
:doc HttpRes#getBodyAsTextRead

/*
Body as a ByteRead (byte stream over the response bytes).
( -- r:ByteRead )
*/
:doc HttpRes#getBodyAsByteRead

/*
Send an HttpReq with a fresh HttpDialer.
( req:HttpReq -- res:HttpRes true | err:s false )
*/
:doc httpDial

/*
Print an httpDial result to stdout.
On success prints status and body string;
on error prints the error message.
( res:HttpRes true | err:s false -- )
*/
:doc printHttpRes
