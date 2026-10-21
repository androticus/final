# Dialing (Get/Post)

**Dialing** means making an outbound HTTP request and reading the response. The flow is always the same:

1. Create an **`HttpReq`** and set at least **`hostAddress.host`**.
2. Optionally set path, query, headers, method, timeout, and body.
3. Send with **`HttpDialer.send`** (or **`httpDial`**).
4. On transport success, inspect **`HttpRes.status`**, headers, and body.

Transport failures (DNS, timeout, TLS, …) use the modal shape `( -- err:s false )`. A completed HTTP exchange — including **4xx/5xx** — returns `( -- res:HttpRes true )`; check `.status` yourself.

---

## Objects

### HttpAddress

Target of the request: scheme, host, and optional port. Every **`HttpReq`** owns one under **`.hostAddress`** (read-only property — mutate its fields).

| Kind | Name | Type / Effect | Description |
| ---- | ---- | ------------- | ----------- |
| class | `new` | `( -- addr )` | empty address |
| prop | `scheme` | String | `"http"` or `"https"`; **empty → `https`** when dialing |
| prop | `host` | String | host name or IP; **required** before send |
| prop | `port` | Int | `0` → scheme default (80 / 443) |

Typical setup parks the request on **ts** and writes the nested host:

```forth
HttpReq.new^ >ts
"checkip.amazonaws.com" t:.hostAddress.host!
ts>
```

### HttpReq

Outgoing request. Only **`hostAddress.host`** is required; everything else has defaults.

| Kind | Name | Type / Effect | Description |
| ---- | ---- | ------------- | ----------- |
| class | `new` | `( -- req )` | fresh req with empty address, query list, and headers |
| prop | `timeoutS` | Int | seconds; `0` → no explicit timeout (see Notes) |
| prop | `hostAddress` | HttpAddress | owned address (mutate fields; property itself is read-only) |
| prop | `method` | String | e.g. `"GET"`, `"POST"`; **empty → `POST` if body set, else `GET`** |
| prop | `path` | String | URL path; leading `/` optional |
| prop | `queryParams` | StringKeyValueList | ordered query pairs |
| prop | `headers` | StringKeyValueList | ordered request headers |
| method | `setBodyFromString` | `( s:String -- )` | set body from UTF-8 text (does not set Content-Type) |
| method | `setBodyFromBytes` | `( bytes:ByteList -- )` | set body from Bytes (`ByteList` carrier; no Content-Type default) |
| method | `setBodyFromTextRead` | `( r:TextRead -- )` | set body from a text stream (does not set Content-Type) |
| method | `setBodyFromByteRead` | `( r:ByteRead -- )` | set body from a byte stream (no Content-Type default) |
| method | `setBodyAsJsonString` | `( json:s -- )` | set body from an already-rendered JSON string and `Content-Type: application/json` |
| method | `clearBody` | `( -- )` | clear body (reuse req for a later GET) |
| method | `appendPathFromList` | `( segments:List -- )` | join String/Int/Uint segments onto `.path` |

Query and headers use **`!{}`** (upsert first key) or **`.add^`** (always append a duplicate). See [StringKeyValueList](../compounds/lists/skvlist.md).

```forth
HttpReq.new^ >ts
"wttr.in" t:.hostAddress.host!
"Denver" t:.path!
"3" "format" t:.queryParams!{}
"text/plain" "Accept" t:.headers!{}
10 t:.timeoutS!
ts>
```

### HttpDialer

Stateless sender. Create one, call **`send`**, then drop it — or use **`httpDial`**, which does that for you.

| Kind | Name | Effect | Description |
| ---- | ---- | ------ | ----------- |
| class | `new` | `( -- dialer )` | new dialer |
| method | `send` | `( req:HttpReq -- res:HttpRes true \| err:s false )` | prepare + transport |

```forth
HttpDialer.new^ >ts
t:.send^          // consumes req from DS
t:drop
```

### HttpRes

Result of a successful transport. There is no public `new`; only **`send`** produces one.

| Kind | Name | Type / Effect | Description |
| ---- | ---- | ------------- | ----------- |
| prop | `status` | Int | HTTP status, e.g. 200, 404 |
| prop | `headers` | StringKeyValueList | one pair per header line (duplicates kept) |
| method | `getBodyAsString` | `( -- s:String )` | body as UTF-8 text |
| method | `getBodyAsBytes` | `( -- bytes:ByteList )` | body as Bytes (`ByteList` carrier) |
| method | `getBodyAsTextRead` | `( -- r:TextRead )` | body as a text stream |
| method | `getBodyAsByteRead` | `( -- r:ByteRead )` | body as a byte stream |

---

## Helpers

Core ships two convenience words (see `0068-http_util.fnl`):

| Word | Effect | Description |
| ---- | ------ | ----------- |
| `httpDial` | `( req:HttpReq -- res:HttpRes true \| err:s false )` | `HttpDialer.new^` + `.send^` |
| `printHttpRes` | `( res:HttpRes true \| err:s false -- )` | print status+body, or `ERROR: …` |

`printHttpRes` consumes the modal pair left by `httpDial` / `.send^`.

---

## Examples

These follow the same progression as `examples/http/dial/http-dial_0-simple.fnl` and `http-dial_1-advanced.fnl`.

### Simple GET (host only)

```forth
: makeCheckIpReq
  HttpReq.new^
  >ts
  "checkip.amazonaws.com" t:.hostAddress.host!
  10 t:.timeoutS!
  ts>
;

makeCheckIpReq httpDial printHttpRes
```

### GET with a path

```forth
: makeZenReq
  HttpReq.new^
  >ts
  10 t:.timeoutS!
  "api.github.com" t:.hostAddress.host!
  "zen" t:.path!
  ts>
;

makeZenReq httpDial printHttpRes
```

### Path, query, and Accept header

```forth
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

makeWttrReq httpDial printHttpRes
```

### JSON response body

`HttpRes` returns the body as text, bytes, or a stream. Parse JSON yourself with `json>` ([Json Parsing](../json/parsing.md)):

```forth
: makeBingoJsonRes
  HttpReq.new^
  >ts
  "httpbingo.org" t:.hostAddress.host!
  "anything/hello" t:.path!
  "bar" "foo" t:.queryParams!{}
  10 t:.timeoutS!
  ts>
;

makeBingoJsonRes httpDial
dup2 printHttpRes
if
  >ts t:.getBodyAsString^ json> t:drop
  // DynamicObject / List on DS — walk as usual
else
  drop          // discard err string; already printed
then
```

### POST JSON

Empty `.method` becomes **`POST`** once a body is set. Set **`Content-Type`** yourself:

```forth
HttpReq.new^ >ts
"httpbingo.org" t:.hostAddress.host!
"post" t:.path!
"application/json" "Content-Type" t:.headers!{}
`{"hello":"world","n":42}` t:.setBodyFromString^
10 t:.timeoutS!
ts>
httpDial printHttpRes
```

To reuse the same req for a later GET, call **`clearBody`** so the method default returns to GET.

### Bodies as streams

You can feed the request from a **TextRead** / **ByteRead**, and read the response the same way. Useful when the payload is already a stream, or when you want to pull the response incrementally ([Streams](../streams/index.md)).

**POST text from a String reader:**

```forth
HttpReq.new^ >ts
"httpbingo.org" t:.hostAddress.host!
"post" t:.path!
"POST" t:.method!
"text/plain" "Content-Type" t:.headers!{}
"hello stream" TextReader.newFromString^ drop
t:.setBodyFromTextRead^
15 t:.timeoutS!
ts>
httpDial printHttpRes
```

**Read a response as text lines:**

```forth
HttpReq.new^ >ts
"httpbingo.org" t:.hostAddress.host!
"stream/3" t:.path!
15 t:.timeoutS!
ts>
httpDial
if
  >ts
  t:.getBodyAsTextRead^ >xs
  x:.readLine^ if $ nl then
  x:.close^ x:drop
  t:drop
else
  drop
then
```

For binary bodies (images, opaque octets), use **`setBodyFromByteRead`** / **`getBodyAsByteRead`** with a **ByteList** reader — see `examples/http/dial/http-dial-demo.fnl` (`dialStreamReqByte`, `dialStreamResByte`).

### Transport error

A bad host yields the false branch (timeout / DNS), not an `HttpRes`:

```forth
HttpReq.new^ >ts
4 t:.timeoutS!
"badbadbadbad123456.com" t:.hostAddress.host!
ts>
httpDial printHttpRes
```

---

## Notes

* **Status vs transport** — `true` means the HTTP exchange finished. Always check `.status` for application success.
* **Timeout** — if `.timeoutS` is `0` at send time, the dialer may apply the session’s advisory run budget; set an explicit seconds value for predictable client timeouts.
* **No cookie jar** — each send is independent; cookies are not persisted across calls.
* **Security** — `HttpReq`, `HttpDialer`, and `HttpRes` participate in the **HttpDial** restriction. Restricted embeds / AI hosts may deny dialing; see [Security Restrictions](../embedding/security-restrictions.md).
* **Examples on disk** — `examples/http/dial/http-dial-demo.fnl`, `http-dial_0-simple.fnl`, `http-dial_1-advanced.fnl`.
