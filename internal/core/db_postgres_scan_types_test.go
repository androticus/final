package core

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // registers "pgx" driver for database/sql
)

// Postgres probe env (export from shell profile; never commit secrets):
//
//	POSTGRES_UNAME  required
//	POSTGRES_PWD    required
//	POSTGRES_HOST   optional, default localhost
//	POSTGRES_PORT   optional, default 5432
//	POSTGRES_DB     optional, default final_demo
//
// See docs/003-design-notes/152.1-db-postgres.md

const pgTypeProbeDDL = `
DROP TABLE IF EXISTS "pgTypeProbe";
CREATE TABLE "pgTypeProbe" (
  "id"               SERIAL PRIMARY KEY,
  "colBoolean"       BOOLEAN,
  "colSmallInt"      SMALLINT,
  "colInteger"       INTEGER,
  "colBigInt"        BIGINT,
  "colReal"          REAL,
  "colDouble"        DOUBLE PRECISION,
  "colNumeric"       NUMERIC(20, 6),
  "colText"          TEXT,
  "colVarchar"       VARCHAR(64),
  "colChar"          CHAR(8),
  "colBytea"         BYTEA,
  "colDate"          DATE,
  "colTime"          TIME,
  "colTimestamp"     TIMESTAMP,
  "colTimestamptz"   TIMESTAMPTZ,
  "colUuid"          UUID,
  "colJson"          JSON,
  "colJsonb"         JSONB
);
`

const pgTypeProbeInsertFull = `
INSERT INTO "pgTypeProbe" (
  "colBoolean", "colSmallInt", "colInteger", "colBigInt",
  "colReal", "colDouble", "colNumeric",
  "colText", "colVarchar", "colChar", "colBytea",
  "colDate", "colTime", "colTimestamp", "colTimestamptz",
  "colUuid", "colJson", "colJsonb"
) VALUES (
  true, 1, 2, 3, 1.25, 2.5, 123456.789012,
  'text-value', 'varchar-value', 'charval!',
  E'\\xDEADBEEF',
  DATE '2026-07-27', TIME '14:30:05',
  TIMESTAMP '2026-07-27 14:30:05',
  TIMESTAMPTZ '2026-07-27 14:30:05-06',
  '550e8400-e29b-41d4-a716-446655440000',
  '{"a": 1}', '{"b": 2}'
);
`

const pgTypeProbeInsertNulls = `INSERT INTO "pgTypeProbe" DEFAULT VALUES;`

const pgTypeProbeSelect = `
SELECT
  "colBoolean", "colSmallInt", "colInteger", "colBigInt",
  "colReal", "colDouble", "colNumeric",
  "colText", "colVarchar", "colChar", "colBytea",
  "colDate", "colTime", "colTimestamp", "colTimestamptz",
  "colUuid", "colJson", "colJsonb"
FROM "pgTypeProbe"
ORDER BY "id";
`

// wantGoType is fmt %T of a non-nil Scan-into-any cell.
// Empty string = observe/log only (no assert) — fill after first run if needed.
// Some PG types commonly arrive as []byte; those stay ByteList after DbIter coerce
// until we add typed handling (e.g. UUID).
var pgProbeWantGoType = map[string]string{
	"colBoolean":     "bool",
	"colSmallInt":    "int64",
	"colInteger":     "int64",
	"colBigInt":      "int64",
	"colReal":        "float64",
	"colDouble":      "float64",
	"colNumeric":     "string",
	"colText":        "string",
	"colVarchar":     "string",
	"colChar":        "string",
	"colBytea":       "[]uint8",
	"colDate":        "time.Time",
	"colTime":        "string", // pgx+any: wall clock as string, not time.Time
	"colTimestamp":   "time.Time",
	"colTimestamptz": "time.Time",
	"colUuid":        "string",
	"colJson":        "[]uint8", // ByteList after DbIter until JSON coerce
	"colJsonb":       "[]uint8", // same
}

func postgresDSNFromEnv(t *testing.T) string {
	t.Helper()
	user := os.Getenv("POSTGRES_UNAME")
	pass := os.Getenv("POSTGRES_PWD")
	if user == "" || pass == "" {
		t.Skip("POSTGRES_UNAME and POSTGRES_PWD not set — skip Postgres probe")
	}
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("POSTGRES_PORT")
	if port == "" {
		port = "5432"
	}
	dbName := os.Getenv("POSTGRES_DB")
	if dbName == "" {
		dbName = "final_demo"
	}
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(user, pass),
		Host:     host + ":" + port,
		Path:     "/" + dbName,
		RawQuery: "sslmode=disable",
	}
	return u.String()
}

func openPostgresProbe(t *testing.T) *sql.DB {
	t.Helper()
	dsn := postgresDSNFromEnv(t)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Ping(); err != nil {
		t.Fatalf("Ping: %v", err)
	}
	return db
}

func scanAllRows(t *testing.T, db *sql.DB) (cols []string, rowsOut [][]any) {
	t.Helper()
	rows, err := db.Query(pgTypeProbeSelect)
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	defer rows.Close()

	cols, err = rows.Columns()
	if err != nil {
		t.Fatalf("Columns: %v", err)
	}
	n := len(cols)
	for rows.Next() {
		vals := make([]any, n)
		ptrs := make([]any, n)
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		// copy slot values (Scan wrote into vals)
		row := make([]any, n)
		copy(row, vals)
		rowsOut = append(rowsOut, row)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err: %v", err)
	}
	return cols, rowsOut
}

func goTypeName(v any) string {
	if v == nil {
		return "nil"
	}
	return fmt.Sprintf("%T", v)
}

// TestPostgresScanAnyTypes probes pgx Scan→any types for the 152.1 matrix.
// Requires local Postgres + POSTGRES_UNAME / POSTGRES_PWD. No secrets in source.
func TestPostgresScanAnyTypes(t *testing.T) {
	db := openPostgresProbe(t)

	if _, err := db.Exec(pgTypeProbeDDL); err != nil {
		t.Fatalf("DDL: %v", err)
	}
	if _, err := db.Exec(pgTypeProbeInsertFull); err != nil {
		t.Fatalf("insert full: %v", err)
	}
	if _, err := db.Exec(pgTypeProbeInsertNulls); err != nil {
		t.Fatalf("insert nulls: %v", err)
	}

	cols, data := scanAllRows(t, db)
	if len(data) != 2 {
		t.Fatalf("row count = %d, want 2 (full + nulls)", len(data))
	}

	t.Log("=== non-null row: column → Go type (Scan into any) ===")
	for i, name := range cols {
		got := goTypeName(data[0][i])
		t.Logf("%s\t%s", name, got)
		want := pgProbeWantGoType[name]
		if want == "" {
			continue // observe-only
		}
		if got != want {
			t.Errorf("%s: Go type = %q, want %q", name, got, want)
		}
	}

	t.Log("=== null row: every payload cell must be nil ===")
	for i, name := range cols {
		if data[1][i] != nil {
			t.Errorf("%s: null row cell = %T %v, want nil", name, data[1][i], data[1][i])
		}
	}

	// Spot-check a few values so we know Scan is not empty shells.
	if b, ok := data[0][0].(bool); !ok || !b {
		t.Errorf("colBoolean = %T %v, want bool true", data[0][0], data[0][0])
	}
	if _, ok := data[0][indexOfCol(cols, "colTimestamp")].(time.Time); !ok {
		t.Errorf("colTimestamp = %T, want time.Time (or revise matrix)", data[0][indexOfCol(cols, "colTimestamp")])
	}
}

func indexOfCol(cols []string, name string) int {
	for i, c := range cols {
		if c == name {
			return i
		}
	}
	return 0
}
