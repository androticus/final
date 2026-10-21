package fyneui

import "testing"

func TestSimEntityList_GetIterSnapshot(t *testing.T) {
	sg := &SimGame{entities: NewSimEntityList()}
	e1 := NewSimEntity()
	e2 := NewSimEntity()
	e3 := NewSimEntity()
	sg.AddEntity(e1)
	sg.AddEntity(e2)

	it := sg.entities.GetIter()
	sg.AddEntity(e3) // concurrent membership change after snapshot
	sg.entities.Remove(e1)

	var got []*SimEntity
	for it.GetNext() {
		got = append(got, it.Current().(*SimEntity))
	}
	if len(got) != 2 || got[0] != e1 || got[1] != e2 {
		t.Fatalf("snapshot iter = %v, want [%p %p]", got, e1, e2)
	}
}

func TestSimEntityList_GetIterEmpty(t *testing.T) {
	it := NewSimEntityList().GetIter()
	if it.GetNext() {
		t.Fatal("empty list: GetNext should be false")
	}
}
