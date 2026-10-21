# SQL Databases

Final’s SQL API is **provider-independent**: one type, **`DbProvider`**, wraps Go [`database/sql`](https://pkg.go.dev/database/sql). A class constant names the driver; the same methods work for every engine.

Supported drivers:

| Constant | Engine | Notes |
| -------- | ------ | ----- |
| `DbProvider.postgres` | Postgres / AWS RDS (pgx) | Placeholders `$1`, `$2`, … |
| `DbProvider.mysql` | MySQL / MariaDB | Placeholders `?`; prefer `parseTime=true` in the DSN |
| `DbProvider.sqlite` | SQLite (modernc, CGo-free) | DSN is a path or `":memory:"` |

There is no ORM and no query builder — you write SQL, bind ordinal args from a `List`, and consume rows.

The **`final`** and **`final-go`** binaries include all three drivers. **`finalgo` / `finalembed`** do not import SQL drivers; the embedding host declares what it needs. The Pi pack includes SQLite.

Opening a connection is gated by **`SR.Db`**. SQLite also needs **`SR.FileRead`** / **`SR.FileWrite`** because it opens a file.

---

## Connection

```forth
DbProvider.postgres dsn DbProvider.new^
// ( driver:s dsn:s|Moniker -- db:DbProvider true | err:s false )
```

| Kind | Name | Effect | Description |
| ---- | ---- | ------ | ----------- |
| class prop | `postgres` / `mysql` / `sqlite` | `( -- driver:s )` | driver name for `new` |
| class | `new` | `( driver:s dsn:s\|Moniker -- db true \| err:s false )` | `sql.Open` + Ping |

The DSN may be a **`String`** or a **`Moniker`** (opaque handle — see [Env and Moniker](env.md)). Keep real passwords out of committed source.

### Postgres with Moniker

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

: connectDb
  DbProvider.postgres createDsnMoniker
  DbProvider.new^
  nif
    dup "Can't open DbProvider: " $ $ nl
    error
  then
;
```

### MySQL / SQLite sketches

```forth
DbProvider.mysql
"user:pass@tcp(127.0.0.1:3306)/final_demo?parseTime=true"
DbProvider.new^

DbProvider.sqlite ":memory:" DbProvider.new^
DbProvider.sqlite "/tmp/demo.db" DbProvider.new^
```

---

## Exec and query

| Kind | Name | Effect | Description |
| ---- | ---- | ------ | ----------- |
| method | `exec` | `( sql:s -- )` | DDL / DML, no binds; panics on error |
| method | `execWithArgs` | `( sql:s args:List -- )` | ordinal binds; panics on error |
| method | `query` | `( sql:s -- iter:DbIter )` | result iterator |
| method | `queryWithArgs` | `( sql:s args:List -- iter:DbIter )` | query with binds |

Args are a **`List` in ordinal order** — build with `( )`. Use a **`ByteList`** for BLOB columns. Placeholder syntax follows the driver (`$N` vs `?`); always pass values by position, not by name.

```forth
: insertCharList   // ( args:List -- )
`
INSERT INTO "hpCharacters" ("firstName", "lastName", "isGood", "firstMovieSeenYear")
VALUES ($1, $2, $3, $4)
`
  hpDb @ >ts
    swap t:.execWithArgs^
  t:drop
;

( "Harry" "Potter" true 2001 ) insertCharList
```

---

## Result rows

`.query^` / `.queryWithArgs^` return a **`DbIter`**: foreach-compatible. Each loop value `i` is a **`StaticObject`** whose fields are the result-set column names (creation order matches the SELECT), minted from the iterator’s shared prototype (`iter.protoObj`). Cell types follow the driver’s Scan values, coerced to Final (`Int`, `String`, `Bool`, `Float`, `ByteList`, `DateTime` where applicable, …). SQL NULL is Final **`Null`**.

Row objects share one schema prototype DO built from the column list. For fixed-schema in-memory grids, see [StaticObject](compounds/staticobject.md) and [Table](compounds/table.md).

```forth
: queryAll
  hpDb @ >ts
`
SELECT "id", "firstName", "lastName", "isGood", "firstMovieSeenYear"
FROM "hpCharacters"
ORDER BY "firstMovieSeenYear", "lastName", "firstName"
`
    t:.query^
  t:drop

  foreach
    i >ts
      ( t:.firstName@ t:.lastName@ t:.firstMovieSeenYear@ )
      "name: %s %s  year first seen: %d\n" fmt $
    t:drop
  /foreach
;
```

### Column names and anonymous queries

Unaliased expressions get **provider-supplied** names. Inspect them with:

```forth
iter .getColumnNames^   // ( -- names:StringList )
```

Prefer explicit aliases in SQL when you want stable Final field names:

```sql
SELECT count(*) AS "n", sum("firstMovieSeenYear") AS "yearSum" ...
```

### NULL helpers

| Word | Effect | Description |
| ---- | ------ | ----------- |
| `val?` | `( val -- val true \| false )` | false if `Null` |
| `val\|def` | `( val default -- val \| default )` | substitute default for `Null` |

---

## Materialize with `iter>list` and `dbIter>table`

These are core utility words (not something you have to define):

```forth
iter>list      // ( iterable -- l:List )
dbIter>table   // ( iter:DbIter -- tbl:Table )
```

`iter>list` drains any `foreach`-compatible source into a **List**. For a query, that is a List of row **StaticObjects**.

```forth
sql hpDb @ .query^ iter>list >jsonp
```

`dbIter>table` builds a **Table** from a query iterator: schema from the iterator’s DynamicObject prototype, rows as StaticObjects. See [Table](compounds/table.md).

```forth
sql hpDb @ .query^ dbIter>table >ts
t:.rows >jsonp $
t:drop
```

---

## Notes

* `DbProvider.new^` is **modal**; `exec` / `query` **panic** if the SQL cannot run.
* Empty result: `foreach` runs zero times; `iter>list` yields an empty `List`; `dbIter>table` yields a Table with schema and no rows.
* Full walkthrough: `examples/db/postgres/harry-potter.fnl`.
* Discoverability: `help DbProvider` in the TUI.

Related: [Env and Moniker](env.md).
