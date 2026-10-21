package core

const (
	DefaultEventLoopBufferSize = 128 // powers of 2, maybe should be larger?
)

var (
	eventLoopRunning = false
)

type EventLoop struct {
	pipe   *Pipe
	closed bool
}

func NewEventLoop(overrideBufferSize int) *EventLoop {
	if overrideBufferSize == 0 {
		overrideBufferSize = DefaultEventLoopBufferSize
	}
	return &EventLoop{pipe: NewPipe(overrideBufferSize)}
}

func GetEventLoop() *EventLoop {
	if TheEventLoop == nil {
		TheEventLoop = NewEventLoop(DefaultEventLoopBufferSize)
	}
	return TheEventLoop
}

func (el *EventLoop) Chan() chan any {
	return el.pipe.channel
}

// Close sets the closed flag so EventLoop#run's foreach can exit.
// We send a dummy event to the channel to wake up the foreach.
func (el *EventLoop) Close() {
	if el == nil || el.closed {
		return
	}
	el.closed = true
	el.pipe.channel <- nil // wake up the foreach
}

func (el *EventLoop) CreateEventFn(handler *MethodRef, index int) HandlerMintEventFn {
	return MetaMintEventFn(el.Chan(), handler, index)
}

func (el *EventLoop) FinalType() *Type {
	return eventLoopType
}

func registerEventLoopMembers(t *Type) {
	t.registerProperty("pipe", "Pipe",
		func(inst any) any {
			return inst.(*EventLoop).pipe
		},
		nil,
	)
	t.registerProperty("closed", "Bool",
		func(inst any) any {
			return inst.(*EventLoop).closed
		},
		nil,
	)
	t.RegisterClassProperty("inst", "EventLoop",
		func(_ any) any {
			return GetEventLoop()
		},
		nil,
	)
	t.RegisterClassProperty("running", "Bool",
		func(_ any) any {
			return eventLoopRunning
		},
		nil,
	)
	t.registerClassProperty("_running", "Bool", // internal use only
		func(_ any) any {
			return eventLoopRunning
		},
		func(_ any, value any) {
			eventLoopRunning = value.(bool)
		},
	)

	// ( -- )
	t.RegisterNativeMethod("close", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*EventLoop).Close()
	})

	// ( handler:MethodRef index:Int -- )
	// Mint an Event onto the loop pipe (nil source, empty params). Safe before #run.
	t.RegisterNativeMethod("mintEvent", func(f *RunFrame, ths *TypeObjPair) {
		index := int(f.RP.DataStack.Pop().(Int))
		handler := f.RP.DataStack.Pop().(*MethodRef)
		ths.Obj.(*EventLoop).CreateEventFn(handler, index)(nil)
	})

	// ( overrideBufferSize:Int -- )
	t.RegisterNativeClassMethod("ensureInstance", func(f *RunFrame, _ *TypeObjPair) {
		overrideBufferSize := int(f.RP.DataStack.Pop().(Int))
		if TheEventLoop == nil {
			TheEventLoop = NewEventLoop(overrideBufferSize)
		}
	})
}

func CreateRegisterEventLoopType() {
	t := NewRegisteringType("EventLoop",
		ReferenceTypeZeroFn,
		MakePointerTypeCheckCoerce[EventLoop]("EventLoop"),
	)
	registerEventLoopMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	eventLoopType = t
}

var (
	TheEventLoop  *EventLoop
	eventLoopType *Type
	_             Typed = (*EventLoop)(nil)
)

func init() {
	CreateRegisterEventLoopType()
}
