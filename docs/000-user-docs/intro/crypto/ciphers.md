# Ciphers (AEAD)

**`Cipher`** binds algorithm + matching **`Key`**. Encrypt/decrypt take a **params DynamicObject**.

| Constant | Notes |
| -------- | ----- |
| `Cipher.aes256gcm` | Prefer on hosts with AES-NI; 12-byte nonce |
| `Cipher.chacha20poly1305` | Strong default without AES-NI; 12-byte nonce |

| Kind | Name | Effect |
| ---- | ---- | ------ |
| class | `new` | `( algo:s key:Key -- cipher )` |
| method | `nonceSize` | `( -- n:i )` |
| method | `generateNonce` | `( -- nonce:ByteList )` |
| method | `encrypt` | `( params:DO -- ciphertext:ByteList )` |
| method | `decrypt` | `( params:DO -- plaintext true \| err:s false )` |

**Encrypt fields:** `plaintext` (req), `nonce` (req), `aad` (opt).  
**Decrypt fields:** `ciphertext` (req), `nonce` (req), `aad` (opt).  
Auth failure: no partial plaintext; modal `false`.

```forth
: demoAead
  msgBytes >zs
  Key.aes256gcm Key.generate^
  Cipher.aes256gcm swap Cipher.new^ >ts
  t:.generateNonce^ >xs
  { z:peek "plaintext" x:peek "nonce" } t:.encrypt^ >ys
  { y:peek "ciphertext" x:peek "nonce" } t:.decrypt^
  nif
    "decrypt failed: " $ $ nl
    error
  then
  // plaintext:ByteList
  t:drop x:drop y:drop z:drop
;
```

Same API for ChaCha20-Poly1305: `Key.chacha20poly1305` + `Cipher.chacha20poly1305`.

Full example: `examples/crypto/1a-seal-and-sign.fnl`.
