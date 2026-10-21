package core

import (
	"testing"
	"time"
)

func testDateTimeType(t *testing.T) *Type {
	t.Helper()
	typ := GetTypeRegistrar().Get("DateTime")
	if typ == nil {
		t.Fatal("DateTime type not registered")
	}
	return typ
}

func testIntervalType(t *testing.T) *Type {
	t.Helper()
	typ := GetTypeRegistrar().Get("Interval")
	if typ == nil {
		t.Fatal("Interval type not registered")
	}
	return typ
}

func TestIntervalClassConstants(t *testing.T) {
	ResetTestRuntime()
	ivT := testIntervalType(t)
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, ivT)

	for _, name := range []string{"nanosecond", "microsecond", "millisecond", "second", "minute", "hour"} {
		pm := ivT.GetClassPropertyMetadata(name)
		if pm == nil {
			t.Fatalf("missing Interval.%s class property", name)
		}
		pm.invokeGet(f, ths)
		iv, ok := rp.DataStack.Pop().(Interval)
		if !ok {
			t.Fatalf("Interval.%s = %T, want Interval", name, rp.DataStack.Peek())
		}
		if iv.Duration <= 0 {
			t.Fatalf("Interval.%s duration = %v, want > 0", name, iv.Duration)
		}
	}
	pm := ivT.GetClassPropertyMetadata("hour")
	f, ths = memberRunFrame(rp, ivT)
	pm.invokeGet(f, ths)
	hour := rp.DataStack.Pop().(Interval)
	if hour.Duration != time.Hour {
		t.Fatalf("Interval.hour = %v, want %v", hour.Duration, time.Hour)
	}
}

func TestDateTimeNewDate(t *testing.T) {
	ResetTestRuntime()
	dtT := testDateTimeType(t)
	rp := NewTestRunProvider()
	parts := NewAnyList()
	parts.AppendFromAny(Int(2026))
	parts.AppendFromAny(Int(7))
	parts.AppendFromAny(Int(4))
	rp.DataStack.Push(parts)
	mm := dtT.ClassMethodByName("newDate")
	f, ths := memberRunFrame(rp, dtT)
	mm.invokeCall(f, ths)
	dt := rp.DataStack.Pop().(time.Time)
	if dt.Year() != 2026 || dt.Month() != time.July || dt.Day() != 4 {
		t.Fatalf("date = %v-%v-%v", dt.Year(), dt.Month(), dt.Day())
	}
	if dt.Hour() != 0 || dt.Minute() != 0 || dt.Second() != 0 {
		t.Fatalf("time = %02d:%02d:%02d, want midnight", dt.Hour(), dt.Minute(), dt.Second())
	}
	if dt.Location() != time.UTC {
		t.Fatalf("location = %v, want UTC", dt.Location())
	}
}

func TestDateTimeNewTimeWithMillis(t *testing.T) {
	ResetTestRuntime()
	dtT := testDateTimeType(t)
	rp := NewTestRunProvider()
	parts := NewAnyList()
	parts.AppendFromAny(Int(14))
	parts.AppendFromAny(Int(30))
	parts.AppendFromAny(Int(0))
	parts.AppendFromAny(Int(500))
	rp.DataStack.Push(parts)
	mm := dtT.ClassMethodByName("newTime")
	f, ths := memberRunFrame(rp, dtT)
	mm.invokeCall(f, ths)
	dt := rp.DataStack.Pop().(time.Time)
	if dt.Hour() != 14 || dt.Minute() != 30 || dt.Second() != 0 {
		t.Fatalf("clock = %02d:%02d:%02d", dt.Hour(), dt.Minute(), dt.Second())
	}
	if dt.Nanosecond() != 500*int(time.Millisecond) {
		t.Fatalf("nsec = %d, want 500ms", dt.Nanosecond())
	}
}

func TestDateTimeNewDateTime(t *testing.T) {
	ResetTestRuntime()
	dtT := testDateTimeType(t)
	rp := NewTestRunProvider()
	parts := NewAnyList()
	for _, v := range []int{2026, 7, 4, 14, 30, 0, 250} {
		parts.AppendFromAny(Int(v))
	}
	rp.DataStack.Push(parts)
	mm := dtT.ClassMethodByName("newDateTime")
	f, ths := memberRunFrame(rp, dtT)
	mm.invokeCall(f, ths)
	dt := rp.DataStack.Pop().(time.Time)
	if dt.Year() != 2026 || dt.Month() != time.July || dt.Day() != 4 {
		t.Fatalf("date = %v-%v-%v", dt.Year(), dt.Month(), dt.Day())
	}
	if dt.Hour() != 14 || dt.Minute() != 30 || dt.Second() != 0 {
		t.Fatalf("clock = %02d:%02d:%02d", dt.Hour(), dt.Minute(), dt.Second())
	}
	if dt.Nanosecond() != 250*int(time.Millisecond) {
		t.Fatalf("nsec = %d, want 250ms", dt.Nanosecond())
	}
}

func TestDateTimeReadonlyProperties(t *testing.T) {
	ResetTestRuntime()
	dtT := testDateTimeType(t)
	dt := time.Date(2026, time.July, 4, 14, 30, 5, 0, time.UTC)
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, dt)

	checkIntProp := func(name string, want Int) {
		t.Helper()
		pm := dtT.PropertyByName(name)
		pm.invokeGet(f, ths)
		got := rp.DataStack.Pop().(Int)
		if got != want {
			t.Fatalf(".%s = %v, want %v", name, got, want)
		}
	}
	checkIntProp("year", Int(2026))
	checkIntProp("month", Int(time.July))
	checkIntProp("day", Int(4))
	checkIntProp("hour", Int(14))
	checkIntProp("minute", Int(30))
	checkIntProp("second", Int(5))
	pm := dtT.PropertyByName("isZero")
	pm.invokeGet(f, ths)
	if got := rp.DataStack.Pop().(Bool); got != Bool(false) {
		t.Fatalf(".isZero = %v, want false", got)
	}
}

func TestNewTypeObjPairDateTime(t *testing.T) {
	ResetTestRuntime()
	dt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	pair := NewTypeObjPair(dt)
	if pair.Typ != dateTimeType {
		t.Fatalf("Typ = %v, want DateTime", pair.Typ.name)
	}
	if pair.Obj.(time.Time) != dt {
		t.Fatalf("Obj = %v, want %v", pair.Obj, dt)
	}
}

func TestDateTimeToStringTimeSentinel(t *testing.T) {
	ResetTestRuntime()
	dtT := testDateTimeType(t)
	dt := time.Date(1, 1, 1, 14, 30, 5, 500*int(time.Millisecond), time.UTC)
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, dt)
	dtT.MethodByName("_toString").invokeCall(f, ths)
	got := string(rp.DataStack.Pop().(String))
	if got != "14.30.05.500" {
		t.Fatalf("_toString = %q, want 14.30.05.500", got)
	}
}

func TestDateTimeToStringISO(t *testing.T) {
	ResetTestRuntime()
	dtT := testDateTimeType(t)
	dt := time.Date(2026, 7, 4, 14, 30, 0, 250*int(time.Millisecond), time.UTC)
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, dt)
	dtT.MethodByName("_toString").invokeCall(f, ths)
	got := string(rp.DataStack.Pop().(String))
	if got != "2026-07-04T14:30:00.250" {
		t.Fatalf("_toString = %q, want 2026-07-04T14:30:00.250", got)
	}
}

func TestDateTimeToStringISOMidnightNoFraction(t *testing.T) {
	ResetTestRuntime()
	dtT := testDateTimeType(t)
	dt := time.Date(2026, 7, 4, 0, 0, 0, 0, time.UTC)
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, dt)
	dtT.MethodByName("_toString").invokeCall(f, ths)
	got := string(rp.DataStack.Pop().(String))
	if got != "2026-07-04T00:00:00" {
		t.Fatalf("_toString = %q, want 2026-07-04T00:00:00", got)
	}
}

func TestDateTimeRenderJson(t *testing.T) {
	ResetTestRuntime()
	dtT := testDateTimeType(t)
	dt := time.Date(2026, 7, 4, 14, 30, 0, 250*int(time.Millisecond), time.UTC)
	// Fractional seconds are dropped on purpose; the rendered form is
	// second-resolution RFC 3339, not json.Marshal's RFC3339Nano.
	want := `"2026-07-04T14:30:00Z"`
	rp := NewTestRunProvider()
	sb := NewStringBuilder()
	rp.ZStack.Push(sb)
	f, ths := memberRunFrame(rp, dt)
	dtT.MethodByName("_renderJson").invokeCall(f, ths)
	if got := sb.ToString(); got != want {
		t.Fatalf("_renderJson = %q, want %q", got, want)
	}
}

func TestNewTypeObjPairInterval(t *testing.T) {
	ResetTestRuntime()
	iv := Interval{time.Hour}
	pair := NewTypeObjPair(iv)
	if pair.Typ != intervalType {
		t.Fatalf("Typ = %v, want Interval", pair.Typ.name)
	}
	if pair.Obj.(Interval) != iv {
		t.Fatalf("Obj = %v, want %v", pair.Obj, iv)
	}
}

func TestIntervalReadonlyUnitProperties(t *testing.T) {
	ResetTestRuntime()
	ivT := testIntervalType(t)
	iv := Interval{50*time.Hour + 30*time.Minute + 1500*time.Millisecond}
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, iv)

	for _, name := range []string{"nanoseconds", "microseconds", "milliseconds", "seconds", "minutes", "hours", "days"} {
		pm := ivT.PropertyByName(name)
		if pm == nil || pm.set != nil {
			t.Fatalf("Interval.%s should be readonly const", name)
		}
	}

	checkInt := func(name string, want Int) {
		t.Helper()
		pm := ivT.PropertyByName(name)
		pm.invokeGet(f, ths)
		got := rp.DataStack.Pop().(Int)
		if got != want {
			t.Fatalf(".%s = %v, want %v", name, got, want)
		}
	}
	checkFloat := func(name string, want Float) {
		t.Helper()
		pm := ivT.PropertyByName(name)
		pm.invokeGet(f, ths)
		got := rp.DataStack.Pop().(Float)
		if got != want {
			t.Fatalf(".%s = %v, want %v", name, got, want)
		}
	}
	d := iv.Duration
	checkInt("nanoseconds", Int(d.Nanoseconds()))
	checkInt("microseconds", Int(d.Microseconds()))
	checkInt("milliseconds", Int(d.Milliseconds()))
	checkFloat("seconds", Float(d.Seconds()))
	checkFloat("minutes", Float(d.Minutes()))
	checkFloat("hours", Float(d.Hours()))
	checkFloat("days", Float(d.Hours()/24))
}

func TestIntervalToString(t *testing.T) {
	ResetTestRuntime()
	ivT := testIntervalType(t)
	iv := Interval{time.Hour + 30*time.Minute}
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, iv)
	ivT.MethodByName("_toString").invokeCall(f, ths)
	got := string(rp.DataStack.Pop().(String))
	if got != iv.Duration.String() {
		t.Fatalf("_toString = %q, want %q", got, iv.Duration.String())
	}
}
