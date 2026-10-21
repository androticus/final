package core

import "testing"

func TestTextWriterInheritsTextWriteMethods(t *testing.T) {
	tw := GetTypeRegistrar().Get("TextWriter")
	if tw == nil {
		t.Fatal("TextWriter not registered")
	}
	if tw.GetParent() != textWriteType {
		t.Fatalf("TextWriter parent = %v, want TextWrite", tw.GetParent())
	}
	if textWriteType.GetParent() != streamReadWriteType {
		t.Fatalf("TextWrite parent = %v, want StreamReadWrite", textWriteType.GetParent())
	}
	if tw.GetMethodMetadata("writeString") == nil {
		t.Fatal("TextWriter should inherit writeString from TextWrite")
	}
	if tw.GetMethodMetadata("writeLine") == nil {
		t.Fatal("TextWriter should inherit writeLine from TextWrite")
	}
	if tw.GetMethodMetadata("writeRune") == nil {
		t.Fatal("TextWriter should inherit writeRune from TextWrite")
	}
	if tw.GetMethodMetadata("close") == nil {
		t.Fatal("TextWriter should inherit close from StreamReadWrite")
	}
	if tw.GetMethodMetadata("flush") == nil {
		t.Fatal("TextWriter should inherit flush from StreamReadWrite")
	}
}

func TestByteWriterInheritsByteWriteAndStreamReadWrite(t *testing.T) {
	bw := GetTypeRegistrar().Get("ByteWriter")
	if bw == nil {
		t.Fatal("ByteWriter not registered")
	}
	if bw.GetParent() != byteWriteType {
		t.Fatalf("ByteWriter parent = %v, want ByteWrite", bw.GetParent())
	}
	if byteWriteType.GetParent() != streamReadWriteType {
		t.Fatalf("ByteWrite parent = %v, want StreamReadWrite", byteWriteType.GetParent())
	}
	if bw.GetMethodMetadata("writeBytes") == nil {
		t.Fatal("ByteWriter should inherit writeBytes from ByteWrite")
	}
	if bw.GetMethodMetadata("flush") == nil {
		t.Fatal("ByteWriter should inherit flush from StreamReadWrite")
	}
}

func TestTextReaderParentsTextRead(t *testing.T) {
	tr := GetTypeRegistrar().Get("TextReader")
	if tr == nil {
		t.Fatal("TextReader not registered")
	}
	if tr.GetParent() != textReadType {
		t.Fatalf("TextReader parent = %v, want TextRead", tr.GetParent())
	}
	if textReadType.GetParent() != streamReadWriteType {
		t.Fatalf("TextRead parent = %v, want StreamReadWrite", textReadType.GetParent())
	}
	if tr.GetMethodMetadata("getNextRune") == nil {
		t.Fatal("TextReader should inherit getNextRune from TextRead")
	}
	if tr.GetMethodMetadata("readLine") == nil {
		t.Fatal("TextReader should declare readLine")
	}
	if tr.GetMethodMetadata("close") == nil {
		t.Fatal("TextReader should inherit close from StreamReadWrite")
	}
}

func TestByteReaderParentsByteRead(t *testing.T) {
	br := GetTypeRegistrar().Get("ByteReader")
	if br == nil {
		t.Fatal("ByteReader not registered")
	}
	if br.GetParent() != byteReadType {
		t.Fatalf("ByteReader parent = %v, want ByteRead", br.GetParent())
	}
	if byteReadType.GetParent() != streamReadWriteType {
		t.Fatalf("ByteRead parent = %v, want StreamReadWrite", byteReadType.GetParent())
	}
	if br.GetMethodMetadata("getNextByte") == nil {
		t.Fatal("ByteReader should inherit getNextByte from ByteRead")
	}
	if br.GetMethodMetadata("readBytes") == nil {
		t.Fatal("ByteReader should inherit readBytes from ByteRead")
	}
}

func TestStringBuilderParentsStreamReadWriteNotTextWrite(t *testing.T) {
	sb := GetTypeRegistrar().Get("StringBuilder")
	if sb == nil {
		t.Fatal("StringBuilder not registered")
	}
	if sb.GetParent() != streamReadWriteType {
		t.Fatalf("StringBuilder parent = %v, want StreamReadWrite", sb.GetParent())
	}
	if sb.GetMethodMetadata("close") == nil {
		t.Fatal("StringBuilder should inherit close from StreamReadWrite")
	}
	if sb.GetMethodMetadata("writeString") == nil {
		t.Fatal("StringBuilder should keep local writeString")
	}
}

func TestLineBufferTextWriterInheritsTextWriteNotTextWriterFactories(t *testing.T) {
	lb := GetTypeRegistrar().Get("LineBufferTextWriter")
	if lb == nil {
		t.Fatal("LineBufferTextWriter not registered")
	}
	if lb.GetParent() != textWriteType {
		t.Fatalf("LineBufferTextWriter parent = %v, want TextWrite", lb.GetParent())
	}
	if lb.GetMethodMetadata("writeString") == nil {
		t.Fatal("LineBufferTextWriter should inherit writeString from TextWrite")
	}
	if lb.GetMethodMetadata("flush") == nil {
		t.Fatal("LineBufferTextWriter should inherit flush from StreamReadWrite")
	}
	if lb.GetMethodMetadata("toString") == nil {
		t.Fatal("LineBufferTextWriter should declare toString")
	}
	if lb.GetClassMethodMetadata("newFromFileAppendOrCreate") != nil {
		t.Fatal("LineBufferTextWriter must not inherit TextWriter file factories")
	}
}

func TestInterpretLineBufferTextWriterWriteAndToString(t *testing.T) {
	ResetTestRuntime()
	task, rp, _ := testParserWithInput(t,
		"3 LineBufferTextWriter.new^ >ts \"ab\" t:.writeLine^ \"cd\" t:.writeLine^ t:.toString^\n")
	if err := interpretSession(task, rp); err != nil {
		t.Fatal(err)
	}
	got, ok := rp.DataStack.Pop().(String)
	if !ok || got != "ab\ncd\n" {
		t.Fatalf("result = %q, want %q", got, "ab\ncd\n")
	}
}
