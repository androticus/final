package fyneui

import (
	"fmt"
	"slices"

	"fyne.io/fyne/v2"
	//"fyne.io/fyne/v2/canvas"
	"github.com/androticus/final/internal/core"
)

type PixSizeIndex = int

const (
	PixSize480x480   PixSizeIndex = iota // 1:1
	PixSize576x576                       // 1:1
	PixSize640x480                       // 4:3
	PixSize768x576                       // 16:9
	PixSize768x768                       // 1:1
	PixSize1024x768                      // 4:3
	PixSize1280x720                      // 16:9
	PixSize1920x1080                     // 16:9
)

var pixSizeConstants = []core.Size{
	{W: 480, H: 480},
	{W: 576, H: 576},
	{W: 640, H: 480},
	{W: 768, H: 576},
	{W: 768, H: 768},
	{W: 1024, H: 768},
	{W: 1280, H: 720},
	{W: 1920, H: 1080},
}

type BoundaryIndex int

const (
	BoundaryIndexLeft BoundaryIndex = iota
	BoundaryIndexTop
	BoundaryIndexRight
	BoundaryIndexBottom
	BoundaryMaskTopBottom BoundaryIndex = 0x01
)

// boundaryNormals is the Go engine table: outward unit normal per side (CG +Y down).
// Index with BoundaryIndex. Do not go through Boundary.normals (List / any).
var boundaryNormals = [4]core.Point{
	{X: -1, Y: 0},
	{X: 0, Y: -1},
	{X: 1, Y: 0},
	{X: 0, Y: 1},
}

func boundaryNormal(idx BoundaryIndex) core.Point {
	return boundaryNormals[idx]
}

// BoundaryMode is a packed playfield-side policy (uint8). Final authors OR
// SimGameOptions.crossAction* / .hitAction* with .crossEvents / .hitEvents.
//
//	bits 0–1  CrossAction  none=0 stop=1 bounce=2
//	bit  2    reserved
//	bit  3    CrossEvents  0x08
//	bits 4–5  HitAction    none=0 stop=0x10 bounce=0x20
//	bit  6    reserved
//	bit  7    HitEvents    0x80
type BoundaryMode uint8

const (
	BoundaryCrossActionNone   BoundaryMode = 0x00
	BoundaryCrossActionStop   BoundaryMode = 0x01
	BoundaryCrossActionBounce BoundaryMode = 0x02
	BoundaryCrossEvents       BoundaryMode = 0x08

	BoundaryHitActionNone   BoundaryMode = 0x00
	BoundaryHitActionStop   BoundaryMode = 0x10
	BoundaryHitActionBounce BoundaryMode = 0x20
	BoundaryHitEvents       BoundaryMode = 0x80

	boundaryCrossActionMask BoundaryMode = 0x03
	boundaryHitActionMask   BoundaryMode = 0x30
)

func (m BoundaryMode) CrossAction() BoundaryMode { return m & boundaryCrossActionMask }

func (m BoundaryMode) HitAction() BoundaryMode { return m & boundaryHitActionMask }

func (m BoundaryMode) CrossEvents() bool { return m&BoundaryCrossEvents != 0 }

func (m BoundaryMode) HitEvents() bool { return m&BoundaryHitEvents != 0 }

type SimGameOptions struct {
	// left, top, right, bottom — packed BoundaryMode; Final: leftBoundaryMode …
	boundaryModes [4]BoundaryMode
	// sparse Hit Matrix keyed by HitArea.kind (min,max); nil until first add
	hitMatrix    map[hitMatrixKey]BoundaryMode
	hasMenuBar   bool
	hasBottomBar bool
	// derived: any side has CrossEvents
	useCrossings bool
	useHits      bool
	// author switch; not derived from sides. Motion without bounds is valid.
	useMotionEngine       bool
	hasKeyHandler         bool
	useKeyAutoRepeat      bool
	hasWindowCloseHandler bool

	hasMouseDownUpHandler     bool
	hasMouseInMovedOutHandler bool
	hasMouseScrolledHandler   bool
}

func NewSimGameOptions() *SimGameOptions {
	return &SimGameOptions{
		hasMenuBar:            false,
		hasBottomBar:          true,
		useMotionEngine:       false,
		hasWindowCloseHandler: false,
	}
}

func (o *SimGameOptions) FinalType() *core.Type {
	return simGameOptionsType
}

func (o *SimGameOptions) setBoundaryMode(idx BoundaryIndex, mode BoundaryMode) {
	if idx < BoundaryIndexLeft || idx > BoundaryIndexBottom {
		panic(fmt.Errorf("setBoundaryMode: invalid index %d", idx))
	}
	o.boundaryModes[idx] = mode
	o.syncDerived()
}

func (o *SimGameOptions) setAllBoundaryModes(mode BoundaryMode) {
	for i := range o.boundaryModes {
		o.boundaryModes[i] = mode
	}
	o.syncDerived()
}

func (o *SimGameOptions) setUseHits(v bool) {
	o.useHits = v
	o.syncDerived()
}

func (o *SimGameOptions) syncDerived() {
	o.useCrossings = false
	for _, m := range o.boundaryModes {
		if m.CrossEvents() {
			o.useCrossings = true
		}
	}
}

// SimGame is a regular Go catalog type. Composition: holds del (user TypedObject);
// _newWithDelegate / newWithDelegate set del; mint (based TypedObject) or the
// public newWithDelegate path writes the child's base/sg field.
type SimGame struct {
	del      *core.TypedObject // user delegate (props + methods)
	clock    *SimClock
	options  *SimGameOptions
	mhe      *SimMHEngine // if useMotionEngine is true
	title    string
	pixSize  core.Size
	pixmap   *core.Pixmap
	painter  *core.Painter // JIT on Painter(); cleared when pixmap replaced
	entities *SimEntityList
	// backgroundColor is the library playfield clear fill (default black).
	backgroundColor core.Color

	view *CanvasImage
	win  *Window

	_timer *core.EventTimer

	timerDispatch *core.MethodRef // SimGame#onTimerEvent; lazy
}

func (sg *SimGame) FinalType() *core.Type {
	return simGameType
}

// NewSimGameWithDelegate mints a SimGame and stores del on it.
// It does not write a field on del — based TypedObject mint sets "base";
// the public newWithDelegate class method still sets del's "sg" for existing examples.
func NewSimGameWithDelegate(del *core.TypedObject) *SimGame {
	if del == nil {
		panic("del is nil")
	}
	sg := &SimGame{
		del:     del,
		clock:   NewSimClock(),
		options: NewSimGameOptions(),
		mhe:     NewSimMHEngine(),
		// a pleasant, legal default
		pixSize:         pixSizeConstants[PixSize768x576],
		entities:        NewSimEntityList(),
		backgroundColor: core.WebColors[core.ColorBlack],
	}
	sg.mhe.sg = sg
	return sg
}

// PixSize is set during init, and is immutable thereafter.
// W and H must be exact integral (discrete pixmap dims).
func (sg *SimGame) SetPixSize(size core.Size) {
	if sg.pixmap != nil {
		panic("pixSize may only be set before init")
	}
	w := core.RequireIntegralDim("SimGame.pixSize width", size.W)
	h := core.RequireIntegralDim("SimGame.pixSize height", size.H)
	// validate it is a bare minimum for the UI to be visible
	if w < 256 || h < 128 {
		panic("pixSize must be at least 256x128 for the UI to be visible")
	}
	sg.pixSize = core.NewSize(w, h)
}

func (sg *SimGame) Start() {
	switch sg.clock.State() {
	case SimClockStateRunning:
		return
	case SimClockStatePaused:
		sg.clock.SetState(SimClockStateRunning)
		// Metronome already running; lastTime was kept current while paused.
		sg.ReprimeTimer()
		if sg.mhe != nil {
			sg.mhe.Start()
		}
		return
	case SimClockStateStopped:
		sg.clock.SetState(SimClockStateRunning)
		sg.clock.Start()
		sg.ReprimeTimer()
		if sg.mhe != nil {
			sg.mhe.Start()
		}
	}
}

func (sg *SimGame) StartPaused() {
	switch sg.clock.State() {
	case SimClockStateRunning:
		sg.Pause()
	case SimClockStatePaused:
		return
	case SimClockStateStopped:
		sg.clock.SetState(SimClockStatePaused)
		sg.clock.Start()
		// EventTimer stays idle until Start (Running). Caller may #doRender
		// for the paused load screen (ticks do not advance while Paused).
	}
}

func (sg *SimGame) Pause() {
	if sg.clock.State() != SimClockStateRunning {
		return
	}
	sg.clock.SetState(SimClockStatePaused)
	// Metronome keeps running; deltas are not accrued while Paused.
	if sg.mhe != nil {
		sg.mhe.Stop()
	}
	if sg._timer != nil {
		sg._timer.Stop()
	}
}

func (sg *SimGame) Stop() {
	if sg.clock.State() == SimClockStateStopped {
		return
	}
	sg.clock.SetState(SimClockStateStopped)
	if sg.mhe != nil {
		sg.mhe.Stop()
		if sg.mhe.avgStepNs != nil {
			sg.mhe.avgStepNs.Reset()
		}
	}
	sg.clock.Stop()
	sg.clock.ResetCostStats()
	if sg._timer != nil {
		sg._timer.Stop()
	}
}

// Close stops the sim and releases timers (EventTimer + SimClock metronome). Idempotent.
// Prefer this from window-close; Stop only pauses scheduling without full metronome teardown.
func (sg *SimGame) Close() {
	sg.clock.SetState(SimClockStateStopped)
	if sg.mhe != nil {
		sg.mhe.Stop()
	}
	if sg.clock != nil {
		sg.clock.Close()
	}
	if sg._timer != nil {
		sg._timer.Close()
		sg._timer = nil
	}
}

// ReprimeTimer schedules the next #onTimer when Running and a scheduleable lane is on.
// autoRender is not a timer lane — needs updateRate > 0 and/or renderRate > 0.
func (sg *SimGame) ReprimeTimer() {
	if sg._timer == nil || sg.clock == nil {
		return
	}
	if sg.clock.State() != SimClockStateRunning {
		return
	}
	if sg.clock.UpdateRate() == 0 && sg.clock.RenderRate() <= 0 {
		return
	}
	sg._timer.EventAfter(sg.clock.ReprimeInterval(sg.clock.EngineSpeed()))
}

// PublishEngineSpeed copies user simSpeed onto clock.engineSpeed once render costs are warm.
// No-op while cold or if the clamped value is unchanged. Called from doRenderInner epilogue
// as Final SimGame#_publishEngineSpeed. CostBasisMs is available for later auto-slow;
// MotionHit will combine it with load signals.
func (sg *SimGame) PublishEngineSpeed() {
	if sg.clock == nil {
		return
	}
	if !sg.clock.IsWarm() {
		return
	}
	sg.clock.SetEngineSpeed(sg.clock.SimSpeed())
}

// we use a counter to track suspended state, because we may have nested suspends
func (sg *SimGame) SetSuspended(suspended bool) {
	if suspended {
		sg.clock.Suspend()
	} else {
		sg.clock.Resume()
	}
}

func (sg *SimGame) SetTitle(title string) {
	sg.title = title
	if sg.win != nil {
		fyne.Do(func() {
			sg.win.SetTitle(title)
		})
	}
}

func (sg *SimGame) RefreshView() {
	if sg.view == nil {
		panic("view is not set")
	}
	sg.view.Refresh() // already wrapped in Do
}

func (sg *SimGame) timerEmit(index int) core.HandlerMintEventFn {
	if sg.timerDispatch == nil {
		sg.timerDispatch = core.NewMethodRef(sg, "onTimerEvent")
	}
	return core.GetEventLoop().CreateEventFn(sg.timerDispatch, index)
}

func (sg *SimGame) TimerEventAfter(index, after int) {
	sg.clock.TimerEventAfter(sg.timerEmit(index), sg, index, after)
}

func (sg *SimGame) TimerEventAt(index, tick int) {
	sg.clock.TimerEventAt(sg.timerEmit(index), sg, index, tick)
}

func (sg *SimGame) TimerEventEvery(index, period int) {
	sg.clock.TimerEventEvery(sg.timerEmit(index), sg, index, period)
}

func (sg *SimGame) CancelTimerEvent(index int) {
	sg.clock.CancelTimerEvent(index)
}

// simWindowContentSize is the inner size to Resize to before SetFixedSize.
// Floor to pixSize so a 1×1 CanvasImage min (pre-refresh / FillOriginal) still
// gets a real GLFW framebuffer — SetFixedSize without Resize shows a blank
// window on some drivers (RDP).
func simWindowContentSize(content fyne.CanvasObject, pix core.Size) fyne.Size {
	sz := fyne.NewSize(float32(pix.W), float32(pix.H))
	if content == nil {
		return sz
	}
	min := content.MinSize()
	if min.Width > sz.Width {
		sz.Width = min.Width
	}
	if min.Height > sz.Height {
		sz.Height = min.Height
	}
	return sz
}

func (sg *SimGame) SetContentFixedSize(content fyne.CanvasObject) {
	if sg.win == nil {
		panic("win is not set")
	}
	fyne.DoAndWait(func() {
		sg.win.SetContent(content)
		sg.win.Resize(simWindowContentSize(content, sg.pixSize))
		sg.win.SetFixedSize(true)
	})
}

// AddEntity sets entity.sg to this game and adds it to entities (marks active).
func (sg *SimGame) AddEntity(entity *SimEntity) {
	entity.sg = sg
	sg.entities.add(entity)
}

// EntitiesWithDrawerSets returns entities with a non-nil drawerSet, sorted by
// ascending zorder (background first). Snapshot; safe to hold across paint.
func (sg *SimGame) EntitiesWithDrawerSets() *core.AnyList {
	snap := sg.entities.Snapshot()
	with := make([]*SimEntity, 0, len(snap))
	for _, e := range snap {
		if e != nil && e.drawerSet != nil {
			with = append(with, e)
		}
	}
	slices.SortStableFunc(with, func(a, b *SimEntity) int {
		return int(a.zorder) - int(b.zorder)
	})
	out := core.NewAnyListWithCapacity(len(with))
	for _, e := range with {
		out.Append(e)
	}
	return out
}

// SetPixmap installs the sim raster buffer and drops any cached Painter.
func (sg *SimGame) SetPixmap(pm *core.Pixmap) {
	sg.pixmap = pm
	sg.painter = nil
}

// Painter returns a Painter bound to sg.pixmap, minting it on first call.
// Panics if pixmap has not been created yet (call after #init).
func (sg *SimGame) Painter() *core.Painter {
	if sg.painter != nil {
		return sg.painter
	}
	if sg.pixmap == nil {
		panic(fmt.Errorf("SimGame.painter: pixmap is nil (init the SimGame first)"))
	}
	sg.painter = core.NewPainterWithPixmap(sg.pixmap)
	return sg.painter
}

func registerSimGameOptionsMembers(t *core.Type) {
	modeClassConst := func(name string, mode BoundaryMode) {
		t.RegisterClassProperty(name, "Int",
			func(_ any) any { return core.Int(mode) },
			nil,
		)
	}
	modeClassConst("crossActionNone", BoundaryCrossActionNone)
	modeClassConst("crossActionStop", BoundaryCrossActionStop)
	modeClassConst("crossActionBounce", BoundaryCrossActionBounce)
	modeClassConst("crossEvents", BoundaryCrossEvents)
	modeClassConst("hitActionNone", BoundaryHitActionNone)
	modeClassConst("hitActionStop", BoundaryHitActionStop)
	modeClassConst("hitActionBounce", BoundaryHitActionBounce)
	modeClassConst("hitEvents", BoundaryHitEvents)

	registerBoundaryMode := func(name string, idx BoundaryIndex) {
		t.RegisterProperty(name, "Int",
			func(inst any) any {
				return core.Int(inst.(*SimGameOptions).boundaryModes[idx])
			},
			func(inst any, value any) {
				inst.(*SimGameOptions).setBoundaryMode(idx, BoundaryMode(uint8(value.(core.Int))))
			},
		)
	}
	registerBoundaryMode("leftBoundaryMode", BoundaryIndexLeft)
	registerBoundaryMode("topBoundaryMode", BoundaryIndexTop)
	registerBoundaryMode("rightBoundaryMode", BoundaryIndexRight)
	registerBoundaryMode("bottomBoundaryMode", BoundaryIndexBottom)

	t.RegisterProperty("hasMenuBar", "Bool",
		func(inst any) any {
			return core.Bool(inst.(*SimGameOptions).hasMenuBar)
		},
		func(inst any, value any) {
			inst.(*SimGameOptions).hasMenuBar = bool(value.(core.Bool))
		},
	)
	t.RegisterProperty("hasBottomBar", "Bool",
		func(inst any) any {
			return core.Bool(inst.(*SimGameOptions).hasBottomBar)
		},
		func(inst any, value any) {
			inst.(*SimGameOptions).hasBottomBar = bool(value.(core.Bool))
		},
	)
	t.RegisterProperty("useCrossings", "Bool",
		func(inst any) any {
			return core.Bool(inst.(*SimGameOptions).useCrossings)
		},
		nil,
	)
	t.RegisterProperty("useHits", "Bool",
		func(inst any) any {
			return core.Bool(inst.(*SimGameOptions).useHits)
		},
		func(inst any, value any) {
			inst.(*SimGameOptions).setUseHits(bool(value.(core.Bool)))
		},
	)
	t.RegisterProperty("useMotionEngine", "Bool",
		func(inst any) any {
			return core.Bool(inst.(*SimGameOptions).useMotionEngine)
		},
		func(inst any, value any) {
			inst.(*SimGameOptions).useMotionEngine = bool(value.(core.Bool))
		},
	)
	t.RegisterProperty("hasKeyHandler", "Bool",
		func(inst any) any {
			return core.Bool(inst.(*SimGameOptions).hasKeyHandler)
		},
		func(inst any, value any) {
			inst.(*SimGameOptions).hasKeyHandler = bool(value.(core.Bool))
		},
	)
	t.RegisterProperty("useKeyAutoRepeat", "Bool",
		func(inst any) any {
			return core.Bool(inst.(*SimGameOptions).useKeyAutoRepeat)
		},
		func(inst any, value any) {
			inst.(*SimGameOptions).useKeyAutoRepeat = bool(value.(core.Bool))
		},
	)
	t.RegisterProperty("hasWindowCloseHandler", "Bool",
		func(inst any) any {
			return core.Bool(inst.(*SimGameOptions).hasWindowCloseHandler)
		},
		func(inst any, value any) {
			inst.(*SimGameOptions).hasWindowCloseHandler = bool(value.(core.Bool))
		},
	)
	t.RegisterProperty("hasMouseDownUpHandler", "Bool",
		func(inst any) any {
			return core.Bool(inst.(*SimGameOptions).hasMouseDownUpHandler)
		},
		func(inst any, value any) {
			inst.(*SimGameOptions).hasMouseDownUpHandler = bool(value.(core.Bool))
		},
	)
	t.RegisterProperty("hasMouseInMovedOutHandler", "Bool",
		func(inst any) any {
			return core.Bool(inst.(*SimGameOptions).hasMouseInMovedOutHandler)
		},
		func(inst any, value any) {
			inst.(*SimGameOptions).hasMouseInMovedOutHandler = bool(value.(core.Bool))
		},
	)
	t.RegisterProperty("hasMouseScrolledHandler", "Bool",
		func(inst any) any {
			return core.Bool(inst.(*SimGameOptions).hasMouseScrolledHandler)
		},
		func(inst any, value any) {
			inst.(*SimGameOptions).hasMouseScrolledHandler = bool(value.(core.Bool))
		},
	)

	// ( mode:i -- )
	t.RegisterNativeMethod("setAllBoundaryModes", func(f *core.RunFrame, ths *core.TypeObjPair) {
		mode := BoundaryMode(uint8(f.RP.DataStack.Pop().(core.Int)))
		ths.Obj.(*SimGameOptions).setAllBoundaryModes(mode)
	})

	// ( kindA:i kindB:i modeBits:i -- )  Hit Matrix cell for HitArea.kind pair
	t.RegisterNativeMethod("addHitMatrixEntry", func(f *core.RunFrame, ths *core.TypeObjPair) {
		mode := BoundaryMode(uint8(f.RP.DataStack.Pop().(core.Int)))
		kb := int32(f.RP.DataStack.Pop().(core.Int))
		ka := int32(f.RP.DataStack.Pop().(core.Int))
		ths.Obj.(*SimGameOptions).AddHitMatrixEntry(ka, kb, mode)
	})
}

// Shared Final List filled from boundaryNormals; authors may mutate it (accepted for now).
var boundaryNormalsList *core.AnyList

func initBoundaryNormalsList() {
	l := core.NewAnyListWithCapacity(4)
	for _, p := range boundaryNormals {
		l.Append(p)
	}
	boundaryNormalsList = l
}

func registerBoundaryMembers(t *core.Type) {
	boundaryClassConst := func(name string, idx BoundaryIndex) {
		t.RegisterClassProperty(name, "Int",
			func(_ any) any { return core.Int(idx) },
			nil,
		)
	}
	boundaryClassConst("left", BoundaryIndexLeft)
	boundaryClassConst("top", BoundaryIndexTop)
	boundaryClassConst("right", BoundaryIndexRight)
	boundaryClassConst("bottom", BoundaryIndexBottom)
	boundaryClassConst("maskTopBottom", BoundaryMaskTopBottom)

	initBoundaryNormalsList()
	t.RegisterClassProperty("normals", "List",
		func(_ any) any { return boundaryNormalsList },
		nil,
	)
}

func CreateRegisterBoundaryType() {
	t := core.NewRegisteringType("Boundary",
		core.ReferenceTypeZeroFn,
		func(value any) any {
			panic(fmt.Errorf("Boundary.checkCoerceType: Boundary is a class-only type (got %T)", value))
		},
	)
	registerBoundaryMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	boundaryType = t
}

func CreateRegisterSimGameOptionsType() {
	t := core.NewRegisteringType("SimGameOptions",
		core.ReferenceTypeZeroFn,
		core.MakePointerTypeCheckCoerce[SimGameOptions]("SimGameOptions"),
	)
	registerSimGameOptionsMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	simGameOptionsType = t
}

// registerSimGameInstanceMembers — public members first; _* internals last. No class members.
func registerSimGameInstanceMembers(t *core.Type) {
	t.RegisterProperty("del", "TypedObject",
		func(inst any) any {
			return inst.(*SimGame).del
		},
		nil,
	)
	t.RegisterProperty("clock", "SimClock",
		func(inst any) any {
			return inst.(*SimGame).clock
		},
		nil,
	)
	t.RegisterProperty("ticks", "SimClock",
		func(inst any) any {
			return inst.(*SimGame).clock.Ticks()
		},
		nil,
	)
	t.RegisterProperty("options", "SimGameOptions",
		func(inst any) any {
			return inst.(*SimGame).options
		},
		nil,
	)
	t.RegisterProperty("mhe", "SimMHEngine",
		func(inst any) any {
			return inst.(*SimGame).mhe
		},
		nil,
	)
	t.RegisterProperty("title", "String",
		func(inst any) any {
			return core.String(inst.(*SimGame).title)
		},
		func(inst any, value any) {
			inst.(*SimGame).SetTitle(string(value.(core.String)))
		},
	)
	t.RegisterProperty("pixSize", "Size",
		func(inst any) any {
			return inst.(*SimGame).pixSize
		},
		func(inst any, value any) {
			inst.(*SimGame).SetPixSize(value.(core.Size))
		},
	)
	// Library playfield clear fill (SimGame_doRenderInner). Default black.
	t.RegisterProperty("backgroundColor", "Color",
		func(inst any) any {
			return inst.(*SimGame).backgroundColor
		},
		func(inst any, value any) {
			inst.(*SimGame).backgroundColor = value.(core.Color)
		},
	)
	t.RegisterProperty("view", "CanvasImage",
		func(inst any) any {
			return inst.(*SimGame).view
		},
		nil,
	)
	t.RegisterProperty("win", "Window",
		func(inst any) any {
			return inst.(*SimGame).win
		},
		nil,
	)
	t.RegisterProperty("pixmap", "Pixmap",
		func(inst any) any {
			return inst.(*SimGame).pixmap
		},
		nil,
	)
	// JIT NewPainterWithPixmap(sg.pixmap) on first get; panics if pixmap nil.
	t.RegisterProperty("painter", "Painter",
		func(inst any) any {
			return inst.(*SimGame).Painter()
		},
		nil,
	)
	t.RegisterProperty("entities", "SimEntityList",
		func(inst any) any {
			return inst.(*SimGame).entities
		},
		nil,
	)

	// Internal use only--Do not use. Registered last in the property set.
	t.RegisterProperty("_view", "CanvasImage",
		func(inst any) any {
			return inst.(*SimGame).view
		},
		func(inst any, value any) {
			inst.(*SimGame).view = value.(*CanvasImage)
		},
	)
	t.RegisterProperty("_win", "Window",
		func(inst any) any {
			return inst.(*SimGame).win
		},
		func(inst any, value any) {
			inst.(*SimGame).win = value.(*Window)
		},
	)
	t.RegisterProperty("_pixmap", "Pixmap",
		func(inst any) any {
			return inst.(*SimGame).pixmap
		},
		func(inst any, value any) {
			inst.(*SimGame).SetPixmap(value.(*core.Pixmap))
		},
	)
	t.RegisterProperty("_timer", "EventTimer",
		func(inst any) any { return inst.(*SimGame)._timer },
		func(inst any, value any) {
			inst.(*SimGame)._timer = value.(*core.EventTimer)
		},
	)

	// ( entity:SimEntity -- )
	t.RegisterNativeMethod("addEntity", func(f *core.RunFrame, ths *core.TypeObjPair) {
		entity := f.RP.DataStack.Pop().(*SimEntity)
		ths.Obj.(*SimGame).AddEntity(entity)
	})
	// ( -- ents:List ) entities with drawerSet, ascending zorder
	t.RegisterNativeMethod("entitiesWithDrawerSets", func(f *core.RunFrame, ths *core.TypeObjPair) {
		f.RP.DataStack.Push(ths.Obj.(*SimGame).EntitiesWithDrawerSets())
	})
	// ( -- )
	t.RegisterNativeMethod("start", func(f *core.RunFrame, ths *core.TypeObjPair) {
		ths.Obj.(*SimGame).Start()
	})
	// ( -- )
	t.RegisterNativeMethod("startPaused", func(f *core.RunFrame, ths *core.TypeObjPair) {
		ths.Obj.(*SimGame).StartPaused()
	})
	// ( -- )
	t.RegisterNativeMethod("pause", func(f *core.RunFrame, ths *core.TypeObjPair) {
		ths.Obj.(*SimGame).Pause()
	})
	// ( -- )
	t.RegisterNativeMethod("stop", func(f *core.RunFrame, ths *core.TypeObjPair) {
		ths.Obj.(*SimGame).Stop()
	})
	// ( -- )
	t.RegisterNativeMethod("close", func(f *core.RunFrame, ths *core.TypeObjPair) {
		ths.Obj.(*SimGame).Close()
	})
	// ( suspended:Bool -- )
	t.RegisterNativeMethod("setSuspended", func(f *core.RunFrame, ths *core.TypeObjPair) {
		ths.Obj.(*SimGame).SetSuspended(bool(f.RP.DataStack.Pop().(core.Bool)))
	})
	// ( -- )
	t.RegisterNativeMethod("reprimeTimer", func(f *core.RunFrame, ths *core.TypeObjPair) {
		ths.Obj.(*SimGame).ReprimeTimer()
	})
	// ( index:i after:i -- )
	t.RegisterNativeMethod("timerEventAfter", func(f *core.RunFrame, ths *core.TypeObjPair) {
		after := int(f.RP.DataStack.Pop().(core.Int))
		index := int(f.RP.DataStack.Pop().(core.Int))
		ths.Obj.(*SimGame).TimerEventAfter(index, after)
	})
	// ( index:i tick:i -- )
	t.RegisterNativeMethod("timerEventAt", func(f *core.RunFrame, ths *core.TypeObjPair) {
		tick := int(f.RP.DataStack.Pop().(core.Int))
		index := int(f.RP.DataStack.Pop().(core.Int))
		ths.Obj.(*SimGame).TimerEventAt(index, tick)
	})
	// ( index:i ticks:i -- )
	t.RegisterNativeMethod("timerEventEvery", func(f *core.RunFrame, ths *core.TypeObjPair) {
		period := int(f.RP.DataStack.Pop().(core.Int))
		index := int(f.RP.DataStack.Pop().(core.Int))
		ths.Obj.(*SimGame).TimerEventEvery(index, period)
	})
	// ( index:i -- )
	t.RegisterNativeMethod("cancelTimerEvent", func(f *core.RunFrame, ths *core.TypeObjPair) {
		index := int(f.RP.DataStack.Pop().(core.Int))
		ths.Obj.(*SimGame).CancelTimerEvent(index)
	})
	// ( -- )
	t.RegisterNativeMethod("_publishEngineSpeed", func(f *core.RunFrame, ths *core.TypeObjPair) {
		ths.Obj.(*SimGame).PublishEngineSpeed()
	})
	// ( -- )
	t.RegisterNativeMethod("refreshView", func(f *core.RunFrame, ths *core.TypeObjPair) {
		ths.Obj.(*SimGame).RefreshView()
	})
	t.RegisterNativeMethod("_setContentFixedSize", func(f *core.RunFrame, ths *core.TypeObjPair) {
		content := f.RP.DataStack.Pop().(fyne.CanvasObject)
		ths.Obj.(*SimGame).SetContentFixedSize(content)
	})
	// ( -- ) engine-owned kindBorder HitRects from HitAction ≠ none
	t.RegisterNativeMethod("_ensureHitBoundaries", func(_ *core.RunFrame, ths *core.TypeObjPair) {
		ths.Obj.(*SimGame).EnsureHitBoundaries()
	})
}

func registerSimGameClassMembers(t *core.Type) {
	// ( del:TypedObject -- game:SimGame ) — Base contract; helpx only (leading _).
	t.RegisterNativeClassMethod("_newWithDelegate", func(f *core.RunFrame, _ *core.TypeObjPair) {
		del := f.RP.DataStack.Pop().(*core.TypedObject)
		f.RP.DataStack.Push(NewSimGameWithDelegate(del))
	})
	// Compat for examples that still declare sg and call this manually.
	// Based TypedObject.new uses _newWithDelegate and writes "base" instead.
	t.RegisterNativeClassMethod("newWithDelegate", func(f *core.RunFrame, _ *core.TypeObjPair) {
		del := f.RP.DataStack.Pop().(*core.TypedObject)
		sg := NewSimGameWithDelegate(del)
		del.SetFieldByName("sg", sg)
		f.RP.DataStack.Push(sg)
	})
	pixSizeClassConst := func(name string, idx PixSizeIndex) {
		sz := pixSizeConstants[idx]
		t.RegisterClassProperty(name, "Size",
			func(_ any) any { return sz },
			nil,
		)
	}
	pixSizeClassConst("PixSize480x480", PixSize480x480)
	pixSizeClassConst("PixSize576x576", PixSize576x576)
	pixSizeClassConst("PixSize640x480", PixSize640x480)
	pixSizeClassConst("PixSize768x576", PixSize768x576)
	pixSizeClassConst("PixSize768x768", PixSize768x768)
	pixSizeClassConst("PixSize1024x768", PixSize1024x768)
	pixSizeClassConst("PixSize1280x720", PixSize1280x720)
	pixSizeClassConst("PixSize1920x1080", PixSize1920x1080)
}

func CreateRegisterSimGameType() {
	t := core.NewRegisteringType("SimGame",
		core.ReferenceTypeZeroFn,
		core.MakePointerTypeCheckCoerce[SimGame]("SimGame"),
	)
	registerSimGameInstanceMembers(t)
	registerSimGameClassMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	simGameType = t
}

var (
	boundaryType       *core.Type
	simGameOptionsType *core.Type
	simGameType        *core.Type
	_                  core.Typed = (*SimGameOptions)(nil)
	_                  core.Typed = (*SimGame)(nil)
)

func init() {
	CreateRegisterBoundaryType()
	CreateRegisterSimGameOptionsType()
	CreateRegisterSimGameType()
}
