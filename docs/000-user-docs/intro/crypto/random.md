# SecureRandom

CSPRNG via Go `crypto/rand` — not `math` RNG. Use for nonces, salts, tokens, and key material.

```forth
// ( n:i -- bytes:ByteList )
32 SecureRandom.bytes^
16 SecureRandom.bytes^   // e.g. Argon2id salt
```

`Cipher.generateNonce^` draws a nonce of the correct length for that AEAD via the same source.
