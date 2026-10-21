package fyneui

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/androticus/final/internal/core"
)

// TableRef holds a Final Table for Fyne TableView callbacks (static data path).
type TableRef struct {
	table *core.Table
}

func NewTableRef(table *core.Table) *TableRef {
	return &TableRef{table: table}
}

func (tr *TableRef) GetTable() *core.Table {
	return tr.table
}

func (tr *TableRef) SetTable(table *core.Table) {
	tr.table = table
}

func (tr *TableRef) RowCount() int {
	if tr.table == nil {
		return 0
	}
	return tr.table.Len()
}

func (tr *TableRef) ColCount() int {
	if tr.table == nil {
		return 0
	}
	return tr.table.ColCount()
}

// CellString returns display text for a Fyne TableCellID.
// Row == -1 → column name; Col == -1 → 1-based row label; both negative → "".
func (tr *TableRef) CellString(row, col int) string {
	if tr.table == nil {
		return ""
	}
	if row < 0 && col < 0 {
		return ""
	}
	if row < 0 {
		names := tr.table.ColumnNames()
		if col < 0 || col >= names.Len() {
			return ""
		}
		return names.At(col)
	}
	if col < 0 {
		return strconv.Itoa(row + 1)
	}
	return tr.table.CellAsString(row, col)
}

func NewTableView(tr *TableRef, headers bool) *widget.Table {
	length := func() (int, int) {
		return tr.RowCount(), tr.ColCount()
	}
	createCell := func() fyne.CanvasObject {
		return widget.NewLabel("width ****") // acts as col width setter
	}
	update := func(id widget.TableCellID, obj fyne.CanvasObject) {
		obj.(*widget.Label).SetText(tr.CellString(id.Row, id.Col))
	}

	tv := widget.NewTable(length, createCell, update)
	if headers {
		tv.ShowHeaderRow = true
		//tv.ShowHeaderColumn = true
		tv.CreateHeader = func() fyne.CanvasObject {
			return widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		}
		tv.UpdateHeader = update
	}
	return tv
}

func registerTableRefMembers(t *core.Type) {
	t.RegisterProperty("table", "Table",
		func(inst any) any {
			return inst.(*TableRef).GetTable()
		},
		func(inst any, value any) {
			inst.(*TableRef).SetTable(value.(*core.Table))
		},
	)

	// ( table:Table -- ref:TableRef )
	t.RegisterNativeClassMethod("new", func(f *core.RunFrame, _ *core.TypeObjPair) {
		table := f.RP.DataStack.Pop().(*core.Table)
		f.RP.DataStack.Push(NewTableRef(table))
	})
}

func CreateRegisterTableRefType() {
	t := core.NewRegisteringType("TableRef",
		core.MakeHostTypeZero[*TableRef](),
		core.MakeHostTypeCheckCoerce[*TableRef]("TableRef"),
	)
	registerTableRefMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	core.RegisterHostType[*TableRef](t)
}

func registerTableViewMembers(t *core.Type) {
	bindSelected := func(f *core.RunFrame, ths *core.TypeObjPair, selected bool) {
		index := int(f.RP.DataStack.Pop().(core.Int))
		handler := f.RP.DataStack.Pop().(*core.MethodRef)
		tv := ths.Obj.(*widget.Table)
		eventFn := core.GetEventLoop().CreateEventFn(handler, index)
		fyne.Do(func() {
			fn := func(id widget.TableCellID) {
				eventFn(tv, core.Int(id.Row), core.Int(id.Col))
			}
			if selected {
				tv.OnSelected = fn
			} else {
				tv.OnUnselected = fn
			}
		})
	}

	// ( handler:MethodRef index:Int -- )
	t.RegisterNativeMethod("setOnSelected", func(f *core.RunFrame, ths *core.TypeObjPair) {
		bindSelected(f, ths, true)
	})
	// ( handler:MethodRef index:Int -- )
	t.RegisterNativeMethod("setOnUnselected", func(f *core.RunFrame, ths *core.TypeObjPair) {
		bindSelected(f, ths, false)
	})

	// Column widths are Fyne device-independent pixels (≈ CSS px at 100% scale).
	// CreateCell's template sets one default for every column; these override per index.
	// Reasonable starting points (tune by eye — theme/DPI change the feel):
	//   60–80   narrow ints / ages / flags
	//   100–160 short names / labels
	//   200–280 longer text / paths
	// Example: ( 120 60 i) tv.setColumnWidths^
	// ( col:Int width:Int -- )
	t.RegisterNativeMethod("setColumnWidth", func(f *core.RunFrame, ths *core.TypeObjPair) {
		width := float32(f.RP.DataStack.Pop().(core.Int))
		col := int(f.RP.DataStack.Pop().(core.Int))
		tv := ths.Obj.(*widget.Table)
		fyne.Do(func() {
			tv.SetColumnWidth(col, width)
		})
	})
	// list[i] → column i width. Build with i): ( 120 60 i)
	// ( widths:IntList -- )
	t.RegisterNativeMethod("setColumnWidths", func(f *core.RunFrame, ths *core.TypeObjPair) {
		widths := f.RP.DataStack.Pop().(*core.IntList)
		tv := ths.Obj.(*widget.Table)
		fyne.Do(func() {
			for i := 0; i < widths.Len(); i++ {
				tv.SetColumnWidth(i, float32(widths.At(i)))
			}
		})
	})

	// ( ref:TableRef headers:Bool -- w:TableView )
	t.RegisterNativeClassMethod("new", func(f *core.RunFrame, _ *core.TypeObjPair) {
		headers := bool(f.RP.DataStack.Pop().(core.Bool))
		ref := f.RP.DataStack.Pop().(*TableRef)
		f.RP.DataStack.Push(NewTableView(ref, headers))
	})
}

func CreateRegisterTableViewType() {
	t := core.NewRegisteringType("TableView",
		core.MakeHostTypeZero[*widget.Table](),
		core.MakeHostTypeCheckCoerce[*widget.Table]("TableView"),
	)
	registerTableViewMembers(t)
	applyCanvasMixin[*widget.Table](t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	core.RegisterHostType[*widget.Table](t)
}

func init() {
	CreateRegisterTableRefType()
	CreateRegisterTableViewType()
}
