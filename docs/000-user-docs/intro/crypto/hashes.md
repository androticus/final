# Hashes

Class-only **`Hash`**. One-shot digests.

| Constant | Use |
| -------- | --- |
| `Hash.md5` | **Non-cryptographic** checksum — cheap object / cache keys only |
| `Hash.sha256` | Default content-id / integrity hash |

```forth
// ( bytes:ByteList|String algo:s -- digest:ByteList )
"abc" Hash.sha256 Hash.hash^ Hex.encode^   // ba7816bf8f01cfea...
blob Hash.md5 Hash.hash^                   // 16-byte checksum
```

Input: **`ByteList`** preferred, or **`String`** as UTF-8 bytes. Output: **`ByteList`** digest; use [Hex / Base64](encoding.md) for text.
