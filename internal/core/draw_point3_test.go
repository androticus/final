package core

import (
	"testing"
)

func TestNewPoint3(t *testing.T) {
	p := NewPoint3(3, -4, 5)
	if p.X != 3 || p.Y != -4 || p.Z != 5 {
		t.Fatalf("NewPoint3 = %v, want (3, -4, 5)", p)
	}
	f := NewPoint3F(1.5, 2.25, -0.5)
	if f.X != 1.5 || f.Y != 2.25 || f.Z != -0.5 {
		t.Fatalf("NewPoint3F = %v, want (1.5, 2.25, -0.5)", f)
	}
	if NewPoint3(1, 2, 3) != NewPoint3F(1, 2, 3) {
		t.Fatal("int and float constructors disagree on integral coordinates")
	}
}

func TestPoint3Round(t *testing.T) {
	// math.Round: halfway cases go away from zero.
	p := NewPoint3F(1.4, 1.5, -1.5)
	if p.Xi() != 1 || p.Yi() != 2 || p.Zi() != -2 {
		t.Fatalf("Xi/Yi/Zi = %d/%d/%d, want 1/2/-2", p.Xi(), p.Yi(), p.Zi())
	}
	q := NewPoint3F(-1.4, 2.5, -2.5)
	if q.Xi() != -1 || q.Yi() != 3 || q.Zi() != -3 {
		t.Fatalf("Xi/Yi/Zi = %d/%d/%d, want -1/3/-3", q.Xi(), q.Yi(), q.Zi())
	}
}

func TestPoint3String(t *testing.T) {
	if got := NewPoint3(1, 2, 3).String(); got != "(1, 2, 3)" {
		t.Fatalf("String = %q, want (1, 2, 3)", got)
	}
	if got := NewPoint3F(1.5, 0, -2.25).String(); got != "(1.5, 0, -2.25)" {
		t.Fatalf("String = %q, want (1.5, 0, -2.25)", got)
	}
}

func TestPoint3Registered(t *testing.T) {
	typ := GetTypeRegistrar().Get("Point3")
	if typ == nil {
		t.Fatal("Point3 type not registered")
	}
	if typ != point3Type {
		t.Fatal("registrar Point3 != point3Type")
	}
	p := NewPoint3(1, 2, 3)
	if p.FinalType() != point3Type {
		t.Fatal("Point3.FinalType mismatch")
	}
	if got := typ.Zero(); got != (Point3{}) {
		t.Fatalf("Zero = %#v, want Point3{}", got)
	}
	if got := typ.CheckCoerceType(p); got != p {
		t.Fatalf("CheckCoerceType(Point3) = %#v", got)
	}
	if got := typ.CheckCoerceType(nil); got != nil {
		t.Fatalf("CheckCoerceType(nil) = %#v, want nil", got)
	}
	defer func() {
		if recover() == nil {
			t.Fatal("CheckCoerceType(Point) should panic")
		}
	}()
	_ = typ.CheckCoerceType(NewPoint(0, 0))
}

func TestPoint3Properties(t *testing.T) {
	typ := GetTypeRegistrar().Get("Point3")
	p := NewPoint3F(1.4, 1.5, -1.5)
	want := map[string]any{
		"x":  Float(1.4),
		"y":  Float(1.5),
		"z":  Float(-1.5),
		"xi": Int(1),
		"yi": Int(2),
		"zi": Int(-2),
	}
	for name, wantVal := range want {
		pm := typ.GetPropertyMetadata(name)
		if pm == nil {
			t.Fatalf("missing Point3.%s", name)
		}
		if pm.set != nil {
			t.Fatalf("Point3.%s should be const", name)
		}
		if got := pm.GetValue(p); got != wantVal {
			t.Fatalf("Point3.%s = %v (%T), want %v", name, got, got, wantVal)
		}
	}
}

func TestPoint3PropertyReadonly(t *testing.T) {
	pm := point3Type.GetPropertyMetadata("x")
	defer func() {
		if recover() == nil {
			t.Fatal("setting Point3.x should panic")
		}
	}()
	pm.SetValue(NewPoint3(1, 2, 3), Float(9))
}

func TestPoint3ZeroClassProperty(t *testing.T) {
	pm := point3Type.GetClassPropertyMetadata("zero")
	if pm == nil {
		t.Fatal("missing Point3.zero")
	}
	if pm.set != nil {
		t.Fatal("Point3.zero should be const")
	}
	if got := pm.GetValue(point3Type); got != (Point3{}) {
		t.Fatalf("Point3.zero = %#v, want Point3{}", got)
	}
}

func TestPoint3ClassMethodNew(t *testing.T) {
	mm := point3Type.ClassMethodByName("new")
	cases := []struct {
		name string
		push func(*Stack[any])
		want Point3
	}{
		{
			"ints",
			func(ds *Stack[any]) {
				ds.Push(Int(3))
				ds.Push(Int(-4))
				ds.Push(Int(5))
			},
			NewPoint3(3, -4, 5),
		},
		{
			"uints",
			func(ds *Stack[any]) {
				ds.Push(Uint(1))
				ds.Push(Uint(2))
				ds.Push(Uint(3))
			},
			NewPoint3(1, 2, 3),
		},
		{
			"floats",
			func(ds *Stack[any]) {
				ds.Push(Float(1.5))
				ds.Push(Float(2.25))
				ds.Push(Float(-0.5))
			},
			NewPoint3F(1.5, 2.25, -0.5),
		},
		{
			"mixed",
			func(ds *Stack[any]) {
				ds.Push(Int(1))
				ds.Push(Uint(2))
				ds.Push(Float(3.5))
			},
			NewPoint3F(1, 2, 3.5),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rp := NewTestRunProvider()
			f, ths := memberRunFrame(rp, point3Type)
			tc.push(rp.DataStack)
			mm.invokeCall(f, ths)
			if rp.DataStack.Len() != 1 {
				t.Fatalf("DS len = %d, want 1", rp.DataStack.Len())
			}
			got := rp.DataStack.Pop().(Point3)
			if got != tc.want {
				t.Fatalf("Point3.new = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestPoint3ClassMethodNewRejectsNonNumeric(t *testing.T) {
	mm := point3Type.ClassMethodByName("new")
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, point3Type)
	rp.DataStack.Push(Int(1))
	rp.DataStack.Push(Int(2))
	rp.DataStack.Push(String("3"))
	defer func() {
		if recover() == nil {
			t.Fatal("Point3.new with a string should panic")
		}
	}()
	mm.invokeCall(f, ths)
}
