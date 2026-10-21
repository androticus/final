package core

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/fogleman/gg"
	xdraw "golang.org/x/image/draw"
)

func asTileImage(v Any) image.Image {
	im, ok := v.(image.Image)
	if !ok || im == nil {
		panic(fmt.Errorf("Fill.tileImage: expected Image or Pixmap, got %T", v))
	}
	return im
}

func (f *Fill) apply(dc *gg.Context, bounds image.Rectangle) {
	switch f.style {
	case FillStyleSolid:
		dc.SetFillStyle(gg.NewSolidPattern(f.color))
	case FillStyleCheckerboard:
		dc.SetFillStyle(&checkerPattern{c0: f.color, c1: f.color2, size: f.checkerSize})
	case FillStyleGradient:
		x0, y0, x1, y1 := f.gradientEnds(bounds)
		g := gg.NewLinearGradient(x0, y0, x1, y1)
		g.AddColorStop(0, f.color)
		g.AddColorStop(1, f.color2)
		dc.SetFillStyle(g)
	case FillStyleTile:
		if f.tileImage == nil {
			panic(fmt.Errorf("Fill.tile: tileImage is nil"))
		}
		dc.SetFillStyle(gg.NewSurfacePattern(f.tilePatternImage(), gg.RepeatBoth))
	default:
		panic(fmt.Errorf("Fill.apply: unexpected style %q", f.style))
	}
}

func (f *Fill) gradientEnds(bounds image.Rectangle) (x0, y0, x1, y1 float64) {
	switch f.gradKind {
	case fillGradientFromTo:
		// From/To are world points; map with the Painter's pixmap viewOrigin (180.4).
		from := f.p.pixmap.MapWorldToView(f.gradFrom)
		to := f.p.pixmap.MapWorldToView(f.gradTo)
		return from.X, from.Y, to.X, to.Y
	case fillGradientDir:
		return aabbDirEnds(bounds, f.gradDX, f.gradDY)
	default:
		panic(fmt.Errorf("Fill.gradient: call setGradientFromTo, setGradientRect, setGradientDegrees, or setGradientDir first"))
	}
}

func (f *Fill) tilePatternImage() image.Image {
	im := unwrapTileImage(f.tileImage)
	z := f.tileZoom
	if z == 1 {
		return im
	}
	if f.tileScaled != nil {
		return f.tileScaled
	}
	b := im.Bounds()
	w := int(math.Round(float64(b.Dx()) * z))
	h := int(math.Round(float64(b.Dy()) * z))
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	dst := image.NewNRGBA(image.Rect(0, 0, w, h))
	xdraw.NearestNeighbor.Scale(dst, dst.Bounds(), im, b, xdraw.Src, nil)
	f.tileScaled = dst
	return dst
}

func unwrapTileImage(im image.Image) image.Image {
	switch t := im.(type) {
	case *Pixmap:
		return t.P
	case *Image:
		return t.Image
	default:
		return im
	}
}

type checkerPattern struct {
	c0, c1 Color
	size   int
}

func (p *checkerPattern) ColorAt(x, y int) color.Color {
	tx := floorDiv(x, p.size)
	ty := floorDiv(y, p.size)
	if (tx+ty)&1 == 0 {
		return p.c0
	}
	return p.c1
}

func floorDiv(a, b int) int {
	q := a / b
	r := a % b
	if r != 0 && a < 0 {
		q--
	}
	return q
}

// aabbDirEnds is center of bounds → hit the box along ±(dx,dy).
func aabbDirEnds(b image.Rectangle, dx, dy float64) (x0, y0, x1, y1 float64) {
	if b.Empty() || b.Dx() == 0 || b.Dy() == 0 {
		panic(fmt.Errorf("Fill.gradient: empty path bounds for direction spec"))
	}
	if dx == 0 && dy == 0 {
		panic(fmt.Errorf("Fill.setGradientDir: zero vector"))
	}
	minX, minY := float64(b.Min.X), float64(b.Min.Y)
	maxX, maxY := float64(b.Max.X), float64(b.Max.Y)
	cx := (minX + maxX) / 2
	cy := (minY + maxY) / 2
	tmin := math.Inf(-1)
	tmax := math.Inf(1)
	if dx != 0 {
		tx1 := (minX - cx) / dx
		tx2 := (maxX - cx) / dx
		if tx1 > tx2 {
			tx1, tx2 = tx2, tx1
		}
		tmin = tx1
		tmax = tx2
	}
	if dy != 0 {
		ty1 := (minY - cy) / dy
		ty2 := (maxY - cy) / dy
		if ty1 > ty2 {
			ty1, ty2 = ty2, ty1
		}
		if ty1 > tmin {
			tmin = ty1
		}
		if ty2 < tmax {
			tmax = ty2
		}
	}
	return cx + tmin*dx, cy + tmin*dy, cx + tmax*dx, cy + tmax*dy
}

func circleAABB(ctr Point, radius float64) image.Rectangle {
	return image.Rect(
		int(math.Floor(ctr.X-radius)),
		int(math.Floor(ctr.Y-radius)),
		int(math.Ceil(ctr.X+radius)),
		int(math.Ceil(ctr.Y+radius)),
	)
}

func ellipseAABB(ctr Point, rx, ry float64) image.Rectangle {
	return image.Rect(
		int(math.Floor(ctr.X-rx)),
		int(math.Floor(ctr.Y-ry)),
		int(math.Ceil(ctr.X+rx)),
		int(math.Ceil(ctr.Y+ry)),
	)
}

// rectAABB is the axis-aligned box for a center+Size rectangle.
func rectAABB(ctr Point, size Size) image.Rectangle {
	hw, hh := size.W/2, size.H/2
	return image.Rect(
		int(math.Floor(ctr.X-hw)),
		int(math.Floor(ctr.Y-hh)),
		int(math.Ceil(ctr.X+hw)),
		int(math.Ceil(ctr.Y+hh)),
	)
}

// rotatedRectAABB is the AABB of the four corners after clockwise rotation about ctr.
func rotatedRectAABB(ctr Point, size Size, rotationDegrees float64) image.Rectangle {
	rad := gg.Radians(rotationDegrees)
	c, s := math.Cos(rad), math.Sin(rad)
	hw, hh := size.W/2, size.H/2
	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)
	for _, dx := range []float64{-hw, hw} {
		for _, dy := range []float64{-hh, hh} {
			// gg clockwise: x' = c*x - s*y, y' = s*x + c*y
			x := ctr.X + c*dx - s*dy
			y := ctr.Y + s*dx + c*dy
			if x < minX {
				minX = x
			}
			if x > maxX {
				maxX = x
			}
			if y < minY {
				minY = y
			}
			if y > maxY {
				maxY = y
			}
		}
	}
	return image.Rect(
		int(math.Floor(minX)),
		int(math.Floor(minY)),
		int(math.Ceil(maxX)),
		int(math.Ceil(maxY)),
	)
}

// rotatedEllipseAABB is the exact AABB of an ellipse after clockwise rotation about ctr.
func rotatedEllipseAABB(ctr Point, rx, ry, rotationDegrees float64) image.Rectangle {
	rad := gg.Radians(rotationDegrees)
	c, s := math.Cos(rad), math.Sin(rad)
	// Same extents for ±θ (cos even, sin² even).
	hx := math.Sqrt(rx*rx*c*c + ry*ry*s*s)
	hy := math.Sqrt(rx*rx*s*s + ry*ry*c*c)
	return image.Rect(
		int(math.Floor(ctr.X-hx)),
		int(math.Floor(ctr.Y-hy)),
		int(math.Ceil(ctr.X+hx)),
		int(math.Ceil(ctr.Y+hy)),
	)
}

var _ gg.Pattern = (*checkerPattern)(nil)
