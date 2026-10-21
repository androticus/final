package core

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestMacHMACSHA256Golden(t *testing.T) {
	// RFC 4231 case 1
	key := bytes.Repeat([]byte{0x0b}, 20)
	data := []byte("Hi There")
	want, _ := hex.DecodeString("b0344c61d8db38535ca8afceaf0bf12b881dc200c9833da726e9376c2e32cff7")
	got := macHash(macAlgoHMACSHA256, key, data)
	if !bytes.Equal(got, want) {
		t.Fatalf("got %x, want %x", got, want)
	}
	if !macVerify(macAlgoHMACSHA256, key, data, want) {
		t.Fatal("verify should succeed")
	}
	bad := append([]byte(nil), want...)
	bad[0] ^= 0xff
	if macVerify(macAlgoHMACSHA256, key, data, bad) {
		t.Fatal("verify should fail on flipped tag")
	}
}

func TestMacHashVerifyClassMethods(t *testing.T) {
	_, rp, f := cryptoFrame(t)
	typ := GetTypeRegistrar().Get("Mac")
	if typ == nil {
		t.Fatal("Mac not registered")
	}
	hashMM := typ.GetClassMethodMetadata("hash")
	verifyMM := typ.GetClassMethodMetadata("verify")
	if hashMM == nil || hashMM.restrict != SR_Crypto || verifyMM == nil || verifyMM.restrict != SR_Crypto {
		t.Fatalf("Mac.hash/verify SR: hash=%#v verify=%#v", hashMM, verifyMM)
	}
	if typ.GetClassPropertyMetadata("hmacSha256") == nil {
		t.Fatal("missing Mac.hmacSha256")
	}

	key := NewByteListFromBytes([]byte("secret"))
	msg := String("payload")
	ths := NewTypeObjPair(typ)

	rp.DataStack.Push(String(macAlgoHMACSHA256))
	rp.DataStack.Push(key)
	rp.DataStack.Push(msg)
	hashMM.invokeCall(f, ths)
	tag := rp.DataStack.Pop().(*ByteList)

	m := hmac.New(sha256.New, []byte("secret"))
	_, _ = m.Write([]byte("payload"))
	want := m.Sum(nil)
	if !bytes.Equal(tag.items, want) {
		t.Fatalf("tag = %x, want %x", tag.items, want)
	}

	rp.DataStack.Push(String(macAlgoHMACSHA256))
	rp.DataStack.Push(key)
	rp.DataStack.Push(msg)
	rp.DataStack.Push(tag)
	verifyMM.invokeCall(f, ths)
	if !bool(rp.DataStack.Pop().(Bool)) {
		t.Fatal("verify failed")
	}
}

func TestMacHashWithKeyType(t *testing.T) {
	_, rp, f := cryptoFrame(t)
	gen := GetTypeRegistrar().Get("Key").GetClassMethodMetadata("generate")
	rp.DataStack.Push(String(keyAlgoAES256GCM))
	gen.invokeCall(f, NewTypeObjPair(GetTypeRegistrar().Get("Key")))
	key := rp.DataStack.Pop().(*Key)

	hashMM := GetTypeRegistrar().Get("Mac").GetClassMethodMetadata("hash")
	rp.DataStack.Push(String(macAlgoHMACSHA256))
	rp.DataStack.Push(key)
	rp.DataStack.Push(String("msg"))
	hashMM.invokeCall(f, NewTypeObjPair(GetTypeRegistrar().Get("Mac")))
	tag := rp.DataStack.Pop().(*ByteList)
	if tag.Len() != sha256.Size {
		t.Fatalf("tag len = %d", tag.Len())
	}
}

func TestMacInterpret(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	runInterpretScript(t, task, rp,
		"Mac.hmacSha256 \"secret\" \"hello\" Mac.hash^ >ts\n"+
			"Mac.hmacSha256 \"secret\" \"hello\" t:peek Mac.verify^\n")
	if !bool(rp.DataStack.Pop().(Bool)) {
		t.Fatal("interpret verify failed")
	}
}

func TestMacHashRestricted(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	task.restrict = SR_Crypto
	rp := NewTestRunProvider()
	mm := GetTypeRegistrar().Get("Mac").GetClassMethodMetadata("hash")
	f := &RunFrame{Task: task, RP: rp}
	defer func() {
		if recover() != "operation not allowed" {
			t.Fatal("expected operation not allowed")
		}
	}()
	mm.invokeCall(f, NewTypeObjPair(GetTypeRegistrar().Get("Mac")))
	t.Fatal("expected panic")
}
