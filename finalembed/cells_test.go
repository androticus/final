package finalembed_test

import (
	"testing"
	"time"

	"github.com/androticus/final/finalembed"
)

func TestToCell(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want any
	}{
		{"nil", nil, nil},
		{"Int pass-through", finalembed.Int(7), finalembed.Int(7)},
		{"String pass-through", finalembed.String("hi"), finalembed.String("hi")},
		{"int32 widen", int32(3), finalembed.Int(3)},
		{"int64 widen", int64(-1), finalembed.Int(-1)},
		{"uint8 widen", uint8(255), finalembed.Uint(255)},
		{"float32 widen", float32(1.5), finalembed.Float(1.5)},
		{"float64 widen", float64(2.5), finalembed.Float(2.5)},
		{"bool widen", true, finalembed.Bool(true)},
		{"string widen", "k", finalembed.String("k")},
		{"DateTime pass-through", time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)},
		{"Color pass-through", finalembed.NewColor(1, 2, 3, 4), finalembed.NewColor(1, 2, 3, 4)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := finalembed.ToCell(tc.in)
			if got != tc.want {
				t.Fatalf("ToCell(%v) = %v (%T), want %v (%T)", tc.in, got, got, tc.want, tc.want)
			}
		})
	}
}

func TestNewAnyKeyValue(t *testing.T) {
	kv := finalembed.NewAnyKeyValue("k", int32(1))
	if kv.Key != finalembed.String("k") {
		t.Fatalf("Key = %v (%T), want String(k)", kv.Key, kv.Key)
	}
	if kv.Value != finalembed.Int(1) {
		t.Fatalf("Value = %v (%T), want Int(1)", kv.Value, kv.Value)
	}
}

type stubIterable struct{}

type stubIterator struct{}

func (stubIterable) GetIter() finalembed.Iterator { return stubIterator{} }

func (stubIterator) GetNext() bool { return false }

func (stubIterator) Current() any { return nil }

func TestIterableInterface(t *testing.T) {
	var _ finalembed.Iterable = stubIterable{}
}
