# Features

This entry contains a whirlwind list of Final's other features. Each it explained in detail in its own entry. Now that you've had a basic introduction to Final, you are free to browse around. However it is recommended to first read the topics in the Foundations section, to get a good grounding of Final's core constructs and operations.

- efficient OpCode-based VM with fast jump-table dispatch
- full support for Types and Objects
- concise, expressive "compound path" syntax for object operations
- auxilliary stacks, for convenient object parking and access
- support for Interators, and convenient `foreach` iterated loop
- excellent assortment of often-used collection types
- `{` … `}` and `sm}` builders for DynamicObject and StringMap; DynamicObject `.clone^` for prototype-style copies; StaticObject via `do.newStatic^` for fixed-schema rows
- TypedObject class constants plus Final methods via `: Type#method … ;` (instance) and `: Type##method … ;` (class)
- Table — columnar collection with schema from DynamicObject / TypedObject / Types; StaticObject rows; stable `sortByOrd` / `sortByName`
- full support for Unicode strings as Utf-8, including efficient Rune iteration; wide variety of str:: operations
- DateTime and Interval support, with formatting via `dt::fmt` (plus constants such as `DateTime.iso`, `Interval.clock`)
- expressive and convenient ( ) operators for creating value list to `fmt` operator
- support for a variety of variants of TextStream and ByteStream interfaces:
  - files
  - strings (reading) StringBuilder (writing)
  - stdio
- File system — `File`, `FileInfo`, `Dir`, `Path`; metadata, copy/move/delete, directory iteration
- Regex, including compiled regex, matches, selectors
- full Json rendering and parsing
- Threading, Pipes, and Atomic threadsafe objects
- Web Dial (Http GET/POST) and Web Listen (`HttpListener` — ServeMux routes, path params, REST methods, Pip-per-route handlers)
- SQL databases via `DbProvider` (Postgres, MySQL, SQLite); iterable result rows as DynamicObjects; `iter>list`
- Env and Moniker — cwd `.env` + process environment; opaque DSN / secret handles with `$KEYNAME` substitution
- Cryptography — Hash (MD5 checksum, SHA-256), Hex/Base64, SecureRandom, Key, Cipher (AES-256-GCM / ChaCha20-Poly1305), Ed25519 / X25519, Mac (HMAC-SHA-256), Kdf (HKDF-SHA-256, Argon2id)
- Linear Algebra: `VecNf` / `Mat2f`, `vec::` / `mat::`, OLS `fit` / `compare` / `predict`
- `try` / `catch` / `finally` / `throw` — recoverable exceptions (invalid JSON `throw`s rather than panicking)
- Modules (local `install` / `import`) and `.fnlp` projects (`loadProject`)
- Session configuration (`sess.config`) and pluggable logging (`logInfo`, …)
- `forget:` reload marks — inclusive wipe back to a mark word for load → try → reload without process relaunch
- Security Restrictions, to restrict operations like file access, web dialing, listening, databases, crypto, threading, and module install/import, for embedded/AI scenarios
- `LockSession` — exclusive Send hold so an agent turn is not interleaved
- full example of a Web-based chat-like app that provides Final as a tool for the Agent
- finalembed / finalgo Go packages for embedding Final in other Go programs, with Session and Send semantics
- VS Code / Cursor Final extension for `.fnl` — TextMate highlighting, 2-space indent, sandwich snippets (local VSIX / F5 install)
- Raspberry Pi 5 support, for writing Pi-based control systems.
- Command line support, OS pipes, shebang scripts (Unix-like only)
- Raster blitting and drawing operations (all Final versions)
- GUI support via Fyne controls, final-gui binary
- support for Games and Simulations via SimGame framework
- Sound support via SoundPlayer and Sound objects

**Near-term Features Planned**

- Language-server features for the editor (dictionary completion, stack-effect hover, diagnostics)
  - build on compiler execution state tracking
- Compiler execution state tracking
  - catch stack bugs: underflow/leak
- Markdown parsing
- Html and Xml rendering
- Interactive Web tutorials, using a sandboxed Final Session

