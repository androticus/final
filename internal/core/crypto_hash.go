package core

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
)

// Hash is a class-only catalog type for one-shot digests (Phase 0/1).
// Algorithm via class constants (DbProvider style): Hash.md5 / Hash.sha256.
// Final: bytes Hash.sha256 Hash.hash^  ( bytes algo -- digest:ByteList )
// MD5 is non-cryptographic (object hashes / cache keys only).
// Design: docs/003-design-notes/154-crypto.md

const (
	hashAlgoMD5    = "md5"
	hashAlgoSHA256 = "sha256"
)

var hashType *Type

func CreateRegisterHashType() {
	t := &Type{
		name:    "Hash",
		getZero: ReferenceTypeZeroFn,
		checkCoerceType: func(value any) any {
			panic(fmt.Errorf("Hash.checkCoerceType: Hash is a class-only type (got %T)", value))
		},
	}
	registerHashMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	hashType = t
}

func registerHashMembers(t *Type) {
	t.registerClassProperty("md5", "String",
		func(_ Any) Any { return String(hashAlgoMD5) },
		nil,
	)
	t.registerClassProperty("sha256", "String",
		func(_ Any) Any { return String(hashAlgoSHA256) },
		nil,
	)

	// ( bytes:ByteList|String algo:s -- digest:ByteList )
	t.registerNativeClassMethodSec("hash", func(f *RunFrame, _ *TypeObjPair) {
		algo := string(f.RP.DataStack.Pop().(String))
		data := cryptoBytesFromStackCell(f.RP.DataStack.Pop())
		f.RP.DataStack.Push(NewByteListFromBytes(hashDigest(algo, data)))
	}, SR_Crypto)
}

func hashDigest(algo string, data []byte) []byte {
	switch algo {
	case hashAlgoMD5:
		sum := md5.Sum(data)
		return sum[:]
	case hashAlgoSHA256:
		sum := sha256.Sum256(data)
		return sum[:]
	default:
		panic(fmt.Errorf("Hash.hash: unknown algorithm %q", algo))
	}
}

func init() {
	CreateRegisterHashType()
}
