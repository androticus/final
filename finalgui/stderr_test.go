package finalgui

import (
	"strings"
	"testing"

	"github.com/androticus/final/finalembed"
	"github.com/androticus/final/internal/core"
)

// SetStderr must be honored for $$; RunEventLoop must not clobber it with os.Stderr.
func TestSetStderrSurvivesForDollarDollar(t *testing.T) {
	core.ResetTestRuntime()
	sess, err := finalembed.NewPersistentSession()
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()

	log := core.NewLineBufferTextWriterSync(32)
	sess.SetStderr(log)

	if _, err := sess.GetSender().Send("\"gui-stderr\\n\" $$", 5); err != nil {
		t.Fatalf("Send: %v", err)
	}
	got := log.ToString()
	if !strings.Contains(got, "gui-stderr") {
		t.Fatalf("standing stderr log = %q, want gui-stderr", got)
	}
}
