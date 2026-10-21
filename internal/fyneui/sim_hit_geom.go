package fyneui

import (
	"math"

	"github.com/androticus/final/internal/core"
)

const (
	hitEps     = 1e-9
	hitSepSlop = 0.5 // pixels of extra separation after bounce
)

// hitGeom is discrete overlap + contact for one ordered pair (a, b).
// n is unit from a toward b; at is WC contact; pen is penetration along n.
type hitGeom struct {
	ok      bool
	nx, ny  float64
	atx, aty float64
	pen     float64
}

func circleWorld(e *SimEntity) (cx, cy, r float64) {
	c := e.hitArea.(*CircleHitArea)
	return e._at.X + c.center.X, e._at.Y + c.center.Y, c.radius
}

func rectWorld(e *SimEntity) (left, top, right, bottom float64) {
	r := e.hitArea.(*RectHitArea).rect
	left = e._at.X + r.TopLeft.X
	top = e._at.Y + r.TopLeft.Y
	right = left + r.Size.W
	bottom = top + r.Size.H
	return
}

func hitTestPair(a, b *SimEntity) hitGeom {
	sa, sb := a.hitArea.Shape(), b.hitArea.Shape()
	switch {
	case sa == HitAreaShapeCircle && sb == HitAreaShapeCircle:
		return hitCircleCircle(a, b)
	case sa == HitAreaShapeCircle && sb == HitAreaShapeRect:
		g := hitCircleRect(a, b) // n is rect→circle; flip to a→b
		if g.ok {
			g.nx, g.ny = -g.nx, -g.ny
		}
		return g
	case sa == HitAreaShapeRect && sb == HitAreaShapeCircle:
		return hitCircleRect(b, a) // n is a(rect)→b(circle)
	case sa == HitAreaShapeRect && sb == HitAreaShapeRect:
		return hitRectRect(a, b)
	default:
		return hitGeom{}
	}
}

func hitCircleCircle(a, b *SimEntity) hitGeom {
	ax, ay, ra := circleWorld(a)
	bx, by, rb := circleWorld(b)
	dx, dy := bx-ax, by-ay
	distSq := dx*dx + dy*dy
	rSum := ra + rb
	if distSq >= rSum*rSum {
		return hitGeom{}
	}
	dist := math.Sqrt(distSq)
	var nx, ny float64
	if dist > hitEps {
		nx, ny = dx/dist, dy/dist
	} else {
		nx, ny = degenerateNormal(a, b)
		dist = 0
	}
	pen := rSum - dist
	return hitGeom{
		ok:   true,
		nx:   nx,
		ny:   ny,
		atx:  ax + nx*ra,
		aty:  ay + ny*ra,
		pen:  pen,
	}
}

func hitCircleRect(circle, rect *SimEntity) hitGeom {
	cx, cy, r := circleWorld(circle)
	left, top, right, bottom := rectWorld(rect)

	inside := cx > left && cx < right && cy > top && cy < bottom
	var closestX, closestY float64
	if inside {
		// Min-penetration axis out of the rect.
		dl, dr := cx-left, right-cx
		dt, db := cy-top, bottom-cy
		minPen := dl
		nx, ny := -1.0, 0.0
		closestX, closestY = left, cy
		if dr < minPen {
			minPen, nx, ny = dr, 1, 0
			closestX, closestY = right, cy
		}
		if dt < minPen {
			minPen, nx, ny = dt, 0, -1
			closestX, closestY = cx, top
		}
		if db < minPen {
			minPen, nx, ny = db, 0, 1
			closestX, closestY = cx, bottom
		}
		// n from rect toward circle (outward from face through center)
		return hitGeom{
			ok:   true,
			nx:   nx,
			ny:   ny,
			atx:  closestX,
			aty:  closestY,
			pen:  r + minPen,
		}
	}

	closestX = clampFloat(cx, left, right)
	closestY = clampFloat(cy, top, bottom)
	dx, dy := cx-closestX, cy-closestY
	distSq := dx*dx + dy*dy
	if distSq >= r*r {
		return hitGeom{}
	}
	dist := math.Sqrt(distSq)
	var nx, ny float64
	if dist > hitEps {
		nx, ny = dx/dist, dy/dist
	} else {
		nx, ny = degenerateNormal(circle, rect)
	}
	return hitGeom{
		ok:   true,
		nx:   nx,
		ny:   ny,
		atx:  closestX,
		aty:  closestY,
		pen:  r - dist,
	}
}

// hitRectRect: AABB overlap; min-penetration ±X/±Y; n from a toward b.
func hitRectRect(a, b *SimEntity) hitGeom {
	al, at, ar, ab := rectWorld(a)
	bl, bt, br, bb := rectWorld(b)
	ox := math.Min(ar, br) - math.Max(al, bl)
	oy := math.Min(ab, bb) - math.Max(at, bt)
	if ox <= 0 || oy <= 0 {
		return hitGeom{}
	}
	acx := (al + ar) * 0.5
	acy := (at + ab) * 0.5
	bcx := (bl + br) * 0.5
	bcy := (bt + bb) * 0.5
	var nx, ny, pen float64
	var atx, aty float64
	if ox < oy {
		pen = ox
		if acx < bcx {
			nx = 1
			atx = ar // a's right face into b
		} else {
			nx = -1
			atx = al
		}
		aty = clampFloat((math.Max(at, bt)+math.Min(ab, bb))*0.5, math.Max(at, bt), math.Min(ab, bb))
		ny = 0
	} else {
		pen = oy
		if acy < bcy {
			ny = 1
			aty = ab
		} else {
			ny = -1
			aty = at
		}
		atx = clampFloat((math.Max(al, bl)+math.Min(ar, br))*0.5, math.Max(al, bl), math.Min(ar, br))
		nx = 0
	}
	return hitGeom{ok: true, nx: nx, ny: ny, atx: atx, aty: aty, pen: pen}
}

func degenerateNormal(a, b *SimEntity) (nx, ny float64) {
	rvx, rvy := a.vx-b.vx, a.vy-b.vy
	rs := math.Hypot(rvx, rvy)
	if rs > hitEps {
		// normalize(-relV) with relV = a-b → from a toward? -relV = b-a, from a toward b if separating...
		return -rvx / rs, -rvy / rs
	}
	dx := a._at.X - b._at.X
	dy := a._at.Y - b._at.Y
	ds := math.Hypot(dx, dy)
	if ds > hitEps {
		return dx / ds, dy / ds
	}
	return 1, 0
}

func clampFloat(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func applyBounce(a, b *SimEntity, nx, ny, pen float64) {
	applyBounceVelocity(a, b, nx, ny)
	applyBounceSeparate(a, b, nx, ny, pen)
}

// applyBounceVelocity exchanges/reflects only when closing along n (a→b).
// Re-applying while separating would flip them back into the contact.
func applyBounceVelocity(a, b *SimEntity, nx, ny float64) {
	aFree := !a.anchored && a.active
	bFree := !b.anchored && b.active
	switch {
	case aFree && bFree:
		rel := (a.vx-b.vx)*nx + (a.vy-b.vy)*ny
		if rel <= hitEps {
			return // separating or resting
		}
		van := a.vx*nx + a.vy*ny
		vbn := b.vx*nx + b.vy*ny
		a.vx += (vbn - van) * nx
		a.vy += (vbn - van) * ny
		b.vx += (van - vbn) * nx
		b.vy += (van - vbn) * ny
	case aFree && !bFree:
		wnx, wny := -nx, -ny // wall → mover
		dot := a.vx*wnx + a.vy*wny
		if dot >= -hitEps {
			return // not approaching the wall
		}
		a.vx -= 2 * dot * wnx
		a.vy -= 2 * dot * wny
	case !aFree && bFree:
		wnx, wny := nx, ny
		dot := b.vx*wnx + b.vy*wny
		if dot >= -hitEps {
			return
		}
		b.vx -= 2 * dot * wnx
		b.vy -= 2 * dot * wny
	}
}

func applyBounceSeparate(a, b *SimEntity, nx, ny, pen float64) {
	aFree := !a.anchored && a.active
	bFree := !b.anchored && b.active
	separatePair(a, b, nx, ny, pen, aFree, bFree)
}

func separatePair(a, b *SimEntity, nx, ny, pen float64, moveA, moveB bool) {
	if pen <= 0 {
		return
	}
	d := pen + hitSepSlop
	switch {
	case moveA && moveB:
		h := d * 0.5
		a.SetAt(core.Point{X: a._at.X - nx*h, Y: a._at.Y - ny*h})
		b.SetAt(core.Point{X: b._at.X + nx*h, Y: b._at.Y + ny*h})
	case moveA:
		a.SetAt(core.Point{X: a._at.X - nx*d, Y: a._at.Y - ny*d})
	case moveB:
		b.SetAt(core.Point{X: b._at.X + nx*d, Y: b._at.Y + ny*d})
	}
}

func applyStop(a, b *SimEntity) {
	if !a.anchored && a.active {
		a.stopped = true
	}
	if !b.anchored && b.active {
		b.stopped = true
	}
}

// orderHitSlots assigns entity1/entity2 by kind→subkind ascending.
func orderHitSlots(a, b *SimEntity) (e1, e2 *SimEntity) {
	if a.kind < b.kind || (a.kind == b.kind && a.subkind < b.subkind) {
		return a, b
	}
	if a.kind > b.kind || (a.kind == b.kind && a.subkind > b.subkind) {
		return b, a
	}
	return a, b
}
