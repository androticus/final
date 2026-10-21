/*
AES-256-GCM seal/open and Ed25519 sign/verify.

Reads the committed fixture message.txt beside this script (path relative
to install/repo root). Writes transient ciphertext and signature under
_tmp/ . Create that directory first: mkdir -p _tmp

Run from repo or Final/ install root, e.g.:
  "examples/crypto/1a-seal-and-sign.fnl" compileFromFile
*/

"examples/crypto/message.txt" const: msgPath
"_tmp/1a-ciphertext.bin" const: ctPath
"_tmp/1a-nonce.hex" const: noncePath
"_tmp/1a-sig.hex" const: sigPath
"_tmp/1a-pub.hex" const: pubPath

0 var: sealedCt
0 var: sealedNonce
0 var: msgSig

/*
Load plaintext bytes from the fixture file.
( -- bytes:ByteList )
*/
: loadMessage
  msgPath File.readBytes^
;

/*
Write bytes to path; clear error if _tmp is missing.
( path:s bytes:ByteList -- )
*/
: writeTmp
  File.writeBytes?^ nif
    "Create _tmp first (mkdir -p _tmp). Write failed: " $ $ nl
    error
  then
;

/*
AEAD round-trip: encrypt message, write ct+nonce under _tmp/, decrypt, compare.
( -- )
*/
: demoAead
  "=== AEAD (AES-256-GCM) ===" $ nl
  loadMessage >zs
  Key.aes256gcm Key.generate^
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
  >ys
  y:peek Hex.encode^ z:peek Hex.encode^ == if
    "round-trip OK (" $ z:.len@ $ " bytes)" $ nl
  else
    "round-trip mismatch" error
  then
  t:drop y:drop z:drop
;

/*
Sign the message with Ed25519; verify; write sig + pub under _tmp/.
generate^ leaves priv then pub (two stack items).
( -- )
*/
: demoSign
  "=== Ed25519 sign / verify ===" $ nl
  loadMessage >zs
  Key.ed25519 Key.generate^ >ys >xs   // y:pub x:priv
  z:peek x:.sign^ msgSig !
  sigPath msgSig @ Hex.encode^ File.writeText^
  pubPath y:.toBytes^ Hex.encode^ File.writeText^
  z:peek msgSig @ y:.verify^ if
    "verify OK" $ nl
  else
    "verify failed" error
  then
  x:drop y:drop z:drop
;

demoAead
demoSign
"wrote " ctPath $ ", " $ noncePath $ ", " $ sigPath $ ", " $ pubPath $ nl
