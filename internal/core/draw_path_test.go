package core

import (
	"testing"
)

func TestPathBuilderMintedAndReadonly(t *testing.T) {
	p := NewPainter(NewSize(16, 16))
	if p.PathBuilder() == nil || p.pathBuilder.mode != pathBuilderReady {
		t.Fatal("pathBuilder should be minted ready")
	}
	if p.PathBuilder() != p.pathBuilder {
		t.Fatal("PathBuilder() mismatch")
	}
	pm := GetTypeRegistrar().Get("Painter").GetPropertyMetadata("pathBuilder")
	if pm == nil || pm.set != nil {
		t.Fatal("Painter.pathBuilder should be readonly const")
	}
}

func TestPathBuilderStartPathStroke(t *testing.T) {
	p := NewPainter(NewSize(64, 32))
	p.Stroke().SetColor(NewColor(255, 0, 0, 255))
	p.Stroke().SetLineWidth(3)
	b := p.PathBuilder()
	b.StartPath()
	b.MoveTo(4, 16)
	b.LineTo(60, 16)
	b.End()
	if !pixelPainted(painterRGBA(p), 32, 16) {
		t.Fatal("startPath lineTo end: expected paint")
	}
	c := painterRGBA(p).RGBAAt(32, 16)
	if c.R < 200 {
		t.Fatalf("path stroke = %v, want red-dominant", c)
	}
	if b.mode != pathBuilderReady {
		t.Fatal("end should return to ready")
	}
}

func TestPathBuilderStartSolidFill(t *testing.T) {
	p := NewPainter(NewSize(64, 64))
	p.Fill().SetStyle(FillStyleSolid)
	p.Fill().SetColor(NewColor(255, 0, 0, 255))
	p.Stroke().SetLineStyle(LineStyleNone)
	b := p.PathBuilder()
	b.StartSolid()
	b.MoveTo(8, 8)
	b.LineTo(56, 8)
	b.LineTo(32, 56)
	b.End()
	c := painterRGBA(p).RGBAAt(32, 24)
	if c.R < 200 || c.A == 0 {
		t.Fatalf("solid fill center = %v, want red", c)
	}
}

func TestPathBuilderReuseAfterEnd(t *testing.T) {
	p := NewPainter(NewSize(64, 32))
	p.Stroke().SetColor(NewColor(0, 0, 255, 255))
	p.Stroke().SetLineWidth(2)
	b := p.PathBuilder()
	b.StartPath()
	b.MoveTo(4, 8)
	b.LineTo(60, 8)
	b.End()
	b.StartPath()
	b.MoveTo(4, 24)
	b.LineTo(60, 24)
	b.End()
	if !pixelPainted(painterRGBA(p), 32, 8) || !pixelPainted(painterRGBA(p), 32, 24) {
		t.Fatal("reuse after end: expected both rows")
	}
}

func TestPathBuilderBezierToPaints(t *testing.T) {
	p := NewPainter(NewSize(64, 64))
	p.Stroke().SetColor(NewColor(0, 0, 0, 255))
	p.Stroke().SetLineWidth(2)
	b := p.PathBuilder()
	b.StartPath()
	b.MoveTo(4, 32)
	b.BezierTo(32, 4, 60, 32)
	b.End()
	if countPainted(painterRGBA(p)) == 0 {
		t.Fatal("bezierTo: no paint")
	}
}

func TestPathBuilderFloatCoordsNative(t *testing.T) {
	typ := GetTypeRegistrar().Get("PathBuilder")
	mm := typ.MethodByName("moveTo")
	ln := typ.MethodByName("lineTo")
	end := typ.MethodByName("end")
	p := NewPainter(NewSize(64, 32))
	p.Stroke().SetLineWidth(3)
	p.Stroke().SetColor(NewColor(255, 0, 0, 255))
	b := p.PathBuilder()
	b.StartPath()
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, b)
	rp.DataStack.Push(Float(4.5))
	rp.DataStack.Push(Float(16))
	mm.invokeCall(f, ths)
	rp.DataStack.Push(Float(60.25))
	rp.DataStack.Push(Float(16))
	ln.invokeCall(f, ths)
	end.invokeCall(f, ths)
	if !pixelPainted(painterRGBA(p), 32, 16) {
		t.Fatal("float moveTo/lineTo: expected paint")
	}
}

func TestPathBuilderEndWithoutStartPanics(t *testing.T) {
	p := NewPainter(NewSize(8, 8))
	defer func() {
		if recover() == nil {
			t.Fatal("end when ready should panic")
		}
	}()
	p.PathBuilder().End()
}

func TestPathBuilderStartWhileActivePanics(t *testing.T) {
	p := NewPainter(NewSize(8, 8))
	p.PathBuilder().StartPath()
	defer func() {
		if recover() == nil {
			t.Fatal("startPath while path should panic")
		}
	}()
	p.PathBuilder().StartSolid()
}

func TestPathBuilderMoveToWhenReadyPanics(t *testing.T) {
	p := NewPainter(NewSize(8, 8))
	defer func() {
		if recover() == nil {
			t.Fatal("moveTo when ready should panic")
		}
	}()
	p.PathBuilder().MoveTo(1, 1)
}

func TestPainterLineWhilePathBuilderActivePanics(t *testing.T) {
	p := NewPainter(NewSize(16, 16))
	p.PathBuilder().StartPath()
	defer func() {
		if recover() == nil {
			t.Fatal("line while pathBuilder active should panic")
		}
	}()
	p.Line(NewPoint(0, 0), NewPoint(4, 4))
}

func TestPainterPushWhilePathBuilderActivePanics(t *testing.T) {
	p := NewPainter(NewSize(16, 16))
	p.PathBuilder().StartSolid()
	defer func() {
		if recover() == nil {
			t.Fatal("push while pathBuilder active should panic")
		}
	}()
	p.Push()
}

func TestPathBuilderArcPaints(t *testing.T) {
	p := NewPainter(NewSize(64, 64))
	p.Stroke().SetColor(NewColor(0, 128, 0, 255))
	p.Stroke().SetLineWidth(3)
	b := p.PathBuilder()
	b.StartPath()
	b.Arc(32, 32, 20, 0, 180)
	b.End()
	if countPainted(painterRGBA(p)) == 0 {
		t.Fatal("arc: no paint")
	}
}
