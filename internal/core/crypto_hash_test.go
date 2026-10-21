package core

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestHashDigestGolden(t *testing.T) {
	// RFC 1321 / FIPS 180-4 style empty and "abc" vectors.
	cases := []struct {
		algo string
		in   []byte
		want string // hex
	}{
		{hashAlgoMD5, []byte(""), "d41d8cd98f00b204e9800998ecf8427e"},
		{hashAlgoMD5, []byte("abc"), "900150983cd24fb0d6963f7d28e17f72"},
		{hashAlgoSHA256, []byte(""), "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{hashAlgoSHA256, []byte("abc"), "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"},
	}
	for _, tc := range cases {
		got := hex.EncodeToString(hashDigest(tc.algo, tc.in))
		if got != tc.want {
			t.Fatalf("%s(%q) = %s, want %s", tc.algo, tc.in, got, tc.want)
		}
	}
}

func TestHashHashClassMethod(t *testing.T) {
	ResetTestRuntime()
	typ := GetTypeRegistrar().Get("Hash")
	if typ == nil {
		t.Fatal("Hash type not registered")
	}
	mm := typ.GetClassMethodMetadata("hash")
	if mm == nil || mm.restrict != SR_Crypto {
		t.Fatalf("Hash.hash restrict = %#v, want SR_Crypto", mm)
	}
	for _, name := range []string{"md5", "sha256"} {
		if typ.GetClassPropertyMetadata(name) == nil {
			t.Fatalf("missing Hash.%s", name)
		}
	}

	task := NewTestTask()
	rp := NewTestRunProvider()
	f := &RunFrame{Task: task, RP: rp}
	ths := NewTypeObjPair(typ)

	rp.DataStack.Push(String("abc"))
	rp.DataStack.Push(String(hashAlgoSHA256))
	mm.invokeCall(f, ths)
	digest := rp.DataStack.Pop().(*ByteList)
	want, _ := hex.DecodeString("ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad")
	if !bytes.Equal(digest.items, want) {
		t.Fatalf("digest = %x, want %x", digest.items, want)
	}

	rp.DataStack.Push(NewByteListFromBytes([]byte("abc")))
	rp.DataStack.Push(String(hashAlgoMD5))
	mm.invokeCall(f, ths)
	digest = rp.DataStack.Pop().(*ByteList)
	want, _ = hex.DecodeString("900150983cd24fb0d6963f7d28e17f72")
	if !bytes.Equal(digest.items, want) {
		t.Fatalf("md5 digest = %x, want %x", digest.items, want)
	}
}

func TestHashHashInterpret(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	runInterpretScript(t, task, rp, "\"abc\" Hash.sha256 Hash.hash^ Hex.encode^\n")
	got := string(rp.DataStack.Pop().(String))
	want := "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestHashHashRestricted(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	task.restrict = SR_Crypto
	rp := NewTestRunProvider()
	mm := GetTypeRegistrar().Get("Hash").GetClassMethodMetadata("hash")
	f := &RunFrame{Task: task, RP: rp}
	defer func() {
		r := recover()
		if r != "operation not allowed" {
			t.Fatalf("panic = %v, want %q", r, "operation not allowed")
		}
	}()
	mm.invokeCall(f, NewTypeObjPair(GetTypeRegistrar().Get("Hash")))
	t.Fatal("expected panic")
}

func TestHexBase64RoundTrip(t *testing.T) {
	ResetTestRuntime()
	task := NewTestTask()
	rp := NewTestRunProvider()
	runInterpretScript(t, task, rp,
		"\"hello\" Hex.encode^ Hex.decode^ Hex.encode^\n"+
			"\"hello\" Base64.encode^ Base64.decode^ Base64.encode^\n")
	b64 := string(rp.DataStack.Pop().(String))
	hx := string(rp.DataStack.Pop().(String))
	if hx != "68656c6c6f" {
		t.Fatalf("hex round-trip encode = %q, want 68656c6c6f", hx)
	}
	if b64 != "aGVsbG8=" {
		t.Fatalf("base64 round-trip encode = %q, want aGVsbG8=", b64)
	}
}

func TestHexEncodeDigest(t *testing.T) {
	ResetTestRuntime()
	typ := GetTypeRegistrar().Get("Hex")
	mm := typ.GetClassMethodMetadata("encode")
	if mm == nil || mm.restrict != SR_Crypto {
		t.Fatalf("Hex.encode restrict = %#v, want SR_Crypto", mm)
	}
	task := NewTestTask()
	rp := NewTestRunProvider()
	f := &RunFrame{Task: task, RP: rp}
	rp.DataStack.Push(NewByteListFromBytes([]byte{0xde, 0xad}))
	mm.invokeCall(f, NewTypeObjPair(typ))
	if got := string(rp.DataStack.Pop().(String)); got != "dead" {
		t.Fatalf("Hex.encode = %q, want dead", got)
	}
}
