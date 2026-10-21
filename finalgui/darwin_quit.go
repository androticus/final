//go:build darwin

package finalgui

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework AppKit -framework Foundation

void installNativeQuitHookC(void);
*/
import "C"

//export goNativeQuit
func goNativeQuit() {
	requestQuit()
}

// installNativeQuitHook rebinds GLFW's app-menu Quit (Cmd-Q) to requestQuit.
// Call after glfw.Init has built the Cocoa menu (Fyne Lifecycle OnStarted).
func installNativeQuitHook() {
	C.installNativeQuitHookC()
}
