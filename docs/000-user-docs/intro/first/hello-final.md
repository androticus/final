# Hello Final!

When you first launch `final` you are in an interactive interpreter. You can enter numbers and strings, operators, or words, and Final will execute it according to its interpretation rules (which are covered later in [Interpreter](../quick/interpreter.md) ).

So one of the first things typically done when introducing a new language is the venerable "Hello World!" example. Let's do it interactively first. Try the following:

```forth
> "Hello from Final!\n" $

Hello from Final!

```

Your string literal got pushed on DS (see below), then `$` "print" printed it out. Easy, peasy!

Let's try a simple arithmetic example. Enter the following, and observe the result:

```forth
> 1 2 + $

3
```

A central construct in Final is the Data Stack (DS). It has no counterpart in conventional languages. Conventional languages use named parameters in functions to pass parameters, and sometimes global variables.

Final uses the DS to hold intermediate operands, pass operands to Final's counterpart of a function, called a `Word`, and return results. When you write a literal value in the interpreter, it gets pushed onto DS.

Some may notice the arithmetic rules in Final work the same as in HP calculators. It is called Reverse Polish Notation (RPN). A nice feature of RPN is that you do not need grouping operators like ( ) to group and order operations, any such expression can be expressed using the RPN/stack-based approach of Final.

This way of thinking does take a bit of getting used to, but it is the "secret sauce" that makes all Final-like languages so uniquely flexible and efficient, in terms of code factoring, and code reusability.

Try out a more complicated example. This requires parentheses in a conventional language, due to order of operations and operator binding. You do it in Final by working inside-out, then left to right at each grouping level: ((3 + 1) * (11 - 6)) / 4 . Tip: you enter operands for things like - and / in reading order, so: a - b is entered as "a b -".

```forth
> 3  1 +  11 6 - * 8 / $

2
```

If you have a sharp eye, you will know that the result is actually 2.5, not 2. The cause is Final's type coercion rules. If all the input operands to arithmetic operators are integers, the result will be an integer, due to integer truncation of rounding. To make it be a float, just change one of the operands to include a decimal point. This causes it to be interpreted as a float, and Final's coercion rules always promote the result of mixed int/float operations to have a float result, so the "float virality" will propagate through. Try it!

```forth
> 3  1.0 +  11 6 - * 8 / $

2.5
```

Now let's write our first word in Final. Words are stored in something called a `Dictionary`; there are at least 2 dictionaries when you start Final:

- the sealed `core` Dictionary that contains the words of the Final language
- a (mostly) empty `local` dictionary in your Session, into which you can add new definitions

Let's write one now, enter this:

```forth
> : helloFromFinal "Hello from Final!\n" $ ;
helloFromFinal

Hello from Final!
```

You just created your first Final word called helloFromFinal, and then invoked it! Here is the anatomy of what is called a `colon definition`

1. opening `:` then a space; this puts Final into compiling mode
2. the name of the new word
3. the literals, operators, and words making up the definition
4. `;` closes the definition and adds it to your local dictionary

The `helloFromFinal` following the definition then invokes it, causing the word to print the text.

Creating new words is as simple as that. When you create new words, you can include words you have previously defined along with core words.
