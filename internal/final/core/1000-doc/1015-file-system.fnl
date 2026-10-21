// File system — Path, Dir, File, FileInfo

/*
Path string algebra (class-only; no Path instances).
Join, split, clean, abs, basename, dirname, extname.
Most members need no security restriction; abs may
read the process cwd (SR.FileSystemRead).
Operands are strings on the data stack; Path is the
type receiver.
eg: `` "out" "q1.md" Path.join^ ``
See also Dir (directories) and File (file endpoints).
*/
:doc Path

/*
Join base and name with the host path separator.
Rebuilds a full endpoint from FileInfo.path + FileInfo.name.
eg: `` "out" "report.md" Path.join^ ``
( base:s name:s -- s )
*/
:doc Path#join

/*
Split one path into components (StringList).
Not PATH-environment SplitList — segment split of a
single path. Absolute paths keep the root as the first
element (e.g. "/").
eg: `` "/var/log" Path.split^ ``
( s -- parts:StringList )
*/
:doc Path#split

/*
Clean redundant . and .. segments.
eg: `` "a/./b/../c" Path.clean^ ``
( s -- s )
*/
:doc Path#clean

/*
Resolve to an absolute path (may use process cwd).
Requires SR.FileSystemRead. Panics if resolution fails.
eg: `` "." Path.abs^ ``
( s -- s )
*/
:doc Path#abs

/*
Leaf name without extension.
Strips any directory prefix first.
"out/reports/q1.md" → "q1".
Leading-dot names stay whole (".gitignore").
Same split rules as File.getBaseExt (base part).
eg: `` "out/q1.md" Path.basename^ ``
( s -- s )
*/
:doc Path#basename

/*
Parent directory of a path string.
Same idea as FileInfo.path for an endpoint.
eg: `` "out/q1.md" Path.dirname^ ``
( s -- s )
*/
:doc Path#dirname

/*
Extension including the leading ".", or "" if none.
Pairs with Path.basename. Leading-dot names have
empty extname.
eg: `` "out/q1.md" Path.extname^ ``
( s -- s )
*/
:doc Path#extname

/*
A directory root path (not an OS open-directory handle).
Class helpers: new, cwd, setCwd, exists?, info, mkdir,
and iter flags. Instance: .string@ and .iter^ .
Iteration is modal — Dir is not Iterable; pass flags
to .iter^ to get a FileSystemIter of FileInfo values.
Flags: Dir.IterFiles, Dir.IterDirs, Dir.IterDeep,
Dir.IterAll (combinable with or / |).
eg: `` "." Dir.new^ Dir.IterFiles t:.iter^ foreach … ``
See Path for string algebra; File for file content.
*/
:doc Dir

/*
Include directories when iterating.
eg: `` Dir.IterDirs Dir.IterFiles or ``
( -- flags:i )
*/
:doc Dir#IterDirs

/*
Include non-directory entries when iterating.
eg: `` Dir.IterFiles t:.iter^ ``
( -- flags:i )
*/
:doc Dir#IterFiles

/*
Recurse into subdirectories (depth-first).
eg: `` Dir.IterFiles Dir.IterDeep or ``
( -- flags:i )
*/
:doc Dir#IterDeep

/*
Dirs | Files | Deep — full recursive tree.
eg: `` Dir.IterAll t:.iter^ ``
( -- flags:i )
*/
:doc Dir#IterAll

/*
Create a Dir from a path string.
Does not require the path to exist until a Dir FS
method or .iter^ runs.
eg: `` "." Dir.new^ ``
( s -- dir:Dir )
*/
:doc Dir#new

/*
Process current working directory as a Dir.
Requires SR.FileSystemRead.
eg: `` Dir.cwd^ .string@ ``
( -- dir:Dir )
*/
:doc Dir#cwd

/*
Change process cwd (chdir). Modal.
Requires SR.FileSystemRead.
eg: `` "/project" Dir.setCwd^ if … else … then ``
( s -- true | err:s false )
*/
:doc Dir#setCwd

/*
True if the path exists (file or directory).
Missing → false. Panics on unexpected FS errors.
Requires SR.FileSystemRead.
eg: `` "config.json" Dir.exists?^ ``
( s -- b )
*/
:doc Dir#exists?

/*
Metadata for a path as FileInfo. Modal.
Requires SR.FileSystemRead. Not named stat.
eg: `` "." Dir.info^ if … else … then ``
( s -- fi:FileInfo true | err:s false )
*/
:doc Dir#info

/*
Create directory and missing parents (MkdirAll).
Modal. Idempotent if path is already a directory.
Fails if path exists as a non-directory.
Requires SR.FileWrite.
eg: `` "out/reports" Dir.mkdir^ if … else … then ``
( s -- true | err:s false )
*/
:doc Dir#mkdir

/*
Canonical path string for this Dir root.
eg: `` t:.string@ ``
( -- s )
*/
:doc Dir#string

/*
FileSystemIter over entries under this Dir.
Flags select files / dirs / deep (see Dir.Iter*).
Does not yield the root itself. Host ReadDir order
(lexical within each directory). Faults if the path
is not a usable directory. Requires SR.FileSystemRead.
foreach accepts the iterator on the data stack.
eg: `` Dir.IterFiles t:.iter^ foreach i >ts … t:drop /foreach ``
( flags:i -- iter )
*/
:doc Dir#iter

/*
Sealed metadata snapshot for a file or directory entry.
Table-friendly columns: path, name, size, modTime,
isDir, isSymlink. path is the parent directory (no
trailing separator except filesystem root); name is
the leaf. Path.join of path and name is the full
endpoint. No basename/extname properties — use
File.getBaseExt on name.
Produced by Dir.info / File.info and by FileSystemIter.
*/
:doc FileInfo

/*
Parent directory of this entry (before name).
( -- s )
*/
:doc FileInfo#path

/*
Final path element (file or directory name).
( -- s )
*/
:doc FileInfo#name

/*
Size in bytes (host-dependent for non-regular files).
In Final-32, use sizeHigh/sizeLow.
( -- n:i )
*/
:doc FileInfo#size

/*
High order 32 bits of unsized size.
Compatability property for Final-32.
( -- n:u )
*/
:doc FileInfo#sizeHigh

/*
Low order 32 bits of unsized size.
Compatability property for Final-32.
( -- n:u )
*/
:doc FileInfo#sizeLow

/*
Modification time.
( -- dt:DateTime )
*/
:doc FileInfo#modTime

/*
True if this entry is a directory.
( -- b )
*/
:doc FileInfo#isDir

/*
True if this entry is a symbolic link.
( -- b )
*/
:doc FileInfo#isSymlink

/*
File endpoints: existence, metadata, whole-file
read/write, delete/move/copy, and name base/ext
helpers. Open File handles are separate (stream
factories). Class methods take path strings.
Prefer *? forms when you want modal errors instead
of panic. See Dir for directories; Path for strings.
*/
:doc File

/*
True if the path exists (file or directory).
Missing → false. Panics on unexpected FS errors.
Requires SR.FileSystemRead. Alias: fileExists? .
eg: `` "notes.txt" File.exists?^ ``
( path:s -- b )
*/
:doc File#exists?

/*
Same as File.exists? (legacy name).
( path:s -- b )
*/
:doc File#fileExists?

/*
Metadata for a path as FileInfo. Modal.
Requires SR.FileSystemRead.
eg: `` "notes.txt" File.info^ if … else … then ``
( path:s -- fi:FileInfo true | err:s false )
*/
:doc File#info

/*
Delete a file or empty directory. Modal.
Non-empty directories fail (no recursive wipe yet).
Requires SR.FileWrite.
eg: `` "tmp/x.txt" File.delete^ if … else … then ``
( srcfilepath:s -- true | err:s false )
*/
:doc File#delete

/*
Rename or relocate. dst is the full destination path.
Fails if dst already exists (no silent clobber).
Requires SR.FileWrite.
eg: `` "a.txt" "b.txt" File.move^ if … else … then ``
( srcfilepath:s dstfilepath:s -- true | err:s false )
*/
:doc File#move

/*
Copy a file into a destination directory.
Result path is dstdir joined with the source leaf
name (including extension). Does not create dstdir
(use Dir.mkdir). Fails if the destination file
already exists. Requires SR.FileRead and SR.FileWrite.
eg: `` "notes.txt" "archive" File.copy^ if … else … then ``
( srcfilepath:s dstdirpath:s -- true | err:s false )
*/
:doc File#copy

/*
Split a name into base and extension (ext without ".").
Strips any directory prefix first. Leading-dot names
are entirely base with empty ext.
eg: `` "archive/q1.md" File.getBaseExt^ ``
( name:s -- base:s ext:s )
*/
:doc File#getBaseExt

/*
Join base and ext into a name. Keeps any path on base.
ext is without the leading "." (a leading "." is
stripped if present). Empty ext → base alone.
eg: `` "out/report" "md" File.makeName^ ``
( base:s ext:s -- name:s )
*/
:doc File#makeName

/*
Read entire file as ByteList. Panics on error.
Requires SR.FileRead.
( path:s -- bytes:ByteList )
*/
:doc File#readBytes

/*
Read entire file as ByteList. Modal.
Requires SR.FileRead.
( path:s -- bytes:ByteList true | err:s false )
*/
:doc File#readBytes?

/*
Create or truncate path; write all bytes. Panics on error.
Requires SR.FileWrite.
( path:s bytes:ByteList -- )
*/
:doc File#writeBytes

/*
Create or truncate path; write all bytes. Modal.
Requires SR.FileWrite.
( path:s bytes:ByteList -- true | err:s false )
*/
:doc File#writeBytes?

/*
Read entire file as String. Panics on error.
Requires SR.FileRead.
( path:s -- s )
*/
:doc File#readText

/*
Read entire file as String. Modal.
Requires SR.FileRead.
( path:s -- s true | err:s false )
*/
:doc File#readText?

/*
Create or truncate path; write UTF-8 text. Panics on error.
Requires SR.FileWrite.
( path:s s -- )
*/
:doc File#writeText

/*
Create or truncate path; write UTF-8 text. Modal.
Requires SR.FileWrite.
( path:s s -- true | err:s false )
*/
:doc File#writeText?
