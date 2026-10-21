package core

import (
	"errors"
	"strings"
	"testing"
)

func TestApiSessionGetSenderSend(t *testing.T) {
	ResetTestRuntime()
	sess, err := NewApiSession(&SessionOptions{Kind: SessionKindLocal})
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()

	if _, err := sess.GetSender().Send("1 2 + drop", 0); err != nil {
		t.Fatalf("Send: %v", err)
	}
}

func TestApiLockSessionUnlockInvalidates(t *testing.T) {
	ResetTestRuntime()
	ls, err := NewApiLockSession(&SessionOptions{Kind: SessionKindLocal})
	if err != nil {
		t.Fatal(err)
	}
	defer ls.Close()

	snd := ls.GetLockedSender()
	if _, err := snd.Send("1 2 + drop", 0); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if err := snd.Unlock(); err != nil {
		t.Fatalf("Unlock: %v", err)
	}
	if _, err := snd.Send("1 drop", 0); !errors.Is(err, ErrLockedSenderInvalid) {
		t.Fatalf("Send after Unlock: %v, want ErrLockedSenderInvalid", err)
	}
	if err := snd.Unlock(); !errors.Is(err, ErrLockedSenderInvalid) {
		t.Fatalf("second Unlock: %v, want ErrLockedSenderInvalid", err)
	}

	snd2 := ls.GetLockedSender()
	if _, err := snd2.Send("3 drop", 0); err != nil {
		t.Fatalf("Send after re-lock: %v", err)
	}
	if err := snd2.Unlock(); err != nil {
		t.Fatalf("Unlock snd2: %v", err)
	}
}

func TestApiSessionTwoGetSendersIndependent(t *testing.T) {
	ResetTestRuntime()
	sess, err := NewApiSession(&SessionOptions{Kind: SessionKindLocal})
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()

	a := sess.GetSender()
	b := sess.GetSender()
	if a == b {
		t.Fatal("GetSender returned the same pointer twice")
	}
	if _, err := a.Send("1 drop", 0); err != nil {
		t.Fatal(err)
	}
	if _, err := b.Send("2 drop", 0); err != nil {
		t.Fatal(err)
	}
}

const pipeMint = "8 Pipe.newWithCapacity^ drop"

func TestApiSenderRestrictOrFloor(t *testing.T) {
	ResetTestRuntime()
	sess, err := NewApiSession(&SessionOptions{Kind: SessionKindLocal})
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()

	// Session-only: floor denies, sender extra 0.
	sess.SetSecurityRestrictions(SR_Threads)
	if _, err := sess.GetSender().Send(pipeMint, 0); err == nil || !strings.Contains(err.Error(), "operation not allowed") {
		t.Fatalf("session floor: err = %v", err)
	}
	if sess.sess.task.restrict != SR_Threads {
		t.Fatalf("task.restrict after floor Send = %#x", sess.sess.task.restrict)
	}

	// Send cannot loosen the floor.
	sess.SetSecurityRestrictions(SR_Threads)
	snd := sess.GetSender()
	snd.SetSecurityRestrictions(0)
	if _, err := snd.Send(pipeMint, 0); err == nil || !strings.Contains(err.Error(), "operation not allowed") {
		t.Fatalf("cannot loosen: err = %v", err)
	}

	// Send-only: floor 0, sender extra denies.
	sess.SetSecurityRestrictions(0)
	snd = sess.GetSender()
	snd.SetSecurityRestrictions(SR_Threads)
	if _, err := snd.Send(pipeMint, 0); err == nil || !strings.Contains(err.Error(), "operation not allowed") {
		t.Fatalf("sender extra: err = %v", err)
	}
	if sess.sess.task.restrict != SR_Threads {
		t.Fatalf("task.restrict after sender extra = %#x", sess.sess.task.restrict)
	}

	// Hybrid OR: floor FileRead | sender Threads.
	sess.SetSecurityRestrictions(SR_FileRead)
	snd = sess.GetSender()
	snd.SetSecurityRestrictions(SR_Threads)
	if _, err := snd.Send(pipeMint, 0); err == nil || !strings.Contains(err.Error(), "operation not allowed") {
		t.Fatalf("OR extra: err = %v", err)
	}
	if sess.sess.task.restrict != (SR_FileRead | SR_Threads) {
		t.Fatalf("task.restrict OR = %#x, want %#x", sess.sess.task.restrict, SR_FileRead|SR_Threads)
	}

	// Two GetSenders do not share extra restrict.
	sess.SetSecurityRestrictions(0)
	deny := sess.GetSender()
	deny.SetSecurityRestrictions(SR_Threads)
	allow := sess.GetSender()
	if _, err := deny.Send(pipeMint, 0); err == nil {
		t.Fatal("deny sender should fail")
	}
	if _, err := allow.Send("1 2 + drop", 0); err != nil {
		t.Fatalf("allow sender: %v", err)
	}
}

func TestApiLockSenderRestrictSticky(t *testing.T) {
	ResetTestRuntime()
	ls, err := NewApiLockSession(&SessionOptions{Kind: SessionKindLocal})
	if err != nil {
		t.Fatal(err)
	}
	defer ls.Close()

	snd := ls.GetLockedSender()
	snd.SetSecurityRestrictions(SR_Threads)
	if _, err := snd.Send(pipeMint, 0); err == nil {
		t.Fatal("expected deny on first Send")
	}
	if _, err := snd.Send(pipeMint, 0); err == nil {
		t.Fatal("expected deny on second Send (sticky)")
	}
	snd.SetSecurityRestrictions(0)
	if _, err := snd.Send("1 drop", 0); err != nil {
		t.Fatalf("cleared extra: %v", err)
	}
	if err := snd.Unlock(); err != nil {
		t.Fatal(err)
	}
}
