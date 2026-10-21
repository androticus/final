# Cryptography

Final exposes a **small, modern crypto API**: digests, AEAD, signing, key agreement, MACs, KDFs, and a CSPRNG. Algorithms are selected like database drivers — **class constants** on one type per kind (`Hash.sha256`, `Cipher.aes256gcm`, …), not a plugin layer and not one Final type per algorithm.

| Type | Role |
| ---- | ---- |
| [Hashes](hashes.md) | Unkeyed digests (MD5 checksum, SHA-256) |
| [Hex / Base64](encoding.md) | Encode digests and key bytes for text |
| [SecureRandom](random.md) | Cryptographic random bytes |
| [Keys](keys.md) | Symmetric and asymmetric key material |
| [Ciphers](ciphers.md) | AEAD encrypt / decrypt |
| [Mac](mac.md) | HMAC-SHA-256 |
| [Kdf](kdf.md) | HKDF-SHA-256, Argon2id |

Policy: recommended algorithms only (no DES, RC4, RSA-PKCS1v1.5, …). APIs are **one-shot** (whole payload → result). Incremental `update` / `digest` is not available.

Privileged members require **`SR.Crypto`**. Examples live under `examples/crypto/`.

### Pattern

```forth
bytes Hash.sha256 Hash.hash^ Hex.encode^
Key.aes256gcm Key.generate^
Cipher.aes256gcm swap Cipher.new^
```

Multi-argument, algorithm-dependent inputs use a **params DynamicObject** (prefer `{ value name … }` for ad-hoc DynamicObjects).
