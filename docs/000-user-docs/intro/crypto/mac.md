# Mac (HMAC)

Keyed authenticity **without** encryption. Distinct from Hash and Kdf. One-shot only.

| Constant | Algorithm |
| -------- | --------- |
| `Mac.hmacSha256` | HMAC-SHA-256 |

| Method | Effect |
| ------ | ------ |
| `Mac.hash` | `( algo key bytes -- tag:ByteList )` |
| `Mac.verify` | `( algo key bytes tag -- bool )` constant-time |

Key may be **`ByteList`**, **`String`**, or **`Key`** (uses key material).

```forth
Mac.hmacSha256 "demo-hmac-secret" msgBytes Mac.hash^ >ts
Mac.hmacSha256 "demo-hmac-secret" msgBytes t:peek Mac.verify^   // true
Mac.hmacSha256 "demo-hmac-secret" "tampered" t:peek Mac.verify^ // false
```

Example: `examples/crypto/1b-hmac-tag.fnl`. Prefer [Env / Moniker](../env.md) for real shared secrets.
