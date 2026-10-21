package core

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql" // registers "mysql" driver
)

// MySQL / MariaDB probe env (DBngin defaults work with no exports):
//
//	MYSQL_UNAME  optional, default root
//	MYSQL_PWD    optional, default empty
//	MYSQL_HOST   optional, default 127.0.0.1
//	MYSQL_PORT   optional, default 3306
//	MYSQL_DB     optional, default final_demo (created if missing)
//
// DSN always includes parseTime=true so DATE/DATETIME/TIMESTAMP Scan as time.Time.
// See docs/003-design-notes/152.3-db-mysql.md

const mysqlTypeProbeDDL = `
CREATE TABLE mysqlTypeProbe (
  id             INT AUTO_INCREMENT PRIMARY KEY,
  colBoolean     BOOLEAN,
  colTinyInt     TINYINT,
  colSmallInt    SMALLINT,
  colInteger     INT,
  colBigInt      BIGINT,
  colFloat       FLOAT,
  colDouble      DOUBLE,
  colDecimal     DECIMAL(20, 6),
  colText        TEXT,
  colVarchar     VARCHAR(64),
  colChar        CHAR(8),
  colBlob        BLOB,
  colBinary      BINARY(4),
  colDate        DATE,
  colTime        TIME,
  colDatetime    DATETIME,
  colTimestamp   TIMESTAMP NULL,
  colUuid        CHAR(36),
  colJson        JSON
);
`

const mysqlTypeProbeInsertFull = `
INSERT INTO mysqlTypeProbe (
  colBoolean, colTinyInt, colSmallInt, colInteger, colBigInt,
  colFloat, colDouble, colDecimal,
  colText, colVarchar, colChar, colBlob, colBinary,
  colDate, colTime, colDatetime, colTimestamp,
  colUuid, colJson
) VALUES (
  TRUE, 1, 2, 3, 4,
  1.25, 2.5, 123456.789012,
  'text-value', 'varchar-value', 'charval!',
  X'DEADBEEF', X'DEADBEEF',
  '2026-07-27', '14:30:05',
  '2026-07-27 14:30:05', '2026-07-27 14:30:05',
  '550e8400-e29b-41d4-a716-446655440000',
  '{"a": 1}'
);
`

const mysqlTypeProbeInsertNulls = `
INSERT INTO mysqlTypeProbe (
  colBoolean, colTinyInt, colSmallInt, colInteger, colBigInt,
  colFloat, colDouble, colDecimal,
  colText, colVarchar, colChar, colBlob, colBinary,
  colDate, colTime, colDatetime, colTimestamp,
  colUuid, colJson
) VALUES (
  NULL, NULL, NULL, NULL, NULL,
  NULL, NULL, NULL,
  NULL, NULL, NULL, NULL, NULL,
  NULL, NULL, NULL, NULL,
  NULL, NULL
);
`

const mysqlTypeProbeSelect = `
SELECT
  colBoolean, colTinyInt, colSmallInt, colInteger, colBigInt,
  colFloat, colDouble, colDecimal,
  colText, colVarchar, colChar, colBlob, colBinary,
  colDate, colTime, colDatetime, colTimestamp,
  colUuid, colJson
FROM mysqlTypeProbe
ORDER BY id;
`

// Locked from first TestMysqlScanAnyTypes run (go-sql-driver/mysql + parseTime=true).
var mysqlProbeWantGoType = map[string]string{
	"colBoolean":   "int64", // TINYINT(1); not Go bool under Scan→any
	"colTinyInt":   "int64",
	"colSmallInt":  "int64",
	"colInteger":   "int64",
	"colBigInt":    "int64",
	"colFloat":     "float32", // narrower than PG/SQLite REAL→float64
	"colDouble":    "float64",
	"colDecimal":   "[]uint8", // ASCII digits; ByteList until String coerce
	"colText":      "[]uint8", // text arrives as bytes (unlike pgx string)
	"colVarchar":   "[]uint8",
	"colChar":      "[]uint8",
	"colBlob":      "[]uint8",
	"colBinary":    "[]uint8",
	"colDate":      "time.Time",
	"colTime":      "[]uint8", // "14:30:05" as bytes
	"colDatetime":  "time.Time",
	"colTimestamp": "time.Time",
	"colUuid":      "[]uint8", // CHAR(36) text as bytes
	"colJson":      "[]uint8",
}

func mysqlDSNParts() (user, pass, host, port, dbName string) {
	user = os.Getenv("MYSQL_UNAME")
	if user == "" {
		user = "root"
	}
	pass = os.Getenv("MYSQL_PWD") // may be empty
	host = os.Getenv("MYSQL_HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	port = os.Getenv("MYSQL_PORT")
	if port == "" {
		port = "3306"
	}
	dbName = os.Getenv("MYSQL_DB")
	if dbName == "" {
		dbName = "final_demo"
	}
	return user, pass, host, port, dbName
}

func mysqlDSN(user, pass, host, port, dbName string) string {
	// user:pass@tcp(host:port)/db?parseTime=true
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, pass, host, port, dbName)
}

func openMysqlProbe(t *testing.T) *sql.DB {
	t.Helper()
	user, pass, host, port, dbName := mysqlDSNParts()

	// Ensure database exists (DBngin fresh instance may have none of our choosing).
	adminDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/?parseTime=true", user, pass, host, port)
	admin, err := sql.Open("mysql", adminDSN)
	if err != nil {
		t.Skipf("mysql sql.Open (admin): %v — skip MySQL probe", err)
	}
	if err := admin.Ping(); err != nil {
		_ = admin.Close()
		t.Skipf("mysql Ping (admin): %v — skip MySQL probe (is DBngin/MySQL on %s:%s?)", err, host, port)
	}
	if _, err := admin.Exec("CREATE DATABASE IF NOT EXISTS `" + dbName + "`"); err != nil {
		_ = admin.Close()
		t.Fatalf("CREATE DATABASE: %v", err)
	}
	_ = admin.Close()

	db, err := sql.Open("mysql", mysqlDSN(user, pass, host, port, dbName))
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Ping(); err != nil {
		t.Fatalf("Ping: %v", err)
	}
	return db
}

func scanMysqlProbeRows(t *testing.T, db *sql.DB) (cols []string, rowsOut [][]any) {
	t.Helper()
	rows, err := db.Query(mysqlTypeProbeSelect)
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

// TestMysqlScanAnyTypes probes go-sql-driver/mysql Scan→any for the 152.3 matrix.
// Defaults target local DBngin (root, no password, 3306). Skips if server unreachable.
func TestMysqlScanAnyTypes(t *testing.T) {
	db := openMysqlProbe(t)

	if _, err := db.Exec(`DROP TABLE IF EXISTS mysqlTypeProbe`); err != nil {
		t.Fatalf("DROP: %v", err)
	}
	if _, err := db.Exec(mysqlTypeProbeDDL); err != nil {
		t.Fatalf("DDL: %v", err)
	}
	if _, err := db.Exec(mysqlTypeProbeInsertFull); err != nil {
		t.Fatalf("insert full: %v", err)
	}
	if _, err := db.Exec(mysqlTypeProbeInsertNulls); err != nil {
		t.Fatalf("insert nulls: %v", err)
	}

	cols, data := scanMysqlProbeRows(t, db)
	if len(data) != 2 {
		t.Fatalf("row count = %d, want 2 (full + nulls)", len(data))
	}

	t.Log("=== non-null row: column → Go type (Scan into any) ===")
	for i, name := range cols {
		got := goTypeName(data[0][i])
		t.Logf("%s\t%s\t%#v", name, got, data[0][i])
		want := mysqlProbeWantGoType[name]
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

	// Spot-checks once types stabilize.
	if want := mysqlProbeWantGoType["colDatetime"]; want == "time.Time" {
		if _, ok := data[0][indexOfCol(cols, "colDatetime")].(time.Time); !ok {
			t.Errorf("colDatetime = %T, want time.Time", data[0][indexOfCol(cols, "colDatetime")])
		}
	}
}

// TestMysqlExecWithArgsByteList binds a ByteList to a BLOB column via DbProvider.
func TestMysqlExecWithArgsByteList(t *testing.T) {
	user, pass, host, port, dbName := mysqlDSNParts()
	adminDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/?parseTime=true", user, pass, host, port)
	admin, err := sql.Open("mysql", adminDSN)
	if err != nil {
		t.Skipf("mysql unavailable: %v", err)
	}
	if err := admin.Ping(); err != nil {
		_ = admin.Close()
		t.Skipf("mysql unavailable: %v", err)
	}
	_, _ = admin.Exec("CREATE DATABASE IF NOT EXISTS `" + dbName + "`")
	_ = admin.Close()

	p, err := NewDbProvider("mysql", mysqlDSN(user, pass, host, port, dbName))
	if err != nil {
		t.Fatalf("NewDbProvider: %v", err)
	}
	p.Exec(`DROP TABLE IF EXISTS mysqlBlobArgs;`)
	p.Exec(`CREATE TABLE mysqlBlobArgs (id INT AUTO_INCREMENT PRIMARY KEY, colBlob BLOB);`)
	args := NewAnyList()
	args.Append(NewByteListFromBytes([]byte{1, 2, 3, 4}))
	p.ExecWithArgs(`INSERT INTO mysqlBlobArgs (colBlob) VALUES (?);`, args)

	rows, err := p.db.Query(`SELECT colBlob FROM mysqlBlobArgs;`)
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
