package fyneui

import (
	"github.com/androticus/final/internal/core"
)

type FyneTreeNode struct {
	ID       int
	Value    string
	Children []*FyneTreeNode
}

func (t *FyneTreeNode) AddChild(child *FyneTreeNode) {
	if t.Children == nil {
		panic("Cannot add child to a leaf node")
	}
	t.Children = append(t.Children, child)
}

func (t *FyneTreeNode) IsBranch() bool {
	return t.Children != nil
}

type FyneTree struct {
	Root   *FyneTreeNode // ID=0
	Index  map[int]*FyneTreeNode
	nextID int
}

func NewFyneTree() *FyneTree {
	ft := &FyneTree{
		Root: &FyneTreeNode{
			ID: 0, Value: "Root", Children: make([]*FyneTreeNode, 0),
		},
		Index:  make(map[int]*FyneTreeNode),
		nextID: 1,
	}
	ft.Index[0] = ft.Root
	return ft
}

func (ft *FyneTree) NewNode(parent *FyneTreeNode, value string, isBranch bool) *FyneTreeNode {
	td := &FyneTreeNode{ID: ft.nextID, Value: value}
	if isBranch {
		td.Children = make([]*FyneTreeNode, 0)
	}
	ft.Index[td.ID] = td
	ft.nextID++
	parent.AddChild(td)
	return td
}

func (ft *FyneTree) NewBranchNode(parent *FyneTreeNode, value string) *FyneTreeNode {
	return ft.NewNode(parent, value, true)
}

func (ft *FyneTree) NewLeafNode(parent *FyneTreeNode, value string) *FyneTreeNode {
	return ft.NewNode(parent, value, false)
}

func registerFyneTreeNodeMembers(t *core.Type) {
	t.RegisterProperty("value", "String",
		func(inst any) any {
			return core.String(inst.(*FyneTreeNode).Value)
		},
		nil,
	)
	t.RegisterProperty("id", "Int",
		func(inst any) any {
			return core.Int(inst.(*FyneTreeNode).ID)
		},
		nil,
	)
}

func CreateRegisterFyneTreeNodeType() {
	t := core.NewRegisteringType("FyneTreeNode",
		core.MakeHostTypeZero[*FyneTreeNode](),
		core.MakeHostTypeCheckCoerce[*FyneTreeNode]("FyneTreeNode"),
	)
	registerFyneTreeNodeMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	core.RegisterHostType[*FyneTreeNode](t)
}

func registerFyneTreeMembers(t *core.Type) {
	t.RegisterProperty("root", "FyneTreeNode",
		func(inst any) any {
			return inst.(*FyneTree).Root
		},
		nil,
	)

	// ( parent:FyneTreeNode value:String -- node:FyneTreeNode )
	t.RegisterNativeMethod("newBranchNode", func(f *core.RunFrame, ths *core.TypeObjPair) {
		value := string(f.RP.DataStack.Pop().(core.String))
		parent := f.RP.DataStack.Pop().(*FyneTreeNode)
		f.RP.DataStack.Push(ths.Obj.(*FyneTree).NewBranchNode(parent, value))
	})
	// ( parent:FyneTreeNode value:String -- node:FyneTreeNode )
	t.RegisterNativeMethod("newLeafNode", func(f *core.RunFrame, ths *core.TypeObjPair) {
		value := string(f.RP.DataStack.Pop().(core.String))
		parent := f.RP.DataStack.Pop().(*FyneTreeNode)
		f.RP.DataStack.Push(ths.Obj.(*FyneTree).NewLeafNode(parent, value))
	})

	// ( -- tree:FyneTree )
	t.RegisterNativeClassMethod("new", func(f *core.RunFrame, _ *core.TypeObjPair) {
		f.RP.DataStack.Push(NewFyneTree())
	})
}

func CreateRegisterFyneTreeType() {
	t := core.NewRegisteringType("FyneTree",
		core.MakeHostTypeZero[*FyneTree](),
		core.MakeHostTypeCheckCoerce[*FyneTree]("FyneTree"),
	)
	registerFyneTreeMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	core.RegisterHostType[*FyneTree](t)
}

func init() {
	CreateRegisterFyneTreeNodeType()
	CreateRegisterFyneTreeType()
}
