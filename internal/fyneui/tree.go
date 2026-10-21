package fyneui

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/androticus/final/internal/core"
)

// TreeRef holds a FyneTree for TreeView callbacks (static data path).
type TreeRef struct {
	tree *FyneTree
}

func NewTreeRef(tree *FyneTree) *TreeRef {
	return &TreeRef{tree: tree}
}

func (tr *TreeRef) GetTree() *FyneTree {
	return tr.tree
}

func (tr *TreeRef) SetTree(tree *FyneTree) {
	tr.tree = tree
}

func (tr *TreeRef) node(uid widget.TreeNodeID) *FyneTreeNode {
	if tr.tree == nil {
		return nil
	}
	id, err := strconv.Atoi(uid)
	if err != nil {
		return nil
	}
	return tr.tree.Index[id]
}

func (tr *TreeRef) ChildUIDs(uid widget.TreeNodeID) []widget.TreeNodeID {
	n := tr.node(uid)
	if n == nil || !n.IsBranch() {
		return nil
	}
	out := make([]widget.TreeNodeID, len(n.Children))
	for i, c := range n.Children {
		out[i] = strconv.Itoa(c.ID)
	}
	return out
}

func (tr *TreeRef) IsBranch(uid widget.TreeNodeID) bool {
	n := tr.node(uid)
	return n != nil && n.IsBranch()
}

func (tr *TreeRef) Label(uid widget.TreeNodeID) string {
	n := tr.node(uid)
	if n == nil {
		return ""
	}
	return n.Value
}

func NewTreeView(tr *TreeRef) *widget.Tree {
	tv := widget.NewTree(
		tr.ChildUIDs,
		tr.IsBranch,
		func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(uid widget.TreeNodeID, branch bool, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(tr.Label(uid))
		},
	)
	tv.Root = "0"
	return tv
}

func registerTreeRefMembers(t *core.Type) {
	t.RegisterProperty("tree", "FyneTree",
		func(inst any) any {
			return inst.(*TreeRef).GetTree()
		},
		func(inst any, value any) {
			inst.(*TreeRef).SetTree(value.(*FyneTree))
		},
	)

	// ( tree:FyneTree -- ref:TreeRef )
	t.RegisterNativeClassMethod("new", func(f *core.RunFrame, _ *core.TypeObjPair) {
		tree := f.RP.DataStack.Pop().(*FyneTree)
		f.RP.DataStack.Push(NewTreeRef(tree))
	})
}

func CreateRegisterTreeRefType() {
	t := core.NewRegisteringType("TreeRef",
		core.MakeHostTypeZero[*TreeRef](),
		core.MakeHostTypeCheckCoerce[*TreeRef]("TreeRef"),
	)
	registerTreeRefMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	core.RegisterHostType[*TreeRef](t)
}

func registerTreeViewMembers(t *core.Type) {
	bindSelected := func(f *core.RunFrame, ths *core.TypeObjPair, selected bool) {
		index := int(f.RP.DataStack.Pop().(core.Int))
		handler := f.RP.DataStack.Pop().(*core.MethodRef)
		tv := ths.Obj.(*widget.Tree)
		eventFn := core.GetEventLoop().CreateEventFn(handler, index)
		fyne.Do(func() {
			fn := func(uid widget.TreeNodeID) {
				id, err := strconv.Atoi(uid)
				if err != nil {
					id = -1
				}
				eventFn(tv, core.Int(id))
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

	// ( ref:TreeRef -- w:TreeView )
	t.RegisterNativeClassMethod("new", func(f *core.RunFrame, _ *core.TypeObjPair) {
		ref := f.RP.DataStack.Pop().(*TreeRef)
		f.RP.DataStack.Push(NewTreeView(ref))
	})
}

func CreateRegisterTreeViewType() {
	t := core.NewRegisteringType("TreeView",
		core.MakeHostTypeZero[*widget.Tree](),
		core.MakeHostTypeCheckCoerce[*widget.Tree]("TreeView"),
	)
	registerTreeViewMembers(t)
	applyCanvasMixin[*widget.Tree](t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	core.RegisterHostType[*widget.Tree](t)
}

func init() {
	CreateRegisterTreeRefType()
	CreateRegisterTreeViewType()
}
