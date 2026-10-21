// Regex — pattern matching (Go RE2)

/*
A compiled regular expression.
Patterns use Go's RE2 syntax (no backreferences or lookaround).
Build once with Regex.new^ and reuse on many subjects,
or use the class one-shot match / find for a single check.
Bad patterns panic (programming error); use isPatternValid?
in the interpreter while drafting a pattern string.
Subject strings come first on the stack; the pattern (or
replacement) is TOS — same "subject then tool" order as
other Final APIs.

Class one-shots (no Regex left on the stack):
  match — ( s pattern:s -- b )
  find  — ( s pattern:s -- found:s true | false )
eg: `` "42" `^\d+$` Regex.match^ ``
eg: `` "hello 42" `\d+` Regex.find^ ``
*/
:doc Regex

/*
Compile a pattern into a Regex.
Panics if the pattern is invalid.
eg: `` `\d+` Regex.new^ ``
( pattern:s -- re:Regex )
*/
:doc Regex#new

/*
True if pattern would compile successfully.
Does not build a Regex — handy in the REPL while
checking a raw string before putting it in source.
( pattern:s -- b )
*/
:doc Regex#isPatternValid?

/*
Escape regexp metacharacters so s matches as a literal
when embedded in a larger pattern.
( s -- q:s )
*/
:doc Regex#quoteMeta

/*
The pattern text of this Regex.
( -- pattern:s )
*/
:doc Regex#pattern

/*
Number of capturing groups (not counting the whole match).
( -- n:i )
*/
:doc Regex#numGroups

/*
Keys for match groups in Go/standard order: index 0 is the
whole match ("0"), then each capturing group.
Named groups use their name; unnamed use "1", "2", ….
Same strings matchObject uses as field names.
Length is .numGroups@ + 1.
Read a group by key with @{} (any string, including "0"),
or by position with @() and an Int. .0@ is not valid.
( -- names:StringList )
*/
:doc Regex#groupNames

/*
True if this Regex matches s.
Also available as class one-shot Regex.match^
(see Regex for that stack effect).
( s -- b )
*/
:doc Regex#match

/*
First match substring in s, or false if none. Modal.
A successful empty match is "" true (not false).
Also available as class one-shot Regex.find^
(see Regex for that stack effect).
( s -- found:s true | false )
*/
:doc Regex#find

/*
All non-overlapping matches in s.
Modal: false when there are none (no empty list).
( s -- parts:StringList true | false )
*/
:doc Regex#findAll

/*
At most n matches in s (n < 0 means all).
Modal: false when there are none.
( s n:i -- parts:StringList true | false )
*/
:doc Regex#findAllN

/*
Byte start and end indexes of the first match in s.
Modal: false when there is no match.
( s -- start:i end:i true | false )
*/
:doc Regex#findIndex

/*
Split s on matches of this Regex.
Always returns a StringList (even when the separator
never appears — then a one-element list of s).
( s -- parts:StringList )
*/
:doc Regex#split

/*
Split s on matches, at most n parts.
( s n:i -- parts:StringList )
*/
:doc Regex#splitN

/*
On match, a DynamicObject of the full match and captures.
Modal: false when s does not match.

Uses standard regex numbering (same as Go / Perl / PCRE):
field "0" is the whole match span; "1", "2", … are
capturing groups (or a (?P<name>…) name when given).
Fields are created in that order, so k @() is Go group k.

Access:
  named groups — .name@ (Final ident) or "name" @{}@
  any group key — @{} with the string key, including
    ordinal forms "0", "1", … (@{} allows any string;
    it is not limited to Final field idents)
  by position — @() with an Int (0 @() = whole match)
  .0@ is invalid: "0" is not a Final identifier

Empty group text is still present as "".
A pattern with no capturing groups still has "0" on success.
( s -- o:DynamicObject true | false )
*/
:doc Regex#matchObject

/*
Replace every match in s with repl.
repl may use Go template escapes ($1, $name, $0, …).
Subject deep, replacement TOS.
( s repl:s -- out:s )
*/
:doc Regex#replaceAll

/*
Replace every match in s with repl as a literal
(no $ expansion in repl).
Subject deep, replacement TOS.
( s repl:s -- out:s )
*/
:doc Regex#replaceAllLiteral
