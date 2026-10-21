// Crypto — digests, AEAD, keys, random (Phase 0 + 1.A)

/*
Class-only one-shot digests. Algorithm via class constants
(Hash.md5, Hash.sha256), same shape as DbProvider drivers.
MD5 is non-cryptographic — cheap object / cache keys only;
use SHA-256 for content-ids and integrity.
Requires SR.Crypto. Incremental update/digest is deferred.
eg: `` "abc" Hash.sha256 Hash.hash^ Hex.encode^ ``
See docs/003-design-notes/154-crypto.md .
*/
:doc Hash

/*
Algorithm name for MD5 (non-crypto checksum).
Pass as the algo argument to Hash.hash^ .
( -- algo:s )
*/
:doc Hash#md5

/*
Algorithm name for SHA-256 (default content-id hash).
Pass as the algo argument to Hash.hash^ .
( -- algo:s )
*/
:doc Hash#sha256

/*
One-shot digest of bytes.
Input is ByteList (preferred) or String (UTF-8 bytes).
Leaves a ByteList digest. Requires SR.Crypto.
eg: `` bytes Hash.sha256 Hash.hash^ ``
( bytes:ByteList|String algo:s -- digest:ByteList )
*/
:doc Hash#hash

/*
Hex encode / decode for digests and key bytes.
Requires SR.Crypto.
*/
:doc Hex

/*
Encode bytes as lowercase hex String.
Input is ByteList or String (UTF-8 bytes).
( bytes:ByteList|String -- hex:s )
*/
:doc Hex#encode

/*
Decode a hex String to ByteList. Panics on invalid hex.
( hex:s -- bytes:ByteList )
*/
:doc Hex#decode

/*
Standard Base64 encode / decode for wire / JSON bodies.
Requires SR.Crypto.
*/
:doc Base64

/*
Encode bytes as standard Base64 String.
( bytes:ByteList|String -- s )
*/
:doc Base64#encode

/*
Decode a standard Base64 String to ByteList. Panics on error.
( s -- bytes:ByteList )
*/
:doc Base64#decode

/*
Keyed message authentication (HMAC). Class-only; one-shot.
Distinct from Hash (needs a shared secret) and from Cipher (no encryption).
Requires SR.Crypto. See examples/crypto/1b-hmac-tag.fnl .
*/
:doc Mac

/*
Algo constant for HMAC-SHA-256.
( -- algo:s )
*/
:doc Mac#hmacSha256

/*
One-shot HMAC. Key is ByteList, String, or Key (uses key material).
eg: `` Mac.hmacSha256 secret bytes Mac.hash^ ``
( algo:s key bytes:ByteList|String -- tag:ByteList )
*/
:doc Mac#hash

/*
Constant-time verify of an HMAC tag.
eg: `` Mac.hmacSha256 secret bytes tag Mac.verify^ ``
( algo:s key bytes:ByteList|String tag:ByteList|String -- bool )
*/
:doc Mac#verify

/*
Key derivation (HKDF-SHA-256, Argon2id). Class-only.
derive takes a params DynamicObject; leaves ByteList
(wrap with Key.fromBytes^ for Cipher). Requires SR.Crypto.
See examples/crypto/1c-derive-key.fnl .
*/
:doc Kdf

/*
Algo constant for HKDF-SHA-256.
( -- algo:s )
*/
:doc Kdf#hkdfSha256

/*
Algo constant for Argon2id (passphrase → key).
( -- algo:s )
*/
:doc Kdf#argon2id

/*
Derive key bytes for algo from a params DynamicObject.

HKDF fields: ikm (req), length (req), info (opt→empty), salt (opt).
Argon2id fields: password (req), salt (req, ≥8 bytes), length (req);
optional time (def 3), memoryKiB (def 65536), parallelism (def 4).

eg: `` Kdf.hkdfSha256 params Kdf.derive^ ``
( algo:s params:DO -- bytes:ByteList )
*/
:doc Kdf#derive

/*
CSPRNG via Go crypto/rand (not math RNG).
Requires SR.Crypto.
eg: `` 32 SecureRandom.bytes^ ``
*/
:doc SecureRandom

/*
Cryptographically secure random bytes.
( n:i -- bytes:ByteList )
*/
:doc SecureRandom#bytes

/*
Algorithm-tagged key material (symmetric or asymmetric half).
generate^ leaves one Key, or priv then pub for Ed25519 / X25519.
Requires SR.Crypto. See examples/crypto/1a-seal-and-sign.fnl .
*/
:doc Key

/*
Algo constant for AES-256-GCM keys (32 bytes).
( -- algo:s )
*/
:doc Key#aes256gcm

/*
Algo constant for ChaCha20-Poly1305 keys (32 bytes).
( -- algo:s )
*/
:doc Key#chacha20poly1305

/*
Algo constant for Ed25519 (generate → priv pub).
( -- algo:s )
*/
:doc Key#ed25519

/*
Algo constant for X25519 (generate → priv pub; agree → shared secret).
( -- algo:s )
*/
:doc Key#x25519

/*
Generate key material for algo.
Symmetric: one Key. Ed25519 / X25519: private then public.
( algo:s -- key ) or ( algo:s -- priv pub )
*/
:doc Key#generate

/*
Wrap bytes as a Key.
Symmetric: 32 bytes. Ed25519: 64 private or 32 public.
X25519: 32 bytes as private (use fromPublic for public).
( algo:s bytes:ByteList|String -- key )
*/
:doc Key#fromBytes

/*
Wrap bytes as a public Key (Ed25519 / X25519, 32 bytes).
( algo:s bytes:ByteList|String -- key )
*/
:doc Key#fromPublic

/*
Export key material as ByteList.
( -- bytes:ByteList )
*/
:doc Key#toBytes

/*
Ed25519 sign (private key receiver).
( msg:ByteList|String -- sig:ByteList )
*/
:doc Key#sign

/*
Ed25519 verify (public key receiver).
( msg:ByteList|String sig:ByteList|String -- bool )
*/
:doc Key#verify

/*
X25519 ECDH (private key receiver; peer is public Key).
( peerPub:Key -- shared:ByteList )
*/
:doc Key#agree

/*
AEAD cipher bound to algo + Key (AES-256-GCM or ChaCha20-Poly1305).
encrypt / decrypt take a params DynamicObject.
Requires SR.Crypto.
*/
:doc Cipher

/*
Algo constant for AES-256-GCM.
( -- algo:s )
*/
:doc Cipher#aes256gcm

/*
Algo constant for ChaCha20-Poly1305.
( -- algo:s )
*/
:doc Cipher#chacha20poly1305

/*
Construct Cipher from matching Key.
eg: `` Key.aes256gcm Key.generate^ Cipher.aes256gcm swap Cipher.new^ ``
( algo:s key:Key -- cipher )
*/
:doc Cipher#new

/*
Nonce size in bytes for this Cipher.
( -- n:i )
*/
:doc Cipher#nonceSize

/*
SecureRandom nonce of the correct length.
( -- nonce:ByteList )
*/
:doc Cipher#generateNonce

/*
Encrypt. Params DO fields: plaintext (req), nonce (req), aad (opt).
( params:DO -- ciphertext:ByteList )
*/
:doc Cipher#encrypt

/*
Decrypt. Params DO fields: ciphertext (req), nonce (req), aad (opt).
Modal: plaintext true | err:s false on auth failure.
( params:DO -- plaintext:ByteList true | err:s false )
*/
:doc Cipher#decrypt
