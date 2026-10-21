# Keys

**`Key`** is opaque, algorithm-tagged material. Capabilities depend on the bound algorithm.

| Constant | `generate^` leaves | Notes |
| -------- | ------------------ | ----- |
| `Key.aes256gcm` | one Key (32 bytes) | feeds `Cipher.aes256gcm` |
| `Key.chacha20poly1305` | one Key (32 bytes) | feeds `Cipher.chacha20poly1305` |
| `Key.ed25519` | **priv then pub** | `.sign^` / `.verify^` |
| `Key.x25519` | **priv then pub** | `.agree^` → shared `ByteList` |

| Kind | Name | Effect |
| ---- | ---- | ------ |
| class | `generate` | `( algo -- key )` or `( algo -- priv pub )` |
| class | `fromBytes` | `( algo bytes -- key )` |
| class | `fromPublic` | `( algo bytes -- key )` public half (Ed25519 / X25519) |
| method | `toBytes` | `( -- bytes:ByteList )` |
| method | `sign` | `( msg -- sig:ByteList )` Ed25519 private |
| method | `verify` | `( msg sig -- bool )` Ed25519 public |
| method | `agree` | `( peerPub:Key -- shared:ByteList )` X25519 private |

`fromBytes`: symmetric 32 bytes; Ed25519 64 = private, 32 = public; X25519 32 = private (use `fromPublic` for public).

### Ed25519

```forth
Key.ed25519 Key.generate^ >ys >xs   // y:pub x:priv
msg x:.sign^ >ts
msg t:peek y:.verify^               // bool
```

### X25519

```forth
Key.x25519 Key.generate^ >ys >xs    // y:pubA x:privA
Key.x25519 Key.generate^
>ts                                 // t:pubB
drop                                // drop privB if unused
t:peek x:.agree^                    // shared:ByteList — feed Kdf, then Cipher
```

See `examples/crypto/1a-seal-and-sign.fnl` and `1c-derive-key.fnl`.
