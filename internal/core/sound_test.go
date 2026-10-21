package core

import (
	"testing"

	"github.com/gopxl/beep/v2"
)

func TestFormatAndSoundTypesRegistered(t *testing.T) {
	ft := GetTypeRegistrar().Get("Format")
	if ft == nil {
		t.Fatal("Format type not registered")
	}
	for _, name := range []string{"sampleRate", "numChannels", "precision"} {
		pm := ft.GetPropertyMetadata(name)
		if pm == nil || pm.set != nil {
			t.Fatalf("Format.%s should be readonly const", name)
		}
	}

	st := GetTypeRegistrar().Get("Sound")
	if st == nil {
		t.Fatal("Sound type not registered")
	}
	pm := st.GetPropertyMetadata("format")
	if pm == nil || pm.set != nil {
		t.Fatal("Sound.format should be readonly const")
	}
	if st.ClassMethodByName("new") == nil {
		t.Fatal("missing Sound.new")
	}
	if st.MethodByName("stop") == nil {
		t.Fatal("missing Sound.stop")
	}
	if st.MethodByName("pause") == nil || st.MethodByName("resume") == nil {
		t.Fatal("missing Sound.pause / resume")
	}
	if pm := st.GetPropertyMetadata("looping"); pm == nil || pm.set != nil {
		t.Fatal("Sound.looping should be readonly const")
	}
}

func TestSoundNewFormatProperties(t *testing.T) {
	typ := GetTypeRegistrar().Get("Sound")
	rp := NewTestRunProvider()
	f, ths := memberRunFrame(rp, typ)
	rp.DataStack.Push(String("assets/sounds/sfx/ui/click.ogg"))
	typ.ClassMethodByName("new").invokeCall(f, ths)
	sound := rp.DataStack.Pop().(*Sound)

	fmt := sound.format
	if fmt.NumChannels < 1 || fmt.Precision < 1 || fmt.SampleRate < 1 {
		t.Fatalf("Format sampleRate=%d numChannels=%d precision=%d looks empty",
			fmt.SampleRate, fmt.NumChannels, fmt.Precision)
	}

	ft := GetTypeRegistrar().Get("Format")
	if ft.GetPropertyMetadata("sampleRate").get(fmt).(Int) != Int(fmt.SampleRate) {
		t.Fatal("Format.sampleRate getter mismatch")
	}
	if ft.GetPropertyMetadata("numChannels").get(fmt).(Int) != Int(fmt.NumChannels) {
		t.Fatal("Format.numChannels getter mismatch")
	}
	if ft.GetPropertyMetadata("precision").get(fmt).(Int) != Int(fmt.Precision) {
		t.Fatal("Format.precision getter mismatch")
	}
}

func TestSoundPlayerMembers(t *testing.T) {
	typ := GetTypeRegistrar().Get("SoundPlayer")
	if typ == nil {
		t.Fatal("SoundPlayer type not registered")
	}
	pm := typ.GetClassPropertyMetadata("inst")
	if pm == nil || pm.set != nil {
		t.Fatal("SoundPlayer.inst should be readonly const class property")
	}
	sr := typ.GetPropertyMetadata("sampleRate")
	if sr == nil || sr.set != nil {
		t.Fatal("SoundPlayer.sampleRate should be readonly const")
	}
	if typ.MethodByName("play") == nil || typ.MethodByName("playLooped") == nil || typ.MethodByName("stop") == nil {
		t.Fatal("missing SoundPlayer.play / playLooped / stop")
	}

	sp := &SoundPlayer{sampleRate: DefaultSampleRate}
	if sr.get(sp).(Int) != Int(DefaultSampleRate) {
		t.Fatalf("sampleRate %v, want %d", sr.get(sp), DefaultSampleRate)
	}
}

func TestSoundStopClearsCtrl(t *testing.T) {
	ctrl := &beep.Ctrl{Streamer: beep.Callback(func() {})}
	s := &Sound{ctrl: ctrl}
	s.Stop()
	if s.ctrl != nil {
		t.Fatal("Sound.ctrl should be nil after stop")
	}
	if ctrl.Streamer != nil {
		t.Fatal("Ctrl.Streamer should be nil after stop")
	}
	s.Stop() // no-op when already stopped
}

func TestSoundPauseResume(t *testing.T) {
	ctrl := &beep.Ctrl{Streamer: beep.Callback(func() {}), Paused: false}
	s := &Sound{ctrl: ctrl}
	if !s.Looping() {
		t.Fatal("expected looping while ctrl set")
	}
	s.Pause()
	if !ctrl.Paused {
		t.Fatal("Pause should set Ctrl.Paused")
	}
	s.Resume()
	if ctrl.Paused {
		t.Fatal("Resume should clear Ctrl.Paused")
	}
	s2 := &Sound{}
	s2.Pause()
	s2.Resume() // no-op without ctrl
}
