package core

import (
	"fmt"
	"slices"
)

func init() {
	CreateRegisterWordFlagsType()
	CreateRegisterWordType()
	CreateRegisterDictionaryType()
	CreateRegisterDictionaryContextType()
}

// Word and Dictionary are Go runtime types, not Final Object/Collection values.
// Final code reaches them only through dictionary axioms; add fields on *Word as needed.

// Dictionary is a map of words by name
// It is append only, unless reset (not common, eg interactive session also clears dictionary)
// It is not threadsafe, because only one task compiles it;
// After it is sealed, everyone can read it.
type Dictionary struct {
	name      string             // for error messages
	moduleId  ModuleID           // the module ID of the dictionary (0=global, 0x7FFFFFFF=local, 1..0x7FFFFFFE=loaded module)
	byOrdinal []*Word            // in order they were added; this is the XT for Exec
	byName    map[string]*Word   // for efficient lookup by name
	maxSize   int                // max number of words; unlimited for global, bounded for local and moules
	building  *Stack[*Word]      // in-flight colon word (depth 1; subwords are not nested building Words)
	dc        *DictionaryContext // owning DC for sealed/locked checks (nil if unbound)
}

func NewDictionary(name string, maxSize int, initCapacity int) *Dictionary {
	return &Dictionary{
		name:      name,
		byOrdinal: make([]*Word, 0, initCapacity),
		byName:    make(map[string]*Word, initCapacity),
		maxSize:   maxSize,
		building:  NewStack[*Word]("building", 8), // 8 nesting levels is enough for most words
	}
}

func (d *Dictionary) Len() int {
	return len(d.byOrdinal)
}

// UserWordsIterator walks a snapshot of dictionary words, yielding those that are
// user-visible for the restrict mask captured at GetUserWords. Current is *Word.
type UserWordsIterator struct {
	words    []*Word
	restrict SecurityRestrictions
	index    int
}

func (it *UserWordsIterator) GetNext() bool {
	for it.index < len(it.words)-1 {
		it.index++
		if it.words[it.index].isUserVisible(it.restrict) {
			return true
		}
	}
	return false
}

func (it *UserWordsIterator) Current() any {
	return it.words[it.index]
}

// GetUserWords returns a pull iterator of user-visible words (IsUserVisible).
// The backing byOrdinal slice is cloned so later Add/RollbackTo do not affect this walk.
func (d *Dictionary) GetUserWords(t *Task) Iterator {
	var restrict SecurityRestrictions
	if t != nil {
		restrict = t.restrict
	}
	return &UserWordsIterator{
		words:    slices.Clone(d.byOrdinal),
		restrict: restrict,
		index:    -1,
	}
}

var _ Iterator = (*UserWordsIterator)(nil)

// Get returns the word by name, or nil if not found
func (d *Dictionary) GetByName(name string) *Word {
	return d.byName[name]
}

func (d *Dictionary) HasName(name string) bool {
	_, ok := d.byName[name]
	return ok
}

// Get returns the word by orddinal (its XT); panics if out of range
func (d *Dictionary) GetByOrdinal(ordinal int) *Word {
	if ordinal < 0 || ordinal >= len(d.byOrdinal) {
		panic(fmt.Errorf("Dictionary.GetByOrdinal: out of range %d", ordinal))
	}
	return d.byOrdinal[ordinal]
}

// XTOf returns the execution token (dictionary ordinal) for w in this dictionary.
// VirtAddr high half is the owning module table slot (w.moduleId); resolve via ModuleTable at run.
func (d *Dictionary) XTOf(w *Word) (XT, bool) {
	if w == nil {
		return 0, false
	}
	xt := EncodeVirt(d.moduleId, w.ordIdx)
	return XT(xt), true
}

// Add adds a word to the dictionary
// Panics if the word already exists(Final doesn't allow redefining words, like Forth did)
func (d *Dictionary) Add(word *Word) {
	if d.dc != nil {
		d.dc.panicIfLockedOrSealed("add word")
	}
	if word == nil {
		panic(fmt.Errorf("Dictionary: Attempt to add nil word to %s", d.name))
	}
	len := len(d.byOrdinal)
	if len >= d.maxSize {
		panic(fmt.Errorf("Dictionary: overflow trying to add to %s", d.name))
	}
	_, ok := d.byName[word.name]
	if ok {
		panic(fmt.Errorf("Dictionary item already exists for name: %s", word.name))
	} else {
		word.moduleId = d.moduleId
		word.ordIdx = len
		d.byOrdinal = append(d.byOrdinal, word)
		d.byName[word.name] = word
	}
}

// RollbackTo removes dictionary entries with ordinal >= start, in reverse ordinal order.
func (d *Dictionary) RollbackTo(start int) {
	if start < 0 || start > len(d.byOrdinal) {
		panic(fmt.Errorf("Dictionary.RollbackTo: start %d out of range, len %d", start, len(d.byOrdinal)))
	}
	for i := len(d.byOrdinal) - 1; i >= start; i-- {
		delete(d.byName, d.byOrdinal[i].name)
	}
	d.byOrdinal = d.byOrdinal[:start]
}

// Streamlined version of Add for Axioms NOTE: MUST be added in OpCode ordinal order
func (d *Dictionary) AddAxiom(w *Word) {
	if w.opCode != OpCode(len(d.byOrdinal)) {
		panic(fmt.Errorf("Invalid Axiom registration: opCode %d does not match ordinal %d", w.opCode, len(d.byOrdinal)))
	}
	d.byOrdinal = append(d.byOrdinal, w)
	d.byName[w.name] = w
}

// PushBuilding pushes an in-flight colon word onto the building stack.
// Dictionary commit (byName and byOrdinal) happens at ; via PopAddBuildingWordToParent → Add.
func (d *Dictionary) PushBuilding(w *Word) {
	if d.dc != nil {
		d.dc.panicIfLockedOrSealed("push building word")
	}
	if w == nil {
		panic(fmt.Errorf("Dictionary.PushBuilding: nil word"))
	}
	// Nested building Words are not supported (187 subwords use a slice on the parent, not PushBuilding).
	if d.building.Len() > 0 {
		panic(fmt.Errorf("Dictionary.PushBuilding: nested words are not supported"))
	}
	if d.HasName(w.name) {
		panic(fmt.Errorf("Dictionary.PushBuilding: name already exists: %s", w.name))
	}
	d.building.Push(w)
}

func (d *Dictionary) PeekBuilding() *Word {
	if d.building.Len() == 0 {
		return nil
	}
	return d.building.Peek()
}

// PopDiscardBuilding pops the building-stack word without dictionary commit (Final method tear-off).
func (d *Dictionary) PopDiscardBuilding() *Word {
	if d.building.Len() == 0 {
		panic(fmt.Errorf("Dictionary.PopDiscardBuilding: building stack is empty"))
	}
	w := d.building.Pop()
	if d.building.Len() != 0 {
		panic(fmt.Errorf("Dictionary.PopDiscardBuilding: nested words are not supported"))
	}
	return w
}

// PopAddBuildingWordToParent pops the building-stack word and Adds it to this dictionary.
// Panics if the building stack is empty or would leave nested building Words.
func (d *Dictionary) PopAddBuildingWordToParent() {
	if d.building.Len() == 0 {
		panic(fmt.Errorf("Dictionary.PopAddBuildingWordToParent: building stack is empty"))
	}
	w := d.building.Pop()
	if d.building.Len() != 0 {
		panic(fmt.Errorf("Dictionary.PopAddBuildingWordToParent: nested words are not supported"))
	}
	if w == nil {
		panic(fmt.Errorf("Dictionary.PopAddBuildingWordToParent: nil word"))
	}
	d.Add(w)
}

func (d *Dictionary) IsBuilding() bool {
	return d.building.Len() > 0
}

// ClearBuilding drops all words on the building stack without committing to the dictionary.
func (d *Dictionary) ClearBuilding() {
	for d.building.Len() > 0 {
		d.building.Pop()
	}
}

// rendered as: ( name:String flags:Int -- word:Word )
func (d *Dictionary) MakeNewWord(name string, flags WordFlags) *Word {
	return NewWord(name, flags)
}

func (d *Dictionary) FinalType() *Type { return GetTypeRegistrar().Get("Dictionary") }

func CreateRegisterDictionaryType() {
	t := &Type{
		name: "Dictionary",
		getZero: func() any {
			return nil
		},
		checkCoerceType: func(value any) any {
			dict, ok := value.(*Dictionary)
			if !ok {
				panic(fmt.Errorf("Dictionary.checkCoerceType: value %T is not *Dictionary", value))
			}
			return dict
		},
	}
	t.registerItemsData("Word",
		nil, nil,
		func(inst Any, key Any) any {
			return inst.(*Dictionary).GetByName(string(key.(String)))
		},
		nil, nil,
	)
	t.registerProperty("len", "Int",
		func(inst Any) Any { return Int(inst.(*Dictionary).Len()) },
		nil,
	)

	t.registerNativeMethod("getUserWordsIter", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(ths.Obj.(*Dictionary).GetUserWords(f.Task))
	})
	t.registerNativeMethod("hasName", func(f *RunFrame, ths *TypeObjPair) {
		name := string(f.RP.DataStack.Pop().(String))
		f.RP.DataStack.Push(Bool(ths.Obj.(*Dictionary).HasName(name)))
	})
	t.registerNativeMethod("pushBuilding", func(f *RunFrame, ths *TypeObjPair) {
		w := f.RP.DataStack.Pop().(*Word)
		ths.Obj.(*Dictionary).PushBuilding(w)
	})
	t.registerNativeMethod("peekBuilding", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(ths.Obj.(*Dictionary).PeekBuilding())
	})
	t.registerNativeMethod("popAddBuildingWordToParent", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*Dictionary).PopAddBuildingWordToParent()
	})
	t.registerNativeMethod("makeNewWord", func(f *RunFrame, ths *TypeObjPair) {
		name := string(f.RP.DataStack.Pop().(String))
		u, ok := integralToUint32(f.RP.DataStack.Pop())
		if !ok {
			panic(fmt.Errorf("Dictionary.makeNewWord: flags expected Int or Uint"))
		}
		f.RP.DataStack.Push(ths.Obj.(*Dictionary).MakeNewWord(name, WordFlags(u)))
	})
	t.registerNativeMethod("add", func(f *RunFrame, ths *TypeObjPair) {
		w := f.RP.DataStack.Pop().(*Word)
		ths.Obj.(*Dictionary).Add(w)
	})
	GetTypeRegistrar().RegisterCurrentType(t)
}
