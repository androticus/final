# Web/Http

Final’s HTTP support covers **client-side dialing** and **server-side listening**:

* **Dialing** — build an **`HttpReq`**, send it with an **`HttpDialer`** (or the `httpDial` helper), and read an **`HttpRes`**. Everyday HTTPS GET/POST against remote hosts.
* **Listening** — bind an **`HttpListener`**, register ServeMux routes, spawn Pip handlers per route, and serve inbound requests.

You already know the pieces these APIs lean on:

* **Class methods** with `^`, property `@` / `!`, and nested paths like `t:.hostAddress.host!` — [Methods](../objects/methods.md), [Compound Path Syntax](../objects/compound-path-syntax.md)
* Parking a request or dialer on **ts** for a short session of member ops — [Writing Clean Final Code](../clean-final-coding.md)
* **StringKeyValueList** for dial query parameters and headers (duplicate keys allowed) — [StringKeyValueList](../compounds/lists/skvlist.md)
* The **modal result** shape `( … -- result true \| err:s false )` — [Writing Clean Final Code](../clean-final-coding.md)
* Parsing a JSON response body with `json>` — [Json](../json/index.md). To *send* JSON, `setBodyAsJsonString` on `HttpReq` / `HttpOutRes` takes an already-rendered JSON string and sets `Content-Type`. Request and response bodies can also be **TextRead** / **ByteRead** streams — [Dialing](dialing.md), [Streams](../streams/index.md).
* **Pips** and **Pipes** for listen-side concurrency — [Threading](../threading/index.md)

| Page | Contents |
| ---- | -------- |
| [Dialing (Get/Post)](dialing.md) | `HttpAddress`, `HttpReq`, `HttpDialer`, `HttpRes`, helpers, worked examples |
| [Listening (Serving)](listening.md) | `HttpListener`, `HttpServeJob`, ServeMux routes, path params, Pip-per-route handlers |

Use `help HttpReq` / `help HttpListener` (and the other type names) in the TUI for the full member list and stack effects.
