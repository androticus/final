forget: reset-datetime

/*
DateTime and Interval formatting via dt::fmt.

Patterns use letter-run codes (not Go's 2006 reference layout).
DateTime patterns are case-sensitive (hh vs HH, a vs A).
Interval patterns are case-insensitive (ToLower before parse).

Run (from repo or Final/ install root):
  "examples/datetime.fnl" compileFromFile
*/


DateTime.now^ const: now

/*
Print a DateTime format pattern and its result for `now`.
( fmt:s -- )
*/
: showDt
  dup $ "  →  " $
  now swap dt::fmt $ nl
;

"=== DateTime (now) ===" $ nl
nl

"--- year ---" $ nl
"yyyy" showDt
"yy" showDt
nl

"--- month ---" $ nl
"M" showDt
"MM" showDt
"MMM" showDt
"MMMM" showDt
nl

"--- day of month ---" $ nl
"d" showDt
"dd" showDt
nl

"--- weekday ---" $ nl
"EEE" showDt
"EEEE" showDt
nl

"--- hours (24h / 12h) ---" $ nl
" 0..23 " $ "HH" showDt
"12..11 " $ "hh" showDt
nl

"--- meridiem ---" $ nl
"hh:mm A" showDt
"hh:mm a" showDt
"A hh:mm" showDt
nl

"--- minutes / seconds ---" $ nl
"mm" showDt
"ss" showDt
nl

"--- fractional seconds ---" $ nl
"ss.sss" showDt
"ss.ssssss" showDt
"ss.sssssssss" showDt
nl

"--- combined / ISO literals (space, T, Z) ---" $ nl
"yyyy-MM-dd" showDt
"yyyy-MM-dd HH:mm:ss" showDt
"yyyy-MM-ddTHH:mm:ssZ" showDt
"yyyy-MM-ddTHH:mm:ss.sssZ" showDt
"EEEE, MMMM d, yyyy" showDt
"MMM d, yyyy" showDt
nl

"--- DateTime pattern constants ---" $ nl
"DateTime.iso       " $ DateTime.iso showDt
"DateTime.date      " $ DateTime.date showDt
"DateTime.time12    " $ DateTime.time12 showDt
"DateTime.time24    " $ DateTime.time24 showDt
"DateTime.dateTime  " $ DateTime.dateTime showDt
"DateTime.longDate  " $ DateTime.longDate showDt
"DateTime.shortDate " $ DateTime.shortDate showDt
nl

/*
1 day + 12h + 30m + 45s + 678901234 ns
*/
24 Interval.hour dt::*
12 Interval.hour dt::* dt::+
30 Interval.minute dt::* dt::+
45 Interval.second dt::* dt::+
678901234 Interval.nanosecond dt::* dt::+
const: interval

/*
Print an Interval format pattern and its result for `interval`.
( fmt:s -- )
*/
: showIv
  dup $ "  →  " $
  interval swap dt::fmt $ nl
;

"=== Interval (1d 12h 30m 45s + 678901234ns) ===" $ nl
nl

"--- days (space-padded) ---" $ nl
"d" showIv
"dd" showIv
"ddd" showIv
nl

"--- clock fields ---" $ nl
"hh" showIv
"mm" showIv
"ss" showIv
nl

"--- cascade (largest unit in pattern absorbs overflow) ---" $ nl
"hh:mm:ss" showIv
"d hh:mm:ss" showIv
"dd hh:mm:ss" showIv
"mm" showIv
nl

"--- fractional seconds ---" $ nl
"ss.sss" showIv
"ss.ssssss" showIv
"ss.sssssssss" showIv
"hh:mm:ss.sss" showIv
"d hh:mm:ss.sssssssss" showIv
nl

"--- sign placeholder (+) → space if ≥0, - if <0 ---" $ nl
"+d hh:mm:ss" showIv
"            →  1 12:30:45\n" $
"+dd hh:mm:ss.sss" showIv
"             →   1 12:30:45\n" $
nl

"--- Interval pattern constants ---" $ nl
"Interval.clock           " $ Interval.clock showIv
"Interval.clockMs         " $ Interval.clockMs showIv
"Interval.daysClock       " $ Interval.daysClock showIv
"Interval.signedDaysClock " $ Interval.signedDaysClock showIv
nl

"--- negative interval (leading - when no + in pattern) ---" $ nl
interval -1 dt::* const: negInterval
: showNeg
  dup $ "  →  " $
  negInterval swap dt::fmt $ nl
;
"d hh:mm:ss" showNeg
"+d hh:mm:ss" showNeg
nl
