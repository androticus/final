package core

import (
	"bytes"
	"testing"
)

func TestKdfHKDFSHA256(t *testing.T) {
	_, rp, f := cryptoFrame(t)
	params := NewDynamicObject()
	params.Set("ikm", NewByteListFromBytes([]byte("ikm-secret-material!!")))
	params.Set("info", String("aes-key"))
	params.Set("length", Int(32))
	params.Set("salt", NewByteListFromBytes([]byte("salt")))

	mm := GetTypeRegistrar().Get("Kdf").GetClassMethodMetadata("derive")
	if mm == nil || mm.restrict != SR_Crypto {
		t.Fatalf("Kdf.derive = %#v", mm)
	}
	rp.DataStack.Push(String(kdfAlgoHKDFSHA256))
	rp.DataStack.Push(params)
	mm.invokeCall(f, NewTypeObjPair(GetTypeRegistrar().Get("Kdf")))
	out := rp.DataStack.Pop().(*ByteList)
	if out.Len() != 32 {
		t.Fatalf("len = %d, want 32", out.Len())
	}

	// Same inputs → same output
	rp.DataStack.Push(String(kdfAlgoHKDFSHA256))
	rp.DataStack.Push(params)
	mm.invokeCall(f, NewTypeObjPair(GetTypeRegistrar().Get("Kdf")))
	out2 := rp.DataStack.Pop().(*ByteList)
	if !bytes.Equal(out.items, out2.items) {
		t.Fatal("HKDF not deterministic")
	}
}

func TestKdfArgon2id(t *testing.T) {
	_, rp, f := cryptoFrame(t)
	params := NewDynamicObject()
	params.Set("password", String("correct horse battery staple"))
	params.Set("salt", NewByteListFromBytes([]byte("saltsalt"))) // 8 bytes
	params.Set("length", Int(32))
	// Fast test params
	params.Set("time", Int(1))
	params.Set("memoryKiB", Int(8))
	params.Set("parallelism", Int(1))

	mm := GetTypeRegistrar().Get("Kdf").GetClassMethodMetadata("derive")
	rp.DataStack.Push(String(kdfAlgoArgon2id))
	rp.DataStack.Push(params)
	mm.invokeCall(f, NewTypeObjPair(GetTypeRegistrar().Get("Kdf")))
	out := rp.DataStack.Pop().(*ByteList)
	if out.Len() != 32 {
		t.Fatalf("len = %d", out.Len())
	}

	rp.DataStack.Push(String(kdfAlgoArgon2id))
	rp.DataStack.Push(params)
	mm.invokeCall(f, NewTypeObjPair(GetTypeRegistrar().Get("Kdf")))
	out2 := rp.DataStack.Pop().(*ByteList)
	if !bytes.Equal(out.items, out2.items) {
		t.Fatal("Argon2id not deterministic")
	}
}

func TestKdfHKDFToCipherRoundTrip(t *testing.T) {
	_, rp, f := cryptoFrame(t)
	// Two X25519 parties → shared → HKDF → AES key → encrypt/decrypt
	gen := GetTypeRegistrar().Get("Key").GetClassMethodMetadata("generate")
	keyT := NewTypeObjPair(GetTypeRegistrar().Get("Key"))

	rp.DataStack.Push(String(keyAlgoX25519))
	gen.invokeCall(f, keyT)
	pubA := rp.DataStack.Pop().(*Key)
	privA := rp.DataStack.Pop().(*Key)

	rp.DataStack.Push(String(keyAlgoX25519))
	gen.invokeCall(f, keyT)
	pubB := rp.DataStack.Pop().(*Key)
	_ = rp.DataStack.Pop().(*Key) // privB unused; A agrees with B's pub

	agree := privA.FinalType().GetMethodMetadata("agree")
	rp.DataStack.Push(pubB)
	agree.invokeCall(f, NewTypeObjPair(privA))
	shared := rp.DataStack.Pop().(*ByteList)
	_ = pubA

	params := NewDynamicObject()
	params.Set("ikm", shared)
	params.Set("info", String("1c-aead"))
	params.Set("length", Int(32))
	derive := GetTypeRegistrar().Get("Kdf").GetClassMethodMetadata("derive")
	rp.DataStack.Push(String(kdfAlgoHKDFSHA256))
	rp.DataStack.Push(params)
	derive.invokeCall(f, NewTypeObjPair(GetTypeRegistrar().Get("Kdf")))
	keyBytes := rp.DataStack.Pop().(*ByteList)

	fromBytes := GetTypeRegistrar().Get("Key").GetClassMethodMetadata("fromBytes")
	rp.DataStack.Push(String(keyAlgoAES256GCM))
	rp.DataStack.Push(keyBytes)
	fromBytes.invokeCall(f, keyT)
	aesKey := rp.DataStack.Pop().(*Key)

	newC := GetTypeRegistrar().Get("Cipher").GetClassMethodMetadata("new")
	rp.DataStack.Push(String(cipherAlgoAES256GCM))
	rp.DataStack.Push(aesKey)
	newC.invokeCall(f, NewTypeObjPair(GetTypeRegistrar().Get("Cipher")))
	c := rp.DataStack.Pop().(*Cipher)

	nonceMM := c.FinalType().GetMethodMetadata("generateNonce")
	nonceMM.invokeCall(f, NewTypeObjPair(c))
	nonce := rp.DataStack.Pop().(*ByteList)

	encParams := NewDynamicObject()
	encParams.Set("plaintext", String("after hkdf"))
	encParams.Set("nonce", nonce)
	enc := c.FinalType().GetMethodMetadata("encrypt")
	rp.DataStack.Push(encParams)
	enc.invokeCall(f, NewTypeObjPair(c))
	ct := rp.DataStack.Pop().(*ByteList)

	decParams := NewDynamicObject()
	decParams.Set("ciphertext", ct)
	decParams.Set("nonce", nonce)
	dec := c.FinalType().GetMethodMetadata("decrypt")
	rp.DataStack.Push(decParams)
	dec.invokeCall(f, NewTypeObjPair(c))
	ok := bool(rp.DataStack.Pop().(Bool))
	plain := rp.DataStack.Pop().(*ByteList)
	if !ok || string(plain.items) != "after hkdf" {
		t.Fatalf("AEAD after HKDF failed: ok=%v plain=%q", ok, plain.items)
	}
}

func TestKdfInterpretHKDF(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	runInterpretScript(t, task, rp,
		"DynamicObject.new^ >ys\n"+
			"\"ikm-material-here!!!!\" y:.ikm!\n"+
			"\"ctx\" y:.info!\n"+
			"32 y:.length!\n"+
			"Kdf.hkdfSha256 y:peek Kdf.derive^ .len@\n")
	if int(rp.DataStack.Pop().(Int)) != 32 {
		t.Fatal("want 32-byte derived key")
	}
}

func TestKdfDeriveRestricted(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	task.restrict = SR_Crypto
	rp := NewTestRunProvider()
	mm := GetTypeRegistrar().Get("Kdf").GetClassMethodMetadata("derive")
	f := &RunFrame{Task: task, RP: rp}
	defer func() {
		if recover() != "operation not allowed" {
			t.Fatal("expected operation not allowed")
		}
	}()
	mm.invokeCall(f, NewTypeObjPair(GetTypeRegistrar().Get("Kdf")))
	t.Fatal("expected panic")
}
