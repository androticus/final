/*
HMAC-SHA-256 tag and verify (crypto Phase 1.B).

Reads message.txt beside this script. Writes the hex tag under _tmp/
(gitignored). Create that directory first: mkdir -p _tmp

Demonstrates a shared-secret MAC (request-auth style), not encryption.
Run from repo or Final/ install root:
  "examples/crypto/1b-hmac-tag.fnl" compileFromFile
*/

"examples/crypto/message.txt" const: msgPath
"_tmp/1b-hmac.hex" const: tagPath

/*
Demo shared secret (do not hard-code secrets in real apps — use Env / Moniker).
*/
"demo-hmac-secret" const: hmacSecret

0 var: macTag

/*
Load plaintext bytes from the fixture file.
( -- bytes:ByteList )
*/
: loadMessage
  msgPath File.readBytes^
;

/*
Tag the message, write hex under _tmp/, verify, then show a failing verify.
( -- )
*/
: demoHmac
  "=== HMAC-SHA-256 ===" $ nl
  loadMessage >zs
  Mac.hmacSha256 hmacSecret z:peek Mac.hash^ macTag !
  tagPath macTag @ Hex.encode^ File.writeText?^ nif
    "Create _tmp first (mkdir -p _tmp). Write failed: " $ $ nl
    error
  then
  Mac.hmacSha256 hmacSecret z:peek macTag @ Mac.verify^ if
    "verify OK (" $ macTag @ .len@ $ " byte tag)" $ nl
  else
    "verify failed" error
  then
  Mac.hmacSha256 hmacSecret "tampered" macTag @ Mac.verify^ nif
    "tamper correctly rejected" $ nl
  else
    "tamper should not verify" error
  then
  z:drop
;

demoHmac
"wrote " tagPath $ nl
