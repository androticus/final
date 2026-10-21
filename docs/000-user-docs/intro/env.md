# Env and Moniker

**Env** reads host configuration. **Moniker** holds a string template that may contain `$KEYNAME` placeholders, filled without putting secrets on the data stack. Together they let scripts connect to databases (and similar services) without baking credentials into `.fnl` source.

---

## Env

Class-only type — no instances. Lookup order:

1. **`./.env`** in the process **current working directory** (re-read on every call; not cached)
2. **Process environment** (`os.LookupEnv`)

A key in `.env` wins over the same key in the process environment. Keys are **case-sensitive**.

| Kind | Name | Effect | Description |
| ---- | ---- | ------ | ----------- |
| class | `get` | `( key:s -- val:s true \| false )` | cleartext value on DS when found |
| class | `subMoniker` | `( mon:Moniker keyname:s -- b )` | replace `$keyname` in moniker; never pushes the secret |

```forth
"POSTGRES_HOST" Env.get^
if
  // val:s on stack
  $ nl
else
  "POSTGRES_HOST not set" $ nl
then
```

Empty string is a **found** value (`"" true`), distinct from missing (`false`).

### `.env` grammar (cwd)

| Line | Meaning |
| ---- | ------- |
| blank | ignored |
| `#` … or `//` … | comment |
| `export VAR=value` | same as `VAR=value` |
| `VAR=value` | binding (split on first `=`; trim key and value) |

Double-quoted values keep interior/trailing spaces after the quotes are removed. No `${…}` expansion, no multiline values. Duplicate keys: earlier binding wins. Missing `.env` is normal — lookup falls through to the process environment.

---

## Moniker

Opaque string handle. Create from a clear template; fill placeholders with `Env.subMoniker^`. Final never prints the payload — diagnostics show `Moniker(<opaque>)`.

| Kind | Name | Effect | Description |
| ---- | ---- | ------ | ----------- |
| class | `new` | `( init:s -- mon:Moniker )` | template or plain string |

Placeholder form: literal `$` + key name (e.g. `$POSTGRES_PWD`). No `${KEY}` form. Call `subMoniker` **once per key**. Unresolved `$…` tokens remain until a source returns `true`.

```forth
"postgres://user:$POSTGRES_PWD@localhost/db" Moniker.new^ >ts
t:peek "POSTGRES_PWD" Env.subMoniker^
if
  // substituted; moniker ready for DbProvider.new^
else
  // try another source, or error
then
```

`Env.subMoniker^` returns **`false`** when the key is missing and leaves the moniker unchanged — suitable for a waterfall of sources.

---

## Use as DSN

`DbProvider.new^` accepts **`String` or `Moniker`** for the DSN. Trusted Go code unwraps the moniker internally; the resolved string never lands on the Final data stack.

```forth
`postgres://$POSTGRES_UNAME@localhost:5432/final_demo?sslmode=disable&password=$POSTGRES_PWD`
const: hpDsn

: createDsnMoniker
  hpDsn Moniker.new^ >ts
  t:peek "POSTGRES_UNAME" Env.subMoniker^ >xs
  t:peek "POSTGRES_PWD" Env.subMoniker^
  xs> && if
    ts>
  else
    t:drop
    "Could not substitute required Postgres parameters" error
  then
;

DbProvider.postgres createDsnMoniker DbProvider.new^
```

Prefer this path over `Env.get^` when the value must stay out of scripts, logs, and `$` / JSON dumps.

---

## When to use which

| Need | API |
| ---- | --- |
| Read a non-secret config string in Final | `Env.get^` |
| Build a DSN / URL that embeds secrets | `Moniker.new^` + `Env.subMoniker^` |
| Pass credentials into `DbProvider` | Moniker (or String if you accept cleartext on DS) |

---

## Non-goals

Env does not write the process environment, walk parent directories for `.env`, or interpolate variables inside `.env` values. Moniker does not expose accessors that return the current string to Final.

See also: [SQL Databases](database.md).
