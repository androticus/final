package core

// ContextMark is a marker for the current context.
type ContextMark struct {
	stacks [csCount]int
	dict   int
	ct     int
	dt     int
}

func (cm *ContextMark) FinalType() *Type {
	return contextMarkType
}

func CreateRegisterContextMarkType() {
	t := &Type{
		name: "ContextMark",
		getZero: func() any {
			return nil
		},
	}
	t.registerProperty("ds", "Int",
		func(inst Any) Any { return inst.(*ContextMark).stacks[csDataStack] },
		nil,
	)
	t.registerProperty("ts", "Int",
		func(inst Any) Any { return inst.(*ContextMark).stacks[csTStack] },
		nil,
	)
	t.registerProperty("xs", "Int",
		func(inst Any) Any { return inst.(*ContextMark).stacks[csXStack] },
		nil,
	)
	t.registerProperty("ys", "Int",
		func(inst Any) Any { return inst.(*ContextMark).stacks[csYStack] },
		nil,
	)
	t.registerProperty("zs", "Int",
		func(inst Any) Any { return inst.(*ContextMark).stacks[csZStack] },
		nil,
	)
	t.registerProperty("os", "Int",
		func(inst Any) Any { return inst.(*ContextMark).stacks[csOStack] },
		nil,
	)
	t.registerProperty("dict", "Int",
		func(inst Any) Any { return inst.(*ContextMark).dict },
		nil,
	)
	t.registerProperty("ct", "Int",
		func(inst Any) Any { return inst.(*ContextMark).ct },
		nil,
	)
	t.registerProperty("dt", "Int",
		func(inst Any) Any { return inst.(*ContextMark).dt },
		nil,
	)
	GetTypeRegistrar().RegisterCurrentType(t)
	contextMarkType = t
}

var contextMarkType *Type

func init() {
	CreateRegisterContextMarkType()
}
