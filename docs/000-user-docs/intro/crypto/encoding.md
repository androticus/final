# Hex and Base64

Class-only helpers for digests, signatures, and key bytes on text channels. Require **`SR.Crypto`**.

### Hex

| Method | Effect |
| ------ | ------ |
| `Hex.encode` | `( bytes:ByteList\|String -- hex:s )` lowercase hex |
| `Hex.decode` | `( hex:s -- bytes:ByteList )` panics on invalid hex |

### Base64

Standard encoding (not raw/URL variant).

| Method | Effect |
| ------ | ------ |
| `Base64.encode` | `( bytes:ByteList\|String -- s )` |
| `Base64.decode` | `( s -- bytes:ByteList )` |

```forth
digest Hex.encode^
digest Base64.encode^
"deadbeef" Hex.decode^
```
