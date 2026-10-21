package core

import "fmt"

// Shared crypto helpers (Phase 0+). Artifact types live in crypto_*.go.
// Design: docs/003-design-notes/154-crypto.md

// cryptoBytesFromStackCell accepts ByteList or String (UTF-8 bytes).
func cryptoBytesFromStackCell(v any) []byte {
	switch x := v.(type) {
	case *ByteList:
		return append([]byte(nil), x.items...)
	case String:
		return []byte(string(x))
	default:
		panic(fmt.Errorf("crypto: expected ByteList or String, got %T", v))
	}
}

// cryptoDORequiredBytes reads a required ByteList|String field from a params DynamicObject.
func cryptoDORequiredBytes(o *DynamicObject, field string) []byte {
	if o == nil {
		panic(fmt.Errorf("crypto: nil params DynamicObject"))
	}
	if !bool(o.HasField(field)) {
		panic(fmt.Errorf("crypto: params missing required field %q", field))
	}
	return cryptoBytesFromStackCell(o.Get(field))
}

// cryptoDOOptionalBytes reads an optional ByteList|String field; missing → nil.
func cryptoDOOptionalBytes(o *DynamicObject, field string) []byte {
	if o == nil || !bool(o.HasField(field)) {
		return nil
	}
	return cryptoBytesFromStackCell(o.Get(field))
}

// cryptoDORequiredInt reads a required Int field from a params DynamicObject.
func cryptoDORequiredInt(o *DynamicObject, field string) int {
	if o == nil {
		panic(fmt.Errorf("crypto: nil params DynamicObject"))
	}
	if !bool(o.HasField(field)) {
		panic(fmt.Errorf("crypto: params missing required field %q", field))
	}
	v := o.Get(field)
	n, ok := v.(Int)
	if !ok {
		panic(fmt.Errorf("crypto: params field %q must be Int, got %T", field, v))
	}
	return int(n)
}

// cryptoDOOptionalInt reads an optional Int field; missing → def.
func cryptoDOOptionalInt(o *DynamicObject, field string, def int) int {
	if o == nil || !bool(o.HasField(field)) {
		return def
	}
	v := o.Get(field)
	n, ok := v.(Int)
	if !ok {
		panic(fmt.Errorf("crypto: params field %q must be Int, got %T", field, v))
	}
	return int(n)
}
