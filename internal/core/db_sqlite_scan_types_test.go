package core

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite" // registers "sqlite" driver (cznic / modernc, cgo-free)
)

// SQLite STRICT probe — in-process tempfile; no secrets.
// See docs/003-design-notes/152.2-db-sqlite.md

const sqliteTypeProbeDDL = `
DROP TABLE IF EXISTS "sqliteTypeProbe";
CREATE TABLE "sqliteTypeProbe" (
  "id"      INTEGER PRIMARY KEY,
  "colInt"  INT,
  "colReal" REAL,
  "colText" TEXT,
  "colBlob" BLOB,
  "colAny"  ANY
) STRICT;
`

const sqliteTypeProbeInsertFull = `
INSERT INTO "sqliteTypeProbe" (
  "colInt", "colReal", "colText", "colBlob", "colAny"
) VALUES (
  42, 1.25, 'text-value', X'DEADBEEF', 99
);
`

const sqliteTypeProbeInsertNulls = `
INSERT INTO "sqliteTypeProbe" (
  "colInt", "colReal", "colText", "colBlob", "colAny"
) VALUES (
  NULL, NULL, NULL, NULL, NULL
);
`

const sqliteTypeProbeSelect = `
SELECT "colInt", "colReal", "colText", "colBlob", "colAny"
FROM "sqliteTypeProbe"
ORDER BY "id";
`

// wantGoType is fmt %T of a non-nil Scan-into-any cell.
var sqliteProbeWantGoType = map[string]string{
	"colInt":  "int64",
	"colReal": "float64",
	"colText": "string",
	"colBlob": "[]uint8",
	"colAny":  "int64", // seeded as integer 99; ANY follows storage class
}

func openSqliteProbe(t *testing.T) *sql.DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "sqliteTypeProbe.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Ping(); err != nil {
		t.Fatalf("Ping: %v", err)
	}
	return db
}

func scanSqliteProbeRows(t *testing.T, db *sql.DB) (cols []string, rowsOut [][]any) {
	t.Helper()
	rows, err := db.Query(sqliteTypeProbeSelect)
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
		row := make([]any, n)
		copy(row, vals)
		rowsOut = append(rowsOut, row)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err: %v", err)
	}
	return cols, rowsOut
}

// TestSqliteScanAnyTypes probes modernc.org/sqlite Scan→any for STRICT INT/REAL/TEXT/BLOB/ANY.
func TestSqliteScanAnyTypes(t *testing.T) {
	db := openSqliteProbe(t)

	if _, err := db.Exec(sqliteTypeProbeDDL); err != nil {
		t.Fatalf("DDL: %v", err)
	}
	if _, err := db.Exec(sqliteTypeProbeInsertFull); err != nil {
		t.Fatalf("insert full: %v", err)
	}
	if _, err := db.Exec(sqliteTypeProbeInsertNulls); err != nil {
		t.Fatalf("insert nulls: %v", err)
	}

	cols, data := scanSqliteProbeRows(t, db)
	if len(data) != 2 {
		t.Fatalf("row count = %d, want 2 (full + nulls)", len(data))
	}

	t.Log("=== non-null row: column → Go type (Scan into any) ===")
	for i, name := range cols {
		got := goTypeName(data[0][i])
		t.Logf("%s\t%s", name, got)
		want := sqliteProbeWantGoType[name]
		if want == "" {
			continue
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

	if n, ok := data[0][indexOfCol(cols, "colInt")].(int64); !ok || n != 42 {
		t.Errorf("colInt = %T %v, want int64 42", data[0][indexOfCol(cols, "colInt")], data[0][indexOfCol(cols, "colInt")])
	}
	blob, ok := data[0][indexOfCol(cols, "colBlob")].([]byte)
	if !ok || len(blob) != 4 || blob[0] != 0xDE || blob[3] != 0xEF {
		t.Errorf("colBlob = %T %v, want []byte DEADBEEF", data[0][indexOfCol(cols, "colBlob")], data[0][indexOfCol(cols, "colBlob")])
	}
}

// TestListAsSQLArgsByteList unwraps ByteList to []uint8 for BLOB binds.
func TestListAsSQLArgsByteList(t *testing.T) {
	args := NewAnyList()
	args.Append(Int(7))
	args.Append(String("hi"))
	args.Append(NewByteListFromBytes([]byte{0xDE, 0xAD}))
	out := listAsSQLArgs(args)
	if len(out) != 3 {
		t.Fatalf("len = %d, want 3", len(out))
	}
	if out[0] != Int(7) {
		t.Errorf("arg0 = %#v, want Int(7)", out[0])
	}
	if out[1] != String("hi") {
		t.Errorf("arg1 = %#v, want String", out[1])
	}
	b, ok := out[2].([]uint8)
	if !ok {
		t.Fatalf("arg2 = %T, want []uint8", out[2])
	}
	if len(b) != 2 || b[0] != 0xDE || b[1] != 0xAD {
		t.Errorf("arg2 bytes = %v, want DE AD", b)
	}
}

// TestSqliteExecWithArgsByteList binds a ByteList to a STRICT BLOB column via DbProvider.
func TestSqliteExecWithArgsByteList(t *testing.T) {
	path := filepath.Join(t.TempDir(), "blobArgs.db")
	p, err := NewDbProvider("sqlite", path)
	if err != nil {
		t.Fatalf("NewDbProvider: %v", err)
	}
	p.Exec(`CREATE TABLE "t" ("id" INTEGER PRIMARY KEY, "colBlob" BLOB) STRICT;`)
	args := NewAnyList()
	args.Append(NewByteListFromBytes([]byte{1, 2, 3, 4}))
	p.ExecWithArgs(`INSERT INTO "t" ("colBlob") VALUES (?);`, args)

	rows, err := p.db.Query(`SELECT "colBlob" FROM "t";`)
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	defer rows.Close()
	if !rows.Next() {
		t.Fatal("expected one row")
	}
	var got any
	if err := rows.Scan(&got); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	b, ok := got.([]byte)
	if !ok || len(b) != 4 || b[0] != 1 || b[3] != 4 {
		t.Fatalf("colBlob = %T %v, want []byte 1 2 3 4", got, got)
	}
}
