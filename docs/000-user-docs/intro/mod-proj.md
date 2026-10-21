# Final Modules and Projects

This release supports **local, leaf** modules and **`.fnlp` projects**. There is no remote repository client yet, and a module may not list other modules in `requires`. Behavior and file format may still change.

Example module: `modules/csv` (Table → CSV/TSV). Example project: `projects/csv-demo`.

## Modules

A module is a directory with a `_MODULE_` JSON manifest (`name`, `version`, `final`, `files`, empty `requires`) plus the listed `.fnl` files.

### `install` / `import`

```forth
// ( name:s opts:DO -- )
"github.com/androticus/final/modules/csv" newdo install

// ( syn:s name:s -- )  empty syn → default prefix (last path segment, here csv)
"" "github.com/androticus/final/modules/csv" import
```

`install` compiles and seals the leaf module (already-installed is a nop). Resolution: optional `opts` field `path`, else a `_REPLACES_` map walked from the working directory, else `name` as a filesystem path.

`import` attaches an installed module to this Session. First attach wins. The default prefix is the last segment of the module name (`csv`). `local` and `core` are reserved.

### Calling module words

Qualified: `prefix\word` (a backslash). After import, a **bare** name also searches attached modules.

```forth
csv\getDefaultCsvOptions
getDefaultCsvOptions        // same, via search order
```

### `using` / `using-default`

```forth
// ( prefixes:StringList -- )
( "csv" s) using            // listed prefixes first in the search order
using-default               // ( -- ) restore Session.modules order
```

### Security

`install` is denied by **`SR.ModuleInstall`**; `import` by **`SR.ModuleImport`**. A host that denies those bits should **pre-install and pre-import** anything the script is allowed to use. See [Security Restrictions](embedding/security-restrictions.md).

## Projects

A **`.fnlp`** file lists project `files` and `requires` (module names + versions). Version is carried but not used for selection yet.

```forth
// ( projectPath:s -- )
"projects/csv-demo/csv-demo.fnlp" loadProject
```

`loadProject` installs and imports each required module, then compiles the project files into the current Local Session. In **final**, **File → Load** on a `.fnlp` does the same.

`loadProject` itself is not SR-tagged; the nested `install` / `import` still honor the Session mask.
