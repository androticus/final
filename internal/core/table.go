package core

import (
	"cmp"
	"fmt"
	"slices"
	"time"
)

// Table is a named-column row collection. Column schema is fixed at construction;
// rows may be appended afterward. Sort is stable (successive sorts compose).
type Table struct {
	columnNames        *StringList
	rows               *List[any]
	canAccessOrdinally bool
	getByOrdinal       GetByOrdinalFn // used when canAccessOrdinally
	getByKey           GetByKeyFn     // used when !canAccessOrdinally
}

func NewTableFromColumnNames(columnNames *StringList) *Table {
	// Both accessors: List rows use @(); Map (and name-keyed) rows use @{}.
	// Sort picks from the first row when both are present.
	return &Table{
		columnNames:        columnNames,
		rows:               NewAnyList(),
		canAccessOrdinally: false,
		getByOrdinal:       tableGetByOrdinalViaTyped,
		getByKey:           tableGetByKeyViaTyped,
	}
}

func NewTableFromDynamicObject(do *DynamicObject) *Table {
	return &Table{
		columnNames:        do.GetPropertyNames(),
		rows:               NewAnyList(),
		canAccessOrdinally: true,
		getByOrdinal:       tableGetByOrdinalDOFamily,
	}
}

// NewTableFromTypedObject builds a Table for rows of a TypedObject-derived type.
// t must have TypedObject as its parent and must expose @(); panics otherwise.
// Distinct from NewTableFromType so this path can assume TO shape (ordinal fields,
// and later e.g. stricter append checks).
func NewTableFromTypedObject(t *Type) *Table {
	if t.GetParent() != typedObjectType {
		panic(fmt.Errorf("Table.newFromTypedObject: type %q parent is not TypedObject", t.name))
	}
	im := t.GetItemsMetadata()
	if im == nil || im.getByOrdinal == nil {
		panic(fmt.Errorf("Table.newFromTypedObject: type %q has no @() accessor", t.name))
	}
	return &Table{
		columnNames:        t.GetPropertyNames(),
		rows:               NewAnyList(),
		canAccessOrdinally: true,
		getByOrdinal:       im.getByOrdinal,
	}
}

func NewTableFromType(t *Type) *Table {
	im := t.GetItemsMetadata()
	var getOrd GetByOrdinalFn
	var getKey GetByKeyFn
	canOrd := false
	if im != nil {
		getOrd = im.getByOrdinal
		getKey = im.getByKey
		canOrd = getOrd != nil
	}
	if getKey == nil {
		getKey = mintPropertyGetByKey(t)
	}
	return &Table{
		columnNames:        t.GetPropertyNames(),
		rows:               NewAnyList(),
		canAccessOrdinally: canOrd,
		getByOrdinal:       getOrd,
		getByKey:           getKey,
	}
}

func mintPropertyGetByKey(t *Type) GetByKeyFn {
	return func(inst Any, key Any) any {
		name := string(key.(String))
		pm := t.GetPropertyMetadata(name)
		if pm == nil || pm.get == nil {
			panic(fmt.Errorf("Table @{}: unknown property %q on %s", name, t.name))
		}
		return pm.get(inst)
	}
}

func tableGetByOrdinalDOFamily(inst Any, idx Any) any {
	i := int(idx.(Int))
	switch o := inst.(type) {
	case *DynamicObject:
		return o.ItemAt(i).Value
	case *StaticObject:
		if i < 0 || i >= len(o.values) {
			panic(fmt.Errorf("Table @(): ordinal %d out of range [0,%d)", i, len(o.values)))
		}
		return o.values[i]
	default:
		panic(fmt.Errorf("Table: row type %T not DynamicObject/StaticObject", inst))
	}
}

func tableGetByOrdinalViaTyped(inst Any, idx Any) any {
	typed, ok := inst.(Typed)
	if !ok {
		panic(fmt.Errorf("Table: row type %T is not Typed", inst))
	}
	im := typed.FinalType().GetItemsMetadata()
	if im == nil || im.getByOrdinal == nil {
		panic(fmt.Errorf("Table: type %s has no @() accessor", typed.FinalType().name))
	}
	return im.GetByOrdinal(inst, idx)
}

func tableGetByKeyViaTyped(inst Any, key Any) any {
	typed, ok := inst.(Typed)
	if !ok {
		panic(fmt.Errorf("Table: row type %T is not Typed", inst))
	}
	t := typed.FinalType()
	if im := t.GetItemsMetadata(); im != nil && im.getByKey != nil {
		return im.GetByKey(inst, key)
	}
	name := string(key.(String))
	pm := t.GetPropertyMetadata(name)
	if pm == nil || pm.get == nil {
		panic(fmt.Errorf("Table @{}: unknown property %q on %s", name, t.name))
	}
	return pm.get(inst)
}

func (t *Table) FinalType() *Type { return tableType }

func (t *Table) ColumnNames() *StringList { return t.columnNames }

func (t *Table) Len() int { return t.rows.Len() }

func (t *Table) ColCount() int { return t.columnNames.Len() }

// Append adds a row. Caller must supply a compatible row object.
func (t *Table) Append(item any) {
	t.rows.Append(item)
}

// At returns the row at ordinal index.
func (t *Table) At(row int) any {
	return t.rows.At(row)
}

func (t *Table) Rows() *List[any] {
	return t.rows
}

// SortByOrd stably reorders rows by column ordinal (index into columnNames).
// sortDirection false = ascending (SortAsc), true = descending (SortDesc).
// Cell values must be Bool, Int, Uint, Float, String, or DateTime,
// and both cells in a comparison must share that type.
func (t *Table) SortByOrd(col int, sortDirection bool) {
	if col < 0 || col >= t.columnNames.Len() {
		panic(fmt.Errorf("Table.SortByOrd: column %d out of range [0,%d)", col, t.columnNames.Len()))
	}
	if t.rows.Len() == 0 {
		return
	}
	useOrd := t.canAccessOrdinally && t.getByOrdinal != nil
	if !useOrd && t.getByOrdinal != nil && t.getByKey != nil {
		// Column-names factory: both paths available — follow first row.
		typed, ok := t.rows.At(0).(Typed)
		if !ok {
			panic(fmt.Errorf("Table.SortByOrd: row type %T is not Typed", t.rows.At(0)))
		}
		im := typed.FinalType().GetItemsMetadata()
		useOrd = im != nil && im.getByOrdinal != nil
	}
	var less func(a, b any) int
	if useOrd {
		if t.getByOrdinal == nil {
			panic("Table.SortByOrd: ordinal access but no @() getter")
		}
		less = mintTableOrdinalCompare(col, t.getByOrdinal, sortDirection)
	} else {
		if t.getByKey == nil {
			panic("Table.SortByOrd: name access but no @{} getter")
		}
		less = mintTableNameCompare(t.columnNames.At(col), t.getByKey, sortDirection)
	}
	slices.SortStableFunc(t.rows.items, less)
}

// SortByName finds name in columnNames and delegates to SortByOrd. Panics if not found.
func (t *Table) SortByName(name string, sortDirection bool) {
	for i := 0; i < t.columnNames.Len(); i++ {
		if t.columnNames.At(i) == name {
			t.SortByOrd(i, sortDirection)
			return
		}
	}
	panic(fmt.Errorf("Table.SortByName: column %q not found", name))
}

func mintTableOrdinalCompare(col int, get GetByOrdinalFn, sortDirection bool) func(a, b any) int {
	idx := Int(col)
	return func(a, b any) int {
		c := compareTableCells(get(a, idx), get(b, idx))
		if sortDirection {
			return -c
		}
		return c
	}
}

func mintTableNameCompare(colName string, get GetByKeyFn, sortDirection bool) func(a, b any) int {
	key := String(colName)
	return func(a, b any) int {
		c := compareTableCells(get(a, key), get(b, key))
		if sortDirection {
			return -c
		}
		return c
	}
}

// compareTableCells demuxes supported scalar types and returns cmp result.
// Panics on unsupported types or a/b type mismatch.
func compareTableCells(a, b any) int {
	switch av := a.(type) {
	case Bool:
		bv, ok := b.(Bool)
		if !ok {
			panicTableCellMismatch(a, b)
		}
		return cmp.Compare(boolToOrd(av), boolToOrd(bv))
	case Int:
		bv, ok := b.(Int)
		if !ok {
			panicTableCellMismatch(a, b)
		}
		return cmp.Compare(av, bv)
	case Uint:
		bv, ok := b.(Uint)
		if !ok {
			panicTableCellMismatch(a, b)
		}
		return cmp.Compare(av, bv)
	case Float:
		bv, ok := b.(Float)
		if !ok {
			panicTableCellMismatch(a, b)
		}
		return cmp.Compare(av, bv)
	case String:
		bv, ok := b.(String)
		if !ok {
			panicTableCellMismatch(a, b)
		}
		return cmp.Compare(av, bv)
	case time.Time:
		bv, ok := b.(time.Time)
		if !ok {
			panicTableCellMismatch(a, b)
		}
		return av.Compare(bv)
	default:
		panic(fmt.Errorf("Table.SortByOrd: unsupported cell type %T", a))
	}
}

func boolToOrd(b Bool) int {
	if b {
		return 1
	}
	return 0
}

func panicTableCellMismatch(a, b any) {
	panic(fmt.Errorf("Table.SortByOrd: cell type mismatch %T vs %T", a, b))
}

func (t *Table) GetIter() Iterator {
	return t.rows.GetIter()
}

// useOrdinalAccess reports whether cells should be read via @() for this table
// (and current first row, when both accessors exist).
func (t *Table) useOrdinalAccess() bool {
	useOrd := t.canAccessOrdinally && t.getByOrdinal != nil
	if !useOrd && t.getByOrdinal != nil && t.getByKey != nil && t.rows.Len() > 0 {
		typed, ok := t.rows.At(0).(Typed)
		if !ok {
			return false
		}
		im := typed.FinalType().GetItemsMetadata()
		useOrd = im != nil && im.getByOrdinal != nil
	}
	return useOrd
}

// CellAt returns the value at (row, col). Panics on out-of-range indices.
func (t *Table) CellAt(row, col int) any {
	if row < 0 || row >= t.Len() {
		panic(fmt.Errorf("Table.CellAt: row %d out of range [0,%d)", row, t.Len()))
	}
	if col < 0 || col >= t.ColCount() {
		panic(fmt.Errorf("Table.CellAt: col %d out of range [0,%d)", col, t.ColCount()))
	}
	rowObj := t.At(row)
	if t.useOrdinalAccess() {
		if t.getByOrdinal == nil {
			panic("Table.CellAt: ordinal access but no @() getter")
		}
		return t.getByOrdinal(rowObj, Int(col))
	}
	if t.getByKey == nil {
		panic("Table.CellAt: name access but no @{} getter")
	}
	return t.getByKey(rowObj, String(t.columnNames.At(col)))
}

// CellAsString returns a display string for the cell at (row, col).
func (t *Table) CellAsString(row, col int) string {
	return FormatAnyAsString(t.CellAt(row, col))
}

// FormatAnyAsString demuxes common Final scalar anys for UI / text display.
func FormatAnyAsString(v any) string {
	if v == nil {
		return ""
	}
	switch x := v.(type) {
	case String:
		return string(x)
	case Bool:
		if x {
			return "true"
		}
		return "false"
	case Int:
		return fmt.Sprintf("%d", x)
	case Uint:
		return fmt.Sprintf("%d", x)
	case Float:
		return fmt.Sprintf("%g", x)
	case time.Time:
		return x.Format(time.RFC3339)
	default:
		return fmt.Sprint(v)
	}
}

func CreateRegisterTableType() {
	typ := &Type{
		name:            "Table",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[Table]("Table"),
	}
	typ.registerProperty("columnNames", "StringList", func(inst Any) Any {
		return inst.(*Table).ColumnNames()
	}, nil)
	typ.registerProperty("rows", "List", func(inst Any) Any {
		return inst.(*Table).Rows()
	}, nil)
	typ.registerNativeMethod("append", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*Table).Append(f.RP.DataStack.Pop())
	})
	typ.registerNativeMethod("at", func(f *RunFrame, ths *TypeObjPair) {
		row := int(f.RP.DataStack.Pop().(Int))
		f.RP.DataStack.Push(ths.Obj.(*Table).At(row))
	})
	// ( row:Int col:Int -- s:String )
	typ.registerNativeMethod("cellAsString", func(f *RunFrame, ths *TypeObjPair) {
		col := int(f.RP.DataStack.Pop().(Int))
		row := int(f.RP.DataStack.Pop().(Int))
		f.RP.DataStack.Push(String(ths.Obj.(*Table).CellAsString(row, col)))
	})
	// ( idx:i sortDirection:b -- )
	typ.registerNativeMethod("sortByOrd", func(f *RunFrame, ths *TypeObjPair) {
		dir := bool(f.RP.DataStack.Pop().(Bool))
		col := int(f.RP.DataStack.Pop().(Int))
		ths.Obj.(*Table).SortByOrd(col, dir)
	})
	// ( name:s sortDirection:b -- )
	typ.registerNativeMethod("sortByName", func(f *RunFrame, ths *TypeObjPair) {
		dir := bool(f.RP.DataStack.Pop().(Bool))
		name := string(f.RP.DataStack.Pop().(String))
		ths.Obj.(*Table).SortByName(name, dir)
	})
	typ.registerClassProperty("SortAsc", "Bool", func(_ Any) Any { return Bool(false) }, nil)
	typ.registerClassProperty("SortDesc", "Bool", func(_ Any) Any { return Bool(true) }, nil)
	// ( names:StringList -- table:Table )
	typ.registerNativeClassMethod("newFromColumnNames", func(f *RunFrame, _ *TypeObjPair) {
		f.RP.DataStack.Push(NewTableFromColumnNames(f.RP.DataStack.Pop().(*StringList)))
	})
	// ( proto:DynamicObject -- table:Table )
	typ.registerNativeClassMethod("newFromDynamicObject", func(f *RunFrame, _ *TypeObjPair) {
		f.RP.DataStack.Push(NewTableFromDynamicObject(f.RP.DataStack.Pop().(*DynamicObject)))
	})
	// ( t:Type -- table:Table ); t must be a TypedObject-derived type with @()
	typ.registerNativeClassMethod("newFromTypedObject", func(f *RunFrame, _ *TypeObjPair) {
		f.RP.DataStack.Push(NewTableFromTypedObject(f.RP.DataStack.Pop().(*Type)))
	})
	// ( t:Type -- table:Table )
	typ.registerNativeClassMethod("newFromType", func(f *RunFrame, _ *TypeObjPair) {
		f.RP.DataStack.Push(NewTableFromType(f.RP.DataStack.Pop().(*Type)))
	})
	GetTypeRegistrar().RegisterCurrentType(typ)
	tableType = typ
}

func init() {
	CreateRegisterTableType()
}

var tableType *Type

var (
	_ Typed    = (*Table)(nil)
	_ Iterable = (*Table)(nil)
)
