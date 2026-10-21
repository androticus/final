package fyneui

// hitMatrixKey is an unordered pair of HitArea.kind values (a <= b).
type hitMatrixKey struct {
	a, b int32
}

func hitMatrixCanonical(ka, kb int32) hitMatrixKey {
	if ka > kb {
		ka, kb = kb, ka
	}
	return hitMatrixKey{a: ka, b: kb}
}

// DefaultHitMatrixMode is bounce with HitEvents off (no matching cell).
const DefaultHitMatrixMode = BoundaryHitActionBounce

// AddHitMatrixEntry stores/overwrites a sparse Hit Matrix cell for HitArea.kind pair.
// Key is unordered: (a,b) and (b,a) share one cell.
func (o *SimGameOptions) AddHitMatrixEntry(ka, kb int32, mode BoundaryMode) {
	if o.hitMatrix == nil {
		o.hitMatrix = make(map[hitMatrixKey]BoundaryMode)
	}
	o.hitMatrix[hitMatrixCanonical(ka, kb)] = mode
}

// LookupHitMatrix resolves pair policy for two HitArea.kind values.
// Priority: exact → one-sided kindAll → (kindAll,kindAll) → default bounce.
// When either kind is kindBorder, kindAll cells are skipped (boundary seed is Phase 4).
func (o *SimGameOptions) LookupHitMatrix(ka, kb int32) BoundaryMode {
	m := o.hitMatrix
	if m == nil {
		return DefaultHitMatrixMode
	}
	exact := hitMatrixCanonical(ka, kb)
	if mode, ok := m[exact]; ok {
		return mode
	}
	involvesBorder := ka == HitKindBorder || kb == HitKindBorder
	if !involvesBorder {
		var oneMode BoundaryMode
		oneCount := 0
		if ka != HitKindAll {
			if mode, ok := m[hitMatrixCanonical(ka, HitKindAll)]; ok {
				oneMode = mode
				oneCount++
			}
		}
		if kb != HitKindAll && kb != ka {
			if mode, ok := m[hitMatrixCanonical(kb, HitKindAll)]; ok {
				oneMode = mode
				oneCount++
			}
		}
		if oneCount == 1 {
			return oneMode
		}
		if mode, ok := m[hitMatrixCanonical(HitKindAll, HitKindAll)]; ok {
			return mode
		}
	}
	return DefaultHitMatrixMode
}
