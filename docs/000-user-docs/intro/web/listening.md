# Listening (Serving)

**Listening** means binding an HTTP server and handling inbound requests. Final’s model is:

1. Create an `HttpListener` and set bind `address` (host + port).
2. For each route: allocate a **request Pipe**, **spawn** a handler Pip that loops on that Pipe, then **register** the ServeMux pattern with `.handle^` (or `.get^` / `.post^` / …).
3. `spawn:` a server Pip that calls `.listen^` — never block the main / REPL pip on listen.
4. Per request, the runtime pushes an `HttpServeJob` on the route’s request Pipe; the handler writes `.res`, then signals `.reply`.

Routing uses Go’s `http.ServeMux` patterns (`METHOD /path/{name}`). Matching, specificity, and default 404/405 follow ServeMux — Final does not implement its own matcher. Path wildcards land on `HttpInReq.pathParams` (e.g. `GET /api/items/{id}` → `x:.req.pathParams.id@`). Query keys land on `HttpInReq.queryParams` (`?id=1`). They are not interchangeable: `/api/items/1` matches a `{id}` route; `/api/items/?id=1` does not turn into a path param.

---



## Objects



### HttpListener

Stateful server. You must `new^` — listening is not a one-shot helper.


| Kind   | Name            | Type / Effect                                    | Description                                       |
| ------ | --------------- | ------------------------------------------------ | ------------------------------------------------- |
| class  | `new`           | `( -- listener )`                                | empty listener; owns ServeMux + route table       |
| prop   | `address`       | HttpAddress                                      | bind address (mutate `.scheme`, `.host`, `.port`) |
| prop   | `routes`        | List                                             | registered `HttpRoute` entries (read-only)        |
| prop   | `readTimeoutS`  | Int                                              | `0` → no extra read timeout                       |
| prop   | `writeTimeoutS` | Int                                              | `0` → no extra write timeout                      |
| method | `handle`        | `( pattern:s name:s xt:XT requestPipe:Pipe -- )` | full ServeMux pattern, e.g. `"GET /health"`       |
| method | `get`           | `( path:s name:s xt:XT requestPipe:Pipe -- )`    | same as `handle` with `"GET "` prepended          |
| method | `post`          | `( path:s name:s xt:XT requestPipe:Pipe -- )`    | `"POST "` + path                                  |
| method | `put`           | `( path:s name:s xt:XT requestPipe:Pipe -- )`    | `"PUT "` + path                                   |
| method | `delete`        | `( path:s name:s xt:XT requestPipe:Pipe -- )`    | `"DELETE "` + path                                |
| method | `listen`        | `( -- true | err:s false )`                      | bind + serve; **blocks the calling pip**          |
| method | `shutdown`      | `( -- true | err:s false )`                      | stop from another pip                             |


`address.port` is **required** before listen. Empty host defaults to `127.0.0.1`. For other methods (e.g. PATCH), pass a full pattern to `.handle^`.

### HttpServeJob

Per-request job object pushed on the route’s request Pipe. There is no public `new`; only the listener creates jobs.


| Kind | Name    | Type       | Description                                      |
| ---- | ------- | ---------- | ------------------------------------------------ |
| prop | `req`   | HttpInReq  | inbound request                                  |
| prop | `res`   | HttpOutRes | outbound response (write this)                   |
| prop | `reply` | Pipe       | **per-request**; push a done token when finished |


**Always** signal `.reply` (including error paths), or the Go side hangs waiting for the response flush.

### HttpInReq

Read-focused inbound request (no public `new`).


| Kind   | Name              | Type / Effect                                       | Description                              |
| ------ | ----------------- | --------------------------------------------------- | ---------------------------------------- |
| prop   | `method`          | String                                              | e.g. `"GET"`                             |
| prop   | `path`            | String                                              | path only (no query), leading `/`        |
| prop   | `pathParams`      | DynamicObject                                       | one String per `{name}` from the pattern |
| prop   | `queryParams`     | DynamicObject                                       | one value per key (no multi-value)       |
| prop   | `headers`         | StringKeyValueList                                  | request headers                          |
| prop   | `host`            | String                                              | Host header / host                       |
| prop   | `remoteAddr`      | String                                              | peer address                             |
| method | `getBodyAsString` | `( -- s )`                                          | body as UTF-8 text                       |
| method | `getBodyAsBytes`  | `( -- bytes:ByteList )`                             | body as Bytes                            |
| method | `getFormParams`   | `( -- form:StringKeyValueList true | err:s false )` | `application/x-www-form-urlencoded` only |




### HttpOutRes

Writable outbound response. Flushed after the reply Pipe signals. Status `0` → **200** at flush. Status is **not** a header.


| Kind   | Name                | Type / Effect           | Description                          |
| ------ | ------------------- | ----------------------- | ------------------------------------ |
| prop   | `status`            | Int                     | HTTP status line; unset/`0` → 200    |
| prop   | `headers`           | StringKeyValueList      | response headers                     |
| method | `setBodyFromString` | `( s:String -- )`       | set body (does not set Content-Type) |
| method | `setBodyFromBytes`  | `( bytes:ByteList -- )` | set body from Bytes                  |
| method | `setBodyAsJsonString` | `( json:s -- )`       | JSON string body + `Content-Type: application/json` |
| method | `clearBody`         | `( -- )`                | clear buffered body                  |


---



## Concurrency model (Pip-per-route)

Native ServeHTTP **never** spawns Final pips. You spawn handlers in Final **before** listen:

```text
setup pip
  → spawn handler loop on requestPipe[i]
  → register pattern → stub closed over i
  → spawn: serverPip → listener.listen^

per HTTP request (Go stub):
  → new reply Pipe
  → push HttpServeJob(req, res, reply) on requestPipe[i]
  → wait on reply → flush .res → return

handler pip:
  → pipe> job → write .res → push done on .reply
```

**One serial loop Pip per route.** Give the request Pipe capacity (e.g. `Pipe.newWithCapacity^`) so bursts can queue while a handler is busy.

Server pip: `.listen^` blocks **that** pip. Main continues; call `.shutdown^` from another pip to stop.

---



## ServeMux patterns and REST params


| Pattern               | Notes                            |
| --------------------- | -------------------------------- |
| `GET /health`         | method + exact path              |
| `GET /api/items/{id}` | path param → `x:.req.pathParams.id@` |
| `POST /api/items`     | method-specific                  |
| `PUT /api/items/{id}` | via `.put^` or `.handle^`        |



| Source          | Access                                                                  |
| --------------- | ----------------------------------------------------------------------- |
| Path `{id}`     | `x:.req.pathParams` DynamicObject (Strings; coerce with `s>i` etc.)     |
| Query           | `x:.req.queryParams` DynamicObject                                      |
| Form urlencoded | `x:.req.getFormParams^`                                                 |
| Headers         | `x:.req.headers` / `x:.res.headers` StringKeyValueList                  |
| Body            | `getBodyAsString` / `getBodyAsBytes`; parse JSON with `json>` if needed |


---



## Example

Matches `examples/http/listen/http-listen_0-simple.fnl` (health + path-param GET).

```forth
20 Pipe.newWithCapacity^ const: healthReqs
20 Pipe.newWithCapacity^ const: getItemReqs

: healthLoop  ( pp:DynamicObject -- )
  >zs
  z:.requestPipe@ >ys
  forever
    y:peek pipe> >xs          // HttpServeJob
    200 x:.res.status!
    "text/plain" "Content-Type" x:.res.headers!{}
    "ok" x:.res.setBodyFromString^
    0 x:.reply@ >pipe         // done token → flush
    x:drop
  /forever
;

: getItemLoop  ( pp:DynamicObject -- )
  >zs
  z:.requestPipe@ >ys
  forever
    y:peek pipe> >xs
    200 x:.res.status!
    "text/plain" "Content-Type" x:.res.headers!{}
    "id was: " x:.req.pathParams.id@ str::+ x:.res.setBodyFromString^
    0 x:.reply@ >pipe
    x:drop
  /forever
;

HttpListener.new^ const: myListener

: bootHandlers  ( -- )
  Pip.newPipParams^ >ts
  healthReqs t:.requestPipe!
  ts> spawn: healthLoop
  Pip.newPipParams^ >ts
  getItemReqs t:.requestPipe!
  ts> spawn: getItemLoop
;

: registerRoutes  ( -- )
  myListener >ts
  "GET /health"         "healthLoop"  "healthLoop"  find .xt nip  healthReqs  t:.handle^
  "GET /api/items/{id}" "getItemLoop" "getItemLoop" find .xt nip  getItemReqs t:.handle^
  t:drop
;

: runServer  ( pp -- )
  drop
  myListener >ts
  "http" t:.address.scheme!
  "127.0.0.1" t:.address.host!
  18080 t:.address.port!
  t:.listen^
  if
    "listener stopped\n" $
  else
    "listen ERROR: %s\n" fmt nip $
  then
  t:drop
;

: spawnServer  ( -- )
  Pip.newPipParams^ spawn: runServer
;

bootHandlers
registerRoutes
spawnServer
```

Then: `GET http://127.0.0.1:18080/health` and `GET http://127.0.0.1:18080/api/items/42`.

---



## Notes

- **Spawn before listen** — handler pips must be running before `.listen^`.
- **Reply Pipe** — capacity-1 Pipe in the job; push any done token (examples use `0`).
- **Status vs transport** — handler owns `.res.status` (4xx/5xx). Reply means “flush,” not success.
- **Security** — listener types participate in `SR.HttpListen`. Spawning server and handler pips needs `SR.Threads`. See [Security Restrictions](../embedding/security-restrictions.md).
- **Dialing** — outbound client APIs are separate; see [Dialing (Get/Post)](dialing.md).
- **Examples on disk** — `examples/http/listen/http-listen_0-simple.fnl`.

Use `help HttpListener` (and `HttpInReq`, `HttpOutRes`, `HttpServeJob`) in the TUI for the full member list.