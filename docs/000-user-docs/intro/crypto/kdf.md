# Kdf

Derive key bytes from IKM or a passphrase. Leaves **`ByteList`** — wrap with `Key.<algo> … Key.fromBytes^` for Cipher.

| Constant | Role |
| -------- | ---- |
| `Kdf.hkdfSha256` | After X25519 agree → AEAD key |
| `Kdf.argon2id` | Passphrase → key |

```forth
// ( algo:s params:DO -- bytes:ByteList )
Kdf.hkdfSha256 params Kdf.derive^
```

### HKDF-SHA-256 params

| Field | Required | Notes |
| ----- | -------- | ----- |
| `ikm` | yes | ByteList \| String |
| `length` | yes | Int (output bytes) |
| `info` | no | default empty |
| `salt` | no | optional |

```forth
{ shared "ikm" "final-aead" "info" 32 "length" }
Kdf.hkdfSha256 swap Kdf.derive^
Key.aes256gcm swap Key.fromBytes^
```

### Argon2id params

| Field | Required | Default if omitted |
| ----- | -------- | ------------------ |
| `password` | yes | — |
| `salt` | yes (≥ 8 bytes) | — |
| `length` | yes | — |
| `time` | no | 3 |
| `memoryKiB` | no | 65536 (64 MiB) |
| `parallelism` | no | 4 |

```forth
16 SecureRandom.bytes^ >xs
{
  "demo-passphrase" "password"
  x:peek "salt"
  32 "length"
}
Kdf.argon2id swap Kdf.derive^
```

Demos may lower `time` / `memoryKiB` / `parallelism` for speed; production should omit them (safe defaults) or set higher costs.

Example: `examples/crypto/1c-derive-key.fnl` (X25519 → HKDF → AES-GCM, plus Argon2id).
