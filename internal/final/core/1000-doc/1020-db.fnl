// SQL database — DbProvider (Go database/sql)

/*
Open SQL connection and run parameterized statements.
Wraps Go database/sql; driver name selects the linked driver
NOTE: note all listed drivers are imported in all runtime
  configurations.
final-go, final-tui cmds import all three
final-tui on Pi imports SQLite only
finalgo, finalembed packages do not import, the embedding app
  must import via: (tested:)
import	_ "github.com/go-sql-driver/mysql" // registers "mysql" driver
import	_ "github.com/jackc/pgx/v5/stdlib" // registers "pgx" driver for database/sql
import	_ "modernc.org/sqlite"             // registers "sqlite" driver (cznic / modernc, cgo-free)
Other drivers for other DBs may work, but have not been tested.  
(use DbProvider.postgres for Postgres / pgx,
DbProvider.mysql for MySQL / MariaDB,
DbProvider.sqlite for SQLite / modernc).
Open is modal; exec and query panic if the SQL cannot run.
query returns a foreach-compatible iterator of row StaticObjects
(column names as fields; SQL NULL is Final Null).
See also list-db helpers: val? val|def iter>list .
Keep real DSNs and passwords out of committed source.
*/
:doc DbProvider

/*
Driver name string for Postgres via pgx/stdlib ("pgx").
Pass as the first argument to DbProvider.new^ .
( -- driver:s )
*/
:doc DbProvider#postgres

/*
Driver name string for MySQL / MariaDB via go-sql-driver ("mysql").
DSN typically user:pass@tcp(host:port)/db?parseTime=true
(parseTime recommended so DATE/DATETIME Scan as DateTime).
Pass as the first argument to DbProvider.new^ .
( -- driver:s )
*/
:doc DbProvider#mysql

/*
Driver name string for SQLite via modernc.org/sqlite ("sqlite").
CGo-free; DSN is a filesystem path (or ":memory:").
Pass as the first argument to DbProvider.new^ .
( -- driver:s )
*/
:doc DbProvider#sqlite

/*
Open a database connection (sql.Open + Ping).
Modal: success leaves the DbProvider and true;
failure leaves an error String and false.
Load USER/PASSWORD (or the whole DSN) from env or
a gitignored config file — not from committed literals.
eg: `` DbProvider.postgres "postgres://…/db?sslmode=disable"
         DbProvider.new^ ``
eg: `` DbProvider.mysql "root@tcp(127.0.0.1:3306)/final_demo?parseTime=true"
         DbProvider.new^ ``
eg: `` DbProvider.sqlite "/tmp/demo.db" DbProvider.new^ ``
( driverName:s dsn:s -- db:DbProvider true | err:s false )
*/
:doc DbProvider#new

/*
Run SQL with no parameters (DDL, INSERT/UPDATE without binds).
Panics if the statement cannot run.
( sql:s -- )
*/
:doc DbProvider#exec

/*
Run SQL with ordinal bind args from a List
(convenience via ( ) ). Panics on error.
Args map to $1, $2, … (Postgres) or ? placeholders
per the driver's parameter style — use ordinal order only.
Pass a ByteList for BLOB columns (unwrapped to []byte).
( sql:s args:List -- )
*/
:doc DbProvider#execWithArgs

/*
Run a query with no parameters.
Returns a foreach-compatible iterator; each i is a
StaticObject with one field per column (creation order).
SQL NULL cells are Final Null — use val? / val|def .
Panics if the query cannot run. Empty result: foreach
runs zero times (or iter>list yields an empty List).
( sql:s -- iter )
*/
:doc DbProvider#query

/*
Run a query with ordinal bind args from a List.
Same row shape and panic rules as query.
( sql:s args:List -- iter )
*/
:doc DbProvider#queryWithArgs
