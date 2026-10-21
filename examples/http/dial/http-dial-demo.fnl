forget: reset-http-dial-demo
`
Simple web sites for Dialer test.
Advanced web sites.
Stream-based web sites (HttpReq/HttpRes TextRead & ByteRead).

Try:

Simple:
dialCheckIp
dialZen
dialBad (then wait for timeout)

Advanced:
dialWttr
dialBingoJsonReq

Streams:
dialStreamReqText   — POST /post body from TextRead (string)
dialStreamReqByte   — POST /upload body from ByteRead (ByteList)
dialStreamResText   — GET /stream/3 drain TextRead (first lines)
dialStreamResByte   — GET /image/png drain ByteRead (PNG magic)

` const: http-dial-demo-blurb

// Request Factories

// ( -- req:HttpReq )
: makeCheckIpReq
  HttpReq.new^
  >ts
  "checkip.amazonaws.com" t:.hostAddress.host!
  10 t:.timeoutS!
  ts>
;

: makeZenReq
  HttpReq.new^
  >ts
  10 t:.timeoutS!
  "api.github.com" t:.hostAddress.host!
  "zen" t:.path!
  ts>
;

: makeBadReq
  HttpReq.new^
  >ts
  4 t:.timeoutS!
  "badbadbadbad123456.com" t:.hostAddress.host!
  ts>
;
/*
// ( res:HttpRes true | err:s false -- ) prints res string, or err string
: printHttpRes
  if
    >ts
    ( t:.status t:.getBodyAsString^ )
    "(status: %d): %s" fmt
    t:drop
  else // error
    ( peek )
    "ERROR: %s" fmt
    nip
  then
  $ nl
;


//( req:HttpReq -- res:HttpRes true | err:s false )

: httpDial
  HttpDialer.new^
  >ts
  t:.send^
  t:drop
  printHttpRes
;
*/

// Dialers

: dialCheckIp
  makeCheckIpReq
  httpDial
  printHttpRes
;

: dialZen
  makeZenReq
  httpDial
  printHttpRes
;

: dialBad
  makeBadReq
  httpDial
  printHttpRes
;

// Advanced

// ( -- req:HttpReq )
: makeWttrReq
  HttpReq.new^
  >ts
  "wttr.in" t:.hostAddress.host!
  "Denver" t:.path!
  "3" "format" t:.queryParams!{}
  "text/plain" "Accept" t:.headers!{}
  10 t:.timeoutS!
  ts>
;

// Json result

: makeBingoJsonReq
  HttpReq.new^
  >ts
  "httpbingo.org" t:.hostAddress.host!
  "anything/hello" t:.path!
  "bar" "foo" t:.queryParams!{}
  "application/json" "Accept" t:.headers!{}
  10 t:.timeoutS!
  ts>
;

// Dialers

: dialWttr
  makeWttrReq
  httpDial
  printHttpRes
;

: dialBingoJsonReq
  makeBingoJsonReq
  httpDial dup2
  printHttpRes
  if 
    // assume it worked, drop the `true` flag
    >ts t:.getBodyAsString^ drop t:drop
  else
    drop
  then
;

// Streams 2x2 for Req and Res, Text and Byte
// Tier 5–6: docs/006-research/05-dialer-test-websites.md

// ( -- req:HttpReq ) POST text/plain from a TextRead
: makeStreamReqText
  HttpReq.new^
  >ts
  "httpbingo.org" t:.hostAddress.host!
  "post" t:.path!
  "POST" t:.method!
  "text/plain" "Content-Type" t:.headers!{}
  15 t:.timeoutS!
  "hello stream" TextReader.newFromString^ drop
  t:.setBodyFromTextRead^
  ts>
;

// ( -- req:HttpReq ) POST octet-stream from a ByteRead over a ByteList
: makeStreamReqByte
  ByteList.new^ >xs
  0 x:.append^
  1 x:.append^
  2 x:.append^
  255 x:.append^
  HttpReq.new^
  >ts
  "httpbingo.org" t:.hostAddress.host!
  "upload" t:.path!
  "POST" t:.method!
  "application/octet-stream" "Content-Type" t:.headers!{}
  15 t:.timeoutS!
  x:peek ByteReader.newFromBytes^ drop
  t:.setBodyFromByteRead^
  ts>
  x:drop
;

// ( -- req:HttpReq ) GET NDJSON-ish stream (drain as TextRead)
: makeStreamResText
  HttpReq.new^
  >ts
  "httpbingo.org" t:.hostAddress.host!
  "stream/3" t:.path!
  15 t:.timeoutS!
  ts>
;

// ( -- req:HttpReq ) GET PNG (drain as ByteRead)
: makeStreamResByte
  HttpReq.new^
  >ts
  "httpbingo.org" t:.hostAddress.host!
  "image/png" t:.path!
  15 t:.timeoutS!
  ts>
;

: dialStreamReqText
  makeStreamReqText
  httpDial
  printHttpRes
;

: dialStreamReqByte
  makeStreamReqByte
  httpDial
  printHttpRes
;

// Print status and the first two lines via getBodyAsTextRead.
: dialStreamResText
  makeStreamResText
  httpDial
  if
    >ts
    ( t:.status )
    "(status: %d)" fmt $ nl
    t:.getBodyAsTextRead^ >xs
    x:.readLine^ if
      ( peek )
      "line1: %s" fmt $ nl
      drop
    else
      "(no line1)" $ nl
    then
    x:.readLine^ if
      ( peek )
      "line2: %s" fmt $ nl
      drop
    else
      "(no line2)" $ nl
    then
    x:.close^
    x:drop
    t:drop
  else
    ( peek )
    "ERROR: %s" fmt
    nip
    $ nl
  then
;

// Print status, PNG magic (4 bytes), and drained length via getBodyAsByteRead.
: dialStreamResByte
  makeStreamResByte
  httpDial
  if
    >ts
    ( t:.status )
    "(status: %d)" fmt $ nl
    t:.getBodyAsByteRead^ >xs
    4 x:.readBytes^ >ys
    ( y:.len@ )
    "magic len=%d bytes: " fmt $
    0 y:@() $ " " $
    1 y:@() $ " " $
    2 y:@() $ " " $
    3 y:@() $ nl
    y:.len@
    begin
      4096 x:.readBytes^ >zs
      z:.len@
      dup 0 == if
        drop
        z:drop
        true
      else
        +
        z:drop
        false
      then
    until
    ( peek )
    "drained total bytes=%d (expect 8090)" fmt $ nl
    drop
    y:drop
    x:.close^
    x:drop
    t:drop
  else
    ( peek )
    "ERROR: %s" fmt
    nip
    $ nl
  then
;

http-dial-demo-blurb $ nl
