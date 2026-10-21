// Modules — install, import, search order, projects

/*
Compile and register a module for this process.
Already-installed names are a no-op.
name is the module identity (optionally name@version;
@version is accepted but not used for selection yet).
opts is a DynamicObject; known fields:
  path — filesystem root of the module (directory that
         contains _MODULE_). When set, skips _REPLACES_
         and treats name as the identity only.
When path is omitted, the root is taken from _REPLACES_
for name, or name itself as a cwd-relative or absolute path.
Fails if the local dictionary is locked or sealed, or if
the module lists requires (dependent modules not yet supported).
( name:s opts:DynamicObject -- )
*/
:doc install

/*
Attach an installed module to this Session so its words
can be found. syn is the prefix used with \; empty syn
uses the module's default prefix (last path segment of
its name). name is the installed module identity.
Fails if the module is not installed, if syn is reserved
(local/core), or if the prefix is already in use.
( syn:s name:s -- )
*/
:doc import

/*
Put the listed module prefixes at the front of the
current word search order.
( prefixes:StringList -- )
*/
:doc using

/*
Restore the default word search order for this Session
(local, then core, then imported modules).
( -- )
*/
:doc using-default

/*
Load a .fnlp project: install and import each requires
entry, then compile the project's files into the local
Session. projectPath must end in .fnlp.
( projectPath:s -- )
*/
:doc loadProject
