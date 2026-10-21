package finaltui

import (
	"testing"
	"time"

	"github.com/androticus/final/finalembed"
)

func TestRaisePromptInFlightPanic(t *testing.T) {
	tui := &Tui{}
	tui.ArmUI()
	go func() {
		<-tui.UIRequests()
		// hold first raise
	}()

	go func() {
		params := finalembed.NewDynamicObject()
		params.Set(DOFieldBodyText.String(), finalembed.String("a"))
		tui.raisePrompt(params)
	}()

	deadline := time.Now().Add(2 * time.Second)
	for !tui.inFlight.Load() && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if !tui.inFlight.Load() {
		t.Fatal("first raise not in flight")
	}

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on second raise")
		}
	}()
	tui.raisePrompt(finalembed.NewDynamicObject())
}
