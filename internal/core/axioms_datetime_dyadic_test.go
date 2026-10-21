package core

import (
	"testing"
	"time"
)

func utcDyadic(y int, mo time.Month, d, h, m, s int) time.Time {
	return time.Date(y, mo, d, h, m, s, 0, time.UTC)
}

func runDtDyadic(t *testing.T, subCode int, op1, op2 any) any {
	t.Helper()
	ResetTestRuntime()
	rp := NewTestRunProvider()
	rp.DataStack.Push(op1)
	rp.DataStack.Push(op2)
	RunOpsWithSubCode(NewTestTask(), rp, subCode, OpDtDyadic, OpStop)
	if rp.DataStack.Len() != 1 {
		t.Fatalf("stack depth = %d, want 1", rp.DataStack.Len())
	}
	return rp.DataStack.Pop()
}

func TestDtDyadicWordsRegistered(t *testing.T) {
	ResetTestRuntime()
	for _, name := range []string{
		"dt::+", "dt::-", "dt::*", "dt::/", "dt::compare", "dt::<", "dt::<=", "dt::>", "dt::>=", "dt::==", "dt::<>",
		"dt::round", "dt::fmt",
	} {
		w := globalDictionaryContext.FindWordByName(name)
		if w == nil {
			t.Fatalf("missing word %q", name)
		}
		if w.opCode != OpDtDyadic {
			t.Fatalf("%q opCode = %v, want OpDtDyadic", name, w.opCode)
		}
	}
}

func TestOpDtDyadicAdd(t *testing.T) {
	t1 := utcDyadic(2026, time.July, 4, 12, 0, 0)
	t2 := utcDyadic(2026, time.July, 4, 14, 0, 0)
	twoH := Interval{2 * time.Hour}

	got := runDtDyadic(t, SC_dt_add, t1, twoH).(time.Time)
	if !got.Equal(t2) {
		t.Fatalf("DateTime + Interval = %v, want %v", got, t2)
	}

	got = runDtDyadic(t, SC_dt_add, twoH, t1).(time.Time)
	if !got.Equal(t2) {
		t.Fatalf("Interval + DateTime = %v, want %v", got, t2)
	}

	sum := runDtDyadic(t, SC_dt_add, Interval{time.Hour}, Interval{30 * time.Minute}).(Interval)
	if sum.Duration != 90*time.Minute {
		t.Fatalf("Interval + Interval = %v, want 90m", sum.Duration)
	}
}

func TestOpDtDyadicSub(t *testing.T) {
	t1 := utcDyadic(2026, time.July, 4, 12, 0, 0)
	t2 := utcDyadic(2026, time.July, 4, 14, 0, 0)
	twoH := Interval{2 * time.Hour}

	got := runDtDyadic(t, SC_dt_sub, t2, twoH).(time.Time)
	if !got.Equal(t1) {
		t.Fatalf("DateTime - Interval = %v, want %v", got, t1)
	}

	elapsed := runDtDyadic(t, SC_dt_sub, t2, t1).(Interval)
	if elapsed.Duration != 2*time.Hour {
		t.Fatalf("DateTime - DateTime = %v, want 2h", elapsed.Duration)
	}

	diff := runDtDyadic(t, SC_dt_sub, Interval{2 * time.Hour}, Interval{30 * time.Minute}).(Interval)
	if diff.Duration != 90*time.Minute {
		t.Fatalf("Interval - Interval = %v, want 90m", diff.Duration)
	}
}

func TestOpDtDyadicMul(t *testing.T) {
	unit := Interval{time.Millisecond}
	want := Interval{100 * time.Millisecond}

	got := runDtDyadic(t, SC_dt_mul, Int(100), unit).(Interval)
	if got.Duration != want.Duration {
		t.Fatalf("Int * Interval = %v, want %v", got.Duration, want.Duration)
	}
	got = runDtDyadic(t, SC_dt_mul, unit, Int(100)).(Interval)
	if got.Duration != want.Duration {
		t.Fatalf("Interval * Int = %v, want %v", got.Duration, want.Duration)
	}
}

func TestOpDtDyadicRound(t *testing.T) {
	// Halfway rounds away from zero (time.Duration.Round).
	got := runDtDyadic(t, SC_dt_round, Interval{1500 * time.Millisecond}, Interval{time.Second}).(Interval)
	if got.Duration != 2*time.Second {
		t.Fatalf("1500ms round 1s = %v, want 2s", got.Duration)
	}
	got = runDtDyadic(t, SC_dt_round, Interval{80 * time.Minute}, Interval{time.Hour}).(Interval)
	if got.Duration != time.Hour {
		t.Fatalf("80m round 1h = %v, want 1h", got.Duration)
	}
	got = runDtDyadic(t, SC_dt_round, Interval{-1500 * time.Millisecond}, Interval{time.Second}).(Interval)
	if got.Duration != -2*time.Second {
		t.Fatalf("-1500ms round 1s = %v, want -2s", got.Duration)
	}
}

func TestOpDtDyadicDiv(t *testing.T) {
	total := Interval{100 * time.Millisecond}
	want := Interval{10 * time.Millisecond}

	got := runDtDyadic(t, SC_dt_div, total, Int(10)).(Interval)
	if got.Duration != want.Duration {
		t.Fatalf("Interval / Int = %v, want %v", got.Duration, want.Duration)
	}
}

func TestOpDtDyadicDivPanics(t *testing.T) {
	iv := Interval{time.Second}
	cases := []struct {
		name string
		op1  any
		op2  any
	}{
		{"n=0", iv, Int(0)},
		{"n=-1", iv, Int(-1)},
		{"Int/Interval", Int(10), iv},
		{"Interval/Interval", iv, iv},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("expected panic")
				}
			}()
			runDtDyadic(t, SC_dt_div, tc.op1, tc.op2)
		})
	}
}

func TestOpDtDyadicCompareOps(t *testing.T) {
	early := utcDyadic(2026, time.July, 4, 10, 0, 0)
	late := utcDyadic(2026, time.July, 4, 12, 0, 0)
	same := early
	short := Interval{30 * time.Minute}
	long := Interval{2 * time.Hour}

	cases := []struct {
		subCode int
		op1     any
		op2     any
		want    any
	}{
		{SC_dt_compare, early, late, Int(-1)},
		{SC_dt_compare, late, early, Int(1)},
		{SC_dt_compare, same, early, Int(0)},
		{SC_dt_lt, early, late, Bool(true)},
		{SC_dt_lte, early, late, Bool(true)},
		{SC_dt_gt, late, early, Bool(true)},
		{SC_dt_gte, same, early, Bool(true)},
		{SC_dt_eq, same, early, Bool(true)},
		{SC_dt_ne, early, late, Bool(true)},
		{SC_dt_compare, short, long, Int(-1)},
		{SC_dt_eq, short, short, Bool(true)},
	}

	for _, tc := range cases {
		got := runDtDyadic(t, tc.subCode, tc.op1, tc.op2)
		if got != tc.want {
			t.Fatalf("subcode %d(%v, %v) = %v, want %v", tc.subCode, tc.op1, tc.op2, got, tc.want)
		}
	}
}

func TestOpDtDyadicInvalidOperandsPanic(t *testing.T) {
	t1 := utcDyadic(2026, time.July, 4, 12, 0, 0)
	t2 := utcDyadic(2026, time.July, 4, 14, 0, 0)
	iv := Interval{time.Hour}

	panics := []struct {
		name    string
		subCode int
		op1     any
		op2     any
	}{
		{"DateTime+DateTime", SC_dt_add, t1, t2},
		{"Interval-DateTime", SC_dt_sub, iv, t1},
		{"Interval*Interval", SC_dt_mul, iv, iv},
		{"Int*Int", SC_dt_mul, Int(2), Int(3)},
		{"DateTime compare Interval", SC_dt_compare, t1, iv},
		{"Interval compare DateTime", SC_dt_compare, iv, t1},
	}

	for _, tc := range panics {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("expected panic")
				}
			}()
			runDtDyadic(t, tc.subCode, tc.op1, tc.op2)
		})
	}
}
