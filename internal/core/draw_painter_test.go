package core

import (
	"image"
	"image/color"
	"image/draw"
	"testing"
)

func painterRGBA(p *Painter) *image.RGBA {
	return p.pixmap.P
}

func pixelPainted(img *image.RGBA, x, y int) bool {
	return img.RGBAAt(x, y).A != 0
}

func countPainted(img *image.RGBA) int {
	n := 0
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if img.RGBAAt(x, y).A != 0 {
				n++
			}
		}
	}
	return n
}

func TestNewPainterMintsRGBAPixmap(t *testing.T) {
	p := NewPainter(NewSize(40, 30))
	img := painterRGBA(p)
	if img.Bounds() != image.Rect(0, 0, 40, 30) {
		t.Fatalf("pixmap bounds %v, want 40x30", img.Bounds())
	}
	if p.Pixmap() != p.pixmap {
		t.Fatal("Pixmap() mismatch")
	}
	if p.FinalType() != painterType {
		t.Fatal("Painter.FinalType mismatch")
	}
}

func TestNewPainterRejectsNonPositiveSize(t *testing.T) {
	for _, sz := range []Size{{0, 10}, {10, 0}, {-1, 8}, {8, -1}} {
		func(sz Size) {
			defer func() {
				if recover() == nil {
					t.Fatalf("NewPainter(%v) should panic", sz)
				}
			}()
			_ = NewPainter(sz)
		}(sz)
	}
}

func TestNewPainterRejectsFractionalSize(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("NewPainter fractional size should panic")
		}
	}()
	_ = NewPainter(NewSizeF(10.5, 20))
}

func TestPainterLineCircleRectStrokePixels(t *testing.T) {
	p := NewPainter(NewSize(64, 64))
	p.Stroke().SetColor(NewColor(255, 0, 0, 255))
	p.Stroke().SetLineWidth(3)

	p.Line(NewPoint(4, 20), NewPoint(60, 20))
	if !pixelPainted(painterRGBA(p), 32, 20) {
		t.Fatal("line: expected paint at (32,20)")
	}

	p.Circle(NewPoint(32, 32), 16)
	if countPainted(painterRGBA(p)) < 2 {
		t.Fatal("circle: expected additional painted pixels")
	}

	p.Rect(NewPoint(20, 48), NewSize(24, 16))
	if !pixelPainted(painterRGBA(p), 20, 40) {
		t.Fatal("rect: expected paint on top edge (20,40)")
	}

	c := painterRGBA(p).RGBAAt(32, 20)
	if c.R == 0 || c.G > c.R/2 || c.B > c.R/2 {
		t.Fatalf("line pixel = %v, want red-dominant", c)
	}
}

func TestPainterRoundedRectStrokePixels(t *testing.T) {
	p := NewPainter(NewSize(64, 64))
	p.Stroke().SetColor(NewColor(0, 255, 0, 255))
	p.Stroke().SetLineWidth(3)
	p.RoundedRect(NewPoint(32, 24), NewSize(48, 32), 8)
	if countPainted(painterRGBA(p)) == 0 {
		t.Fatal("roundedRect: no paint")
	}
	if !pixelPainted(painterRGBA(p), 32, 8) {
		t.Fatal("roundedRect: expected paint on top edge (32,8)")
	}
}

func TestPainterRoundedRectRejectsNegativeRadius(t *testing.T) {
	p := NewPainter(NewSize(16, 16))
	defer func() {
		if recover() == nil {
			t.Fatal("roundedRect radius=-1 should panic")
		}
	}()
	p.RoundedRect(NewPoint(4, 4), NewSize(8, 8), -1)
}

func TestPainterEllipseStrokePixels(t *testing.T) {
	p := NewPainter(NewSize(64, 64))
	p.Stroke().SetColor(NewColor(0, 0, 255, 255))
	p.Stroke().SetLineWidth(3)
	p.Ellipse(NewPoint(32, 32), 24, 12)
	if countPainted(painterRGBA(p)) == 0 {
		t.Fatal("ellipse: no paint")
	}
	if !pixelPainted(painterRGBA(p), 56, 32) && !pixelPainted(painterRGBA(p), 55, 32) {
		t.Fatal("ellipse: expected paint near right vertex")
	}
}

func TestPainterEllipseRadiiCoercion(t *testing.T) {
	typ := GetTypeRegistrar().Get("Painter")
	mm := typ.MethodByName("ellipse")
	p := NewPainter(NewSize(48, 48))
	p.Stroke().SetLineWidth(2)

	invoke := func(rx, ry any) {
		rp := NewTestRunProvider()
		f, ths := memberRunFrame(rp, p)
		rp.DataStack.Push(NewPoint(24, 24))
		rp.DataStack.Push(rx)
		rp.DataStack.Push(ry)
		mm.invokeCall(f, ths)
	}
	invoke(Int(16), Int(8))
	invoke(Float(14.5), Uint(10))
	if countPainted(painterRGBA(p)) == 0 {
		t.Fatal("ellipse: Int/Float/Uint radii produced no paint")
	}
}

func TestPainterEllipseRejectsNegativeRadii(t *testing.T) {
	p := NewPainter(NewSize(16, 16))
	defer func() {
		if recover() == nil {
			t.Fatal("ellipse rx=-1 should panic")
		}
	}()
	p.Ellipse(NewPoint(8, 8), -1, 4)
}

func TestPainterCircleRadiusCoercion(t *testing.T) {
	typ := GetTypeRegistrar().Get("Painter")
	mm := typ.MethodByName("circle")
	p := NewPainter(NewSize(48, 48))
	p.Stroke().SetLineWidth(2)

	invoke := func(r any) {
		rp := NewTestRunProvider()
		f, ths := memberRunFrame(rp, p)
		rp.DataStack.Push(NewPoint(24, 24))
		rp.DataStack.Push(r)
		mm.invokeCall(f, ths)
	}
	invoke(Int(12))
	invoke(Float(10.5))
	invoke(Uint(8))
	if countPainted(painterRGBA(p)) == 0 {
		t.Fatal("circle: Int/Float/Uint radius produced no paint")
	}
}

func TestPainterCircleRejectsNegativeRadius(t *testing.T) {
	p := NewPainter(NewSize(16, 16))
	defer func() {
		if recover() == nil {
			t.Fatal("circle radius=-1 should panic")
		}
	}()
	p.Circle(NewPoint(8, 8), -1)
}

func TestPainterRoundedRectRadiusCoercion(t *testing.T) {
	typ := GetTypeRegistrar().Get("Painter")
	mm := typ.MethodByName("roundedRect")
	p := NewPainter(NewSize(48, 48))
	p.Stroke().SetLineWidth(2)

	invoke := func(r any) {
		rp := NewTestRunProvider()
		f, ths := memberRunFrame(rp, p)
		rp.DataStack.Push(NewPoint(4, 4))
		rp.DataStack.Push(NewSize(40, 32))
		rp.DataStack.Push(r)
		mm.invokeCall(f, ths)
	}
	invoke(Int(6))
	invoke(Float(5.5))
	invoke(Uint(4))
	if countPainted(painterRGBA(p)) == 0 {
		t.Fatal("roundedRect: Int/Float/Uint radius produced no paint")
	}
}

func TestPainterRoundedRectFillDistinctFromStroke(t *testing.T) {
	p := NewPainter(NewSize(64, 64))
	p.Fill().SetStyle(FillStyleSolid)
	p.Fill().SetColor(NewColor(255, 0, 0, 255))
	p.Stroke().SetColor(NewColor(0, 0, 255, 255))
	p.Stroke().SetLineWidth(4)
	p.RoundedRect(NewPoint(32, 32), NewSize(48, 48), 6)
	c := painterRGBA(p).RGBAAt(32, 32)
	if c.R < 200 || c.B > 80 {
		t.Fatalf("fill center = %v, want red-dominant", c)
	}
}

func TestPainterEllipseFillDistinctFromStroke(t *testing.T) {
	p := NewPainter(NewSize(64, 64))
	p.Fill().SetStyle(FillStyleSolid)
	p.Fill().SetColor(NewColor(255, 0, 0, 255))
	p.Stroke().SetColor(NewColor(0, 0, 255, 255))
	p.Stroke().SetLineWidth(4)
	p.Ellipse(NewPoint(32, 32), 20, 14)
	c := painterRGBA(p).RGBAAt(32, 32)
	if c.R < 200 || c.B > 80 {
		t.Fatalf("fill center = %v, want red-dominant", c)
	}
}

func TestPainterPixmapAcceptsBlit(t *testing.T) {
	p := NewPainter(NewSize(8, 8))
	green := NewColor(0, 255, 0, 255)
	draw.Draw(p.Pixmap().P, image.Rect(0, 0, 4, 4), &image.Uniform{green}, image.Point{}, draw.Src)
	got := painterRGBA(p).RGBAAt(1, 1)
	want := color.RGBAModel.Convert(green).(color.RGBA)
	if got != want {
		t.Fatalf("blit pixel = %v, want %v", got, want)
	}
}

func TestPainterClearWipesInPlace(t *testing.T) {
	p := NewPainter(NewSize(32, 32))
	pix := p.Pixmap()
	buf := painterRGBA(p)
	pen := NewColor(255, 0, 0, 255)
	p.Stroke().SetColor(pen)
	p.Line(NewPoint(2, 16), NewPoint(30, 16))
	if countPainted(buf) == 0 {
		t.Fatal("line: expected paint before clear")
	}
	p.Clear()
	if p.Pixmap() != pix {
		t.Fatal("clear must not replace Pixmap")
	}
	if painterRGBA(p) != buf {
		t.Fatal("clear must not replace P")
	}
	if countPainted(buf) != 0 {
		t.Fatal("clear: expected no paint")
	}
	if p.Stroke().Color() != pen {
		t.Fatalf("clear must not change stroke.color, got %v", p.Stroke().Color())
	}
	p.Line(NewPoint(2, 16), NewPoint(30, 16))
	if countPainted(buf) == 0 {
		t.Fatal("draw after clear: expected paint")
	}
}

func TestNewPainterWithPixmapUsesExistingPixmap(t *testing.T) {
	pm := NewPixmap(30, 40)
	ink := NewColor(255, 0, 0, 255)
	pm.Put(10, 20, ink)
	p := NewPainterWithPixmap(pm)
	if p.Pixmap() != pm {
		t.Fatal("newWithPixmap should bind the given Pixmap")
	}
	if p.Pixmap().P != pm.P {
		t.Fatal("newWithPixmap should use the same RGBA buffer")
	}
	p.Line(NewPoint(0, 0), NewPoint(5, 5))
	if !pixelPainted(painterRGBA(p), 2, 2) {
		t.Fatal("draw should land on the shared pixmap")
	}
	got := pm.P.RGBAAt(20, 10)
	want := color.RGBAModel.Convert(ink).(color.RGBA)
	if got != want {
		t.Fatalf("pre-seeded cell = %v, want %v", got, want)
	}
}

func TestNewPainterWithPixmapNilPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("NewPainterWithPixmap(nil) should panic")
		}
	}()
	_ = NewPainterWithPixmap(nil)
}

func TestPainterSetPixmapRebinds(t *testing.T) {
	p := NewPainter(NewSize(16, 16))
	pm := NewPixmap(8, 8)
	p.SetPixmap(pm)
	if p.Pixmap() != pm {
		t.Fatal("setPixmap should rebind")
	}
	p.Stroke().SetColor(NewColor(0, 255, 0, 255))
	p.Line(NewPoint(0, 0), NewPoint(7, 0))
	if !pixelPainted(pm.P, 4, 0) {
		t.Fatal("draw after setPixmap should hit the new buffer")
	}
	if countPainted(painterRGBA(p)) != countPainted(pm.P) {
		t.Fatal("Painter buffer should match bound Pixmap")
	}
}

func TestPainterSetPixmapNilPanics(t *testing.T) {
	p := NewPainter(NewSize(4, 4))
	defer func() {
		if recover() == nil {
			t.Fatal("setPixmap(nil) should panic")
		}
	}()
	p.SetPixmap(nil)
}

func TestPainterNewWithPixmapClassMethod(t *testing.T) {
	typ := GetTypeRegistrar().Get("Painter")
	pm := NewPixmap(6, 4)
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, typ)
	rp.DataStack.Push(pm)
	typ.ClassMethodByName("newWithPixmap").invokeCall(f, ths)
	p := rp.DataStack.Pop().(*Painter)
	if p.Pixmap() != pm {
		t.Fatalf("Painter.newWithPixmap = %v, want %p", p.Pixmap(), pm)
	}
}

func TestPainterPixmapPropertySet(t *testing.T) {
	typ := GetTypeRegistrar().Get("Painter")
	p := NewPainter(NewSize(4, 4))
	pm := NewPixmap(2, 2)
	prop := typ.GetPropertyMetadata("pixmap")
	if prop == nil || prop.set == nil {
		t.Fatal("pixmap should be a writable property")
	}
	prop.set(p, pm)
	if p.Pixmap() != pm {
		t.Fatal("pixmap property set should rebind")
	}
}

func TestPainterPixmapPropertySetNilPanics(t *testing.T) {
	p := NewPainter(NewSize(4, 4))
	prop := painterType.GetPropertyMetadata("pixmap")
	defer func() {
		if recover() == nil {
			t.Fatal("pixmap property set nil should panic")
		}
	}()
	prop.set(p, nil)
}

func TestPainterNewClassMethod(t *testing.T) {
	typ := GetTypeRegistrar().Get("Painter")
	if typ == nil {
		t.Fatal("Painter type not registered")
	}
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, typ)
	rp.DataStack.Push(Int(12))
	rp.DataStack.Push(Int(8))
	typ.ClassMethodByName("new").invokeCall(f, ths)
	p := rp.DataStack.Pop().(*Painter)
	img := painterRGBA(p)
	if img.Bounds() != image.Rect(0, 0, 12, 8) {
		t.Fatalf("Painter.new 12 8 bounds %v", img.Bounds())
	}

	pm := typ.GetPropertyMetadata("pixmap")
	if pm == nil || pm.set == nil {
		t.Fatal("pixmap should be writable")
	}
}

func TestPainterRegularPolygonTriangle(t *testing.T) {
	p := NewPainter(NewSize(64, 64))
	p.Stroke().SetColor(NewColor(255, 0, 0, 255))
	p.Stroke().SetLineWidth(3)
	p.RegularPolygon(3, NewPoint(32, 32), 20)
	if countPainted(painterRGBA(p)) == 0 {
		t.Fatal("regularPolygon sides=3: no paint")
	}
	if !pixelPainted(painterRGBA(p), 32, 12) && !pixelPainted(painterRGBA(p), 32, 13) {
		t.Fatal("regularPolygon sides=3: expected top vertex")
	}
}

func TestPainterRegularPolygonStrokePixels(t *testing.T) {
	p := NewPainter(NewSize(64, 64))
	p.Stroke().SetColor(NewColor(0, 0, 255, 255))
	p.Stroke().SetLineWidth(3)
	before := countPainted(painterRGBA(p))
	p.RegularPolygon(5, NewPoint(32, 32), 20)
	if countPainted(painterRGBA(p)) <= before {
		t.Fatal("regularPolygon: expected painted pixels")
	}
	// rotation 0, odd n: a vertex up (negative Y)
	if !pixelPainted(painterRGBA(p), 32, 12) && !pixelPainted(painterRGBA(p), 32, 13) {
		t.Fatal("regularPolygon: expected paint near top vertex (32,12)")
	}
}

func TestPainterRegularPolygonRotationCoercion(t *testing.T) {
	typ := GetTypeRegistrar().Get("Painter")
	mm := typ.MethodByName("regularPolygonRot")
	p := NewPainter(NewSize(48, 48))
	p.Stroke().SetLineWidth(2)

	invoke := func(radius, rot any) {
		rp := NewTestRunProvider()
		f, ths := memberRunFrame(rp, p)
		rp.DataStack.Push(Int(6))
		rp.DataStack.Push(NewPoint(24, 24))
		rp.DataStack.Push(radius)
		rp.DataStack.Push(rot)
		mm.invokeCall(f, ths)
	}
	invoke(Int(12), Int(0))
	invoke(Float(11.5), Float(15))
	invoke(Uint(10), Uint(30))
	if countPainted(painterRGBA(p)) == 0 {
		t.Fatal("regularPolygonRot: Int/Float/Uint radius/rotation produced no paint")
	}
}

func TestPainterRegularPolygonSidesIntOnly(t *testing.T) {
	typ := GetTypeRegistrar().Get("Painter")
	mm := typ.MethodByName("regularPolygon")
	p := NewPainter(NewSize(48, 48))
	defer func() {
		if recover() == nil {
			t.Fatal("regularPolygon sides Float should panic")
		}
	}()
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, p)
	rp.DataStack.Push(Float(6))
	rp.DataStack.Push(NewPoint(24, 24))
	rp.DataStack.Push(Int(12))
	mm.invokeCall(f, ths)
}

func TestPainterRegularPolygonRejectsFewSides(t *testing.T) {
	p := NewPainter(NewSize(16, 16))
	defer func() {
		if recover() == nil {
			t.Fatal("regularPolygon sides=2 should panic")
		}
	}()
	p.RegularPolygon(2, NewPoint(8, 8), 4)
}

func TestPainterRectRotZeroMatchesBase(t *testing.T) {
	base := NewPainter(NewSize(64, 64))
	rot0 := NewPainter(NewSize(64, 64))
	for _, p := range []*Painter{base, rot0} {
		p.Stroke().SetColor(NewColor(255, 0, 0, 255))
		p.Stroke().SetLineWidth(2)
	}
	ctr, sz := NewPoint(32, 32), NewSize(24, 16)
	base.Rect(ctr, sz)
	rot0.RectRot(ctr, sz, 0)
	if countPainted(painterRGBA(base)) == 0 || countPainted(painterRGBA(base)) != countPainted(painterRGBA(rot0)) {
		t.Fatalf("RectRot(0) paint count %d vs base %d", countPainted(painterRGBA(rot0)), countPainted(painterRGBA(base)))
	}
}

func TestPainterRectRotNonZeroPaints(t *testing.T) {
	p := NewPainter(NewSize(64, 64))
	p.Stroke().SetColor(NewColor(0, 0, 255, 255))
	p.Stroke().SetLineWidth(3)
	p.RectRot(NewPoint(32, 32), NewSize(40, 10), 45)
	if countPainted(painterRGBA(p)) == 0 {
		t.Fatal("RectRot(45): expected paint")
	}
	aabb := rotatedRectAABB(NewPoint(32, 32), NewSize(40, 10), 45)
	if aabb.Empty() || aabb.Dx() < 20 || aabb.Dy() < 20 {
		t.Fatalf("rotatedRectAABB too small: %v", aabb)
	}
}

func TestPainterEllipseRotNonZeroPaints(t *testing.T) {
	p := NewPainter(NewSize(64, 64))
	p.Stroke().SetColor(NewColor(0, 255, 0, 255))
	p.Stroke().SetLineWidth(2)
	p.EllipseRot(NewPoint(32, 32), 24, 8, 30)
	if countPainted(painterRGBA(p)) == 0 {
		t.Fatal("EllipseRot: expected paint")
	}
}

func TestPainterRoundedRectRotNonZeroPaints(t *testing.T) {
	p := NewPainter(NewSize(64, 64))
	p.Stroke().SetColor(NewColor(255, 128, 0, 255))
	p.Stroke().SetLineWidth(2)
	p.RoundedRectRot(NewPoint(32, 32), NewSize(36, 20), 4, 25)
	if countPainted(painterRGBA(p)) == 0 {
		t.Fatal("RoundedRectRot: expected paint")
	}
}

func TestPainterRegularPolygonRotPaints(t *testing.T) {
	p := NewPainter(NewSize(64, 64))
	p.Stroke().SetColor(NewColor(128, 0, 255, 255))
	p.Stroke().SetLineWidth(2)
	p.RegularPolygonRot(5, NewPoint(32, 32), 18, 36)
	if countPainted(painterRGBA(p)) == 0 {
		t.Fatal("RegularPolygonRot: expected paint")
	}
}

func TestPainterStateDefaults(t *testing.T) {
	p := NewPainter(NewSize(16, 16))
	if p.Stroke() == nil || p.Fill() == nil || p.Font() == nil {
		t.Fatal("ps adjuncts nil")
	}
	if p.Stroke().Color() != WebColors[ColorBlack] {
		t.Fatal("stroke.color default")
	}
	if p.Stroke().LineWidth() != defaultPainterLineWidth {
		t.Fatal("stroke.lineWidth default")
	}
	if p.Stroke().LineCap() != LineCapRound {
		t.Fatal("stroke.lineCap default")
	}
	if p.Stroke().LineJoin() != LineJoinRound {
		t.Fatal("stroke.lineJoin default")
	}
	if p.Stroke().LineStyle() != LineStyleSolid {
		t.Fatal("stroke.lineStyle default")
	}
	if p.Fill().Style() != FillStyleNone {
		t.Fatal("fill.style default none")
	}
	if p.Font().Color() != WebColors[ColorBlack] {
		t.Fatal("font.color default")
	}
	typ := GetTypeRegistrar().Get("Painter")
	for _, name := range []string{"stroke", "fill", "font"} {
		pm := typ.GetPropertyMetadata(name)
		if pm == nil || pm.set != nil {
			t.Fatalf("Painter.%s should be readonly const", name)
		}
	}
}

func TestPainterFillSolidDistinctFromStroke(t *testing.T) {
	p := NewPainter(NewSize(64, 64))
	p.Fill().SetStyle(FillStyleSolid)
	p.Fill().SetColor(NewColor(255, 0, 0, 255))
	p.Stroke().SetColor(NewColor(0, 0, 255, 255))
	p.Stroke().SetLineWidth(4)
	p.Circle(NewPoint(32, 32), 20)
	c := painterRGBA(p).RGBAAt(32, 32)
	if c.R < 200 || c.B > 80 {
		t.Fatalf("fill center = %v, want red-dominant", c)
	}
}

func TestPainterLineStyleNoneSkipsStroke(t *testing.T) {
	p := NewPainter(NewSize(32, 32))
	p.Stroke().SetColor(NewColor(255, 0, 0, 255))
	p.Stroke().SetLineStyle(LineStyleNone)
	p.Line(NewPoint(2, 16), NewPoint(30, 16))
	if countPainted(painterRGBA(p)) != 0 {
		t.Fatal("line with LineStyle.none should not paint")
	}
}

func countOpaqueOnRow(img *image.RGBA, y, x0, x1 int) (painted, clear int) {
	for x := x0; x < x1; x++ {
		if img.RGBAAt(x, y).A > 128 {
			painted++
		} else {
			clear++
		}
	}
	return painted, clear
}

func TestPainterLineStyleDashPatternsHaveGaps(t *testing.T) {
	ink := NewColor(255, 0, 0, 255)
	for _, st := range []LineStyle{LineStyleDot, LineStyleDash, LineStyleDotDash} {
		p := NewPainter(NewSize(80, 32))
		p.Stroke().SetColor(ink)
		p.Stroke().SetLineWidth(2)
		p.Stroke().SetLineStyle(st)
		p.Line(NewPoint(4, 16), NewPoint(76, 16))
		painted, clear := countOpaqueOnRow(painterRGBA(p), 16, 8, 72)
		if painted == 0 || clear == 0 {
			t.Fatalf("%s: painted=%d clear=%d, want gaps", st, painted, clear)
		}
	}
}

func TestPainterLineStyleSolidHasNoGaps(t *testing.T) {
	p := NewPainter(NewSize(80, 32))
	p.Stroke().SetColor(NewColor(255, 0, 0, 255))
	p.Stroke().SetLineWidth(2)
	p.Stroke().SetLineStyle(LineStyleSolid)
	p.Line(NewPoint(4, 16), NewPoint(76, 16))
	painted, clear := countOpaqueOnRow(painterRGBA(p), 16, 8, 72)
	if painted == 0 || clear != 0 {
		t.Fatalf("solid: painted=%d clear=%d", painted, clear)
	}
}

func TestPainterLineStyleUnknownPanics(t *testing.T) {
	p := NewPainter(NewSize(8, 8))
	defer func() {
		if recover() == nil {
			t.Fatal("unknown LineStyle should panic")
		}
	}()
	p.Stroke().SetLineStyle(LineStyle("wavy"))
}

func TestStrokeLineWidthFlexNumeric(t *testing.T) {
	typ := GetTypeRegistrar().Get("Stroke")
	pm := typ.GetPropertyMetadata("lineWidth")
	p := NewPainter(NewSize(8, 8))
	s := p.Stroke()
	pm.set(s, Int(4))
	if s.LineWidth() != 4 {
		t.Fatalf("lineWidth from Int %v", s.LineWidth())
	}
	pm.set(s, Float(2.5))
	if s.LineWidth() != 2.5 {
		t.Fatalf("lineWidth from Float %v", s.LineWidth())
	}
}

func TestClonePainterStateValueCopies(t *testing.T) {
	p := NewPainter(NewSize(8, 8))
	oldStroke, oldFill, oldFont := p.Stroke(), p.Fill(), p.Font()
	oldFace := oldFont.Face()
	p.Stroke().SetColor(NewColor(1, 2, 3, 255))
	p.Stroke().SetLineWidth(3)
	p.Fill().SetStyle(FillStyleSolid)
	p.Fill().SetColor(NewColor(4, 5, 6, 255))

	cl := ClonePainterState(p.ps)
	if cl.stroke == oldStroke || cl.fill == oldFill || cl.font == oldFont {
		t.Fatal("clone must mint new Stroke/Fill/Font")
	}
	if cl.stroke.p != p || cl.fill.p != p || cl.font.p != p {
		t.Fatal("clone adjuncts must keep the same Painter")
	}
	if cl.stroke.Color() != oldStroke.Color() || cl.stroke.LineWidth() != oldStroke.LineWidth() {
		t.Fatal("cloned Stroke values")
	}
	if cl.fill.Style() != oldFill.Style() || cl.fill.Color() != oldFill.Color() {
		t.Fatal("cloned Fill values")
	}
	if cl.font.Family() != oldFont.Family() || cl.font.Size() != oldFont.Size() || cl.font.Face() != oldFace {
		t.Fatal("cloned Font values share face until rebuild")
	}
}

func TestPainterPushPopRestoresAdjunctsAndDraw(t *testing.T) {
	p := NewPainter(NewSize(64, 64))
	black := NewColor(0, 0, 0, 255)
	red := NewColor(255, 0, 0, 255)
	p.Stroke().SetColor(black)
	p.Stroke().SetLineWidth(1)

	oldStroke, oldFill, oldFont := p.Stroke(), p.Fill(), p.Font()
	oldFace := oldFont.Face()
	if p.stack != nil {
		t.Fatal("stack should be nil until first push")
	}

	p.Line(NewPoint(4, 16), NewPoint(60, 16))
	pre := painterRGBA(p).RGBAAt(32, 16)

	p.Push()
	if p.stack == nil || p.stack.Len() != 1 {
		t.Fatal("first push should create stack and push old ps")
	}
	if p.Stroke() == oldStroke || p.Fill() == oldFill || p.Font() == oldFont {
		t.Fatal("push should switch to cloned adjuncts")
	}
	if p.Font().Face() != oldFace {
		t.Fatal("clone must share face until rebuild")
	}

	p.Stroke().SetColor(red)
	p.Stroke().SetLineWidth(8)
	p.Fill().SetStyle(FillStyleSolid)
	p.Fill().SetColor(red)
	p.Font().SetSize(28)
	if p.Font().Face() == oldFace {
		t.Fatal("SetSize after push should rebuild current face")
	}
	if oldFont.Face() != oldFace || oldFont.Size() != defaultFontSize {
		t.Fatal("stacked Font must keep old face and size")
	}
	if oldStroke.Color() != black || oldStroke.LineWidth() != 1 {
		t.Fatal("stacked Stroke must keep old values")
	}

	p.Line(NewPoint(4, 8), NewPoint(60, 8))
	mid := painterRGBA(p).RGBAAt(32, 8)
	if mid.R < 200 || mid.G > 80 || mid.B > 80 {
		t.Fatalf("pushed stroke line = %v, want red-dominant", mid)
	}

	p.Pop()
	if p.Stroke() != oldStroke || p.Fill() != oldFill || p.Font() != oldFont {
		t.Fatal("pop should restore previous adjuncts")
	}
	if p.Font().Face() != oldFace || p.Font().Size() != defaultFontSize {
		t.Fatal("pop should restore Font face and size")
	}
	if p.stack.Len() != 0 {
		t.Fatal("pop should empty the stack")
	}

	p.Line(NewPoint(4, 40), NewPoint(60, 40))
	post := painterRGBA(p).RGBAAt(32, 40)
	if post != pre {
		t.Fatalf("draw after pop = %v, want pre-push %v", post, pre)
	}
}

func TestPainterPushPopNested(t *testing.T) {
	p := NewPainter(NewSize(8, 8))
	s0 := p.Stroke()
	p.Push()
	s1 := p.Stroke()
	p.Push()
	s2 := p.Stroke()
	if s0 == s1 || s1 == s2 {
		t.Fatal("each push should mint a new Stroke")
	}
	p.Pop()
	if p.Stroke() != s1 {
		t.Fatal("inner pop")
	}
	p.Pop()
	if p.Stroke() != s0 {
		t.Fatal("outer pop")
	}
}

func TestPainterPopWithoutPushPanics(t *testing.T) {
	p := NewPainter(NewSize(8, 8))
	defer func() {
		if recover() == nil {
			t.Fatal("pop with no stack should panic")
		}
	}()
	p.Pop()
}

func TestPainterPopEmptyStackPanics(t *testing.T) {
	p := NewPainter(NewSize(8, 8))
	p.Push()
	p.Pop()
	defer func() {
		if recover() == nil {
			t.Fatal("pop with empty stack should panic")
		}
	}()
	p.Pop()
}

func TestPainterPushPopNativeMethods(t *testing.T) {
	typ := GetTypeRegistrar().Get("Painter")
	p := NewPainter(NewSize(8, 8))
	old := p.Stroke()
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, p)
	typ.MethodByName("push").invokeCall(f, ths)
	if p.Stroke() == old {
		t.Fatal("Painter.push should switch Stroke")
	}
	typ.MethodByName("pop").invokeCall(f, ths)
	if p.Stroke() != old {
		t.Fatal("Painter.pop should restore Stroke")
	}
}

func TestPainterPushPopRestoresFontDraw(t *testing.T) {
	ink := NewColor(0, 0, 0, 255)
	ref := NewPainter(NewSize(128, 64))
	ref.Font().SetColor(ink)
	ref.Text("Ag", NewPoint(8, 40))
	want := countPainted(painterRGBA(ref))

	p := NewPainter(NewSize(128, 64))
	p.Font().SetColor(ink)
	p.Push()
	p.Font().SetSize(28)
	p.Pop()
	p.Text("Ag", NewPoint(8, 40))
	got := countPainted(painterRGBA(p))
	if got != want {
		t.Fatalf("text after pop painted %d, want size-12 %d", got, want)
	}
}

func TestStrokeLineJoinSet(t *testing.T) {
	p := NewPainter(NewSize(16, 16))
	if p.Stroke().LineJoin() != LineJoinRound {
		t.Fatal("default round")
	}
	p.Stroke().SetLineJoin(LineJoinBevel)
	if p.Stroke().LineJoin() != LineJoinBevel {
		t.Fatal("lineJoin bevel")
	}
}

func TestLineCapLineStyleFillStyleConsts(t *testing.T) {
	lc := GetTypeRegistrar().Get("LineCap")
	if lc.GetClassPropertyMetadata("round").get(nil).(LineCap) != LineCapRound {
		t.Fatal("LineCap.round")
	}
	lj := GetTypeRegistrar().Get("LineJoin")
	if lj.GetClassPropertyMetadata("round").get(nil).(LineJoin) != LineJoinRound {
		t.Fatal("LineJoin.round")
	}
	if lj.GetClassPropertyMetadata("bevel").get(nil).(LineJoin) != LineJoinBevel {
		t.Fatal("LineJoin.bevel")
	}
	ls := GetTypeRegistrar().Get("LineStyle")
	if ls.GetClassPropertyMetadata("none").get(nil).(LineStyle) != LineStyleNone {
		t.Fatal("LineStyle.none")
	}
	if ls.GetClassPropertyMetadata("dot").get(nil).(LineStyle) != LineStyleDot {
		t.Fatal("LineStyle.dot")
	}
	if ls.GetClassPropertyMetadata("dash").get(nil).(LineStyle) != LineStyleDash {
		t.Fatal("LineStyle.dash")
	}
	if ls.GetClassPropertyMetadata("dotDash").get(nil).(LineStyle) != LineStyleDotDash {
		t.Fatal("LineStyle.dotDash")
	}
	fs := GetTypeRegistrar().Get("FillStyle")
	if fs.GetClassPropertyMetadata("solid").get(nil).(FillStyle) != FillStyleSolid {
		t.Fatal("FillStyle.solid")
	}
}

func TestPainterPixelWritesRGBA(t *testing.T) {
	p := NewPainter(NewSize(8, 8))
	ink := NewColor(255, 0, 0, 255)
	p.Stroke().SetColor(ink)
	p.Pixel(NewPoint(3, 4))
	got := painterRGBA(p).RGBAAt(3, 4)
	want := color.RGBAModel.Convert(ink).(color.RGBA)
	if got != want {
		t.Fatalf("pixel = %v, want %v", got, want)
	}
	if countPainted(painterRGBA(p)) != 1 {
		t.Fatal("pixel should write exactly one cell")
	}
	if p.Stroke().Color() != ink {
		t.Fatal("pixel must not change stroke.color")
	}
}

func TestPainterPixelLineStyleNoneNoOp(t *testing.T) {
	p := NewPainter(NewSize(8, 8))
	p.Stroke().SetColor(NewColor(255, 0, 0, 255))
	p.Stroke().SetLineStyle(LineStyleNone)
	p.Pixel(NewPoint(2, 2))
	if countPainted(painterRGBA(p)) != 0 {
		t.Fatal("pixel with LineStyle.none should not paint")
	}
}

func TestPainterPixelOutOfRangeNoOp(t *testing.T) {
	p := NewPainter(NewSize(8, 8))
	p.Stroke().SetColor(NewColor(255, 0, 0, 255))
	for _, at := range []Point{{-1, 0}, {0, -1}, {8, 0}, {0, 8}} {
		p.Pixel(at) // clipped — no panic
	}
	if countPainted(painterRGBA(p)) != 0 {
		t.Fatal("out-of-range pixel should not paint")
	}
}

func TestPainterPixelNativeMethod(t *testing.T) {
	typ := GetTypeRegistrar().Get("Painter")
	mm := typ.MethodByName("pixel")
	p := NewPainter(NewSize(8, 8))
	ink := NewColor(0, 0, 255, 255)
	p.Stroke().SetColor(ink)

	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, p)
	rp.DataStack.Push(NewPoint(1, 2))
	mm.invokeCall(f, ths)
	got := painterRGBA(p).RGBAAt(1, 2)
	want := color.RGBAModel.Convert(ink).(color.RGBA)
	if got != want {
		t.Fatalf("pixel Point = %v, want %v", got, want)
	}

	rp2 := NewTestRunProvider()
	f2, ths2 := memberRunFrame(rp2, p)
	rp2.DataStack.Push(Int(5))
	rp2.DataStack.Push(Int(6))
	mm.invokeCall(f2, ths2)
	got = painterRGBA(p).RGBAAt(5, 6)
	if got != want {
		t.Fatalf("pixel x y = %v, want %v", got, want)
	}
}

func TestPainterPixelCWritesExplicitColor(t *testing.T) {
	p := NewPainter(NewSize(8, 8))
	strokeInk := NewColor(0, 0, 255, 255)
	poke := NewColor(255, 0, 0, 255)
	p.Stroke().SetColor(strokeInk)
	p.Stroke().SetLineStyle(LineStyleNone)
	p.PixelC(poke, NewPoint(3, 4))
	got := painterRGBA(p).RGBAAt(3, 4)
	want := color.RGBAModel.Convert(poke).(color.RGBA)
	if got != want {
		t.Fatalf("pixelC = %v, want %v", got, want)
	}
	if p.Stroke().Color() != strokeInk {
		t.Fatal("pixelC must not change stroke.color")
	}
	if p.Stroke().LineStyle() != LineStyleNone {
		t.Fatal("pixelC must not change lineStyle")
	}
}

func TestPainterPixelCOutOfRangeNoOp(t *testing.T) {
	p := NewPainter(NewSize(8, 8))
	ink := NewColor(255, 0, 0, 255)
	p.PixelC(ink, NewPoint(8, 0)) // outside — clipped, no panic
	p.Pixmap().SetViewOrigin(NewPoint(100, 100))
	p.PixelC(ink, NewPoint(0, 0)) // world maps to (-100,-100) — clipped
	if got := painterRGBA(p).RGBAAt(0, 0); got.A != 0 {
		t.Fatalf("out-of-view pixelC should leave (0,0) untouched, got %v", got)
	}
}

func TestPainterPixelCNativeColorDeepest(t *testing.T) {
	typ := GetTypeRegistrar().Get("Painter")
	mm := typ.MethodByName("pixelC")
	p := NewPainter(NewSize(8, 8))
	ink := NewColor(0, 255, 0, 255)
	want := color.RGBAModel.Convert(ink).(color.RGBA)

	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, p)
	rp.DataStack.Push(ink)
	rp.DataStack.Push(ink) // dup: keep a copy under the poke
	rp.DataStack.Push(NewPoint(1, 1))
	mm.invokeCall(f, ths)
	if painterRGBA(p).RGBAAt(1, 1) != want {
		t.Fatal("pixelC Point")
	}
	if rp.DataStack.Len() != 1 || rp.DataStack.Peek().(Color) != ink {
		t.Fatal("pixelC should leave the unduped Color")
	}
	rp.DataStack.Push(Int(2))
	rp.DataStack.Push(Int(3))
	mm.invokeCall(f, ths)
	if painterRGBA(p).RGBAAt(2, 3) != want {
		t.Fatal("pixelC x y")
	}
	if rp.DataStack.Len() != 0 {
		t.Fatal("last pixelC should consume Color")
	}
}
