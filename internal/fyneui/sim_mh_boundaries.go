package fyneui

import (
	"github.com/androticus/final/internal/core"
)

// EnsureHitBoundaries creates/caches 0–4 engine-owned kindBorder HitRect slabs
// for sides with HitAction ≠ none. Not on SimGame.entities; Phase 4 splices
// BorderEntities into the hit-test snapshot.
//
// Idempotent: rebuilds when pixSize or the HitAction side mask changes.
// Geometry: flush outside playable [0,w)×[0,h); thickness ≈ half-span; slabs
// overshoot endpoints so adjacent walls overlap at corners.
func (e *SimMHEngine) EnsureHitBoundaries() {
	sg := e.sg
	if sg == nil || sg.options == nil {
		return
	}
	// pixSize is discrete (SetPixSize requires integral); use int cells for slabs.
	w := core.RequireIntegralDim("SimMHEngine.pixSize width", sg.pixSize.W)
	h := core.RequireIntegralDim("SimMHEngine.pixSize height", sg.pixSize.H)
	mask := hitActionSideMask(sg.options.boundaryModes)
	if e.borderW == w && e.borderH == h && e.borderHitMask == mask && e.bordersBuilt {
		return
	}
	e.clearBorderEntities()
	e.borderW = w
	e.borderH = h
	e.borderHitMask = mask
	e.bordersBuilt = true
	if mask == 0 || w < 1 || h < 1 {
		return
	}
	tx := w / 2
	if tx < 1 {
		tx = 1
	}
	ty := h / 2
	if ty < 1 {
		ty = 1
	}
	// Overshoot by tx/ty on each end so corner regions are covered by both slabs.
	modes := sg.options.boundaryModes
	if modes[BoundaryIndexLeft].HitAction() != BoundaryHitActionNone {
		e.addBorderSlab(sg, BoundaryIndexLeft,
			core.NewPoint(-tx, -ty),
			core.NewSize(tx, h+2*ty),
		)
	}
	if modes[BoundaryIndexTop].HitAction() != BoundaryHitActionNone {
		e.addBorderSlab(sg, BoundaryIndexTop,
			core.NewPoint(-tx, -ty),
			core.NewSize(w+2*tx, ty),
		)
	}
	if modes[BoundaryIndexRight].HitAction() != BoundaryHitActionNone {
		e.addBorderSlab(sg, BoundaryIndexRight,
			core.NewPoint(w, -ty),
			core.NewSize(tx, h+2*ty),
		)
	}
	if modes[BoundaryIndexBottom].HitAction() != BoundaryHitActionNone {
		e.addBorderSlab(sg, BoundaryIndexBottom,
			core.NewPoint(-tx, h),
			core.NewSize(w+2*tx, ty),
		)
	}
}

// BorderEntities returns the cached engine-owned border slabs (0–4).
// Callers must not mutate the slice; treat entities as read-only for Hits.
func (e *SimMHEngine) BorderEntities() []*SimEntity {
	return e.borders
}

func (e *SimMHEngine) clearBorderEntities() {
	for _, b := range e.borders {
		if b != nil {
			b.active = false
			b.sg = nil
		}
	}
	e.borders = e.borders[:0]
}

func (e *SimMHEngine) addBorderSlab(sg *SimGame, side BoundaryIndex, topLeft core.Point, size core.Size) {
	ent := NewSimEntity()
	ent.kind = EntityKindBorder
	ent.subkind = int32(side) // which side (for debug / optional handler use)
	ent.anchored = true
	ent.visible = false
	ent.enabled = true
	ent.sg = sg
	ent.SetAt(topLeft)
	ent.SetHitArea(NewRectHitArea(HitKindBorder, core.NewRect(core.Point{}, size)))
	ent.active = true // after HitArea; not on SimGame.entities
	e.borders = append(e.borders, ent)
}

func hitActionSideMask(modes [4]BoundaryMode) uint8 {
	var m uint8
	for i := BoundaryIndexLeft; i <= BoundaryIndexBottom; i++ {
		if modes[i].HitAction() != BoundaryHitActionNone {
			m |= 1 << uint(i)
		}
	}
	return m
}

// EnsureHitBoundaries rebuilds engine border HitRects from options + pixSize.
func (sg *SimGame) EnsureHitBoundaries() {
	if sg.mhe != nil {
		sg.mhe.EnsureHitBoundaries()
	}
}
