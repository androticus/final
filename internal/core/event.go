package core

type Event struct {
	Source  any // optional, Event-context specific
	Handler *MethodRef
	Index   int      // optional, Event-context specific
	Params  *AnyList // the global empty list if no params
}

// HandlerMintEventFn is a function that creates and sends a new Event in an Event source.
type HandlerMintEventFn func(source any, params ...any)

func ParamsToList(params ...any) *AnyList {
	if len(params) == 0 {
		return GetEmptyAnyList()
	}
	list := NewAnyList()
	for _, param := range params {
		list.Append(param)
	}
	return list
}

func MetaMintEventFn(channel chan any, handler *MethodRef, index int) HandlerMintEventFn {
	return func(source any, params ...any) {
		channel <- &Event{
			Source:  source,
			Handler: handler,
			Index:   index,
			Params:  ParamsToList(params...),
		}
	}
}

func (e *Event) FinalType() *Type {
	return eventType
}

func registerEventMembers(t *Type) {
	t.registerProperty("source", "Any",
		func(inst any) any { return inst.(*Event).Source },
		nil,
	)
	t.registerProperty("handler", "MethodRef",
		func(inst any) any { return inst.(*Event).Handler },
		nil,
	)
	t.registerProperty("index", "Int",
		func(inst any) any { return inst.(*Event).Index },
		nil,
	)
	t.registerProperty("params", "List",
		func(inst any) any { return inst.(*Event).Params },
		nil,
	)
}

func CreateRegisterEventType() {
	t := NewRegisteringType("Event",
		ReferenceTypeZeroFn,
		MakePointerTypeCheckCoerce[Event]("Event"),
	)
	registerEventMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	eventType = t
}

var (
	eventType *Type
	_         Typed = (*Event)(nil)
)

func init() {
	CreateRegisterEventType()
}
