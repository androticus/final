# Stack Operators

Final has two kinds of stack operators:
* explicit stack selection
* Context Selector (cs) stack selection

As explained in the previous section, cs-aware operators referring to the Data Stack do not use a selector, DS is the default.

Some of these operators have been inherited from Forth, which as explained earlier, often required programmers to engage in mind-numbing stack value juggling; Final's design and coding conventions have greatly reduced the need for such juggling, so don't worry about words like `rot` or `pick` and such--they don't find much use in Final. `dup` `drop` `peek` and the interstack movement operations make up the vast majority of Final's stack repertoire.

The term `tos` means Top of Stack.

## Explicit Stack Selection Operators

These operators explicitly state the stacks involved. All the to/from ones imply DS as the unspecified half. These operators are used frequently.

| Op | Effect | Description |
| -- | ------ | ----------- |
| `>ts` | ( val:a -- ) t( -- val ) | move value from DS to TS |
| `ts>` | ( -- val:a ) t( val -- ) | move value from TS to DS |
| `tdrop` | t( val:a -- ) | discard tos of TS |
| `>xs` | ( val:a -- ) x( -- val ) | move value from DS to XS |
| `xs>` | ( -- val:a ) x( val -- ) | move value from XS to DS |
| `xdrop` | x( val:a -- ) | discard tos of XS |
| `>ys` | ( val:a -- ) y( -- val ) | move value from DS to YS |
| `ys>` | ( -- val:a ) y( val -- ) | move value from YS to DS |
| `ydrop` | y( val:a -- ) | discard tos of YS |
| `>zs` | ( val:a -- ) z( -- val ) | move value from DS to ZS |
| `zs>` | ( -- val:a ) z( val -- ) | move value from ZS to DS |
| `zdrop` | z( val:a -- ) | discard tos of ZS |

## CS-Aware Stack Operators

These operators act on the stack selected by a Context Selector (`t:`, `x:`, `y:`, `z:`, or none for DS).

| Op | Effect | Description |
| -- | ------ | ----------- |
| `dup` | cs( val:a -- val val ) | duplicate the value on tos |
| `dup2` | cs( a:a b:a -- a b a b ) | duplicate the top two values |
| `drop` | cs( val:a -- ) | discard tos |
| `drop2` | cs( a:a b:a -- ) | discard the top two values |
| `swap` | cs( a:a b:a -- b a ) | swap the top two values |
| `over` | cs( a:a b:a -- a b a ) | copy the second value to tos |
| `nip` | cs( a:a b:a -- b ) | discard the second value |
| `tuck` | cs( a:a b:a -- b a b ) | copy tos under the second value |
| `pick` | cs( … +n:i -- … val ) | copy the nth value to tos (0 = tos) |
| `rot` | cs( a:a b:a c:a -- b c a ) | rotate the top three values |
| `peek` | cs( val:a -- val ) ( -- val ) | copy tos of CS onto DS |
| `ds>` | cs( val:a -- ) ( -- val ) | move tos of CS onto DS |
| `>ds` | ( val:a -- ) cs( -- val ) | move tos of DS onto CS |
| `depth` | cs( -- n:i ) | depth of the selected stack |
