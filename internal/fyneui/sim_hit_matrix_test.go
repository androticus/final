package fyneui

import "testing"

func TestHitMatrix_AddOverwriteCanonical(t *testing.T) {
	o := NewSimGameOptions()
	stop := BoundaryHitActionStop | BoundaryHitEvents
	bounce := BoundaryHitActionBounce

	o.AddHitMatrixEntry(1, 2, stop)
	if got := o.LookupHitMatrix(2, 1); got != stop {
		t.Fatalf("unordered lookup: got 0x%02x want 0x%02x", got, stop)
	}
	o.AddHitMatrixEntry(2, 1, bounce)
	if got := o.LookupHitMatrix(1, 2); got != bounce {
		t.Fatalf("overwrite: got 0x%02x want 0x%02x", got, bounce)
	}
}

func TestHitMatrix_Default(t *testing.T) {
	o := NewSimGameOptions()
	if got := o.LookupHitMatrix(0, 1); got != DefaultHitMatrixMode {
		t.Fatalf("empty map default: got 0x%02x want 0x%02x", got, DefaultHitMatrixMode)
	}
	o.AddHitMatrixEntry(3, 4, BoundaryHitActionStop)
	if got := o.LookupHitMatrix(0, 1); got != DefaultHitMatrixMode {
		t.Fatalf("unrelated pair default: got 0x%02x want 0x%02x", got, DefaultHitMatrixMode)
	}
}

func TestHitMatrix_KindAllPriority(t *testing.T) {
	o := NewSimGameOptions()
	exact := BoundaryHitActionStop | BoundaryHitEvents
	oneSide := BoundaryHitActionStop
	allAll := BoundaryHitActionNone

	o.AddHitMatrixEntry(HitKindAll, HitKindAll, allAll)
	o.AddHitMatrixEntry(1, HitKindAll, oneSide)
	o.AddHitMatrixEntry(1, 2, exact)

	if got := o.LookupHitMatrix(1, 2); got != exact {
		t.Fatalf("exact beats wildcards: got 0x%02x want 0x%02x", got, exact)
	}
	if got := o.LookupHitMatrix(1, 9); got != oneSide {
		t.Fatalf("one-sided kindAll: got 0x%02x want 0x%02x", got, oneSide)
	}
	if got := o.LookupHitMatrix(5, 6); got != allAll {
		t.Fatalf("(kindAll,kindAll): got 0x%02x want 0x%02x", got, allAll)
	}
}

func TestHitMatrix_BothOneSidedFallsThrough(t *testing.T) {
	o := NewSimGameOptions()
	allAll := BoundaryHitActionNone
	o.AddHitMatrixEntry(1, HitKindAll, BoundaryHitActionStop)
	o.AddHitMatrixEntry(2, HitKindAll, BoundaryHitActionBounce)
	o.AddHitMatrixEntry(HitKindAll, HitKindAll, allAll)
	if got := o.LookupHitMatrix(1, 2); got != allAll {
		t.Fatalf("both one-sided → (all,all): got 0x%02x want 0x%02x", got, allAll)
	}
}

func TestHitMatrix_BorderNeverViaKindAll(t *testing.T) {
	o := NewSimGameOptions()
	o.AddHitMatrixEntry(HitKindAll, HitKindAll, BoundaryHitActionStop|BoundaryHitEvents)
	o.AddHitMatrixEntry(1, HitKindAll, BoundaryHitActionStop|BoundaryHitEvents)

	if got := o.LookupHitMatrix(1, HitKindBorder); got != DefaultHitMatrixMode {
		t.Fatalf("kindAll must not match border: got 0x%02x want default 0x%02x", got, DefaultHitMatrixMode)
	}
	if got := o.LookupHitMatrix(HitKindBorder, 5); got != DefaultHitMatrixMode {
		t.Fatalf("border vs other: got 0x%02x want default 0x%02x", got, DefaultHitMatrixMode)
	}

	exact := BoundaryHitActionStop | BoundaryHitEvents
	o.AddHitMatrixEntry(1, HitKindBorder, exact)
	if got := o.LookupHitMatrix(HitKindBorder, 1); got != exact {
		t.Fatalf("exact border cell: got 0x%02x want 0x%02x", got, exact)
	}
}
