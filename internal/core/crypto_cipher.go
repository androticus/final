package core

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"

	"golang.org/x/crypto/chacha20poly1305"
)

// Cipher is an AEAD bound to algorithm + key.
// encrypt / decrypt take a params DynamicObject (plaintext|ciphertext, nonce, optional aad).
// Design: docs/003-design-notes/154-crypto.md Phase 1.A

const (
	cipherAlgoAES256GCM        = keyAlgoAES256GCM
	cipherAlgoChaCha20Poly1305 = keyAlgoChaCha20Poly1305
	cipherNonceSize            = 12
)

// Cipher holds a Go AEAD for AES-256-GCM or ChaCha20-Poly1305.
type Cipher struct {
	algo string
	aead cipher.AEAD
}

func (c *Cipher) FinalType() *Type { return cipherType }

var cipherType *Type

func CreateRegisterCipherType() {
	t := &Type{
		name:            "Cipher",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[Cipher]("Cipher"),
	}
	registerCipherMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	cipherType = t
}

func registerCipherMembers(t *Type) {
	t.registerClassProperty("aes256gcm", "String",
		func(_ Any) Any { return String(cipherAlgoAES256GCM) }, nil)
	t.registerClassProperty("chacha20poly1305", "String",
		func(_ Any) Any { return String(cipherAlgoChaCha20Poly1305) }, nil)

	// ( algo:s key:Key -- cipher )
	t.registerNativeClassMethodSec("new", func(f *RunFrame, _ *TypeObjPair) {
		key := f.RP.DataStack.Pop().(*Key)
		algo := string(f.RP.DataStack.Pop().(String))
		f.RP.DataStack.Push(newCipher(algo, key))
	}, SR_Crypto)

	// ( -- n:i ) nonce size in bytes for this Cipher's AEAD
	t.registerNativeMethodSec("nonceSize", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(Int(ths.Obj.(*Cipher).aead.NonceSize()))
	}, SR_Crypto)

	// ( -- nonce:ByteList ) SecureRandom nonce of the correct length
	t.registerNativeMethodSec("generateNonce", func(f *RunFrame, ths *TypeObjPair) {
		n := ths.Obj.(*Cipher).aead.NonceSize()
		f.RP.DataStack.Push(NewByteListFromBytes(secureRandomBytes(n)))
	}, SR_Crypto)

	// ( params:DO -- ciphertext:ByteList )
	// Fields: plaintext (req), nonce (req), aad (opt)
	t.registerNativeMethodSec("encrypt", func(f *RunFrame, ths *TypeObjPair) {
		params := f.RP.DataStack.Pop().(*DynamicObject)
		c := ths.Obj.(*Cipher)
		plaintext := cryptoDORequiredBytes(params, "plaintext")
		nonce := cryptoDORequiredBytes(params, "nonce")
		aad := cryptoDOOptionalBytes(params, "aad")
		if len(nonce) != c.aead.NonceSize() {
			panic(fmt.Errorf("Cipher.encrypt: nonce must be %d bytes, got %d", c.aead.NonceSize(), len(nonce)))
		}
		out := c.aead.Seal(nil, nonce, plaintext, aad)
		f.RP.DataStack.Push(NewByteListFromBytes(out))
	}, SR_Crypto)

	// ( params:DO -- plaintext:ByteList true | err:s false )
	// Fields: ciphertext (req), nonce (req), aad (opt)
	t.registerNativeMethodSec("decrypt", func(f *RunFrame, ths *TypeObjPair) {
		params := f.RP.DataStack.Pop().(*DynamicObject)
		c := ths.Obj.(*Cipher)
		ciphertext := cryptoDORequiredBytes(params, "ciphertext")
		nonce := cryptoDORequiredBytes(params, "nonce")
		aad := cryptoDOOptionalBytes(params, "aad")
		if len(nonce) != c.aead.NonceSize() {
			panic(fmt.Errorf("Cipher.decrypt: nonce must be %d bytes, got %d", c.aead.NonceSize(), len(nonce)))
		}
		plain, err := c.aead.Open(nil, nonce, ciphertext, aad)
		if err != nil {
			f.RP.DataStack.Push(String("authentication failed"))
			f.RP.DataStack.Push(Bool(false))
			return
		}
		f.RP.DataStack.Push(NewByteListFromBytes(plain))
		f.RP.DataStack.Push(Bool(true))
	}, SR_Crypto)
}

func newCipher(algo string, key *Key) *Cipher {
	if key == nil {
		panic(fmt.Errorf("Cipher.new: nil key"))
	}
	if key.algo != algo {
		panic(fmt.Errorf("Cipher.new: key algorithm %q does not match cipher %q", key.algo, algo))
	}
	if !key.isPrivate {
		panic(fmt.Errorf("Cipher.new: requires symmetric / private key material"))
	}
	if len(key.material) != keySymmetricLen {
		panic(fmt.Errorf("Cipher.new: key must be %d bytes, got %d", keySymmetricLen, len(key.material)))
	}
	var aead cipher.AEAD
	var err error
	switch algo {
	case cipherAlgoAES256GCM:
		block, e := aes.NewCipher(key.material)
		if e != nil {
			panic(fmt.Errorf("Cipher.new aes256gcm: %w", e))
		}
		aead, err = cipher.NewGCM(block)
	case cipherAlgoChaCha20Poly1305:
		aead, err = chacha20poly1305.New(key.material)
	default:
		panic(fmt.Errorf("Cipher.new: unknown algorithm %q", algo))
	}
	if err != nil {
		panic(fmt.Errorf("Cipher.new %s: %w", algo, err))
	}
	return &Cipher{algo: algo, aead: aead}
}

func init() {
	CreateRegisterCipherType()
}

var _ Typed = (*Cipher)(nil)
