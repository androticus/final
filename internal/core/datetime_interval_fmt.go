package core

import (
	"fmt"
	"strings"
	"time"
	"unicode"
)

// DateTime / Interval pattern formatting for dt::fmt — see docs/003-design-notes/158-datetime-format.md.
// Users never see Go reference layouts or Duration.String lexicons.

func runDtDyadicFmt(ds *Stack[any]) {
	fmtVal := ds.Pop()
	val := ds.Pop()
	pat, ok := fmtVal.(String)
	if !ok {
		panic(fmt.Errorf("dt::fmt: format must be String, got %T", fmtVal))
	}
	switch v := val.(type) {
	case time.Time:
		ds.Push(String(formatDateTimePattern(v, string(pat))))
	case Interval:
		ds.Push(String(formatIntervalPattern(v, string(pat))))
	default:
		panic(fmt.Errorf("dt::fmt: value must be DateTime or Interval, got %T", val))
	}
}

func padZeroInt(n, width int) string {
	if n < 0 {
		n = -n
	}
	s := fmt.Sprintf("%d", n)
	if len(s) >= width {
		return s
	}
	return strings.Repeat("0", width-len(s)) + s
}

func padSpaceInt(n, width int) string {
	if n < 0 {
		n = -n
	}
	s := fmt.Sprintf("%d", n)
	if len(s) >= width {
		return s
	}
	return strings.Repeat(" ", width-len(s)) + s
}

func fracDigitsFromNsec(nsec, width int) string {
	if width < 1 {
		width = 1
	}
	if width > 9 {
		width = 9
	}
	div := 1
	for i := 0; i < 9-width; i++ {
		div *= 10
	}
	return padZeroInt(nsec/div, width)
}

func hour12(h int) int {
	h = h % 12
	if h == 0 {
		return 12
	}
	return h
}

func formatDateTimePattern(t time.Time, pattern string) string {
	t = t.UTC()
	var b strings.Builder
	i := 0
	for i < len(pattern) {
		c := pattern[i]
		switch c {
		case 'y':
			n := countRun(pattern, i, 'y')
			switch n {
			case 2:
				b.WriteString(padZeroInt(t.Year()%100, 2))
			case 4:
				b.WriteString(padZeroInt(t.Year(), 4))
			default:
				panic(fmt.Errorf("dt::fmt DateTime: unsupported year run %q (use yy or yyyy)", pattern[i:i+n]))
			}
			i += n
		case 'M':
			n := countRun(pattern, i, 'M')
			switch n {
			case 1:
				b.WriteString(fmt.Sprintf("%d", int(t.Month())))
			case 2:
				b.WriteString(padZeroInt(int(t.Month()), 2))
			case 3:
				b.WriteString(t.Month().String()[:3])
			case 4:
				b.WriteString(t.Month().String())
			default:
				panic(fmt.Errorf("dt::fmt DateTime: unsupported month run %q (use M, MM, MMM, or MMMM)", pattern[i:i+n]))
			}
			i += n
		case 'd':
			n := countRun(pattern, i, 'd')
			switch n {
			case 1:
				b.WriteString(fmt.Sprintf("%d", t.Day()))
			case 2:
				b.WriteString(padZeroInt(t.Day(), 2))
			default:
				panic(fmt.Errorf("dt::fmt DateTime: unsupported day run %q (use d or dd)", pattern[i:i+n]))
			}
			i += n
		case 'E':
			n := countRun(pattern, i, 'E')
			name := t.Weekday().String()
			switch n {
			case 3:
				b.WriteString(name[:3])
			case 4:
				b.WriteString(name)
			default:
				panic(fmt.Errorf("dt::fmt DateTime: unsupported weekday run %q (use EEE or EEEE)", pattern[i:i+n]))
			}
			i += n
		case 'H':
			n := countRun(pattern, i, 'H')
			b.WriteString(padZeroInt(t.Hour(), max(n, 2)))
			i += n
		case 'h':
			n := countRun(pattern, i, 'h')
			b.WriteString(padZeroInt(hour12(t.Hour()), max(n, 2)))
			i += n
		case 'm':
			n := countRun(pattern, i, 'm')
			b.WriteString(padZeroInt(t.Minute(), max(n, 2)))
			i += n
		case 's':
			secRun, fracRun := scanSecondsAndFrac(pattern, i)
			b.WriteString(padZeroInt(t.Second(), max(secRun, 2)))
			i += secRun
			if fracRun > 0 {
				b.WriteByte('.')
				b.WriteString(fracDigitsFromNsec(t.Nanosecond(), fracRun))
				i += 1 + fracRun // '.' + s run
			}
		case 'a':
			if t.Hour() < 12 {
				b.WriteString("am")
			} else {
				b.WriteString("pm")
			}
			i++
		case 'A':
			if t.Hour() < 12 {
				b.WriteString("AM")
			} else {
				b.WriteString("PM")
			}
			i++
		case 'T', 'Z':
			// ISO literals (design note 158) — not tokens
			b.WriteByte(c)
			i++
		case '\'':
			lit, next := scanQuotedLiteral(pattern, i)
			b.WriteString(lit)
			i = next
		default:
			if unicode.IsLetter(rune(c)) {
				panic(fmt.Errorf("dt::fmt DateTime: unknown token starting at %q", pattern[i:]))
			}
			b.WriteByte(c)
			i++
		}
	}
	return b.String()
}

func formatIntervalPattern(iv Interval, pattern string) string {
	pat := strings.ToLower(pattern)
	neg := iv.Duration < 0
	abs := iv.Duration
	if neg {
		abs = -abs
	}

	hasDays := strings.Contains(pat, "d")
	hasHours := strings.Contains(pat, "h")
	hasMins := strings.Contains(pat, "m")
	hasSecs := strings.Contains(pat, "s")
	hasSignSlot := strings.Contains(pat, "+")

	if !hasDays && !hasHours && !hasMins && !hasSecs {
		panic(fmt.Errorf("dt::fmt Interval: pattern %q has no duration fields", pattern))
	}

	days, hours, mins, secs, nsecPart := cascadeInterval(abs, hasDays, hasHours, hasMins, hasSecs)

	var b strings.Builder
	if neg && !hasSignSlot {
		b.WriteByte('-')
	}

	i := 0
	for i < len(pat) {
		c := pat[i]
		switch c {
		case 'd':
			n := countRun(pat, i, 'd')
			b.WriteString(padSpaceInt(int(days), n))
			i += n
		case 'h':
			n := countRun(pat, i, 'h')
			b.WriteString(padZeroInt(int(hours), max(n, 2)))
			i += n
		case 'm':
			n := countRun(pat, i, 'm')
			b.WriteString(padZeroInt(int(mins), max(n, 2)))
			i += n
		case 's':
			secRun, fracRun := scanSecondsAndFrac(pat, i)
			b.WriteString(padZeroInt(int(secs), max(secRun, 2)))
			i += secRun
			if fracRun > 0 {
				b.WriteByte('.')
				b.WriteString(fracDigitsFromNsec(int(nsecPart), fracRun))
				i += 1 + fracRun
			}
		case '+':
			if neg {
				b.WriteByte('-')
			} else {
				b.WriteByte(' ')
			}
			i++
		case '\'':
			lit, next := scanQuotedLiteral(pat, i)
			b.WriteString(lit)
			i = next
		default:
			if unicode.IsLetter(rune(c)) {
				panic(fmt.Errorf("dt::fmt Interval: unknown token starting at %q", pattern[i:]))
			}
			b.WriteByte(c)
			i++
		}
	}
	return b.String()
}

func cascadeInterval(abs time.Duration, hasDays, hasHours, hasMins, hasSecs bool) (days, hours, mins, secs, nsecPart int64) {
	rem := abs.Nanoseconds()
	const (
		dayNanos  = int64(24 * time.Hour)
		hourNanos = int64(time.Hour)
		minNanos  = int64(time.Minute)
		secNanos  = int64(time.Second)
	)
	if hasDays {
		days = rem / dayNanos
		rem %= dayNanos
	}
	if hasHours {
		hours = rem / hourNanos
		rem %= hourNanos
	}
	if hasMins {
		mins = rem / minNanos
		rem %= minNanos
	}
	if hasSecs {
		secs = rem / secNanos
		nsecPart = rem % secNanos
	}
	return
}

func countRun(s string, i int, ch byte) int {
	n := 0
	for i+n < len(s) && s[i+n] == ch {
		n++
	}
	return n
}

// scanSecondsAndFrac: at a run of 's', detect optional '.' + 's' fraction.
// Returns (secondsRunLen, fracRunLen). fracRunLen==0 means no fraction; caller advances by secRun only.
// When fracRunLen>0, total span is secRun + 1 + fracRun.
func scanSecondsAndFrac(s string, i int) (secRun, fracRun int) {
	secRun = countRun(s, i, 's')
	j := i + secRun
	if j < len(s) && s[j] == '.' && j+1 < len(s) && s[j+1] == 's' {
		fracRun = countRun(s, j+1, 's')
	}
	return secRun, fracRun
}

func scanQuotedLiteral(s string, i int) (lit string, next int) {
	// pattern[i] == '\''
	i++
	var b strings.Builder
	for i < len(s) {
		if s[i] == '\'' {
			if i+1 < len(s) && s[i+1] == '\'' {
				b.WriteByte('\'')
				i += 2
				continue
			}
			return b.String(), i + 1
		}
		b.WriteByte(s[i])
		i++
	}
	panic(fmt.Errorf("dt::fmt: unterminated quoted literal"))
}

func registerDateTimeFormatConstants(t *Type) {
	for _, c := range []struct {
		name, pat string
	}{
		{"iso", "yyyy-MM-ddTHH:mm:ss.sssZ"},
		{"date", "yyyy-MM-dd"},
		{"time12", "hh:mm:ss A"},
		{"time24", "HH:mm:ss"},
		{"dateTime", "yyyy-MM-dd HH:mm:ss"},
		{"longDate", "EEEE, MMMM d, yyyy"},
		{"shortDate", "MMM d, yyyy"},
	} {
		name, pat := c.name, c.pat
		t.registerClassProperty(name, "String", func(_ Any) Any {
			return String(pat)
		}, nil)
	}
}

func registerIntervalFormatConstants(t *Type) {
	for _, c := range []struct {
		name, pat string
	}{
		{"clock", "hh:mm:ss"},
		{"clockMs", "hh:mm:ss.sss"},
		{"daysClock", "d hh:mm:ss"},
		{"signedDaysClock", "+d hh:mm:ss"},
	} {
		name, pat := c.name, c.pat
		t.registerClassProperty(name, "String", func(_ Any) Any {
			return String(pat)
		}, nil)
	}
}
