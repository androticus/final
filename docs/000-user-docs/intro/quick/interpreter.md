# Interpreter

When you first start the Final TUI, Final is in intepreter mode. This is an interactive request/response type of interface. When you start a colon definition with `:` Final goes into compiling mode. We will discuss the basic lexical rules of Final text, which applies both to the iterpreter and compiler, then cover the interpreter itself.

## Final Lexical Rules

Final source is Utf-8, of which regular ASCII-7 is a proper subset. Note that Final does not accept "Code Pages", which are the early DOS and Windows era "hack" for creating overlays in the top 128 bits of ASCII. Such sources will be mis-interpeted by Final.

Final text is parsed as tokens, which are whitespace separated values as defined by the parsing rules. Individual Final operators and words must be separated by whitespace, unlike languages which permit whitespace eliding in many contexts. (The one exception to this is Final's compound object path syntax, covered later.)

Valid whitespace is: space, tab ("\t"), newline ("\n" aka linefeed), or carriage return ("\r" aka return.) In an operation counting or splitting lines, Final accepts "\r\n" Windows-style newlines to be a single end-of-line marker. It is suggested that you uniformly use the single "\n" form for both Final source files, and textual data files. No other whitespace, including Unicode whitespace characters, are accepted.

These are the types of tokens in Final
* Null literal: `null`;
* Bool literal: `true` or `false` in lowercase;
* Int literal: any valid decimal, octal, hex, or binary number, without a . decimal  
   eg, "10" "0x7F" "0b101010" "-49"  
   note: numbers may not be preceded with `+`, only `-`;
* Uint literal: a valid positive integer, with "u" affix, eg. 25u;
* Float literal: a valid number, that contains a decimal point; exponential notation is allowed;
* Quoted String: a textual value enclosed in quotes; Go-style escapes `\n` `\t` `\xHH` `\uHHHH` `\UHHHHHHHH`; literal Unicode (emoji, non-English) is accepted in the source unless it is a dangerous character (see below). Dangerous characters may appear in the string *value* only via `\x` / `\u` / `\U`;
* Raw String: a textual value enclosed in back ticks: `the rain in spain`; the entire text is interpreted literally, including line breaks; no escapes. Dangerous Unicode in a raw string is a parse error;
* Unicode Rune: a single rune in single quotes; printable characters, C-style escapes, `\xHH`, `\uHHHH`, `\UHHHHHHHH`;
* // Comment: the // characters signify a comment, to the end of the line;
* /* */ Comment: a comment that can span multiple lines; it can include embedded // comments (which are just considered part of the whole comment text) but cannot include nested slash-star comments;
* Compound Path accessor: a valid compound path accessor (covered later);
* an identifier, which is either the name of an existing word, or based on the previous operator such as `var:` will become a new word; identifiers may not start with a number, and must be printable Unicode characters.

### Identifiers

Identifiers in Final are far more permissive than in most other languages--they can contain combinations of letters, numbers, and most special characters.

Most of the special characters are used by Final to convey a family of related operations, so before going wild and using them in your identifiers, take the time to learn the lexicon, and ask if your new word belongs in that group.

Some characters and patterns are reserved by Final, and may not be used in identiers. Here is a summary of the rules, or disallowed characters:
* an identifier cannot begin with a number
* the characters must be Printable/non-whitespace Unicode characters
* the following characters cannot be used: `.` `@` `!` `\` `^` `~` `"` `` ` `` `'`
* Source text may not contain "Dangerous" characters: Trojan-Source / invisible Unicode (Bidi marks, zero-width characters, variation selectors, and similar). The parser stops at that code prints out the value in hex. If you absolutely must use those characters, put them in a quoted string or rune using `\x`, `\u`, or `\U`.

**Naming Guidelines**

Final follows a standard of using camelCase for most identifiers. Exceptions are Types and constants, which are in PascalCase. You may also choose to use hyphen-naming instead of camelCase if you wish.

Final encourages the use of code factoring to keep individual words small, and to convey the purpose of a piece of code by way of its name, as opposed to relying on comments. To avoid name collisions in your User dictionary, it is suggested to factor by keeping the main word name, then an underscore, followed by the factored name. [Subwords](../control-flow/subwords.md) within a word, are also supported, and often more convenient.

## The Final Interpreter

Final's interpreter uses simple rules. It skips whitespace, then reads a token. Based on the type of the token, it does the following:

1. Comment: discard it
2. Literal: push it on DS
3. Identifier: look that up in the current Dictionary search path
   if it is found, then the word is executed
4. else signal an error

If the word is `:` then Final is put into Compiling state. It leaves compiling state and re-enters interpreting state after the `;`.

So for you, the user, this is very convenient, because you can use the interpreter for interactive purposes, including compiling and testing new words. Likewise, a Final source file can intermix immediate, interpretive behavior with compilation of new words. This has a number of useful use-cases, which we will mention as we get to the features involved.

Let's take a quick look at compiling state, and how Final compiles new words. Then we'll move on to a quick intro to Final's control flow blocks, and its collection types.