package core

import (
	"sync"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/speaker"
)

const DefaultSampleRate = 44100

type SoundPlayer struct {
	sampleRate beep.SampleRate
}

func (sp *SoundPlayer) FinalType() *Type { return soundPlayerType }

func NewSoundPlayer() *SoundPlayer {
	soundPlayerOnce.Do(func() {
		sr := beep.SampleRate(DefaultSampleRate)
		if err := speaker.Init(sr, sr.N(time.Second/10)); err != nil {
			panic(err)
		}
		soundPlayer = &SoundPlayer{sampleRate: sr}
	})
	return soundPlayer
}

func (sp *SoundPlayer) Play(sound *Sound) {
	speaker.Play(sound.StreamerAt(sp.sampleRate))
}

func (sp *SoundPlayer) PlayLooped(sound *Sound) {
	sound.Stop()
	ctrl := &beep.Ctrl{Streamer: sound.LoopedStreamerAt(sp.sampleRate)}
	sound.ctrl = ctrl
	speaker.Play(ctrl)
}

func (sp *SoundPlayer) Stop() {
	speaker.Clear()
}

func registerSoundPlayerMembers(t *Type) {
	t.registerClassProperty("inst", "SoundPlayer",
		func(_ Any) Any { return NewSoundPlayer() },
		nil,
	)
	t.registerProperty("sampleRate", "Int",
		func(inst Any) Any { return Int(inst.(*SoundPlayer).sampleRate) },
		nil,
	)
	// ( sound:Sound -- )
	t.registerNativeMethod("play", func(f *RunFrame, ths *TypeObjPair) {
		sound := f.RP.DataStack.Pop().(*Sound)
		ths.Obj.(*SoundPlayer).Play(sound)
	})
	// ( sound:Sound -- )
	t.registerNativeMethod("playLooped", func(f *RunFrame, ths *TypeObjPair) {
		sound := f.RP.DataStack.Pop().(*Sound)
		ths.Obj.(*SoundPlayer).PlayLooped(sound)
	})
	// ( -- )
	t.registerNativeMethod("stop", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*SoundPlayer).Stop()
	})
}

func CreateRegisterSoundPlayerType() {
	t := NewRegisteringType("SoundPlayer",
		ReferenceTypeZeroFn,
		MakePointerTypeCheckCoerce[SoundPlayer]("SoundPlayer"),
	)
	registerSoundPlayerMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	soundPlayerType = t
}

var (
	soundPlayer     *SoundPlayer
	soundPlayerOnce sync.Once
	soundPlayerType *Type
)

func init() {
	CreateRegisterSoundPlayerType()
}

var _ Typed = (*SoundPlayer)(nil)
