# Compound Builder and Filler Operators

Final provides a concise, convenient set of operators to build or fill many of the most-used collections and objects. The operators work interactively in the interpreter, or can be compiled into code. There are three families of operators, each starts with an opener operator, followed by zero or more elements, then a close operator. Close operators can have variants that control the exact type of the collection from the family that is rendered or filled.

- `(` opens a List or Set builder
- `{` opens an Object or Map builder
- `[` opens a Vec or Matrix *filler* (note: not a builder)

When you enter the opening operator, Final marks the location on DS. The closing operator then gathers up all the items placed on DS after the mark, and puts them into the collection. Easy-peasy!

The operators are best explained by example.

## List and Set Builders

There are several List terminators:

- `)` builds a List; the elements can be any Final type
- `s)` builds a StringList; the elements must be of String type
- `i)` builds an IntList; the elements must be of Int type
- `b)` builds a ByteList; the element must be of Int type, 0-255

And two Set terminators:

- `set)` builds a Set; the elements can be any **comparable** type: numerics or String
- `sset)` builds a StringSet; the elements must be of String type

```forth
// ( -- l:List )
( 1 2 "buckle" )

// ( -- sl:StringList )
( "my" "shoe" s)

// ( -- il:IntList )
( 1 2 3 4 5 i)

// ( -- bl:ByteList )
( 0 1 2 255 b)
```

(The set operators behave identically, but produce the requested set type.)

## Object Builders

The Object operators build a new DynamicObject, or a new StaticObject created from a supplied DynamicObject prototype, or a StringMap. The items must be in pairs: the first element is the value, the second element is the field name.

```forth
// ( -- do:DynamicObject )
{
  1 "one"
  2 "two"
}

// Creates a StaticObject minted from proto
// In this example, we use the previous DO as the prototype
// The field names used must be defined in the DO
// ( proto:DynamicObject -- so:StaticObject)
{ 
  1.0 "one"
  2.0 "two"
so}

// Creates a StringMap
// ( -- sm:StringMap )
{ 
  1 "one"
  2 "two"
sm}
```

## Vec and Matrix Fillers

Unlike the `(` and `{` operators, the `[` operators fill an instance you supply; they do not create a new instance. The pattern of use is creating the instance with the `new` factory operator, then using the Paren operators to provide the values. Vec values are provided in 0-based order; Matrix values must be supplied in row-major order, i.e. for a 2D Matrix, the first row supplies col0 to col(cols-1), etc.

NOTES:

- every value based on the length of the Vec/Mat must be supplied, otherwise the closer will fault;
- the values must be compatible with the cell type of the matrix; no numeric coercion is done.

```forth
// fill a Mat2f Matrix
// ( mat:Mat2f )
2 3 Mat2f.new^
[ 
  0. 1. 2.
  3. 4. 5.
]

// fill a Vec
// ( vec:Vec )
3 VecNf.new^
[ 1. 2. 3. ]
```

Full Vec/Mat API, operators, and OLS: [Linear Algebra](../la/index.md).

## Nesting Operators

You can nest operators to build complex compound objects. There is no practical limit to the nesting level. The rules are:

- the operators must be in family pairs
- you cannot span inner closers across outer opener boundaries eg. no ( 1 "two" { 3 "three" ) }
- the inner collection you build must be a valid type for the outer collection you are building


## Word Invocations while building/filling

You are allowed to invoke (interpret) or compile (compiling) a word inside unclosed Paren blocks. This can be useful to factor repeated patterns, perhaps with a computation, into a worker word.
 ```
: gen3 0 2 range foreach i /foreach ;

( -2 -1 gen3 ) // adds 0 1 2
 ```

## Tips, Cautions, and "Gotchas"

### Tips

You can use code inside the operators, and this is often useful. For example, to get the value of the "num" field from an object on xs you can do this:

```forth
( ... x:.num@ ... )
```

You can do computations inside the operators. For example, to fetch two values and add them, putting the sum on a list slot, you can do this:

```forth
( ... x:.num@ y:.num@ + ... )
```



### Cautions and Gotchas

Final provides comprehensive nesting, semantic, and closing checking on block control structures inside Colon words. But in order to retain the flexibility of interactive as well as compiled operation of Paren operators, Final does almost no static checking of the nesting/closing rules of the Paren operators. Therefore, it is incumbent on the developer or agent to take special care in use of the operators.

In general, Final doesn't care how or how much you mutate DS during the build, **as long as you do not drop any values that were on DS after the opening operator.** Doing so will have undefined effects, and will likely result in a a stack fault. eg.

```forth
// ( val:a -- ? ) will likely fault
( drop ... )
```

The operators only work on DS items added **after** the opening operator; they do not include any values already on DS. If you want to include eg. DS.tos, you could do this:

```forth
// ( myVal:a -- myVal:a l:List )
( peek ... )
```

If you wanted the top 2 values, in order you could do this:

```forth
// ( val0:a val1:a -- val0:a val1:a l:List )
( over over  ... )
```

Any more than that, requires moving eg. to an aux stack.
