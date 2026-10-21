// Type coercion — panic forms
//
// Delegate to the modal ? axioms. On success leave the converted value
// (flag dropped). On failure, error with a clear message.
// Type probes (Int?, Bool?, …) are intentional flag-only words — no
// non-? wrappers here.

// Representation casts (compatible primitives; no string parsing)

: >Bool
  >Bool? not if
    "cannot coerce to Bool" error
  then
;

: >Int
  >Int? not if
    "cannot coerce to Int" error
  then
;

: >Uint
  >Uint? not if
    "cannot coerce to Uint" error
  then
;

: >Float
  >Float? not if
    "cannot coerce to Float" error
  then
;

// Semantic casts (representation + string parsing where applicable)

: >>Null
  >>Null? not if
    "cannot coerce to Null" error
  then
;

: >>Bool
  >>Bool? not if
    "cannot coerce to Bool" error
  then
;

: >>Int
  >>Int? not if
    "cannot coerce to Int" error
  then
;

: >>Uint
  >>Uint? not if
    "cannot coerce to Uint" error
  then
;

: >>Float
  >>Float? not if
    "cannot coerce to Float" error
  then
;

: >>String
  >>String? not if
    "cannot coerce to String" error
  then
;
