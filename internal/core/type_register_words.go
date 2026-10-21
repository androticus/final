package core

import "fmt"

// typeWords holds core-module type dictionary entries; appended after axioms on every init/reset.
var typeWords []*Word

func isCoreModuleDC(dc *DictionaryContext) bool {
	return dc != nil && dc.moduleId == CoreModuleSlot
}

// addTypeDictionaryWord adds a type proxy word (OpPushType) to dc.dict.
// Core-module words are tracked in typeWords for ResetTestRuntime rewind.
func addTypeDictionaryWord(dc *DictionaryContext, typ *Type) *Word {
	if dc == nil || dc.dict == nil {
		panic(fmt.Errorf("addTypeDictionaryWord: nil dictionary context or dictionary"))
	}
	w := NewTypeWord(typ, dc.moduleId)
	dc.dict.Add(w)
	if isCoreModuleDC(dc) {
		typeWords = append(typeWords, w)
	}
	return w
}

// registerTypeWords adds one WordFlagType dictionary word per registered *Type.
// Call after axioms are loaded and ResolveBuiltInTypeRegistrations has run.
func registerTypeWords(dc *DictionaryContext) {
	if dc == nil || dc.tr == nil {
		panic(fmt.Errorf("registerTypeWords: nil dictionary context or TypeRegistrar"))
	}
	types := dc.tr.Types()
	if isCoreModuleDC(dc) {
		typeWords = make([]*Word, 0, len(types))
	}
	for _, typ := range types {
		if typ == nil {
			panic(fmt.Errorf("registerTypeWords: nil type at idx %d", len(typeWords)))
		}
		addTypeDictionaryWord(dc, typ)
	}
}
