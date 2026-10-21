package core

import (
	"strings"
	"testing"
)

func TestTypePropertyMethodNameCollisionPanics(t *testing.T) {
	t.Parallel()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for property/method name collision")
		}
		if !strings.Contains(r.(error).Error(), "collides") {
			t.Fatalf("panic = %v", r)
		}
	}()
	probe := &Type{name: "Probe"}
	probe.AddProperty("x", &PropertyMetadata{name: "x"}, "Any")
	probe.AddMethod("x", &MethodMetadata{name: "x", call: func(_ *RunFrame, _ *TypeObjPair) {}})
}
