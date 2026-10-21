package fyneui

import (
	"testing"

	"github.com/androticus/final/internal/core"
)

func TestModalCompleteOneShot(t *testing.T) {
	w := beginModalWait()
	w.complete(btnOk, nil)
	w.complete(btnCancel, "ignored")
	res := w.wait()
	if res == nil {
		t.Fatal("wait returned nil")
	}
	if res.CloserIndex != btnOk {
		t.Fatalf("CloserIndex=%d, want %d", res.CloserIndex, btnOk)
	}
	if res.Data != nil {
		t.Fatalf("Data=%v, want nil", res.Data)
	}
}

func TestModalWaitAfterComplete(t *testing.T) {
	w := beginModalWait()
	go func() {
		w.complete(btnOk, nil)
		w.complete(btnCancel, "ignored")
	}()
	res := w.wait()
	if res.CloserIndex != btnOk {
		t.Fatalf("CloserIndex=%d, want %d", res.CloserIndex, btnOk)
	}
}

func TestModalInFlightPanics(t *testing.T) {
	w := beginModalWait()
	t.Cleanup(func() {
		w.complete(btnCancel, nil)
		_ = w.wait()
	})
	defer func() {
		if recover() == nil {
			t.Fatal("second beginModalWait must panic")
		}
	}()
	_ = beginModalWait()
}

func TestInformationDialogShowRegistered(t *testing.T) {
	typ := core.GetTypeRegistrar().Get("InformationDialog")
	if typ == nil {
		t.Fatal("InformationDialog missing")
	}
	if typ.GetClassMethodMetadata("show") == nil {
		t.Fatal("InformationDialog.show missing")
	}
}

func TestConfirmationDialogShowRegistered(t *testing.T) {
	typ := core.GetTypeRegistrar().Get("ConfirmationDialog")
	if typ == nil {
		t.Fatal("ConfirmationDialog missing")
	}
	if typ.GetClassMethodMetadata("show") == nil {
		t.Fatal("ConfirmationDialog.show missing")
	}
}

func TestColorDialogShowRegistered(t *testing.T) {
	typ := core.GetTypeRegistrar().Get("ColorDialog")
	if typ == nil {
		t.Fatal("ColorDialog missing")
	}
	if typ.GetClassMethodMetadata("show") == nil {
		t.Fatal("ColorDialog.show missing")
	}
}

func TestFileDialogShowRegistered(t *testing.T) {
	typ := core.GetTypeRegistrar().Get("FileDialog")
	if typ == nil {
		t.Fatal("FileDialog missing")
	}
	for _, name := range []string{"showOpenFile", "showSaveText", "showSaveBytes", "showOpenFolder"} {
		if typ.GetClassMethodMetadata(name) == nil {
			t.Fatalf("FileDialog.%s missing", name)
		}
	}
}

func TestFormShowModalRegistered(t *testing.T) {
	typ := core.GetTypeRegistrar().Get("Form")
	if typ == nil {
		t.Fatal("Form missing")
	}
	if typ.GetMethodMetadata("showModal") == nil {
		t.Fatal("Form.showModal missing")
	}
}

func TestConfirmCloserIndex(t *testing.T) {
	if got := confirmCloserIndex(true); got != btnOk {
		t.Fatalf("yes → %d, want %d", got, btnOk)
	}
	if got := confirmCloserIndex(false); got != btnCancel {
		t.Fatalf("no → %d, want %d", got, btnCancel)
	}
}
