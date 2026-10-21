package core

import "fmt"

// BlitParamInherit (-1) on a BlitParams override field means use the Pixmap.blit base.
const BlitParamInherit = -1

// Mask values (setters; not inherit).
const (
	BlitMaskOff = 0
	BlitMaskOn  = 1
	BlitMaskDef = BlitMaskOff
)

const (
	BlitOpDef     = int(BlitStomp)
	BlitCellPxDef = 1
)

// BlitParams is the blit policy bag: Pixmap.blit holds a concrete base;
// per-call overrides start as all Inherit via BlitParams.new.
// Dest overhang always uses intersection clip (like image/draw) — not a param.
type BlitParams struct {
	mask   int
	op     int
	cellPx int
}

func (p *BlitParams) FinalType() *Type { return blitParamsType }

// NewBlitParamsDefaults is the concrete Pixmap.blit base (no Inherit).
func NewBlitParamsDefaults() *BlitParams {
	return &BlitParams{
		mask:   BlitMaskDef,
		op:     BlitOpDef,
		cellPx: BlitCellPxDef,
	}
}

// NewBlitParams is a per-source override bag (all Inherit). Final: BlitParams.new.
func NewBlitParams() *BlitParams {
	return &BlitParams{
		mask:   BlitParamInherit,
		op:     BlitParamInherit,
		cellPx: BlitParamInherit,
	}
}

func (p *BlitParams) Mask() int   { return p.mask }
func (p *BlitParams) Op() int     { return p.op }
func (p *BlitParams) CellPx() int { return p.cellPx }

func (p *BlitParams) SetMask(v int) {
	requireBlitMask(v)
	p.mask = v
}

func (p *BlitParams) SetOp(v int) {
	requireBlitParamsOp(v)
	p.op = v
}

func (p *BlitParams) SetCellPx(v int) {
	requireBlitParamsCellPx(v)
	p.cellPx = v
}

func requireBlitMask(v int) {
	if v == BlitMaskOff || v == BlitMaskOn {
		return
	}
	panic(fmt.Errorf("BlitParams.mask: want 0 (off) or 1 (on), got %d", v))
}

func requireBlitParamsOp(v int) {
	if v == int(BlitBlend) || v == int(BlitStomp) {
		return
	}
	panic(fmt.Errorf("BlitParams.op: want 0 (blend) or 1 (stomp), got %d", v))
}

func requireBlitParamsCellPx(v int) {
	if v >= 1 {
		return
	}
	panic(fmt.Errorf("BlitParams.cellPx: want >= 1, got %d", v))
}

func blitResolveParam(base, over, def int) int {
	if over != BlitParamInherit {
		return over
	}
	if base != BlitParamInherit {
		return base
	}
	return def
}

// resolveBlitParams merges Pixmap.params base with a per-call override.
// Both base and over are required (panic if nil) — call only from BlitWithOverrides.
// Inherit (-1) on override uses base; Inherit on base uses defined defaults.
// Result never contains Inherit. Not exposed as a Final method.
func resolveBlitParams(base, over *BlitParams) *BlitParams {
	if base == nil || over == nil {
		panic(fmt.Errorf("resolveBlitParams: base and over are both required"))
	}
	return &BlitParams{
		mask:   blitResolveParam(base.mask, over.mask, BlitMaskDef),
		op:     blitResolveParam(base.op, over.op, BlitOpDef),
		cellPx: blitResolveParam(base.cellPx, over.cellPx, BlitCellPxDef),
	}
}

func registerBlitParamsMembers(t *Type) {
	t.registerProperty("mask", "Int",
		func(inst Any) Any { return Int(inst.(*BlitParams).Mask()) },
		func(inst Any, value Any) { inst.(*BlitParams).SetMask(int(value.(Int))) },
	)
	t.registerProperty("op", "Int",
		func(inst Any) Any { return Int(inst.(*BlitParams).Op()) },
		func(inst Any, value Any) { inst.(*BlitParams).SetOp(int(value.(Int))) },
	)
	t.registerProperty("cellPx", "Int",
		func(inst Any) Any { return Int(inst.(*BlitParams).CellPx()) },
		func(inst Any, value Any) { inst.(*BlitParams).SetCellPx(int(value.(Int))) },
	)
	registerNamedConst(t, "Int", "ParamUseDefault", Int(BlitParamInherit))
	registerNamedConst(t, "Int", "MaskOn", Int(BlitMaskOn))
	registerNamedConst(t, "Int", "MaskOff", Int(BlitMaskOff))
	registerNamedConst(t, "Int", "OpStomp", Int(BlitStomp))
	registerNamedConst(t, "Int", "OpBlend", Int(BlitBlend))
	// ( -- p:BlitParams ) all fields Inherit; set only fields to override
	t.registerNativeClassMethod("new", func(f *RunFrame, _ *TypeObjPair) {
		f.RP.DataStack.Push(NewBlitParams())
	})
}

func CreateRegisterBlitParamsType() {
	t := NewRegisteringType("BlitParams",
		ReferenceTypeZeroFn,
		MakePointerTypeCheckCoerce[BlitParams]("BlitParams"),
	)
	registerBlitParamsMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	blitParamsType = t
}

var (
	blitParamsType *Type
	_              Typed = (*BlitParams)(nil)
)

func init() {
	CreateRegisterBlitParamsType()
}
