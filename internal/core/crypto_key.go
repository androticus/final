package core

import (
	"crypto/ecdh"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
)

// Key is algorithm-tagged key material (symmetric or asymmetric half).
// generate^ leaves one Key (symmetric) or priv then pub (Ed25519 / X25519).
// Design: docs/003-design-notes/154-crypto.md Phase 1.A

const (
	keyAlgoAES256GCM        = "aes256gcm"
	keyAlgoChaCha20Poly1305 = "chacha20poly1305"
	keyAlgoEd25519          = "ed25519"
	keyAlgoX25519           = "x25519"
	keySymmetricLen         = 32
)

// Key holds opaque key bytes and algorithm. Do not log contents.
type Key struct {
	algo      string
	material  []byte
	isPrivate bool // true for symmetric and private halves; false for public
}

func (k *Key) FinalType() *Type { return keyType }

var keyType *Type

func CreateRegisterKeyType() {
	t := &Type{
		name:            "Key",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[Key]("Key"),
	}
	registerKeyMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	keyType = t
}

func registerKeyMembers(t *Type) {
	t.registerClassProperty("aes256gcm", "String",
		func(_ Any) Any { return String(keyAlgoAES256GCM) }, nil)
	t.registerClassProperty("chacha20poly1305", "String",
		func(_ Any) Any { return String(keyAlgoChaCha20Poly1305) }, nil)
	t.registerClassProperty("ed25519", "String",
		func(_ Any) Any { return String(keyAlgoEd25519) }, nil)
	t.registerClassProperty("x25519", "String",
		func(_ Any) Any { return String(keyAlgoX25519) }, nil)

	// ( algo:s -- key ) or ( algo:s -- priv pub )
	t.registerNativeClassMethodSec("generate", func(f *RunFrame, _ *TypeObjPair) {
		algo := string(f.RP.DataStack.Pop().(String))
		switch algo {
		case keyAlgoAES256GCM, keyAlgoChaCha20Poly1305:
			f.RP.DataStack.Push(newSymmetricKey(algo, secureRandomBytes(keySymmetricLen)))
		case keyAlgoEd25519:
			pub, priv, err := ed25519.GenerateKey(rand.Reader)
			if err != nil {
				panic(fmt.Errorf("Key.generate ed25519: %w", err))
			}
			f.RP.DataStack.Push(&Key{algo: algo, material: append([]byte(nil), priv...), isPrivate: true})
			f.RP.DataStack.Push(&Key{algo: algo, material: append([]byte(nil), pub...), isPrivate: false})
		case keyAlgoX25519:
			priv, pub := generateX25519KeyPair()
			f.RP.DataStack.Push(priv)
			f.RP.DataStack.Push(pub)
		default:
			panic(fmt.Errorf("Key.generate: unknown algorithm %q", algo))
		}
	}, SR_Crypto)

	// ( algo:s bytes:ByteList|String -- key )
	// Symmetric: 32 bytes. Ed25519: 64 = private, 32 = public. X25519: 32 = private.
	t.registerNativeClassMethodSec("fromBytes", func(f *RunFrame, _ *TypeObjPair) {
		data := cryptoBytesFromStackCell(f.RP.DataStack.Pop())
		algo := string(f.RP.DataStack.Pop().(String))
		f.RP.DataStack.Push(keyFromBytes(algo, data, false))
	}, SR_Crypto)

	// ( algo:s bytes:ByteList|String -- key ) — public half (Ed25519 / X25519, 32 bytes)
	t.registerNativeClassMethodSec("fromPublic", func(f *RunFrame, _ *TypeObjPair) {
		data := cryptoBytesFromStackCell(f.RP.DataStack.Pop())
		algo := string(f.RP.DataStack.Pop().(String))
		f.RP.DataStack.Push(keyFromBytes(algo, data, true))
	}, SR_Crypto)

	// ( -- bytes:ByteList )
	t.registerNativeMethodSec("toBytes", func(f *RunFrame, ths *TypeObjPair) {
		k := ths.Obj.(*Key)
		f.RP.DataStack.Push(NewByteListFromBytes(append([]byte(nil), k.material...)))
	}, SR_Crypto)

	// ( msg:ByteList|String -- sig:ByteList ) — Ed25519 private only
	t.registerNativeMethodSec("sign", func(f *RunFrame, ths *TypeObjPair) {
		k := ths.Obj.(*Key)
		msg := cryptoBytesFromStackCell(f.RP.DataStack.Pop())
		if k.algo != keyAlgoEd25519 || !k.isPrivate {
			panic(fmt.Errorf("Key.sign: requires Ed25519 private key"))
		}
		sig := ed25519.Sign(ed25519.PrivateKey(k.material), msg)
		f.RP.DataStack.Push(NewByteListFromBytes(sig))
	}, SR_Crypto)

	// ( msg:ByteList|String sig:ByteList|String -- bool ) — Ed25519 public only
	t.registerNativeMethodSec("verify", func(f *RunFrame, ths *TypeObjPair) {
		k := ths.Obj.(*Key)
		sig := cryptoBytesFromStackCell(f.RP.DataStack.Pop())
		msg := cryptoBytesFromStackCell(f.RP.DataStack.Pop())
		if k.algo != keyAlgoEd25519 || k.isPrivate {
			panic(fmt.Errorf("Key.verify: requires Ed25519 public key"))
		}
		ok := ed25519.Verify(ed25519.PublicKey(k.material), msg, sig)
		f.RP.DataStack.Push(Bool(ok))
	}, SR_Crypto)

	// ( peerPub:Key -- shared:ByteList ) — X25519 private only
	t.registerNativeMethodSec("agree", func(f *RunFrame, ths *TypeObjPair) {
		k := ths.Obj.(*Key)
		peer := f.RP.DataStack.Pop().(*Key)
		if k.algo != keyAlgoX25519 || !k.isPrivate {
			panic(fmt.Errorf("Key.agree: requires X25519 private key receiver"))
		}
		if peer.algo != keyAlgoX25519 || peer.isPrivate {
			panic(fmt.Errorf("Key.agree: peer must be X25519 public key"))
		}
		f.RP.DataStack.Push(NewByteListFromBytes(x25519Agree(k.material, peer.material)))
	}, SR_Crypto)
}

func newSymmetricKey(algo string, material []byte) *Key {
	if len(material) != keySymmetricLen {
		panic(fmt.Errorf("Key: %s key must be %d bytes, got %d", algo, keySymmetricLen, len(material)))
	}
	return &Key{algo: algo, material: append([]byte(nil), material...), isPrivate: true}
}

// keyFromBytes builds a Key. forcePublic selects the public half when length is ambiguous (X25519).
func keyFromBytes(algo string, data []byte, forcePublic bool) *Key {
	switch algo {
	case keyAlgoAES256GCM, keyAlgoChaCha20Poly1305:
		if forcePublic {
			panic(fmt.Errorf("Key.fromPublic: not valid for symmetric algorithm %q", algo))
		}
		return newSymmetricKey(algo, data)
	case keyAlgoEd25519:
		switch len(data) {
		case ed25519.PrivateKeySize:
			if forcePublic {
				panic(fmt.Errorf("Key.fromPublic: Ed25519 public key must be %d bytes", ed25519.PublicKeySize))
			}
			return &Key{algo: algo, material: append([]byte(nil), data...), isPrivate: true}
		case ed25519.PublicKeySize:
			return &Key{algo: algo, material: append([]byte(nil), data...), isPrivate: false}
		default:
			panic(fmt.Errorf("Key.fromBytes ed25519: want %d (private) or %d (public) bytes, got %d",
				ed25519.PrivateKeySize, ed25519.PublicKeySize, len(data)))
		}
	case keyAlgoX25519:
		if len(data) != 32 {
			panic(fmt.Errorf("Key.fromBytes x25519: want 32 bytes, got %d", len(data)))
		}
		if forcePublic {
			return &Key{algo: algo, material: append([]byte(nil), data...), isPrivate: false}
		}
		// Validate as private key material.
		curve := ecdh.X25519()
		if _, err := curve.NewPrivateKey(data); err != nil {
			panic(fmt.Errorf("Key.fromBytes x25519 private: %w", err))
		}
		return &Key{algo: algo, material: append([]byte(nil), data...), isPrivate: true}
	default:
		panic(fmt.Errorf("Key.fromBytes: unknown algorithm %q", algo))
	}
}

func generateX25519KeyPair() (priv, pub *Key) {
	curve := ecdh.X25519()
	p, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		panic(fmt.Errorf("Key.generate x25519: %w", err))
	}
	priv = &Key{algo: keyAlgoX25519, material: append([]byte(nil), p.Bytes()...), isPrivate: true}
	pub = &Key{algo: keyAlgoX25519, material: append([]byte(nil), p.PublicKey().Bytes()...), isPrivate: false}
	return priv, pub
}

func x25519Agree(privBytes, peerPubBytes []byte) []byte {
	curve := ecdh.X25519()
	priv, err := curve.NewPrivateKey(privBytes)
	if err != nil {
		panic(fmt.Errorf("Key.agree: private key: %w", err))
	}
	peer, err := curve.NewPublicKey(peerPubBytes)
	if err != nil {
		panic(fmt.Errorf("Key.agree: peer public key: %w", err))
	}
	shared, err := priv.ECDH(peer)
	if err != nil {
		panic(fmt.Errorf("Key.agree: %w", err))
	}
	return shared
}

func init() {
	CreateRegisterKeyType()
}

var _ Typed = (*Key)(nil)
