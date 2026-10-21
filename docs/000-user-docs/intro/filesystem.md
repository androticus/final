# File System

Final’s filesystem API covers **paths**, **metadata**, **directories**, and **whole-file** helpers. Sequential streaming I/O (line/byte readers and writers) lives under [File Streams](streams/file.md) — this page does not repeat open/read/write stream factories.

You already know the pieces these APIs lean on:

* Class methods with `^` — [Methods](objects/methods.md)
* Modal `( … -- result true \| err:s false )` — [Writing Clean Final Code](clean-final-coding.md)
* `foreach` over iterators — [Iterators: foreach](control-flow/iter-foreach.md)

| Type | Role |
| ---- | ---- |
| **Path** | Class-only path-string algebra (join, split, clean, …) |
| **File** | Exists / info / copy / move / delete; whole-file read/write; name helpers |
| **Dir** | Directory root value; cwd; mkdir; filesystem iteration |
| **FileInfo** | Metadata snapshot (name, size, modTime, isDir, …) |

Use `help File` / `help Dir` / `help Path` / `help FileInfo` in the TUI for the full member list.

---

## Path (class-only)

Operands are **strings** on the data stack. There are no Path instances.

| Method | Effect | Description |
| ------ | ------ | ----------- |
| `join` | `( base:s name:s -- s )` | join with host separator |
| `split` | `( s -- list:StringList )` | path components (absolute keeps root) |
| `clean` | `( s -- s )` | drop redundant `.` / `..` |
| `abs` | `( s -- s )` | absolute path (uses process cwd) |
| `basename` | `( s -- s )` | leaf **without** extension |
| `dirname` | `( s -- s )` | parent directory |
| `extname` | `( s -- s )` | extension **with** leading `.`, or empty |

```forth
"out" "report.md" Path.join^          // "out/report.md" (host sep)
"a/./b/../c" Path.clean^              // "a/c"
"out/reports/q1.md" Path.basename^    // "q1"
"out/reports/q1.md" Path.extname^     // ".md"
"/var/log/syslog" Path.split^ foreach i $ sp /foreach
```

---

## FileInfo

Snapshot returned by `File.info` / `Dir.info` and by directory iteration. No public `new`.

| Prop | Type | Description |
| ---- | ---- | ----------- |
| `path` | String | path for this entry |
| `name` | String | leaf name |
| `size` | Int | size in bytes |
| `modTime` | DateTime | modification time |
| `isDir` | Bool | directory? |
| `isSymlink` | Bool | symlink? |

---

## File (class methods)

Most ops take a **path string**. Mutators use the modal shape unless noted.

| Method | Effect | Description |
| ------ | ------ | ----------- |
| `exists?` / `fileExists?` | `( path:s -- b )` | missing → `false` |
| `info` | `( path:s -- fi:FileInfo true \| err:s false )` | metadata |
| `delete` | `( path:s -- true \| err:s false )` | file or **empty** directory |
| `move` | `( src:s dst:s -- true \| err:s false )` | rename/relocate; **dst is full path**; no clobber |
| `copy` | `( src:s dstdir:s -- true \| err:s false )` | copy into directory; keeps leaf name; no clobber |
| `getBaseExt` | `( name:s -- base:s ext:s )` | leaf split; **ext without** dot |
| `makeName` | `( base:s ext:s -- name:s )` | join base+ext; empty ext → base alone |
| `readText` / `readText?` | `( path:s -- s )` / modal | whole file as String |
| `writeText` / `writeText?` | `( path:s s -- )` / modal | create/truncate UTF-8 |
| `readBytes` / `readBytes?` | `( path:s -- bytes )` / modal | whole file as ByteList |
| `writeBytes` / `writeBytes?` | `( path:s bytes -- )` / modal | create/truncate bytes |

Bare read/write forms **panic** on I/O error; `?` forms are modal. Stream open/read/write: [File Streams](streams/file.md).

```forth
"archive/report.final.md" File.getBaseExt^   // base="report.final" ext="md"
"out/report" "md" File.makeName^             // "out/report.md"
"." File.exists?^
"." File.info^ if >ts t:.name@ $ sp t:.isDir@ $ nl t:drop else $ nl then

"tmp/notes.txt" "Apple" File.writeText?^
"tmp/notes.txt" "tmp/archive" File.copy^
"tmp/archive/notes.txt" "tmp/archive/notes-2.txt" File.move^
"tmp/archive/notes-2.txt" File.delete^
```

---

## Dir

A **Dir** is a directory root value (path), not an open OS handle.

| Kind | Name | Effect | Description |
| ---- | ---- | ------ | ----------- |
| class | `new` | `( path:s -- dir )` | Dir for path |
| class | `cwd` | `( -- dir )` | process working directory |
| class | `setCwd` | `( path:s -- true \| err:s false )` | change process cwd |
| class | `exists?` | `( path:s -- b )` | path exists? |
| class | `info` | `( path:s -- fi true \| err:s false )` | metadata |
| class | `mkdir` | `( path:s -- true \| err:s false )` | `MkdirAll`; idempotent if already a dir |
| prop | `string` | String | canonical path string for this Dir |
| method | `iter` | `( flags:i -- iter )` | `FileSystemIter` of `FileInfo` under this Dir |

### Iteration flags (class properties on Dir)

Combine with bitwise OR:

| Flag | Meaning |
| ---- | ------- |
| `Dir.IterDirs` | include directories |
| `Dir.IterFiles` | include files (non-dirs) |
| `Dir.IterDeep` | recurse |
| `Dir.IterAll` | dirs + files + deep |

```forth
"." Dir.new^ >ts
t:.string@ $ nl

"tmp/demo" Dir.mkdir^
"tmp/demo" Dir.new^ >ts
Dir.IterFiles t:.iter^
foreach
  i >xs
  x:.name@ $ nl
  x:drop
/foreach
t:drop
```

Deep walk (pick a **small** root — recursive walks can be large):

```forth
: filesysIter  ( root:s -- )
  Dir.new^ >ts
  Dir.IterAll t:.iter^
  foreach
    i >ts
      t:.isDir if ">" else " " then
      "%s %9d %v %s %s\n" fmt $
      drop
    t:drop
  /foreach
  t:drop
;

"." filesysIter
```

---

## Security

Filesystem members are gated by **`SR.FileRead`**, **`SR.FileWrite`**, and **`SR.FileSystemRead`** (exists/info/iter/cwd vs content open). Restricted Sessions may deny them. See [Security Restrictions](embedding/security-restrictions.md).

## Examples on disk

* `examples/file-system/path.fnl`
* `examples/file-system/file.fnl`
* `examples/file-system/dir.fnl`
* `examples/file-system/filesys-iter.fnl`
