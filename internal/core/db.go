package core

import (
	"database/sql"
)

// SQL drivers are blank-imported by the host, not by core: cmd/final-go and
// cmd/final link mysql+pgx+sqlite; cmd/final-pi links sqlite only.
// Embedders of finalgo/finalembed import the engines they need (or none).

// DbCellCoercer maps a post–layer-A cell value to a Final cell (provider second pass).
type DbCellCoercer func(v any) any

// Provider-specific second-pass maps keyed by sql.ColumnType.DatabaseTypeName().
// nil / missing key → leave layer-A result. TINYINT omitted on purpose (BOOLEAN synonym
// is indistinguishable; leave as Int; Truthy at call sites).
var mysqlDbCellCoercers = map[string]DbCellCoercer{
	"VARCHAR": coerceDbTypeToFinalType,
	"CHAR":    coerceDbTypeToFinalType,
	"TEXT":    coerceDbTypeToFinalType,
	"DECIMAL": coerceDbTypeToFinalType,
	"TIME":    coerceDbTypeToFinalType,
	"JSON":    coerceDbTypeToFinalType,
	// BLOB / BINARY: stay ByteList (layer-A)
}

type DbProvider struct {
	db       *sql.DB
	coercers map[string]DbCellCoercer // nil when driver needs only layer-A
}

func NewDbProvider(driverName, dsn string) (*DbProvider, error) {
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, err
	}
	// sql.Open does not dial; Ping fails fast on bad host/auth/unreachable DB.
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	p := &DbProvider{db: db}
	switch driverName {
	case "mysql":
		p.coercers = mysqlDbCellCoercers
	}
	return p, nil
}

// Exec runs query with no parameters. Panics if the statement cannot run.
func (p *DbProvider) Exec(query string) {
	if _, err := p.db.Exec(query); err != nil {
		panic(err)
	}
}

// ExecWithArgs runs query with ordinal args from a Final List. Panics on error.
func (p *DbProvider) ExecWithArgs(query string, args *AnyList) {
	if _, err := p.db.Exec(query, listAsSQLArgs(args)...); err != nil {
		panic(err)
	}
}

// Query runs a query with no parameters. Panics on error.
func (p *DbProvider) Query(query string) *DbIter {
	rows, err := p.db.Query(query)
	if err != nil {
		panic(err)
	}
	return NewDbIter(rows, p.coercers)
}

// QueryWithArgs runs a query with ordinal args from a Final List. Panics on error.
func (p *DbProvider) QueryWithArgs(query string, args *AnyList) *DbIter {
	rows, err := p.db.Query(query, listAsSQLArgs(args)...)
	if err != nil {
		panic(err)
	}
	return NewDbIter(rows, p.coercers)
}

func (p *DbProvider) FinalType() *Type {
	return dbProviderType
}

// listAsSQLArgs converts a Final List of bind args to []any for database/sql.
// ByteList cells become the inner []uint8 (BLOB-friendly).
func listAsSQLArgs(args *AnyList) []any {
	if args == nil {
		return nil
	}
	n := args.Len()
	out := make([]any, n)
	for i := 0; i < n; i++ {
		v := args.AtAsAny(i)
		switch x := v.(type) {
		case *ByteList:
			out[i] = x.items
		default:
			out[i] = v
		}
	}
	return out
}

// coerceCellLayerA is the shared Scan→Final pass (all drivers).
// Integral widths from drivers are normalized to Final Int / Uint (Go optimizes these switches).
func coerceCellLayerA(v any) any {
	if v == nil {
		return nil
	}
	switch x := v.(type) {
	case int8:
		return Int(x)
	case uint8:
		return Int(x)
	case int16:
		return Int(x)
	case uint16:
		return Int(x)
	case int32:
		return Int(x)
	case uint32:
		return Int(x)
	case int64:
		return Int(x)
	case int:
		return Int(x)
	case uint64:
		return Uint(x)
	case uint:
		return Uint(x)
	case float32:
		return Float(x)
	case float64:
		return Float(x)
	case []byte:
		// Caller owns the slice; ByteList takes ownership.
		return NewByteListFromBytes(x)
	case bool:
		return Bool(x)
	case string:
		return String(x)
	default:
		return v
	}
}

// coerceDbTypeToFinalType is a provider second-pass coercer: layer-A cell → Final cell.
// Input may be ByteList, raw []byte, String, or later other shapes (e.g. string→DateTime).
// Current MySQL text-ish entries: bytes / ByteList / String → String.
func coerceDbTypeToFinalType(v any) any {
	switch x := v.(type) {
	case *ByteList:
		return String(string(x.items))
	case []byte:
		return String(string(x))
	case String:
		return x
	default:
		return v
	}
}

type DbIter struct {
	rows        *sql.Rows
	scanVals    []any
	scanPtrs    []any // each element is *any pointing at scanVals[i]
	protoObj    *DynamicObject
	currObj     *StaticObject
	coercers    map[string]DbCellCoercer // nil ⇒ no second pass; no ColumnTypes fetch
	dbTypeNames []string                 // DatabaseTypeName per ordinal; only if coercers != nil
}

func NewDbIter(rows *sql.Rows, coercers map[string]DbCellCoercer) *DbIter {
	it := &DbIter{rows: rows, coercers: coercers}
	cols, err := it.rows.Columns()
	if err != nil {
		panic(err)
	}
	n := len(cols)
	// Create the Scan Values and Pointers
	// These get reused for each row
	it.scanVals = make([]any, n)
	it.scanPtrs = make([]any, n)
	for i := range it.scanVals {
		it.scanPtrs[i] = &it.scanVals[i]
	}

	// Prototype DO: column names in SELECT order; each row is NewStaticObject(proto)
	fields := NewSFM[any](n)
	for _, name := range cols {
		fields.Add(name, nil)
	}
	it.protoObj = &DynamicObject{fields: fields}

	// Column type names only when a provider second pass exists.
	if coercers != nil {
		cts, err := it.rows.ColumnTypes()
		if err != nil {
			panic(err)
		}
		it.dbTypeNames = make([]string, n)
		for i, ct := range cts {
			it.dbTypeNames[i] = ct.DatabaseTypeName()
		}
	}
	return it
}

func (it *DbIter) GetNext() bool {
	hasNext := it.rows.Next()
	if !hasNext {
		it.currObj = nil
		return false
	}
	// Scan the row into the Scan Values
	if err := it.rows.Scan(it.scanPtrs...); err != nil {
		panic(err)
	}

	// Mint a StaticObject for the current row from the shared prototype
	it.currObj = NewStaticObject(it.protoObj)

	// Layer-A: shared driver any → Final cell
	for i, val := range it.scanVals {
		it.currObj.values[i] = coerceCellLayerA(val)
	}

	// Provider second pass: DatabaseTypeName → coercer
	if it.coercers != nil {
		for i := range it.scanVals {
			fn := it.coercers[it.dbTypeNames[i]]
			if fn == nil {
				continue
			}
			it.currObj.values[i] = fn(it.currObj.values[i])
		}
	}
	return true
}

func (it *DbIter) Current() Any {
	return it.currObj
}

func (it *DbIter) FinalType() *Type {
	return dbIterType
}

func (it *DbIter) GetColumnNames() *StringList {
	cols, err := it.rows.Columns()
	if err != nil {
		panic(err)
	}
	return NewStringListFromSlice(cols)
}

func registerDbIterMembers(t *Type) {
	t.registerProperty("protoObj", "DynamicObject",
		func(inst Any) Any { return inst.(*DbIter).protoObj },
		nil,
	)
	t.registerNativeMethod("getColumnNames", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(ths.Obj.(*DbIter).GetColumnNames())
	})
}

func CreateRegisterDbIterType() {
	t := &Type{
		name:            "DbIter",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[DbIter]("DbIter"),
	}
	registerDbIterMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	dbIterType = t
}

var dbIterType *Type

func registerDbProviderMembers(t *Type) {
	t.registerClassProperty("postgres", "String",
		func(_ Any) Any { return String("pgx") },
		nil,
	)
	t.registerClassProperty("mysql", "String",
		func(_ Any) Any { return String("mysql") },
		nil,
	)
	t.registerClassProperty("sqlite", "String",
		func(_ Any) Any { return String("sqlite") },
		nil,
	)
	// ( driverName:s dsn:s|Moniker -- db:DbProvider true | err:s false )
	// SR applied dynamically: SR_Db always; sqlite also FileRead|FileWrite (creates/opens a file).
	t.registerNativeClassMethod("new", func(f *RunFrame, _ *TypeObjPair) {
		dsn := dsnFromStackCell(f.RP.DataStack.Pop())
		driverName := string(f.RP.DataStack.Pop().(String))
		need := SR_Db
		if driverName == "sqlite" {
			need |= SR_FileRead | SR_FileWrite
		}
		assertMemberRestrict(f, need)
		p, err := NewDbProvider(driverName, dsn)
		pushStreamOpenResult(f, p, err)
	})
	// ( sql:s -- )
	t.registerNativeMethod("exec", func(f *RunFrame, ths *TypeObjPair) {
		query := string(f.RP.DataStack.Pop().(String))
		ths.Obj.(*DbProvider).Exec(query)
	})
	// ( sql:s args:List -- )
	t.registerNativeMethod("execWithArgs", func(f *RunFrame, ths *TypeObjPair) {
		args := f.RP.DataStack.Pop().(*AnyList)
		query := string(f.RP.DataStack.Pop().(String))
		ths.Obj.(*DbProvider).ExecWithArgs(query, args)
	})
	// ( sql:s -- iter:DbIter )
	t.registerNativeMethod("query", func(f *RunFrame, ths *TypeObjPair) {
		query := string(f.RP.DataStack.Pop().(String))
		f.RP.DataStack.Push(ths.Obj.(*DbProvider).Query(query))
	})
	// ( sql:s args:List -- iter:DbIter )
	t.registerNativeMethod("queryWithArgs", func(f *RunFrame, ths *TypeObjPair) {
		args := f.RP.DataStack.Pop().(*AnyList)
		query := string(f.RP.DataStack.Pop().(String))
		f.RP.DataStack.Push(ths.Obj.(*DbProvider).QueryWithArgs(query, args))
	})
}

func CreateRegisterDbProviderType() {
	t := &Type{
		name:            "DbProvider",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[DbProvider]("DbProvider"),
	}
	registerDbProviderMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	dbProviderType = t
}

var dbProviderType *Type

func init() {
	CreateRegisterDbProviderType()
	CreateRegisterDbIterType()
}

var _ Typed = (*DbProvider)(nil)
var _ Typed = (*DbIter)(nil)
var _ Iterator = (*DbIter)(nil)
