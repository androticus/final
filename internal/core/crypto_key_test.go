package core

import (
	"bytes"
	"crypto/ed25519"
	"testing"
)

func cryptoFrame(t *testing.T) (*Task, *RunProvider, *RunFrame) {
	t.Helper()
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	return task, rp, &RunFrame{Task: task, RP: rp}
}

func TestSecureRandomBytes(t *testing.T) {
	_, rp, f := cryptoFrame(t)
	mm := GetTypeRegistrar().Get("SecureRandom").GetClassMethodMetadata("bytes")
	if mm == nil || mm.restrict != SR_Crypto {
		t.Fatalf("SecureRandom.bytes restrict = %#v", mm)
	}
	rp.DataStack.Push(Int(16))
	mm.invokeCall(f, NewTypeObjPair(GetTypeRegistrar().Get("SecureRandom")))
	b := rp.DataStack.Pop().(*ByteList)
	if b.Len() != 16 {
		t.Fatalf("len = %d, want 16", b.Len())
	}
}

func TestKeyGenerateAESAndCipherRoundTrip(t *testing.T) {
	_, rp, f := cryptoFrame(t)
	gen := GetTypeRegistrar().Get("Key").GetClassMethodMetadata("generate")
	rp.DataStack.Push(String(keyAlgoAES256GCM))
	gen.invokeCall(f, NewTypeObjPair(GetTypeRegistrar().Get("Key")))
	key := rp.DataStack.Pop().(*Key)
	if key.algo != keyAlgoAES256GCM || len(key.material) != 32 {
		t.Fatalf("bad key: algo=%s len=%d", key.algo, len(key.material))
	}

	newC := GetTypeRegistrar().Get("Cipher").GetClassMethodMetadata("new")
	rp.DataStack.Push(String(cipherAlgoAES256GCM))
	rp.DataStack.Push(key)
	newC.invokeCall(f, NewTypeObjPair(GetTypeRegistrar().Get("Cipher")))
	c := rp.DataStack.Pop().(*Cipher)

	nonceMM := c.FinalType().GetMethodMetadata("generateNonce")
	nonceMM.invokeCall(f, NewTypeObjPair(c))
	nonce := rp.DataStack.Pop().(*ByteList)

	params := NewDynamicObject()
	params.Set("plaintext", NewByteListFromBytes([]byte("hello aead")))
	params.Set("nonce", nonce)
	params.Set("aad", String("meta"))

	enc := c.FinalType().GetMethodMetadata("encrypt")
	rp.DataStack.Push(params)
	enc.invokeCall(f, NewTypeObjPair(c))
	ct := rp.DataStack.Pop().(*ByteList)

	decParams := NewDynamicObject()
	decParams.Set("ciphertext", ct)
	decParams.Set("nonce", nonce)
	decParams.Set("aad", String("meta"))
	dec := c.FinalType().GetMethodMetadata("decrypt")
	rp.DataStack.Push(decParams)
	dec.invokeCall(f, NewTypeObjPair(c))
	if !bool(rp.DataStack.Pop().(Bool)) {
		t.Fatal("decrypt failed")
	}
	plain := rp.DataStack.Pop().(*ByteList)
	if !bytes.Equal(plain.items, []byte("hello aead")) {
		t.Fatalf("plain = %q", plain.items)
	}

	// Bit-flip → auth fail
	bad := append([]byte(nil), ct.items...)
	bad[0] ^= 0xff
	badParams := NewDynamicObject()
	badParams.Set("ciphertext", NewByteListFromBytes(bad))
	badParams.Set("nonce", nonce)
	rp.DataStack.Push(badParams)
	dec.invokeCall(f, NewTypeObjPair(c))
	if bool(rp.DataStack.Pop().(Bool)) {
		t.Fatal("expected auth failure")
	}
}

func TestKeyGenerateChaChaRoundTrip(t *testing.T) {
	_, rp, f := cryptoFrame(t)
	gen := GetTypeRegistrar().Get("Key").GetClassMethodMetadata("generate")
	rp.DataStack.Push(String(keyAlgoChaCha20Poly1305))
	gen.invokeCall(f, NewTypeObjPair(GetTypeRegistrar().Get("Key")))
	key := rp.DataStack.Pop().(*Key)

	newC := GetTypeRegistrar().Get("Cipher").GetClassMethodMetadata("new")
	rp.DataStack.Push(String(cipherAlgoChaCha20Poly1305))
	rp.DataStack.Push(key)
	newC.invokeCall(f, NewTypeObjPair(GetTypeRegistrar().Get("Cipher")))
	c := rp.DataStack.Pop().(*Cipher)

	nonceMM := c.FinalType().GetMethodMetadata("generateNonce")
	nonceMM.invokeCall(f, NewTypeObjPair(c))
	nonce := rp.DataStack.Pop().(*ByteList)

	params := NewDynamicObject()
	params.Set("plaintext", String("chacha"))
	params.Set("nonce", nonce)
	enc := c.FinalType().GetMethodMetadata("encrypt")
	rp.DataStack.Push(params)
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
	if !ok || string(plain.items) != "chacha" {
		t.Fatalf("chacha round-trip failed ok=%v plain=%q", ok, plain.items)
	}
}

func TestEd25519SignVerify(t *testing.T) {
	_, rp, f := cryptoFrame(t)
	gen := GetTypeRegistrar().Get("Key").GetClassMethodMetadata("generate")
	rp.DataStack.Push(String(keyAlgoEd25519))
	gen.invokeCall(f, NewTypeObjPair(GetTypeRegistrar().Get("Key")))
	pub := rp.DataStack.Pop().(*Key)
	priv := rp.DataStack.Pop().(*Key)
	if !priv.isPrivate || pub.isPrivate {
		t.Fatal("expected priv then pub on stack")
	}
	if len(priv.material) != ed25519.PrivateKeySize || len(pub.material) != ed25519.PublicKeySize {
		t.Fatalf("sizes priv=%d pub=%d", len(priv.material), len(pub.material))
	}

	sign := priv.FinalType().GetMethodMetadata("sign")
	rp.DataStack.Push(String("androticus hello"))
	sign.invokeCall(f, NewTypeObjPair(priv))
	sig := rp.DataStack.Pop().(*ByteList)

	verify := pub.FinalType().GetMethodMetadata("verify")
	rp.DataStack.Push(String("androticus hello"))
	rp.DataStack.Push(sig)
	verify.invokeCall(f, NewTypeObjPair(pub))
	if !bool(rp.DataStack.Pop().(Bool)) {
		t.Fatal("verify should succeed")
	}

	rp.DataStack.Push(String("androticus hello!"))
	rp.DataStack.Push(sig)
	verify.invokeCall(f, NewTypeObjPair(pub))
	if bool(rp.DataStack.Pop().(Bool)) {
		t.Fatal("verify should fail on different message")
	}
}

func TestX25519Agree(t *testing.T) {
	_, rp, f := cryptoFrame(t)
	gen := GetTypeRegistrar().Get("Key").GetClassMethodMetadata("generate")
	typ := NewTypeObjPair(GetTypeRegistrar().Get("Key"))

	rp.DataStack.Push(String(keyAlgoX25519))
	gen.invokeCall(f, typ)
	pubA := rp.DataStack.Pop().(*Key)
	privA := rp.DataStack.Pop().(*Key)

	rp.DataStack.Push(String(keyAlgoX25519))
	gen.invokeCall(f, typ)
	pubB := rp.DataStack.Pop().(*Key)
	privB := rp.DataStack.Pop().(*Key)

	agree := privA.FinalType().GetMethodMetadata("agree")
	rp.DataStack.Push(pubB)
	agree.invokeCall(f, NewTypeObjPair(privA))
	sharedAB := rp.DataStack.Pop().(*ByteList)

	rp.DataStack.Push(pubA)
	agree.invokeCall(f, NewTypeObjPair(privB))
	sharedBA := rp.DataStack.Pop().(*ByteList)

	if !bytes.Equal(sharedAB.items, sharedBA.items) || len(sharedAB.items) == 0 {
		t.Fatalf("shared secrets differ or empty: %x vs %x", sharedAB.items, sharedBA.items)
	}
}

func TestCrypto1AInterpretSeal(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	// Build cipher, encrypt via params DO, leave hex ciphertext.
	runInterpretScript(t, task, rp,
		"Key.aes256gcm Key.generate^\n"+
			"Cipher.aes256gcm swap Cipher.new^ >ts\n"+
			"t:.generateNonce^ >xs\n"+
			"DynamicObject.new^ >ys\n"+
			"\"hi\" y:.plaintext!\n"+
			"x:peek y:.nonce!\n"+
			"y:peek t:.encrypt^ Hex.encode^\n")
	hexCT := string(rp.DataStack.Pop().(String))
	if len(hexCT) < 2 {
		t.Fatalf("empty ciphertext hex %q", hexCT)
	}
}

func TestKeyGenerateRestricted(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	task.restrict = SR_Crypto
	rp := NewTestRunProvider()
	mm := GetTypeRegistrar().Get("Key").GetClassMethodMetadata("generate")
	f := &RunFrame{Task: task, RP: rp}
	defer func() {
		if recover() != "operation not allowed" {
			t.Fatal("expected operation not allowed")
		}
	}()
	mm.invokeCall(f, NewTypeObjPair(GetTypeRegistrar().Get("Key")))
	t.Fatal("expected panic")
}
