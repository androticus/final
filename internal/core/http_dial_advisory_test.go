package core

import (
	"testing"
	"time"
)

func TestHttpSendAppliesAdvisoryTimeoutWhenReqUnset(t *testing.T) {
	task := NewTestTask()
	task.advisoryTimeoutS = 30
	f := &RunFrame{Task: task, RP: NewTestRunProvider()}

	req := NewHttpReq()
	req.hostAddress.host = "example.com"
	req.timeoutS = 0
	if req.timeoutS == 0 {
		req.timeoutS = f.AdvisoryTimeoutS()
	}

	gpr, err := NewHttpDialer().PrepareRequest(req)
	if err != nil {
		t.Fatal(err)
	}
	want := 30 * time.Second
	if gpr.client.Timeout != want {
		t.Fatalf("client.Timeout = %v, want %v", gpr.client.Timeout, want)
	}
}

func TestHttpSendKeepsExplicitReqTimeoutOverAdvisory(t *testing.T) {
	task := NewTestTask()
	task.advisoryTimeoutS = 30
	f := &RunFrame{Task: task, RP: NewTestRunProvider()}

	req := NewHttpReq()
	req.hostAddress.host = "example.com"
	req.timeoutS = 10
	if req.timeoutS == 0 {
		req.timeoutS = f.AdvisoryTimeoutS()
	}

	gpr, err := NewHttpDialer().PrepareRequest(req)
	if err != nil {
		t.Fatal(err)
	}
	want := 10 * time.Second
	if gpr.client.Timeout != want {
		t.Fatalf("client.Timeout = %v, want %v", gpr.client.Timeout, want)
	}
}
