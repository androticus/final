package finalgo

import (
	"strings"
	"testing"
)

type mockSender struct {
	lastScript string
	response   string
	err        error
}

func (m *mockSender) SetSecurityRestrictions(restrictions SecurityRestrictions) {
	_ = restrictions
}

func (m *mockSender) Send(script string, timeoutS int) (string, error) {
	m.lastScript = script
	_ = timeoutS
	return m.response, m.err
}

func TestFinalTool_staticGetters(t *testing.T) {
	ft := NewFinalTool()
	if ft.Name() != "final" {
		t.Fatalf("name = %q", ft.Name())
	}
	if !strings.Contains(ft.Description(), "RPN") {
		t.Fatalf("description = %q", ft.Description())
	}
	if !strings.Contains(string(ft.ParametersJSON()), `"script"`) {
		t.Fatalf("params = %s", ft.ParametersJSON())
	}
}

func TestFinalTool_Dispatch(t *testing.T) {
	ft := NewFinalTool()
	snd := &mockSender{response: "7\n"}
	content, err := ft.Dispatch(snd, "call-1", "final", `{"script":"3 4 + $"}`)
	if err != nil {
		t.Fatal(err)
	}
	if content != "7\n" {
		t.Fatalf("content = %q", content)
	}
	if snd.lastScript != "3 4 + $" {
		t.Fatalf("script = %q", snd.lastScript)
	}
}

func TestFinalTool_Dispatch_wrongName(t *testing.T) {
	ft := NewFinalTool()
	_, err := ft.Dispatch(&mockSender{}, "id", "other", `{}`)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFinalTool_Dispatch_runError(t *testing.T) {
	ft := NewFinalTool()
	snd := &mockSender{err: errTestFault}
	content, err := ft.Dispatch(snd, "id", "final", `{"script":"1 0 /"}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(content, "[error]") {
		t.Fatalf("content = %q", content)
	}
}

var errTestFault = &testFault{"divide by zero"}

type testFault struct{ msg string }

func (e *testFault) Error() string { return e.msg }
