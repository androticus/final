package fyneui

import (
	"math"
	"reflect"

	"github.com/androticus/final/internal/core"
)

// hitPairKey is a canonical unordered HitArea pointer pair.
type hitPairKey struct {
	lo, hi uintptr
}

func hitPairKeyOf(a, b HitArea) hitPairKey {
	pa := reflect.ValueOf(a).Pointer()
	pb := reflect.ValueOf(b).Pointer()
	if pa > pb {
		pa, pb = pb, pa
	}
	return hitPairKey{lo: pa, hi: pb}
}

// ResolveHitPairMode returns HitAction/HitEvents policy for two entities' HitAreas.
// Border pairs: exact matrix cell, else seed from the border side's packed mode.
func (o *SimGameOptions) ResolveHitPairMode(ea, eb *SimEntity) BoundaryMode {
	ka := ea.hitArea.Kind()
	kb := eb.hitArea.Kind()
	if ka == HitKindBorder || kb == HitKindBorder {
		if m, ok := o.lookupHitMatrixExact(ka, kb); ok {
			return m
		}
		var border *SimEntity
		switch {
		case ka == HitKindBorder && kb != HitKindBorder:
			border = ea
		case kb == HitKindBorder && ka != HitKindBorder:
			border = eb
		default:
			return DefaultHitMatrixMode
		}
		side := BoundaryIndex(border.subkind)
		if side < BoundaryIndexLeft || side > BoundaryIndexBottom {
			return DefaultHitMatrixMode
		}
		return o.boundaryModes[side]
	}
	return o.LookupHitMatrix(ka, kb)
}

func (o *SimGameOptions) lookupHitMatrixExact(ka, kb int32) (BoundaryMode, bool) {
	if o.hitMatrix == nil {
		return 0, false
	}
	m, ok := o.hitMatrix[hitMatrixCanonical(ka, kb)]
	return m, ok
}

// stepHits runs the Hits pass: pair tests, contacts, matrix resolve, optional SimHits.
func (e *SimMHEngine) stepHits(gameplay []*SimEntity) *SimHits {
	sg := e.sg
	if sg == nil || sg.options == nil {
		return nil
	}
	if e.contacts == nil {
		e.contacts = make(map[hitPairKey]uint64)
	}
	e.contactGen++
	gen := e.contactGen

	hitters := make([]*SimEntity, 0, len(gameplay)+len(e.borders))
	for _, ent := range gameplay {
		if ent != nil && ent.active && ent.enabled && ent.hitArea != nil {
			hitters = append(hitters, ent)
		}
	}
	for _, b := range e.borders {
		if b != nil && b.active && b.hitArea != nil {
			hitters = append(hitters, b)
		}
	}

	var hits *SimHits
	n := len(hitters)
	for i := 0; i < n; i++ {
		a := hitters[i]
		for j := i + 1; j < n; j++ {
			b := hitters[j]
			if a.kind == EntityKindBorder && b.kind == EntityKindBorder {
				continue
			}
			g := hitTestPair(a, b)
			if !g.ok {
				continue
			}
			key := hitPairKeyOf(a.hitArea, b.hitArea)
			_, existed := e.contacts[key]
			e.contacts[key] = gen
			mode := sg.options.ResolveHitPairMode(a, b)
			action := mode.HitAction()
			if existed {
				// stay: no Event; still depenetrate / fix closing velocity so
				// a later rock nudge cannot tunnel through a wall contact.
				if action == BoundaryHitActionBounce {
					applyBounce(a, b, g.nx, g.ny, g.pen)
				}
				continue
			}
			// enter
			switch action {
			case BoundaryHitActionNone:
				// track contact only
			case BoundaryHitActionStop:
				applyStop(a, b)
			case BoundaryHitActionBounce:
				applyBounce(a, b, g.nx, g.ny, g.pen)
			}
			if action != BoundaryHitActionNone && mode.HitEvents() {
				hits = ensureHits(hits)
				hits.Add(makeSimHit(a, b, g))
			}
		}
	}

	// Sweep exits
	for k, stamped := range e.contacts {
		if stamped != gen {
			delete(e.contacts, k)
		}
	}
	return hits
}

func ensureHits(h *SimHits) *SimHits {
	if h == nil {
		return NewSimHits()
	}
	return h
}

func makeSimHit(a, b *SimEntity, g hitGeom) *SimHit {
	e1, e2 := orderHitSlots(a, b)
	nx, ny := g.nx, g.ny
	// g.n is from a toward b; SimHit normal is from entity2 toward entity1.
	if e1 == a {
		nx, ny = -nx, -ny
	}
	return NewSimHit(e1, e2,
		core.NewPointF(g.atx, g.aty),
		normalToPoint(nx, ny),
	)
}

func normalToPoint(nx, ny float64) core.Point {
	// Prefer axis unit when nearly axis-aligned; else milliscale.
	if math.Abs(nx) >= math.Abs(ny) {
		if math.Abs(ny) < 1e-3 {
			if nx < 0 {
				return core.Point{X: -1, Y: 0}
			}
			return core.Point{X: 1, Y: 0}
		}
	} else if math.Abs(nx) < 1e-3 {
		if ny < 0 {
			return core.Point{X: 0, Y: -1}
		}
		return core.Point{X: 0, Y: 1}
	}
	return core.NewPointF(nx*1000, ny*1000)
}
