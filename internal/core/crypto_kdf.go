package core

import (
	"crypto/sha256"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/hkdf"
)

// Kdf is a class-only key-derivation catalog (Phase 1.C).
// derive takes algo + params DynamicObject; leaves ByteList (wrap with Key.fromBytes for Cipher).
// Design: docs/003-design-notes/154-crypto.md

const (
	kdfAlgoHKDFSHA256 = "hkdfSha256"
	kdfAlgoArgon2id   = "argon2id"

	// Argon2id defaults when optional cost fields are omitted (OWASP-ish interactive).
	argon2DefaultTime        = 3
	argon2DefaultMemoryKiB   = 64 * 1024 // 64 MiB
	argon2DefaultParallelism = 4
)

var kdfType *Type

func CreateRegisterKdfType() {
	t := &Type{
		name:    "Kdf",
		getZero: ReferenceTypeZeroFn,
		checkCoerceType: func(value any) any {
			panic(fmt.Errorf("Kdf.checkCoerceType: Kdf is a class-only type (got %T)", value))
		},
	}
	registerKdfMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	kdfType = t
}

func registerKdfMembers(t *Type) {
	t.registerClassProperty("hkdfSha256", "String",
		func(_ Any) Any { return String(kdfAlgoHKDFSHA256) }, nil)
	t.registerClassProperty("argon2id", "String",
		func(_ Any) Any { return String(kdfAlgoArgon2id) }, nil)

	// ( algo:s params:DO -- bytes:ByteList )
	t.registerNativeClassMethodSec("derive", func(f *RunFrame, _ *TypeObjPair) {
		params := f.RP.DataStack.Pop().(*DynamicObject)
		algo := string(f.RP.DataStack.Pop().(String))
		f.RP.DataStack.Push(NewByteListFromBytes(kdfDerive(algo, params)))
	}, SR_Crypto)
}

func kdfDerive(algo string, params *DynamicObject) []byte {
	switch algo {
	case kdfAlgoHKDFSHA256:
		return kdfHKDFSHA256(params)
	case kdfAlgoArgon2id:
		return kdfArgon2id(params)
	default:
		panic(fmt.Errorf("Kdf.derive: unknown algorithm %q", algo))
	}
}

// HKDF params: ikm (req), length (req), info (opt → empty), salt (opt → nil).
func kdfHKDFSHA256(params *DynamicObject) []byte {
	ikm := cryptoDORequiredBytes(params, "ikm")
	length := cryptoDORequiredInt(params, "length")
	if length < 0 {
		panic(fmt.Errorf("Kdf.hkdfSha256: length must be >= 0, got %d", length))
	}
	info := cryptoDOOptionalBytes(params, "info")
	if info == nil {
		info = []byte{}
	}
	salt := cryptoDOOptionalBytes(params, "salt")
	r := hkdf.New(sha256.New, ikm, salt, info)
	out := make([]byte, length)
	if _, err := io.ReadFull(r, out); err != nil {
		panic(fmt.Errorf("Kdf.hkdfSha256: %w", err))
	}
	return out
}

// Argon2id params: password (req), salt (req), length (req);
// optional time, memoryKiB, parallelism with documented defaults.
func kdfArgon2id(params *DynamicObject) []byte {
	password := cryptoDORequiredBytes(params, "password")
	salt := cryptoDORequiredBytes(params, "salt")
	length := cryptoDORequiredInt(params, "length")
	if length < 1 {
		panic(fmt.Errorf("Kdf.argon2id: length must be >= 1, got %d", length))
	}
	if len(salt) < 8 {
		panic(fmt.Errorf("Kdf.argon2id: salt must be at least 8 bytes, got %d", len(salt)))
	}
	time := uint32(cryptoDOOptionalInt(params, "time", argon2DefaultTime))
	memory := uint32(cryptoDOOptionalInt(params, "memoryKiB", argon2DefaultMemoryKiB))
	threads := uint8(cryptoDOOptionalInt(params, "parallelism", argon2DefaultParallelism))
	if time < 1 || memory < 1 || threads < 1 {
		panic(fmt.Errorf("Kdf.argon2id: time/memoryKiB/parallelism must be >= 1"))
	}
	return argon2.IDKey(password, salt, time, memory, threads, uint32(length))
}

func init() {
	CreateRegisterKdfType()
}
