forget: reset

/* HttpListen serial handler pips + ServeMux routes
"examples/http/listen/http-listen_0-simple.fnl" compileFromFile

Boot: spawn handler loops, register routes, spawn server pip.
Then GET  http://127.0.0.1:18080/health
     GET  http://127.0.0.1:18080/api/items/42
     POST http://127.0.0.1:18080/api/items
*/
/*
Request pipe for each endpoint handler.
Provide a buffer so Listener won't get blocked waiting for
a busy endpoint handler to service the requests.
*/
20 Pipe.newWithCapacity^ const: healthReqs
20 Pipe.newWithCapacity^ const: getItemReqs
20 Pipe.newWithCapacity^ const: postItemReqs

/*
Each endpoint handler is spawned into its own Pip.
A handler is a loop, waiting on a request, servicing it,
then informing the dispatcher the response is ready to 
be sent back to the caller.
( pp:DynamicObject -- )
*/
: healthLoop
  >zs // pip params for lifetime of handler
  z:.requestPipe@ >ys // req pipe on ys for lifetime
  forever
    y:peek pipe> >xs // HttpServeJob
    200 x:.res.status!
    "text/plain" "Content-Type" x:.res.headers!{}
    "ok" x:.res.setBodyFromString^
    0 x:.reply@ >pipe
    x:drop
  /forever
;

// ( pp:DynamicObject -- )
: getItemLoop
  >zs
  z:.requestPipe@ >ys
  forever
    y:peek pipe> >xs
    //TODO: jsonize this, using setBodyAsJson, using the id
    200 x:.res.status!
    "text/plain" "Content-Type" x:.res.headers!{}
    "id was: " x:.req.pathParams.id@ str::+ x:.res.setBodyFromString^
    0 x:.reply@ >pipe
    x:drop
  /forever
;

// ( pp:DynamicObject -- )
: postItemLoop
  >zs
  z:.requestPipe@ >ys
  forever
    y:peek pipe> >xs
    // the body should be a json object, with no id
    "Content-Type" x:.req.headers@{}
    "application/json" == 
    if
      /*TODO: convert to json, set id to 43
      x:.req.getBodyAsJson^
      >ts  43 t:.id!  ts> 
      x.res.setBodyAsJson^
      // note: the setBodyAsJson should do this:
      "application/json" "Content-Type" x:.res.headers!{}
      */
      201 x:.res.status! // Created
    else
      415 x:.res.status! // Unsupported Media Type
    then
    0 x:.reply@ >pipe
    x:drop
  /forever
;

HttpListener.new^ const: myListener

/*
Launch the route handler Pips.
These should exist and be running before calling listen.
( )
*/
: bootHandlers
  Pip.newPipParams^ >ts
  healthReqs t:.requestPipe!
  ts> spawn: healthLoop
  Pip.newPipParams^ >ts
  getItemReqs t:.requestPipe!
  ts> spawn: getItemLoop
;

/*
Register the routes, specify the spawn handler.
See https://pkg.go.dev/net/http#ServeMux for patterns.
( )
*/
: registerRoutes
  myListener >ts
  "GET /health"         "healthLoop"  "healthLoop"  find .xt nip  healthReqs  t:.handle^
  "GET /api/items/{id}" "getItemLoop" "getItemLoop" find .xt nip  getItemReqs t:.handle^
//"POST /api/items"     "postItemLoop" "postItemLoop" find .xt nip  postItemReqs t:.handle^
  t:drop
;

/*
Run the server.
A server must be a Pip, because it blocks when it enters the 
listen code.
( pp:PipParams )
*/
: runServer
  drop // we don't use pp
  myListener >ts
  "http" t:.address.scheme!
  "127.0.0.1" t:.address.host!
  18080 t:.address.port!
  t:.listen^ //  enters the Go Serve loop, not expected to exit
  if // normal stop
    "listener stopped\n" $
  else // error occured trying to listen
    ( peek )
    "listen ERROR: %s\n" fmt
    nip $
  then
  t:drop
;

: spawnServer
  Pip.newPipParams^ spawn: runServer
;

bootHandlers
registerRoutes
spawnServer
"listening on http://127.0.0.1:18080 — GET /health and /api/items/{id}\n" $
