// Streams, builders, and files

/*
Accumulate Utf-8 text, then take .toString^ .
Used by Json pretty printing and general string building.
*/
:doc StringBuilder

/*
Create an empty StringBuilder (compact, no indent).
( -- sb:StringBuilder )
*/
:doc StringBuilder#new

/*
Create a pretty StringBuilder with the given indent string
(e.g. two spaces).
( indent:String -- sb:StringBuilder )
*/
:doc StringBuilder#newWithPretty

/*
Append a Unicode code point (as Int).
( r:i -- )
*/
:doc StringBuilder#writeRune

/*
Append a String followed by a newline.
( s:String -- )
*/
:doc StringBuilder#writeLine

/*
Append a String as-is.
( s:String -- )
*/
:doc StringBuilder#writeString

/*
Append a String with current pretty indentation.
( s:String -- )
*/
:doc StringBuilder#writeStringPretty

/*
Append a space when pretty mode is on.
( -- )
*/
:doc StringBuilder#writeSpacePretty

/*
Append newline when pretty, else a space.
( -- )
*/
:doc StringBuilder#writeWhitespacePretty

/*
Increase pretty indent nesting.
( -- )
*/
:doc StringBuilder#nest

/*
Decrease pretty indent nesting.
( -- )
*/
:doc StringBuilder#unnest

/*
Return the accumulated text.
( -- s:String )
*/
:doc StringBuilder#toString

/*
True when built with newWithPretty.
*/
:doc StringBuilder#pretty

/*
Indent string used in pretty mode.
*/
:doc StringBuilder#indentString

/*
Shared Final parent for stream close / flush.
Methods probe Go Closeable / Flushable and no-op when absent.
Readers typically only Close; writers may Flush and Close.
*/
:doc StreamReadWrite

/*
Release the stream when the concrete value is Closeable.
( -- )
*/
:doc StreamReadWrite#close

/*
Flush buffered output when the concrete value is Flushable.
( -- )
*/
:doc StreamReadWrite#flush

/*
Capability type: Utf-8 text write (writeRune / writeString / writeLine).
Parents StreamReadWrite. Concrete writers (TextWriter, LineBufferTextWriter, …)
and other TextWrite implementers coerce here.
*/
:doc TextWrite

/*
Write a Unicode code point (as Int).
( r:i -- )
*/
:doc TextWrite#writeRune

/*
Write a String followed by a newline.
( s:String -- )
*/
:doc TextWrite#writeLine

/*
Write a String as-is.
( s:String -- )
*/
:doc TextWrite#writeString

/*
Capability type: raw byte write (writeNextByte / writeBytes).
Parents StreamReadWrite.
*/
:doc ByteWrite

/*
Write one byte (Int or Uint).
( b:iu -- )
*/
:doc ByteWrite#writeNextByte

/*
Write all bytes from a ByteList.
( bytes:ByteList -- )
*/
:doc ByteWrite#writeBytes

/*
Capability type: Utf-8 text read (getNextRune).
Parents StreamReadWrite. Concrete TextReader subclasses this;
readLine stays on TextReader (shell helper over GetNextRune).
*/
:doc TextRead

/*
Next Unicode code point.
( -- r:i true | false )
*/
:doc TextRead#getNextRune

/*
Capability type: raw byte read (getNextByte / readBytes).
Parents StreamReadWrite.
*/
:doc ByteRead

/*
Next byte as Int.
( -- b:i true | false )
*/
:doc ByteRead#getNextByte

/*
Read up to n bytes into a ByteList.
( n:i -- bytes:ByteList )
*/
:doc ByteRead#readBytes

/*
Read Utf-8 text as runes or lines from a file or String.
Parents TextRead → StreamReadWrite.
*/
:doc TextReader

/*
Open a TextReader on a file path. Panics on error.
( path:String -- reader:TextReader )
*/
:doc TextReader#newFromFile

/*
Open a TextReader on a file path.
( path:String -- reader:TextReader true | err:s false )
*/
:doc TextReader#newFromFile?

/*
Open a TextReader over an in-memory String.
( s:String -- reader:TextReader true | err:s false )
*/
:doc TextReader#newFromString

/*
Next Unicode code point.
( -- r:i true | false )
*/
:doc TextReader#getNextRune

/*
Next line (without the newline).
( -- line:s true | false )
*/
:doc TextReader#readLine

/*
Close the reader.
( -- )
*/
:doc TextReader#close

/*
Write Utf-8 text to a file.
Parents TextWrite → StreamReadWrite.
*/
:doc TextWriter

/*
Open for append, creating the file if needed. Panics on error.
( path:String -- writer:TextWriter )
*/
:doc TextWriter#newFromFileAppendOrCreate

/*
Open for append, creating the file if needed.
( path:String -- writer:TextWriter true | err:s false )
*/
:doc TextWriter#newFromFileAppendOrCreate?

/*
Open for write, truncating or creating. Panics on error.
( path:String -- writer:TextWriter )
*/
:doc TextWriter#newFromFileTruncateOrCreate

/*
Open for write, truncating or creating.
( path:String -- writer:TextWriter true | err:s false )
*/
:doc TextWriter#newFromFileTruncateOrCreate?

/*
Write a Unicode code point (as Int).
( r:i -- )
*/
:doc TextWriter#writeRune

/*
Write a String followed by a newline.
( s:String -- )
*/
:doc TextWriter#writeLine

/*
Write a String as-is.
( s:String -- )
*/
:doc TextWriter#writeString

/*
Flush buffered output.
( -- )
*/
:doc TextWriter#flush

/*
Close the writer.
( -- )
*/
:doc TextWriter#close

/*
Read raw bytes from a file or ByteList.
Parents ByteRead → StreamReadWrite.
*/
:doc ByteReader

/*
Open a ByteReader on a file path. Panics on error.
( path:String -- reader:ByteReader )
*/
:doc ByteReader#newFromFile

/*
Open a ByteReader on a file path.
( path:String -- reader:ByteReader true | err:s false )
*/
:doc ByteReader#newFromFile?

/*
Open a ByteReader over Bytes (ByteList carrier).
( bytes:ByteList -- reader:ByteReader true | err:s false )
*/
:doc ByteReader#newFromBytes

/*
Next byte as Int.
( -- b:i true | false )
*/
:doc ByteReader#getNextByte

/*
Read up to n bytes into a ByteList.
( n:i -- bytes:ByteList )
*/
:doc ByteReader#readBytes

/*
Close the reader.
( -- )
*/
:doc ByteReader#close

/*
Write raw bytes to a file or ByteList.
Parents ByteWrite → StreamReadWrite.
*/
:doc ByteWriter

/*
Open for append, creating the file if needed. Panics on error.
( path:String -- writer:ByteWriter )
*/
:doc ByteWriter#newFromFileAppendOrCreate

/*
Open for append, creating the file if needed.
( path:String -- writer:ByteWriter true | err:s false )
*/
:doc ByteWriter#newFromFileAppendOrCreate?

/*
Open for write, truncating or creating. Panics on error.
( path:String -- writer:ByteWriter )
*/
:doc ByteWriter#newFromFileTruncateOrCreate

/*
Open for write, truncating or creating.
( path:String -- writer:ByteWriter true | err:s false )
*/
:doc ByteWriter#newFromFileTruncateOrCreate?

/*
Write into an existing Bytes buffer (ByteList carrier).
( bytes:ByteList -- writer:ByteWriter true | err:s false )
*/
:doc ByteWriter#newFromBytes

/*
Write one byte (Int or Uint).
( b:iu -- )
*/
:doc ByteWriter#writeNextByte

/*
Write all bytes from a ByteList.
( bytes:ByteList -- )
*/
:doc ByteWriter#writeBytes

/*
Flush buffered output.
( -- )
*/
:doc ByteWriter#flush

/*
Close the writer.
( -- )
*/
:doc ByteWriter#close

// File / Dir / Path / FileInfo — see 1015-file-system.fnl

