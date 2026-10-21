package fyneui

import "github.com/androticus/final/internal/core"

type SimHitEvent struct {
	asOfTick int
	hits     *SimHits
}

func NewSimHitEvent(asOfTick int, hits *SimHits) *SimHitEvent {
	return &SimHitEvent{asOfTick: asOfTick, hits: hits}
}

func (e *SimHitEvent) FinalType() *core.Type { return simHitEventType }

type SimCrossingEvent struct {
	asOfTick  int
	crossings *SimCrossings
}

func NewSimCrossingEvent(asOfTick int, crossings *SimCrossings) *SimCrossingEvent {
	return &SimCrossingEvent{asOfTick: asOfTick, crossings: crossings}
}

func (e *SimCrossingEvent) FinalType() *core.Type { return simCrossingEventType }

func registerSimHitEventMembers(t *core.Type) {
	t.RegisterProperty("asOfTick", "Int",
		func(inst any) any { return core.Int(inst.(*SimHitEvent).asOfTick) },
		nil,
	)
	t.RegisterProperty("hits", "SimHits",
		func(inst any) any { return inst.(*SimHitEvent).hits },
		nil,
	)
}

func registerSimCrossingEventMembers(t *core.Type) {
	t.RegisterProperty("asOfTick", "Int",
		func(inst any) any { return core.Int(inst.(*SimCrossingEvent).asOfTick) },
		nil,
	)
	t.RegisterProperty("crossings", "SimCrossings",
		func(inst any) any { return inst.(*SimCrossingEvent).crossings },
		nil,
	)
}

func CreateRegisterSimHitEventType() {
	t := core.NewRegisteringType("SimHitEvent",
		core.ReferenceTypeZeroFn,
		core.MakePointerTypeCheckCoerce[SimHitEvent]("SimHitEvent"),
	)
	registerSimHitEventMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	simHitEventType = t
}

func CreateRegisterSimCrossingEventType() {
	t := core.NewRegisteringType("SimCrossingEvent",
		core.ReferenceTypeZeroFn,
		core.MakePointerTypeCheckCoerce[SimCrossingEvent]("SimCrossingEvent"),
	)
	registerSimCrossingEventMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	simCrossingEventType = t
}

var (
	simHitEventType      *core.Type
	simCrossingEventType *core.Type
	_                    core.Typed = (*SimHitEvent)(nil)
	_                    core.Typed = (*SimCrossingEvent)(nil)
)

func init() {
	CreateRegisterSimHitEventType()
	CreateRegisterSimCrossingEventType()
}
