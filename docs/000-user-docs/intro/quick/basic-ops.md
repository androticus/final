# Basic Operators

Final has the basic set of operators that most languages provide, for things like arithmetic, boolean logic, bitwise operations, compares, and so on. Each family is briefly described below, with the operators. Type `help <op>` in the prompt to see a description of the operator, and its stack effects.

The type coercion/promotion rules are also listed for each family. A decision has been made in Final to not try to coerce/promote arithmetic operations mixing Int and Uint. The reason is that there are many cases where coercion cannot be inferred. The user must decide how the two values are to be interpreted, and apply the `>Int` or `>Uint` bitwise coercion operator that makes sense in the context. Mixing the two types in many operators causes a runtime error. If you are unsure, try mixed values first in the interpreter, before including it in your code.

Try out some of the operators in each category in the UI!

## Arithmetic

### Dyadic

`+` `-` `*` `/` `%` `modrem`

### Monadic

`abs` (`Uint` is nop)<br>`neg` (`Uint` panics)

**Promotion / coercion**

* Same type → result is that type (`Int`/`Int`, `Uint`/`Uint`, `Float`/`Float`).
* Any operand is `Float` → the other is converted to `Float`; result is `Float`.
* Mixed `Int` and `Uint` → **runtime error**. Cast one side with `>Int` or `>Uint` first.
* `%` and `modrem` require matching integral types (`Int`/`Int` or `Uint`/`Uint`); mixed `Int`/`Uint` → **runtime error**.

## Boolean

### Dyadic

`&&` `||`

### Monadic

`not`

**Promotion / coercion**

* Operands must be **Truthy**: `Bool`, `Int`, `Uint`, `Float`, or `String`.
* Numeric zero and empty `String` are false; anything else in that set is true; `Bool` is itself.
* Other types (lists, maps, objects, …) → **runtime error**.

## Bitwise

### Dyadic

`&` `|` `xor` `bitset` `bitclear` `bittoggle`<br>
`<<` -- logical shift left<br>
`>>` -- arithmetic shift right<br>
`rotl` `rotr` -- rotate left/right

### Monadic

`bitnot`

**Promotion / coercion**

* Operands must be integral (`Int` or `Uint`); non-integral → **runtime error**.
* Dyadic `&` `|` `xor`: both the same type → that type; otherwise result is `Int` (any `Uint` is reinterpreted as `Int`). Never auto-promotes `Int` → `Uint`.
* `bitnot`: result matches the operand’s type.
* Flag mutators (`bitset` `bitclear` `bittoggle`): result matches the flags value being updated.
* Shifts and rotates: result matches the value being shifted; the count may be `Int` or `Uint` (non-negative). Counts are automatically capped by the runtime to 64 bits.

## Equality and Comparison

### Dyadic

`==` `<>` `<` `<=` `>` `>=`

### Monadic

`==0` `<>0` `<0` `<=0` `>0` `>=0`

**Promotion / coercion**

* Result is always `Bool`.
* Both operands must be in the **same family**, or → **runtime error**:
  * **String** — both `String` (ordering is Unicode byte ordering, see: [Strings](../strings/index.md).
  * **Numeric** — both numeric (`Int` / `Uint` / `Float`), with the same promotion rules as arithmetic (including: mixed `Int`/`Uint` → **runtime error**).
  * **Truthy** — both Truthy, but not both Numeric and not both String (each coerced to Bool, then compared).
* Examples: `1 true ==` is fine (Truthy); `1.0 1u ==` is Numeric; `"a" 1 ==` → **runtime error**.
* Zero-compare forms push a typed `0` matching the value (`Int`/`Uint`/`Float`), then compare — so `Uint` stays unsigned.

## Casts

Use these when you need an explicit type change or a membership test. The `?` forms leave a success flag instead of raising an error; the forms without `?` raise an error on failure. Details: `help` on each word.

### Type probes

`Null?` `Bool?` `Int?` `Uint?` `Float?` `String?`

Exact runtime type only — no conversion. Leave a `Bool`. ( value -- is:b )

### Representation casts

`>Bool` `>Int` `>Uint` `>Float`  
`>Bool?` `>Int?` `>Uint?` `>Float?`

Compatible primitives only — **no** string parsing.

### Semantic casts

`>>Null` `>>Bool` `>>Int` `>>Uint` `>>Float` `>>String`  
`>>Null?` `>>Bool?` `>>Int?` `>>Uint?` `>>Float?` `>>String?`

Like representation casts, plus string parsing where it applies (e.g. `"42" >>Int`).

## Trancendentals and Related

Final provides a set of trancendental operators and constants, such as `sin`, `cos`, `Math.pi` etc. The operators take any type of numeric value, and typically return a Float. Enter `list-math` for the list, then `help` on each operator for its description and stack effects.

`Float` provides some useful constants like Float.NaN and tests like Float.isNaN^ and Float.isInf^ to help deal with the thorny corner cases of IEEE-754 NaNs and Infs. use `help Float` to list them.

### Monadic

| Op | Description |
| -- | ----------- |
| `acos` | arccosine; result in radians |
| `acosh` | inverse hyperbolic cosine |
| `asin` | arcsine; result in radians |
| `asinh` | inverse hyperbolic sine |
| `atan` | arctangent; result in radians |
| `atanh` | inverse hyperbolic tangent |
| `cbrt` | cube root |
| `ceil` | smallest integer Float not less than the argument |
| `cos` | cosine of radians argument |
| `cosh` | hyperbolic cosine |
| `exp` | e raised to the power of the argument |
| `floor` | largest integer Float not greater than the argument |
| `log` | natural logarithm (base e) |
| `log10` | base-10 logarithm |
| `log2` | base-2 logarithm |
| `round` | nearest integer Float (half away from zero) |
| `sin` | sine of radians argument |
| `sinh` | hyperbolic sine |
| `sqrt` | square root |
| `tan` | tangent of radians argument |
| `tanh` | hyperbolic tangent |
| `trunc` | truncate toward zero to an integer Float |

### Dyadic

| Op | Description |
| -- | ----------- |
| `pow` | raise x to the power y (`x:n y:n -- f`) |

### Contants and Helpers on Float

| Const/Op | Stack | Description |
| --- | --- | --- |
.maxFloat | ( -- f ) | Largest finite Float
.minNonZero | ( -- f ) | Smallest positivive non-zero
.+Inf | ( -- f ) |  IEEE-754 positive Infinity sentinel (1)
.+Inf | ( -- f ) |  IEEE-754 negative Infinity sentinel (1)
.NaN | ( -- NaN:f ) | IEEE-754 Not A Number sentinel (2)
.isInf | ( a:n -- b ) | true if the value is an IEEE-754 +Inf or -Inf value
.isNaN | ( a:n -- b ) | true if the value is an IEEE-754 NaN

Notes:
1. Can be compared to valid Float values; use .isInf to test for +-Inf;
1. .NaN is only for producing the NaN sentinel, if needed; use .isNaN to test for NaN; all float operations using NaN as an input return NaN; NaN <> NaN in comparisons

### Contants on Math

| Const | Stack | Description |
| --- | --- | --- |
.e | ( -- f ) | Base of natural logarithms (e)
.pi | ( -- f ) | Pi
.ln2 | ( -- f ) |   Natural log of 2
.ln10 | ( -- f ) |  Natural log of 10
