package fyneui

import "testing"

func TestEffectDrawer_NewCyclerRecipe(t *testing.T) {
	d := NewEffectDrawer(8, 40, 5, 0)
	c := d.NewCycler()
	if c.phaseCount != 8 || c.phaseRate != 40 || c.phaseOffset != 5 || c.cyclesToRun != 0 {
		t.Fatalf("recipe: count=%d rate=%d off=%d run=%d",
			c.phaseCount, c.phaseRate, c.phaseOffset, c.cyclesToRun)
	}
	c.Advance(0)
	c.Advance(45) // past offset 5 → elapsed 40 → phase 1
	if c.currentPhase != 1 {
		t.Fatalf("phase=%d want 1", c.currentPhase)
	}
}

func TestEffectDrawer_FiniteRecipe(t *testing.T) {
	d := NewEffectDrawer(8, 50, 0, 2)
	c := d.NewCycler()
	if c.cyclesToRun != 2 || c.phaseCount != 8 || c.phaseRate != 50 {
		t.Fatalf("got count=%d rate=%d run=%d", c.phaseCount, c.phaseRate, c.cyclesToRun)
	}
}

func TestEffectDrawer_HandleRenderPanicsWithoutPaint(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	d := NewEffectDrawer(1, 1, 0, 0)
	d.HandleRender(nil)
}

func TestEffectDrawer_TypeRegistered(t *testing.T) {
	if effectDrawerType == nil || effectDrawerType.GetPropertyMetadata("phaseCount") == nil {
		t.Fatal("EffectDrawer not registered")
	}
	if effectDrawerType.GetMethodMetadata("newCycler") == nil {
		t.Fatal("missing newCycler")
	}
	if effectDrawerType.GetClassMethodMetadata("newWithRender") == nil {
		t.Fatal("missing newWithRender")
	}
}
