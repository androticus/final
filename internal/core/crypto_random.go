package core

import (
	"crypto/rand"
	"fmt"
)

// SecureRandom is a class-only CSPRNG (crypto/rand). No math/rand.
// Final: n SecureRandom.bytes^  ( n:i -- bytes:ByteList )
// Design: docs/003-design-notes/154-crypto.md Phase 1.A

var secureRandomType *Type

func CreateRegisterSecureRandomType() {
	t := &Type{
		name:    "SecureRandom",
		getZero: ReferenceTypeZeroFn,
		checkCoerceType: func(value any) any {
			panic(fmt.Errorf("SecureRandom.checkCoerceType: SecureRandom is a class-only type (got %T)", value))
		},
	}
	registerSecureRandomMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	secureRandomType = t
}

func registerSecureRandomMembers(t *Type) {
	// ( n:i -- bytes:ByteList )
	t.registerNativeClassMethodSec("bytes", func(f *RunFrame, _ *TypeObjPair) {
		n := int(f.RP.DataStack.Pop().(Int))
		if n < 0 {
			panic(fmt.Errorf("SecureRandom.bytes: n must be >= 0, got %d", n))
		}
		f.RP.DataStack.Push(NewByteListFromBytes(secureRandomBytes(n)))
	}, SR_Crypto)
}

func secureRandomBytes(n int) []byte {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Errorf("SecureRandom.bytes: %w", err))
	}
	return b
}

func init() {
	CreateRegisterSecureRandomType()
}
