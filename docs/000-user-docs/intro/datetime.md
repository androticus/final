# DateTime and Interval

The [Quick Tour](quick/basic-types.md) deferred these two basic types here. **DateTime** is a calendar date and/or clock time; **Interval** is a length of time (a duration). Both are **immutable** sealed values.

They are **not** part of the numeric family. Ordinary `+` `-` `*` `/` and compares do **not** apply — use the **`dt::`** family below. Invalid operand pairs panic.

Related material already covered elsewhere:

* **`$`** can print a DateTime ([stdin/out Streams](streams/stdio.md)).
* Json rendering emits DateTime as an RFC 3339 string ([Json Rendering](json/rendering.md)).
* Pattern formatting for DateTime and Interval is covered in [Formatting with `dt::fmt`](#formatting-with-dtfmt) below; general `%` message formatting is in [Output Formatting](out-formatting.md).

Type `list-datetime` for all the DateTime/Interval related words, then e.g. `help DateTime`, `help Interval`, `help dt::+`, or `help dt::fmt` in the UI for the detailed documentation and stack effects.

---

## DateTime

Factories create values in **UTC**. **`DateTime.now` returns UTC** as well (the current instant).

Construction takes a **`List` of `Int` components** — typically written with `( … )` — rather than a long stack of loose numbers.

| Kind | Name | Effect | Description |
| ---- | ---- | ------ | ----------- |
| class | `now` | `( -- dt:DateTime )` | current instant (**UTC**) |
| class | `newDate` | `( parts:List -- dt:DateTime )` | year, month (1–12), day → midnight UTC |
| class | `newTime` | `( parts:List -- dt:DateTime )` | hour, minute, second, optional ms → clock-only |
| class | `newDateTime` | `( parts:List -- dt:DateTime )` | year…second, optional ms → full instant UTC |
| prop | `year` | Int | calendar year |
| prop | `month` | Int | month 1–12 |
| prop | `day` | Int | day of month |
| prop | `hour` | Int | hour 0–23 |
| prop | `minute` | Int | minute 0–59 |
| prop | `second` | Int | second 0–59 |
| prop | `nanosecond` | Int | nanoseconds within the second |
| prop | `isZero` | Bool | `true` if the zero DateTime |
| method | `since` | `( -- iv:Interval )` | elapsed Interval from this instant until now |

```forth
( 2026 7 4 ) DateTime.newDate^ $ nl
( 14 30 0 ) DateTime.newTime^ $ nl
( 2026 7 4 14 30 0 250 ) DateTime.newDateTime^ $ nl
DateTime.now^ $ nl
```

Park a DateTime on **ts** when you need several property reads:

```forth
DateTime.now^ >ts
t:.year@ $ sp t:.month@ $ sp t:.day@ $ nl
t:drop
```

---

## Interval

An Interval is a **duration**. Build one from unit class constants (then scale with `dt::*`), or from a List of clock-length components.

| Kind | Name | Effect | Description |
| ---- | ---- | ------ | ----------- |
| class | `nanosecond` | `( -- iv:Interval )` | one nanosecond |
| class | `microsecond` | `( -- iv:Interval )` | one microsecond |
| class | `millisecond` | `( -- iv:Interval )` | one millisecond |
| class | `second` | `( -- iv:Interval )` | one second |
| class | `minute` | `( -- iv:Interval )` | one minute |
| class | `hour` | `( -- iv:Interval )` | one hour |
| class | `newInterval` | `( parts:List -- iv:Interval )` | hour, minute, second, optional ms |
| prop | `nanoseconds` | Int | total nanoseconds |
| prop | `microseconds` | Int | total microseconds (truncated) |
| prop | `milliseconds` | Int | total milliseconds (truncated) |
| prop | `seconds` | Float | total seconds, including fraction |
| prop | `minutes` | Float | total minutes, including fraction |
| prop | `hours` | Float | total hours, including fraction |
| prop | `days` | Float | total days, including fraction (`hours / 24`) |

```forth
Interval.second $ nl
5 Interval.second dt::* $ nl          // 5s
5 Interval.second dt::* .milliseconds@ $ nl  // 5000
( 1 30 0 ) Interval.newInterval^ $ nl // 1h30m0s
( 0 0 5 500 ) Interval.newInterval^ $ nl
```

Unit constants are class properties (same style as `Math.pi` / `Int.maxInt` — see [Properties](objects/properties.md)).

---

## `dt::` operators

All take operands on **DS**. Same-kind compares only (two DateTimes, or two Intervals). Mixed kinds and unsupported arithmetic pairs **panic**.

| Word | Effect | Valid pairs / notes |
| ---- | ------ | ------------------- |
| `dt::+` | `( a b -- result )` | `DateTime`+`Interval` → `DateTime` (either order); `Interval`+`Interval` → `Interval` |
| `dt::-` | `( a b -- result )` | `DateTime`−`Interval` → `DateTime`; `DateTime`−`DateTime` → `Interval`; `Interval`−`Interval` → `Interval` |
| `dt::*` | `( a b -- iv:Interval )` | `Int`×`Interval` (either order) |
| `dt::/` | `( iv:Interval n:i -- iv:Interval )` | divide by positive Int (truncates toward zero); panics if `n <= 0` |
| `dt::compare` | `( a b -- c:i )` | −1 / 0 / 1 (before / equal / after) |
| `dt::<` `dt::<=` `dt::>` `dt::>=` `dt::==` `dt::<>` | `( a b -- c:b )` | ordering / equality; both either `DateTime` or `Interval` |
| `dt::round` | `( iv:Interval roundTo:Interval -- rounded:Interval )` | nearest multiple of `roundTo` |
| `dt::fmt` | `( val:DateTime\|Interval fmt:s -- result:s )` | pattern format (see below) |

---

## Formatting with `dt::fmt`

[`$`](streams/stdio.md) prints a DateTime or Interval’s **natural** string. For a composed layout — ISO, a prose date, a stopwatch clock — use **`dt::fmt`** with a **pattern String**.

```text
dt::fmt  // ( val:DateTime|Interval  fmt:s -- result:s )
```

- One word for both types; the lexicon is chosen from **`val`’s type**.
- `fmt` is always a pattern String (or a class constant that holds one). Wrong tokens for that type panic.
- Patterns use letter-run codes familiar from app languages (`yyyy-MM-dd`, `HH:mm:ss`).

```forth
DateTime.now^ "yyyy-MM-ddTHH:mm:ss.sssZ" dt::fmt $ nl
DateTime.now^ DateTime.iso dt::fmt $ nl
5 Interval.minute dt::* "HH:mm:ss" dt::fmt $ nl
Interval.second Interval.clock dt::fmt $ nl    // 00:00:01
```

General message formatting (`fmt` / `%d` / `%s`) is separate — see [Output Formatting](out-formatting.md).

### DateTime patterns

**Case-sensitive.** `MM` ≠ `mm`; `HH` ≠ `hh`; `a` ≠ `A`. Values are formatted in **UTC**.

| Token / fragment | Meaning | Example |
| ---------------- | ------- | ------- |
| `yyyy` | 4-digit year | `2026` |
| `yy` | 2-digit year | `26` |
| `M` | month 1–12 (unpadded) | `7` |
| `MM` | month 01–12 | `07` |
| `MMM` | short month name (English) | `Jul` |
| `MMMM` | long month name (English) | `July` |
| `d` | day of month (unpadded) | `4` |
| `dd` | day of month 01–31 | `04` |
| `EEE` | short weekday (English) | `Sat` |
| `EEEE` | long weekday (English) | `Saturday` |
| `HH` | hour 00–23 (24-hour) | `14` |
| `hh` | hour 01–12 (12-hour) | `02` |
| `mm` | minute | `30` |
| `ss` | whole seconds | `05` |
| `ss.sss` … | seconds + fraction (`s` run = digit width; ms / µs / ns) | `05.250` |
| `a` | meridiem lower | `am` / `pm` |
| `A` | meridiem upper | `AM` / `PM` |

**Hours / meridiem:** `hh` does **not** add AM/PM by itself — include `a` or `A` when you want it.

**Fractional seconds:** after `ss`, an optional `.` plus a run of `s` sets the digit width (`ss.sss` = milliseconds, up to nine `s` for nanoseconds). A `.` elsewhere in the pattern is a literal.

**Literals (no quotes needed):** space, `T`, `Z`, and usual separators (`-`, `:`, `/`, `,`).  
`Z` is the letter Z (UTC designator in ISO-style patterns), not a computed offset.

**Quoted literals:** wrap text in `'…'` when you need a letter or `.` that would otherwise be a token (e.g. `ss'.'` for a literal dot after seconds with no fraction). `''` inside quotes is a single `'`.

```forth
DateTime.now^ "EEEE, MMMM d, yyyy" dt::fmt $ nl
DateTime.now^ "hh:mm A" dt::fmt $ nl
DateTime.now^ "yyyy-MM-ddTHH:mm:ss.sssZ" dt::fmt $ nl
```

#### DateTime format constants

Class properties (same style as `Math.pi`) — each pushes a pattern String:

| Constant | Pattern |
| -------- | ------- |
| `DateTime.iso` | `yyyy-MM-ddTHH:mm:ss.sssZ` |
| `DateTime.date` | `yyyy-MM-dd` |
| `DateTime.time12` | `hh:mm:ss A` |
| `DateTime.time24` | `HH:mm:ss` |
| `DateTime.dateTime` | `yyyy-MM-dd HH:mm:ss` |
| `DateTime.longDate` | `EEEE, MMMM d, yyyy` |
| `DateTime.shortDate` | `MMM d, yyyy` |

```forth
DateTime.now^ DateTime.iso dt::fmt $ nl
DateTime.now^ DateTime.longDate dt::fmt $ nl
```

### Interval patterns

**Case-insensitive** (the runtime lowercases the pattern before parsing). You may write `HH` or `hh` — same meaning. No meridiem, no calendar month/year names, no weekday tokens.

| Token (after lower) | Meaning | Padding |
| ------------------- | ------- | ------- |
| `d` / `dd` / … | **days** (24h chunks only — not months/years) | width = run length; **space**-padded on the left (`dd` + 5 → `" 5"`; `d` → `"5"`) |
| `h` / `hh` / … | **hours** | min width 2, **zero**-padded; **widens** if the value needs more digits |
| `m` / `mm` / … | minutes | same zero-pad / widen rule |
| `s` / `ss` / … | seconds | same; optional `.` + `s`… fraction (same idea as DateTime) |
| `+` | **sign placeholder** | ≥0 → space; &lt;0 → `-` (column-aligned) |
| space, `:`, other non-letters | literals | — |

**Cascade:** only units that **appear in the pattern** participate. Among those, the **largest** unit absorbs overflow; smaller units are remainders. Absent units are not filled in silently.

| Duration | Pattern | Output |
| -------- | ------- | ------ |
| 50h | `hh:mm` | `50:00` |
| 50h | `d hh:mm` | `2 02:00` |
| 50h | `dd hh:mm` | ` 2 02:00` |
| 50h | `mm` | `3000` |

**Sign:**

- Format the **absolute** duration with the numeric fields.
- If the pattern contains `+`, emit space when `val ≥ 0`, `-` when `val < 0`, at that position.
- If there is **no** `+` and the value is negative, prefix a single `-`.
- Zero is non-negative → space when `+` is present.

```forth
( 1 30 0 ) Interval.newInterval^ "hh:mm:ss" dt::fmt $ nl     // 01:30:00
50 Interval.hour dt::* "d hh:mm" dt::fmt $ nl                // 2 02:00
50 Interval.hour dt::* "dd hh:mm" dt::fmt $ nl               //  2 02:00
1250 Interval.millisecond dt::* "ss.sss" dt::fmt $ nl        // 01.250

// negative: leading - when no + in pattern
5 Interval.hour dt::* -1 dt::* "d hh:mm:ss" dt::fmt $ nl     // -0 05:00:00
5 Interval.hour dt::* -1 dt::* "+d hh:mm:ss" dt::fmt $ nl    // -0 05:00:00
```

#### Interval format constants

| Constant | Pattern | Role |
| -------- | ------- | ---- |
| `Interval.clock` | `hh:mm:ss` | stopwatch-style |
| `Interval.clockMs` | `hh:mm:ss.sss` | with milliseconds |
| `Interval.daysClock` | `d hh:mm:ss` | day + clock |
| `Interval.signedDaysClock` | `+d hh:mm:ss` | aligned sign |

```forth
( 1 30 0 ) Interval.newInterval^ Interval.clock dt::fmt $ nl
50 Interval.hour dt::* Interval.daysClock dt::fmt $ nl
```

### Token overlap

The same letters appear in both lexicons with **different** meanings (`hh` is 12-hour on a DateTime, hours-of-duration on an Interval). **`val`’s type** selects which table applies. Do not pass a DateTime ISO pattern to an Interval (or the reverse).

---

## Examples

### Elapsed time

Capture a start instant, do work, then subtract with `dt::-`. From a parked start, `.since^` is the same idea (elapsed until now):

```forth
DateTime.now^ >ts
// … work …
DateTime.now^ ts> dt::- $ nl    // elapsed Interval
```

### Deadline from now

Advance the current instant by a duration, then compare:

```forth
DateTime.now^
5 Interval.minute dt::* dt::+ >ts   // deadline = now + 5m

DateTime.now^ t:peek dt::< $ nl     // true while still before deadline
t:drop
```

### Average of several Intervals

Scale and divide with `dt::*` / `dt::/` (metrics-style averaging):

```forth
( 0 0 10 ) Interval.newInterval^    // 10s
( 0 0 20 ) Interval.newInterval^    // 20s
dt::+
2 dt::/ $ nl                        // 15s
```

---

## Notes

* **Immutability** — `dt::+` / `dt::-` and friends return **new** values; inputs are unchanged.
* **No `DateTime + DateTime`** — elapsed time is `dt::-` (later minus earlier), not addition.
* **Printing** — `$` uses the DateTime natural string (ISO-style for full dates; clock-only for `newTime` values). Interval prints in a compact duration form (e.g. `1h30m0s`). For controlled layouts, use [`dt::fmt`](#formatting-with-dtfmt) and the pattern tables above.
* **Json** — only DateTime participates in `>json` / `json>` as a string; see [Json](json/index.md).
