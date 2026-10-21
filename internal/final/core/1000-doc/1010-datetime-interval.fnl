// DateTime and Interval

// DateTime

/*
A calendar date and/or clock time.
Values are immutable.
Factories build in UTC.
Use DateTime.now^ for the current instant.
*/
:doc DateTime

/*
Current date and time.
( -- dt:DateTime )
*/
:doc DateTime#now

/*
Build a date at midnight UTC from a List of Ints:
year, month (1-12), day.
eg: `( 2026 7 4 ) DateTime.newDate^`
( parts:List -- dt:DateTime )
*/
:doc DateTime#newDate

/*
Build a clock-only time from a List of Ints:
hour, minute, second, and optional milliseconds.
eg: `( 14 30 0 ) DateTime.newTime^`
eg: `( 14 30 0 500 ) DateTime.newTime^`
( parts:List -- dt:DateTime )
*/
:doc DateTime#newTime

/*
Build a full date and time from a List of Ints:
year, month (1-12), day, hour, minute, second,
and optional milliseconds.
eg: `( 2026 7 4 14 30 0 ) DateTime.newDateTime^`
eg: `( 2026 7 4 14 30 0 250 ) DateTime.newDateTime^`
( parts:List -- dt:DateTime )
*/
:doc DateTime#newDateTime

/*
Year of the date.
( dt:DateTime -- year:i )
*/
:doc DateTime#year

/*
Month of the year, 1 through 12.
( dt:DateTime -- month:i )
*/
:doc DateTime#month

/*
Day of the month.
( dt:DateTime -- day:i )
*/
:doc DateTime#day

/*
Hour of the day, 0 through 23.
( dt:DateTime -- hour:i )
*/
:doc DateTime#hour

/*
Minute of the hour, 0 through 59.
( dt:DateTime -- minute:i )
*/
:doc DateTime#minute

/*
Second of the minute, 0 through 59.
( dt:DateTime -- second:i )
*/
:doc DateTime#second

/*
Nanoseconds within the second.
( dt:DateTime -- nsec:i )
*/
:doc DateTime#nanosecond

/*
True if this is the zero DateTime.
( dt:DateTime -- b )
*/
:doc DateTime#isZero

/*
Elapsed Interval from this DateTime until now.
( dt:DateTime -- iv:Interval )
*/
:doc DateTime#since

/*
ISO 8601 date/time standard.
eg. 2026-08-01T04:21:51.986Z
*/
:doc DateTime#iso

/*
Numeric calendar date.
eg. 2026-08-01
*/
:doc DateTime#date

/*
12-hour clock with AM/PM.
eg. 04:21:51 AM
*/
:doc DateTime#time12

/*
24-hour clock time.
eg. 04:21:51
*/
:doc DateTime#time24

/*
Numeric date and 24-hour time.
eg. 2026-08-01 04:21:51
*/
:doc DateTime#dateTime

/*
Long date format, day/month spelled out.
eg. Saturday, August 1, 2026
*/
:doc DateTime#longDate

/*
Short date with abbreviated month.
eg. Aug 1, 2026
*/
:doc DateTime#shortDate

// Interval

/*
A length of time (a duration), not a clock reading.
Values are immutable.
Unit constants such as Interval.second give one unit.
eg. to create an interval of 1 second:
`1 Interval.second * ( -- it:Interval )
*/
:doc Interval

/*
One nanosecond.
( -- iv:Interval )
*/
:doc Interval#nanosecond

/*
One microsecond.
( -- iv:Interval )
*/
:doc Interval#microsecond

/*
One millisecond.
( -- iv:Interval )
*/
:doc Interval#millisecond

/*
One second.
eg: `Pipe.newTicker^ Interval.second swap pipe::tickerStart`

( -- iv:Interval )
*/
:doc Interval#second

/*
One minute.
( -- iv:Interval )
*/
:doc Interval#minute

/*
One hour.
( -- iv:Interval )
*/
:doc Interval#hour

/*
Total duration as nanoseconds.
*/
:doc Interval#nanoseconds

/*
Total duration as microseconds (truncated).
*/
:doc Interval#microseconds

/*
Total duration as milliseconds (truncated).
*/
:doc Interval#milliseconds

/*
Total duration as seconds, including fraction.
*/
:doc Interval#seconds

/*
Total duration as minutes, including fraction.
*/
:doc Interval#minutes

/*
Total duration as hours, including fraction.
*/
:doc Interval#hours

/*
Total duration as days, including fraction (hours / 24).
*/
:doc Interval#days

/*
Build an Interval from a List of Ints:
hour, minute, second, and optional milliseconds.
eg: `( 1 30 0 ) Interval.newInterval^`
eg: `( 0 0 5 500 ) Interval.newInterval^`
( parts:List -- iv:Interval )
*/
:doc Interval#newInterval

/*
Stopwatch-style clock duration.
eg. 36:30:45
*/
:doc Interval#clock

/*
Clock duration with milliseconds.
eg. 36:30:45.678
*/
:doc Interval#clockMs

/*
Days plus clock duration.
eg. 1 12:30:45
*/
:doc Interval#daysClock

/*
Signed days plus clock; leading space if non-negative.
eg.  1 12:30:45
*/
:doc Interval#signedDaysClock

// DateTime / Interval axioms

/*
Add DateTime and Interval, or two Intervals.
Valid pairs:
  DateTime Interval -- DateTime
  Interval DateTime -- DateTime
  Interval Interval -- Interval
Other pairs panic.
( a b -- result )
*/
:doc dt::+

/*
Subtract Interval from DateTime, DateTime from DateTime,
or Interval from Interval.
Valid pairs:
  DateTime Interval -- DateTime
  DateTime DateTime -- Interval
  Interval Interval -- Interval
Other pairs panic.
( a b -- result )
*/
:doc dt::-

/*
Scale an Interval by an Int (either order).
Valid pairs:
  Int Interval -- Interval
  Interval Int -- Interval
Other pairs panic.
( a b -- iv:Interval )
*/
:doc dt::*

/*
Divide an Interval by a positive Int (truncates toward zero).
Panics if n <= 0 or operands are not Interval then Int.
( iv:Interval n:i -- iv:Interval )
*/
:doc dt::/

/*
Format a DateTime or Interval with a pattern String.
DateTime patterns are case-sensitive (hh vs HH, a vs A).
Interval patterns are case-insensitive.
Literals include space, T, and Z (for ISO).
See DateTime.iso and Interval.clock for curated patterns.
Run examples/format/datetime.fnl .
( val:DateTime|Interval fmt:s -- result:s )
*/
:doc dt::fmt

/*
Compare two DateTimes or two Intervals.
Result is -1, 0, or 1 (before, equal, after).
Operands must be the same kind; mixed pairs panic.
( a b -- c:i )
*/
:doc dt::compare

/*
True if a is before b.
Operands must both be DateTime or both Interval.
( a b -- c:b )
*/
:doc dt::<

/*
True if a is before or equal to b.
Operands must both be DateTime or both Interval.
( a b -- c:b )
*/
:doc dt::<=

/*
True if a is after b.
Operands must both be DateTime or both Interval.
( a b -- c:b )
*/
:doc dt::>

/*
True if a is after or equal to b.
Operands must both be DateTime or both Interval.
( a b -- c:b )
*/
:doc dt::>=

/*
True if a equals b.
Operands must both be DateTime or both Interval.
( a b -- c:b )
*/
:doc dt::==

/*
True if a does not equal b.
Operands must both be DateTime or both Interval.
( a b -- c:b )
*/
:doc dt::<>

/*
Round an Interval to the nearest multiple of an Interval.
eg: round to the nearest second or hour.
( iv:Interval roundTo:Interval -- rounded:Interval )
*/
:doc dt::round
