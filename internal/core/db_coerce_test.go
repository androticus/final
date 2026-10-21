package core

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/go-sql-driver/mysql" // registers "mysql" driver
)

func TestCoerceCellLayerA(t *testing.T) {
	if got := coerceCellLayerA(nil); got != nil {
		t.Fatalf("nil = %v", got)
	}
	if got := coerceCellLayerA(int64(42)); got != Int(42) {
		t.Fatalf("int64 = %T %v, want Int 42", got, got)
	}
	if got := coerceCellLayerA(float32(1.25)); got != Float(1.25) {
		t.Fatalf("float32 = %T %v, want Float 1.25", got, got)
	}
	if got := coerceCellLayerA(float64(2.5)); got != Float(2.5) {
		t.Fatalf("float64 = %T %v, want Float 2.5", got, got)
	}
	if got := coerceCellLayerA(true); got != Bool(true) {
		t.Fatalf("bool = %T %v", got, got)
	}
	if got := coerceCellLayerA("hi"); got != String("hi") {
		t.Fatalf("string = %T %v", got, got)
	}
	b := coerceCellLayerA([]byte{1, 2, 3})
	bl, ok := b.(*ByteList)
	if !ok || bl.Len() != 3 || bl.items[0] != 1 {
		t.Fatalf("[]byte = %T %#v", b, b)
	}
}

func TestCoerceDbTypeToFinalType(t *testing.T) {
	if got := coerceDbTypeToFinalType(NewByteListFromBytes([]byte("café"))); got != String("café") {
		t.Fatalf("ByteList = %T %v", got, got)
	}
	if got := coerceDbTypeToFinalType([]byte("x")); got != String("x") {
		t.Fatalf("[]byte = %T %v", got, got)
	}
	if got := coerceDbTypeToFinalType(String("already")); got != String("already") {
		t.Fatalf("String = %T %v", got, got)
	}
}

// TestMysqlDbIterProviderCoerce checks layer-A + MySQL second pass via DbProvider.Query.
func TestMysqlDbIterProviderCoerce(t *testing.T) {
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
	if _, err := admin.Exec("CREATE DATABASE IF NOT EXISTS `" + dbName + "`"); err != nil {
		_ = admin.Close()
		t.Fatalf("CREATE DATABASE: %v", err)
	}
	_ = admin.Close()

	p, err := NewDbProvider("mysql", mysqlDSN(user, pass, host, port, dbName))
	if err != nil {
		t.Fatalf("NewDbProvider: %v", err)
	}
	if p.coercers == nil {
		t.Fatal("mysql provider should have coercers")
	}

	p.Exec(`DROP TABLE IF EXISTS mysqlCoerceProbe`)
	p.Exec(`
CREATE TABLE mysqlCoerceProbe (
  id INT AUTO_INCREMENT PRIMARY KEY,
  colBoolean BOOLEAN,
  colTinyInt TINYINT,
  colText TEXT,
  colVarchar VARCHAR(32),
  colDecimal DECIMAL(10, 2),
  colBlob BLOB,
  colFloat FLOAT,
  colDouble DOUBLE,
  colJson JSON
)`)
	p.Exec(`
INSERT INTO mysqlCoerceProbe (
  colBoolean, colTinyInt, colText, colVarchar, colDecimal, colBlob, colFloat, colDouble, colJson
) VALUES (
  TRUE, 42, 'hello', 'world', 12.34, X'DEADBEEF', 1.25, 2.5, '{"a":1}'
)`)

	it := p.Query(`SELECT colBoolean, colTinyInt, colText, colVarchar, colDecimal, colBlob, colFloat, colDouble, colJson FROM mysqlCoerceProbe`)
	if !it.GetNext() {
		t.Fatal("expected a row")
	}
	row := it.Current().(*StaticObject)

	assertCell := func(name string, wantType string, check func(any) bool) {
		t.Helper()
		v := row.Get(name)
		if !check(v) {
			t.Errorf("%s = %T %#v, want %s", name, v, v, wantType)
		}
	}

	// TINYINT / BOOLEAN stay Int (no Bool coerce — DatabaseTypeName is always TINYINT)
	assertCell("colBoolean", "Int", func(v any) bool { return v == Int(1) })
	assertCell("colTinyInt", "Int", func(v any) bool { return v == Int(42) })

	assertCell("colText", "String", func(v any) bool { return v == String("hello") })
	assertCell("colVarchar", "String", func(v any) bool { return v == String("world") })
	assertCell("colDecimal", "String", func(v any) bool { return v == String("12.34") })
	assertCell("colJson", "String", func(v any) bool {
		s, ok := v.(String)
		if !ok {
			return false
		}
		return string(s) == `{"a": 1}` || string(s) == `{"a":1}`
	})

	assertCell("colBlob", "ByteList", func(v any) bool {
		bl, ok := v.(*ByteList)
		return ok && bl.Len() == 4 && bl.items[0] == 0xDE
	})

	assertCell("colFloat", "Float", func(v any) bool {
		f, ok := v.(Float)
		return ok && f == Float(float32(1.25))
	})
	assertCell("colDouble", "Float", func(v any) bool { return v == Float(2.5) })
}
