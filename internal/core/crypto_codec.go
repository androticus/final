package core

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// Hex and Base64 are class-only encode/decode helpers for digests and key bytes.
// Design: docs/003-design-notes/154-crypto.md (Phase 0 Codec)

var hexType *Type
var base64Type *Type

func CreateRegisterHexType() {
	t := &Type{
		name:    "Hex",
		getZero: ReferenceTypeZeroFn,
		checkCoerceType: func(value any) any {
			panic(fmt.Errorf("Hex.checkCoerceType: Hex is a class-only type (got %T)", value))
		},
	}
	registerHexMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	hexType = t
}

func registerHexMembers(t *Type) {
	// ( bytes:ByteList|String -- hex:s )
	t.registerNativeClassMethodSec("encode", func(f *RunFrame, _ *TypeObjPair) {
		data := cryptoBytesFromStackCell(f.RP.DataStack.Pop())
		f.RP.DataStack.Push(String(hex.EncodeToString(data)))
	}, SR_Crypto)

	// ( hex:s -- bytes:ByteList )
	t.registerNativeClassMethodSec("decode", func(f *RunFrame, _ *TypeObjPair) {
		s := string(f.RP.DataStack.Pop().(String))
		b, err := hex.DecodeString(s)
		if err != nil {
			panic(fmt.Errorf("Hex.decode: %w", err))
		}
		f.RP.DataStack.Push(NewByteListFromBytes(b))
	}, SR_Crypto)
}

func CreateRegisterBase64Type() {
	t := &Type{
		name:    "Base64",
		getZero: ReferenceTypeZeroFn,
		checkCoerceType: func(value any) any {
			panic(fmt.Errorf("Base64.checkCoerceType: Base64 is a class-only type (got %T)", value))
		},
	}
	registerBase64Members(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	base64Type = t
}

func registerBase64Members(t *Type) {
	// ( bytes:ByteList|String -- s )
	t.registerNativeClassMethodSec("encode", func(f *RunFrame, _ *TypeObjPair) {
		data := cryptoBytesFromStackCell(f.RP.DataStack.Pop())
		f.RP.DataStack.Push(String(base64.StdEncoding.EncodeToString(data)))
	}, SR_Crypto)

	// ( s -- bytes:ByteList )
	t.registerNativeClassMethodSec("decode", func(f *RunFrame, _ *TypeObjPair) {
		s := string(f.RP.DataStack.Pop().(String))
		b, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			panic(fmt.Errorf("Base64.decode: %w", err))
		}
		f.RP.DataStack.Push(NewByteListFromBytes(b))
	}, SR_Crypto)
}

func init() {
	CreateRegisterHexType()
	CreateRegisterBase64Type()
}
