package core

import (
	"testing"
	"time"
)

func TestDtFmtWordRegistered(t *testing.T) {
	ResetTestRuntime()
	w := globalDictionaryContext.FindWordByName("dt::fmt")
	if w == nil {
		t.Fatal("missing word dt::fmt")
	}
	if w.opCode != OpDtDyadic {
		t.Fatalf("dt::fmt opCode = %v, want OpDtDyadic", w.opCode)
	}
	if w.subCode != SC_dt_fmt {
		t.Fatalf("dt::fmt subCode = %d, want %d", w.subCode, SC_dt_fmt)
	}
}

func TestFormatDateTimePattern(t *testing.T) {
	// Friday Jul 4, 2026 14:30:05.250123456 UTC
	tm := time.Date(2026, time.July, 4, 14, 30, 5, 250123456, time.UTC)
	cases := []struct {
		pat, want string
	}{
		{"yyyy", "2026"},
		{"yy", "26"},
		{"MM", "07"},
		{"M", "7"},
		{"MMM", "Jul"},
		{"MMMM", "July"},
		{"d", "4"},
		{"dd", "04"},
		{"EEE", "Sat"},
		{"EEEE", "Saturday"},
		{"HH", "14"},
		{"hh", "02"},
		{"mm", "30"},
		{"ss", "05"},
		{"ss.sss", "05.250"},
		{"ss.ssssss", "05.250123"},
		{"ss.sssssssss", "05.250123456"},
		{"hh:mm A", "02:30 PM"},
		{"hh:mm a", "02:30 pm"},
		{"yyyy-MM-ddTHH:mm:ss.sssZ", "2026-07-04T14:30:05.250Z"},
		{"yyyy-MM-dd HH:mm:ss", "2026-07-04 14:30:05"},
		{"EEEE, MMMM d, yyyy", "Saturday, July 4, 2026"},
		{"MMM d, yyyy", "Jul 4, 2026"},
	}
	for _, tc := range cases {
		got := formatDateTimePattern(tm, tc.pat)
		if got != tc.want {
			t.Fatalf("pattern %q = %q, want %q", tc.pat, got, tc.want)
		}
	}

	am := time.Date(2026, time.July, 4, 0, 15, 0, 0, time.UTC)
	if got := formatDateTimePattern(am, "hh:mm A"); got != "12:15 AM" {
		t.Fatalf("midnight 12h = %q, want 12:15 AM", got)
	}
	noon := time.Date(2026, time.July, 4, 12, 0, 0, 0, time.UTC)
	if got := formatDateTimePattern(noon, "hh A"); got != "12 PM" {
		t.Fatalf("noon 12h = %q, want 12 PM", got)
	}
}

func TestFormatDateTimeConstants(t *testing.T) {
	ResetTestRuntime()
	tm := time.Date(2026, time.July, 4, 14, 30, 5, 250000000, time.UTC)
	for _, name := range []string{"iso", "date", "time12", "time24", "dateTime", "longDate", "shortDate"} {
		pat := invokeClassConst(t, dateTimeType, name).(String)
		_ = formatDateTimePattern(tm, string(pat))
	}
	iso := invokeClassConst(t, dateTimeType, "iso").(String)
	got := formatDateTimePattern(tm, string(iso))
	want := "2026-07-04T14:30:05.250Z"
	if got != want {
		t.Fatalf("DateTime.iso = %q, want %q", got, want)
	}
}

func TestFormatIntervalPattern(t *testing.T) {
	fiftyH := Interval{50 * time.Hour}
	cases := []struct {
		pat string
		iv  Interval
		want string
	}{
		{"hh:mm", fiftyH, "50:00"},
		{"d hh:mm", fiftyH, "2 02:00"},
		{"dd hh:mm", fiftyH, " 2 02:00"},
		{"mm", fiftyH, "3000"},
		{"HH:mm:ss", Interval{time.Hour + 30*time.Minute}, "01:30:00"}, // case-insensitive
		{"ss.sss", Interval{1250 * time.Millisecond}, "01.250"},
		{"+d hh:mm:ss", Interval{2*24*time.Hour + 3*time.Hour}, " 2 03:00:00"},
		{"+d hh:mm:ss", Interval{-(2*24*time.Hour + 3*time.Hour)}, "-2 03:00:00"},
		{"d hh:mm:ss", Interval{-(2*24*time.Hour + 3*time.Hour)}, "-2 03:00:00"},
	}
	for _, tc := range cases {
		got := formatIntervalPattern(tc.iv, tc.pat)
		if got != tc.want {
			t.Fatalf("pattern %q val %v = %q, want %q", tc.pat, tc.iv.Duration, got, tc.want)
		}
	}
}

func TestOpDtFmt(t *testing.T) {
	tm := time.Date(2026, time.July, 4, 14, 30, 5, 0, time.UTC)
	got := runDtDyadic(t, SC_dt_fmt, tm, String("yyyy-MM-dd")).(String)
	if got != "2026-07-04" {
		t.Fatalf("dt::fmt DateTime = %q, want 2026-07-04", got)
	}
	iv := Interval{90 * time.Minute}
	got = runDtDyadic(t, SC_dt_fmt, iv, String("hh:mm:ss")).(String)
	if got != "01:30:00" {
		t.Fatalf("dt::fmt Interval = %q, want 01:30:00", got)
	}
}

func TestOpDtFmtPanics(t *testing.T) {
	tm := time.Date(2026, time.July, 4, 14, 0, 0, 0, time.UTC)
	ResetTestRuntime()
	rp := NewTestRunProvider()
	rp.DataStack.Push(tm)
	rp.DataStack.Push(String("yyyy-WW"))
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for unknown DateTime token")
		}
	}()
	RunOpsWithSubCode(NewTestTask(), rp, SC_dt_fmt, OpDtDyadic, OpStop)
}

func TestFormatDateTimeUnknownTokenPanics(t *testing.T) {
	tm := time.Date(2026, time.July, 4, 14, 0, 0, 0, time.UTC)
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	_ = formatDateTimePattern(tm, "yyyy-WW")
}

func TestJuly4Weekday(t *testing.T) {
	tm := time.Date(2026, time.July, 4, 14, 30, 5, 0, time.UTC)
	// sanity: 2026-07-04 is Saturday
	if tm.Weekday() != time.Saturday {
		t.Fatalf("weekday = %v, want Saturday (fix fixture)", tm.Weekday())
	}
}
