package core

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
)

// Mac is a class-only keyed MAC (Phase 1.B). One-shot only.
// Final: Mac.hmacSha256 key bytes Mac.hash^  ( algo key bytes -- tag )
//        Mac.hmacSha256 key bytes tag Mac.verify^  ( algo key bytes tag -- bool )
// Design: docs/003-design-notes/154-crypto.md

const macAlgoHMACSHA256 = "hmacSha256"

var macType *Type

func CreateRegisterMacType() {
	t := &Type{
		name:    "Mac",
		getZero: ReferenceTypeZeroFn,
		checkCoerceType: func(value any) any {
			panic(fmt.Errorf("Mac.checkCoerceType: Mac is a class-only type (got %T)", value))
		},
	}
	registerMacMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	macType = t
}

func registerMacMembers(t *Type) {
	t.registerClassProperty("hmacSha256", "String",
		func(_ Any) Any { return String(macAlgoHMACSHA256) },
		nil,
	)

	// ( algo:s key:ByteList|String bytes:ByteList|String -- tag:ByteList )
	t.registerNativeClassMethodSec("hash", func(f *RunFrame, _ *TypeObjPair) {
		data := cryptoBytesFromStackCell(f.RP.DataStack.Pop())
		key := cryptoMacKeyFromStackCell(f.RP.DataStack.Pop())
		algo := string(f.RP.DataStack.Pop().(String))
		f.RP.DataStack.Push(NewByteListFromBytes(macHash(algo, key, data)))
	}, SR_Crypto)

	// ( algo:s key:ByteList|String bytes:ByteList|String tag:ByteList|String -- bool )
	t.registerNativeClassMethodSec("verify", func(f *RunFrame, _ *TypeObjPair) {
		tag := cryptoBytesFromStackCell(f.RP.DataStack.Pop())
		data := cryptoBytesFromStackCell(f.RP.DataStack.Pop())
		key := cryptoMacKeyFromStackCell(f.RP.DataStack.Pop())
		algo := string(f.RP.DataStack.Pop().(String))
		f.RP.DataStack.Push(Bool(macVerify(algo, key, data, tag)))
	}, SR_Crypto)
}

// cryptoMacKeyFromStackCell accepts ByteList|String or *Key (uses key material).
func cryptoMacKeyFromStackCell(v any) []byte {
	switch x := v.(type) {
	case *Key:
		return append([]byte(nil), x.material...)
	default:
		return cryptoBytesFromStackCell(v)
	}
}

func macHash(algo string, key, data []byte) []byte {
	switch algo {
	case macAlgoHMACSHA256:
		m := hmac.New(sha256.New, key)
		_, _ = m.Write(data)
		return m.Sum(nil)
	default:
		panic(fmt.Errorf("Mac.hash: unknown algorithm %q", algo))
	}
}

func macVerify(algo string, key, data, tag []byte) bool {
	want := macHash(algo, key, data)
	if len(want) != len(tag) {
		return false
	}
	return subtle.ConstantTimeCompare(want, tag) == 1
}

func init() {
	CreateRegisterMacType()
}
