package fyneui

import (
	"fmt"
	"sync"
)

// ModalCallbackResult is the Go-only payload of one modal wait.
// Final natives unwrap it onto the data stack; this type is not registered.
type ModalCallbackResult struct {
	CloserIndex int // 0=btnCancel, 1=btnOk, 2+=user
	Data        any
}

// modalWait is one show/wait occurrence: buffer-1 channel, no helper goroutine.
type modalWait struct {
	ch   chan *ModalCallbackResult
	once sync.Once
}

var (
	modalMu       sync.Mutex
	modalInFlight bool
)

func beginModalWait() *modalWait {
	modalMu.Lock()
	defer modalMu.Unlock()
	if modalInFlight {
		panic(fmt.Errorf("modal: already open"))
	}
	modalInFlight = true
	return &modalWait{ch: make(chan *ModalCallbackResult, 1)}
}

func (m *modalWait) complete(index int, data any) {
	m.once.Do(func() {
		m.ch <- &ModalCallbackResult{CloserIndex: index, Data: data}
	})
}

func (m *modalWait) wait() *ModalCallbackResult {
	defer func() {
		modalMu.Lock()
		modalInFlight = false
		modalMu.Unlock()
	}()
	return <-m.ch
}

func confirmCloserIndex(yes bool) int {
	if yes {
		return btnOk
	}
	return btnCancel
}
