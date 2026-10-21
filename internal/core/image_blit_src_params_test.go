package core

import "testing"

func TestBlitParamsNew(t *testing.T) {
	p := NewBlitParams()
	if p.Mask() != BlitParamInherit || p.Op() != BlitParamInherit || p.CellPx() != BlitParamInherit {
		t.Fatalf("new = %+v, want all %d", *p, BlitParamInherit)
	}
	reg := GetTypeRegistrar().Get("BlitParams")
	if reg.GetClassMethodMetadata("new") == nil {
		t.Fatal("missing BlitParams.new")
	}
	if reg.GetClassMethodMetadata("newDefaults") != nil || reg.GetClassMethodMetadata("newOverride") != nil {
		t.Fatal("old factories should not be registered")
	}
}

func TestBlitParamsDefaultsGo(t *testing.T) {
	p := NewBlitParamsDefaults()
	if p.Mask() != BlitMaskOff || p.Op() != int(BlitStomp) || p.CellPx() != 1 {
		t.Fatalf("defaults = %+v, want off/stomp/1", *p)
	}
}

func TestBlitParamsClassConsts(t *testing.T) {
	reg := GetTypeRegistrar().Get("BlitParams")
	want := map[string]Int{
		"ParamUseDefault": Int(BlitParamInherit),
		"MaskOn":          Int(BlitMaskOn),
		"MaskOff":         Int(BlitMaskOff),
		"OpStomp":         Int(BlitStomp),
		"OpBlend":         Int(BlitBlend),
	}
	for name, v := range want {
		pm := reg.GetClassPropertyMetadata(name)
		if pm == nil {
			t.Fatalf("missing BlitParams.%s", name)
		}
		got := pm.get(nil).(Int)
		if got != v {
			t.Fatalf("BlitParams.%s = %d, want %d", name, got, v)
		}
		if pm.set != nil {
			t.Fatalf("BlitParams.%s should be readonly const", name)
		}
	}
	for _, gone := range []string{"DimClip", "DimWrap", "dimX", "dimY"} {
		if reg.GetClassPropertyMetadata(gone) != nil || reg.GetPropertyMetadata(gone) != nil {
			t.Fatalf("BlitParams.%s should be removed", gone)
		}
	}
}

func TestBlitParamsSetters(t *testing.T) {
	p := NewBlitParams()
	p.SetMask(BlitMaskOn)
	p.SetOp(int(BlitBlend))
	p.SetCellPx(4)
	if p.Mask() != BlitMaskOn || p.Op() != int(BlitBlend) || p.CellPx() != 4 {
		t.Fatalf("after sets = %+v", *p)
	}
}

func TestBlitParamsSettersRejectInherit(t *testing.T) {
	p := NewBlitParams()
	for _, tc := range []struct {
		name string
		fn   func()
	}{
		{"mask", func() { p.SetMask(BlitParamInherit) }},
		{"op", func() { p.SetOp(BlitParamInherit) }},
		{"cellPx", func() { p.SetCellPx(BlitParamInherit) }},
		{"cellPx 0", func() { p.SetCellPx(0) }},
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("%s: expected panic", tc.name)
				}
			}()
			tc.fn()
		}()
	}
}

func TestResolveBlitParamsInheritOverride(t *testing.T) {
	base := NewBlitParamsDefaults()
	over := NewBlitParams()
	over.SetMask(BlitMaskOn)
	r := resolveBlitParams(base, over)
	if r.Mask() != BlitMaskOn || r.Op() != int(BlitStomp) || r.CellPx() != 1 {
		t.Fatalf("resolved %+v", *r)
	}
}

func TestResolveBlitParamsRequiresBaseAndOver(t *testing.T) {
	base := NewBlitParamsDefaults()
	over := NewBlitParams()
	for _, tc := range []struct {
		name string
		base *BlitParams
		over *BlitParams
	}{
		{"nil over", base, nil},
		{"nil base", nil, over},
		{"both nil", nil, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("expected panic")
				}
			}()
			resolveBlitParams(tc.base, tc.over)
		})
	}
}

func TestResolveBlitParamsBaseInheritUsesDefinedDefault(t *testing.T) {
	base := &BlitParams{
		mask: BlitParamInherit, op: BlitParamInherit, cellPx: BlitParamInherit,
	}
	r := resolveBlitParams(base, NewBlitParams())
	if r.Mask() != BlitMaskOff || r.Op() != int(BlitStomp) || r.CellPx() != 1 {
		t.Fatalf("resolved %+v, want defined defaults", *r)
	}
}

func TestResolveBlitParamsOverrideWins(t *testing.T) {
	base := NewBlitParamsDefaults()
	over := NewBlitParams()
	over.SetOp(int(BlitBlend))
	r := resolveBlitParams(base, over)
	if r.Op() != int(BlitBlend) {
		t.Fatalf("resolved %+v", *r)
	}
}

func TestBlitSrcNew1AndSet(t *testing.T) {
	pal := NewPalette1(NewColor(0, 0, 0, 255), NewColor(255, 255, 255, 255))
	mat := NewMat2i1(2, 2)
	src := NewBlitSrc1(pal, mat)
	if src.Mat() != mat || src.Pal() != pal {
		t.Fatal("new1 mat/pal")
	}
	mat2 := NewMat2i1(3, 3)
	src.SetMat(mat2)
	if src.Mat() != mat2 {
		t.Fatal("setMat")
	}
	pal2 := NewPalette1(NewColor(1, 2, 3, 255), NewColor(4, 5, 6, 255))
	src.SetPal(pal2)
	if src.Pal() != pal2 {
		t.Fatal("setPal")
	}
}

func TestBlitSrcNewCNoPal(t *testing.T) {
	mat := NewMat2C(2, 2)
	src := NewBlitSrcC(mat)
	if src.Mat() != mat || src.Pal() != nil {
		t.Fatal("newC")
	}
	defer func() {
		if recover() == nil {
			t.Fatal("setPal on newC: expected panic")
		}
	}()
	src.SetPal(NewPalette1(Color{}, Color{}))
}

func TestBlitSrcSetMatWrongKindPanics(t *testing.T) {
	src := NewBlitSrc1(
		NewPalette1(Color{}, Color{}),
		NewMat2i1(1, 1),
	)
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	src.SetMat(NewMat2i4(1, 1))
}

func TestBlitSrcFactoriesRegistered(t *testing.T) {
	reg := GetTypeRegistrar().Get("BlitSrc")
	for _, name := range []string{"new1", "new4", "new8", "newC"} {
		if reg.GetClassMethodMetadata(name) == nil {
			t.Fatalf("missing BlitSrc.%s", name)
		}
	}
	for _, name := range []string{"mat", "pal"} {
		pm := reg.GetPropertyMetadata(name)
		if pm == nil {
			t.Fatalf("missing BlitSrc.%s property", name)
		}
		if pm.set == nil {
			t.Fatalf("BlitSrc.%s should be writable", name)
		}
	}
	if reg.GetMethodMetadata("setMat") != nil || reg.GetMethodMetadata("setPal") != nil {
		t.Fatal("setMat/setPal methods removed; use .mat! / .pal!")
	}
}

func TestBlitSrcNew4AndNew8(t *testing.T) {
	p4 := NewPalette4FromSlice(WebColors, 255)
	m4 := NewMat2i4(2, 2)
	s4 := NewBlitSrc4(p4, m4)
	if s4.Mat() != m4 || s4.Pal() != p4 {
		t.Fatal("new4")
	}
	p8 := NewPalette8FromSlice(make([]Color, 256))
	m8 := NewMat2u8(2, 2)
	s8 := NewBlitSrc8(p8, m8)
	if s8.Mat() != m8 || s8.Pal() != p8 {
		t.Fatal("new8")
	}
}
