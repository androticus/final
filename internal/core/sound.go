package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/gopxl/beep/v2/vorbis"
	"github.com/gopxl/beep/v2/wav"
)

const Quality = 6 // offline resample-into-buffer; 6 is beep's "very good" offline quality

// Format is beep.Format as a Final value type. SampleRate is exposed as Int.
type Format beep.Format

func (f Format) FinalType() *Type { return formatType }

// Sound holds a decoded file until EnsureSampleRate, then a Buffer at the player rate.
// Loop cannot wrap *Resampler (no Seek). Buffer.Streamer is a StreamSeeker, so Loop and
// overlapping one-shots both clone from the buffer after a single resample.
type Sound struct {
	streamer beep.StreamSeekCloser // decoder; closed after buffering
	format   Format                // original file format / sample rate
	buffer   *beep.Buffer          // player-rate PCM, set by EnsureSampleRate
	ctrl     *beep.Ctrl            // set while this Sound is looping; nil after stop
}

func (s *Sound) FinalType() *Type { return soundType }

func DecodeSound(path string, typ string) (beep.StreamSeekCloser, beep.Format, error) {
	f, ferr := os.Open(path)
	if ferr != nil {
		return nil, beep.Format{}, ferr
	}

	var streamer beep.StreamSeekCloser
	var format beep.Format
	var err error

	switch typ {
	case ".ogg":
		streamer, format, err = vorbis.Decode(f)
	case ".wav":
		streamer, format, err = wav.Decode(f)
	case ".mp3":
		streamer, format, err = mp3.Decode(f)
	default:
		f.Close()
		return nil, beep.Format{}, fmt.Errorf("Unsupported sound file extension: %s", typ)
	}
	if err != nil {
		f.Close()
		return nil, beep.Format{}, err
	}
	// Decoder owns f; close via streamer.Close(), not f.Close().
	return streamer, format, nil
}

func NewSound(path string) *Sound {
	// assume the extension is correct
	ext := strings.ToLower(filepath.Ext(path))
	streamer, format, err := DecodeSound(path, ext)
	if err != nil {
		panic(err)
	}
	return &Sound{
		streamer: streamer,
		format:   Format(format),
	}
}

// EnsureSampleRate drains the decoder once into a Buffer at playerSampleRate.
// Resample runs here if needed, then the decoder is closed. Later Play/Loop
// clone StreamSeekers from the buffer (cheap, no resample).
func (s *Sound) EnsureSampleRate(playerSampleRate beep.SampleRate) {
	if s.buffer != nil {
		if s.buffer.Format().SampleRate != playerSampleRate {
			panic(fmt.Errorf("Sound already buffered at %d Hz, not %d", s.buffer.Format().SampleRate, playerSampleRate))
		}
		return
	}
	playerFmt := beep.Format(s.format)
	playerFmt.SampleRate = playerSampleRate
	buf := beep.NewBuffer(playerFmt)
	var src beep.Streamer = s.streamer
	if beep.SampleRate(s.format.SampleRate) != playerSampleRate {
		src = beep.Resample(Quality, beep.SampleRate(s.format.SampleRate), playerSampleRate, s.streamer)
	}
	buf.Append(src)
	if err := s.streamer.Close(); err != nil {
		panic(err)
	}
	s.streamer = nil
	s.buffer = buf
}

// StreamerAt is a fresh StreamSeeker over the buffered samples (player rate).
func (s *Sound) StreamerAt(playerSampleRate beep.SampleRate) beep.StreamSeeker {
	s.EnsureSampleRate(playerSampleRate)
	return s.buffer.Streamer(0, s.buffer.Len())
}

// LoopedStreamerAt loops a clone of the buffer.
func (s *Sound) LoopedStreamerAt(playerSampleRate beep.SampleRate) beep.Streamer {
	return beep.Loop(-1, s.StreamerAt(playerSampleRate))
}

// Looping is true while this Sound owns an active playLooped Ctrl.
func (s *Sound) Looping() bool {
	return s.ctrl != nil
}

// Pause freezes a looping Sound (beep.Ctrl.Paused). No-op if not looping.
func (s *Sound) Pause() {
	if s.ctrl == nil {
		return
	}
	speaker.Lock()
	s.ctrl.Paused = true
	speaker.Unlock()
}

// Resume unfreezes a looping Sound. No-op if not looping.
func (s *Sound) Resume() {
	if s.ctrl == nil {
		return
	}
	speaker.Lock()
	s.ctrl.Paused = false
	speaker.Unlock()
}

// Stop ends this Sound's loop, if any. One-shot plays are not tracked.
func (s *Sound) Stop() {
	if s.ctrl == nil {
		return
	}
	speaker.Lock()
	s.ctrl.Streamer = nil
	speaker.Unlock()
	s.ctrl = nil
}

func registerFormatMembers(t *Type) {
	t.registerProperty("sampleRate", "Int",
		func(inst Any) Any { return Int(inst.(Format).SampleRate) },
		nil,
	)
	t.registerProperty("numChannels", "Int",
		func(inst Any) Any { return Int(inst.(Format).NumChannels) },
		nil,
	)
	t.registerProperty("precision", "Int",
		func(inst Any) Any { return Int(inst.(Format).Precision) },
		nil,
	)
}

func CreateRegisterFormatType() {
	t := NewRegisteringType("Format",
		MakeHostTypeZero[Format](),
		MakeHostTypeCheckCoerce[Format]("Format"),
	)
	registerFormatMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	formatType = t
}

func registerSoundMembers(t *Type) {
	t.registerProperty("format", "Format",
		func(inst Any) Any { return inst.(*Sound).format },
		nil,
	)
	t.registerProperty("looping", "Bool",
		func(inst Any) Any { return Bool(inst.(*Sound).Looping()) },
		nil,
	)
	// ( path:String -- sound:Sound )
	t.registerNativeClassMethod("new", func(f *RunFrame, _ *TypeObjPair) {
		path := string(f.RP.DataStack.Pop().(String))
		f.RP.DataStack.Push(NewSound(path))
	})
	// ( -- ) freeze loop if any
	t.registerNativeMethod("pause", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*Sound).Pause()
	})
	// ( -- ) unfreeze loop if any
	t.registerNativeMethod("resume", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*Sound).Resume()
	})
	// ( -- )
	t.registerNativeMethod("stop", func(f *RunFrame, ths *TypeObjPair) {
		ths.Obj.(*Sound).Stop()
	})
}

func CreateRegisterSoundType() {
	t := NewRegisteringType("Sound",
		ReferenceTypeZeroFn,
		MakePointerTypeCheckCoerce[Sound]("Sound"),
	)
	registerSoundMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	soundType = t
}

var (
	formatType *Type
	soundType  *Type
)

func init() {
	CreateRegisterFormatType()
	CreateRegisterSoundType()
}

var (
	_ Typed = (*Format)(nil)
	_ Typed = (*Sound)(nil)
)
