/*
X25519 agree → HKDF-SHA-256 → AES-256-GCM, plus Argon2id passphrase → key
(crypto Phase 1.C).

Reads message.txt beside this script. Writes derived-key hex and ciphertext
under _tmp/ (gitignored). Create that directory first: mkdir -p _tmp

Run from repo or Final/ install root:
  "examples/crypto/1c-derive-key.fnl" compileFromFile
*/

"examples/crypto/message.txt" const: msgPath
"_tmp/1c-aes-key.hex" const: aesKeyPath
"_tmp/1c-ciphertext.bin" const: ctPath
"_tmp/1c-nonce.hex" const: noncePath
"_tmp/1c-argon2-key.hex" const: argonKeyPath

0 var: derivedKey
0 var: sealedCt
0 var: sealedNonce

/*
Load plaintext bytes from the fixture file.
( -- bytes:ByteList )
*/
: loadMessage
  msgPath File.readBytes^
;

/*
Write bytes; clear error if _tmp is missing.
( path:s bytes:ByteList -- )
*/
: writeTmp
  File.writeBytes?^ nif
    "Create _tmp first (mkdir -p _tmp). Write failed: " $ $ nl
    error
  then
;

/*
Two ephemeral X25519 parties, HKDF the shared secret to an AES key,
seal message.txt, open it back.
( -- )
*/
: demoX25519HkdfAead
  "=== X25519 → HKDF → AES-GCM ===" $ nl
  loadMessage >zs
  Key.x25519 Key.generate^ >ys >xs      // y:pubA x:privA
  Key.x25519 Key.generate^              // privB pubB
  >ts                                   // t:pubB ; privB TOS (unused after)
  drop
  t:peek x:.agree^ >ys                  // y:shared
  DynamicObject.new^ >ts
  y:peek t:.ikm!
  "final-1c-aead" t:.info!
  32 t:.length!
  Kdf.hkdfSha256 t:peek Kdf.derive^ derivedKey !
  aesKeyPath derivedKey @ Hex.encode^ File.writeText^
  Key.aes256gcm derivedKey @ Key.fromBytes^
  Cipher.aes256gcm swap Cipher.new^ >ts
  t:.generateNonce^ sealedNonce !
  DynamicObject.new^ >ys
  z:peek y:.plaintext!
  sealedNonce @ y:.nonce!
  y:peek t:.encrypt^ sealedCt !
  ctPath sealedCt @ writeTmp
  noncePath sealedNonce @ Hex.encode^ File.writeText^
  DynamicObject.new^ >ys
  sealedCt @ y:.ciphertext!
  sealedNonce @ y:.nonce!
  y:peek t:.decrypt^ nif
    "decrypt failed: " $ $ nl
    error
  then
  Hex.encode^ z:peek Hex.encode^ == if
    "HKDF AEAD round-trip OK (" $ z:.len@ $ " bytes)" $ nl
  else
    "HKDF AEAD mismatch" error
  then
  t:drop x:drop y:drop z:drop
;

/*
Argon2id: passphrase + random salt → 32-byte key (light cost for a quick demo).
Production defaults are higher when time/memoryKiB/parallelism are omitted.
( -- )
*/
: demoArgon2id
  "=== Argon2id (passphrase → key) ===" $ nl
  16 SecureRandom.bytes^ >xs
  DynamicObject.new^ >ys
  "demo-passphrase" y:.password!
  x:peek y:.salt!
  32 y:.length!
  1 y:.time!
  8 y:.memoryKiB!
  1 y:.parallelism!
  Kdf.argon2id y:peek Kdf.derive^ >ts
  argonKeyPath t:peek Hex.encode^ File.writeText^
  "derived " $ t:.len@ $ " bytes (salt " $ x:.len@ $ " bytes)" $ nl
  t:drop x:drop y:drop
;

demoX25519HkdfAead
demoArgon2id
"wrote " aesKeyPath $ ", " $ ctPath $ ", " $ noncePath $ ", " $ argonKeyPath $ nl
