// RichText

/*
Formatted text, plain or Markdown.
Also has Canvas.visible and Canvas.refresh.
*/
:doc RichText

/*
Replace the contents with plain text.
( text:s -- )
*/
:doc RichText#setText

/*
Replace the contents by parsing Markdown.
( content:s -- )
*/
:doc RichText#parseMarkdown

/*
Create a RichText with plain text.
( text:s -- rt:RichText )
*/
:doc RichText#newWithText

/*
Create a RichText from Markdown.
( content:s -- rt:RichText )
*/
:doc RichText#newFromMarkdown
