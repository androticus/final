package fyneui

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"github.com/androticus/final/internal/core"
)

func dialogParent(v any) *Window {
	w, ok := v.(*Window)
	if !ok || w == nil {
		panic(fmt.Errorf("dialog: parent must be Window, got %T", v))
	}
	return w
}

func colorFromFyne(c color.Color) core.Color {
	n := color.NRGBAModel.Convert(c).(color.NRGBA)
	return core.NewColor(int(n.R), int(n.G), int(n.B), int(n.A))
}

func pathFromURI(u fyne.URI) string {
	if u == nil {
		return ""
	}
	return u.Path()
}

func showFyneDialog(d dialog.Dialog) {
	fyne.Do(func() {
		d.Show()
	})
}

func closeReaderPath(reader fyne.URIReadCloser, err error) string {
	if err != nil || reader == nil {
		return ""
	}
	path := pathFromURI(reader.URI())
	_ = reader.Close()
	return path
}

func folderPath(uri fyne.ListableURI, err error) string {
	if err != nil || uri == nil {
		return ""
	}
	return pathFromURI(uri)
}

func completeSaveWriter(w *modalWait, writer fyne.URIWriteCloser, err error, text bool) {
	if err != nil || writer == nil {
		w.complete(btnCancel, nil)
		return
	}
	path := pathFromURI(writer.URI())
	if text {
		tw, werr := core.NewTextWriterFromWriter(writer)
		if werr != nil {
			_ = writer.Close()
			w.complete(btnCancel, nil)
			return
		}
		w.complete(btnOk, &saveWriterResult{path: path, writer: tw})
		return
	}
	bw, werr := core.NewByteWriterFromWriter(writer)
	if werr != nil {
		_ = writer.Close()
		w.complete(btnCancel, nil)
		return
	}
	w.complete(btnOk, &saveWriterResult{path: path, writer: bw})
}

type saveWriterResult struct {
	path   string
	writer any
}

func pushSaveWriterResult(f *core.RunFrame, res *ModalCallbackResult) {
	if res.CloserIndex != btnOk {
		f.RP.DataStack.Push(core.Bool(false))
		return
	}
	sr := res.Data.(*saveWriterResult)
	f.RP.DataStack.Push(core.String(sr.path))
	f.RP.DataStack.Push(sr.writer)
	f.RP.DataStack.Push(core.Bool(true))
}

// --- InformationDialog ---

func registerInformationDialogMembers(t *core.Type) {
	// ( parent:Window title:s message:s -- )
	t.RegisterNativeClassMethod("show", func(f *core.RunFrame, _ *core.TypeObjPair) {
		message := string(f.RP.DataStack.Pop().(core.String))
		title := string(f.RP.DataStack.Pop().(core.String))
		parent := dialogParent(f.RP.DataStack.Pop())
		w := beginModalWait()
		var d dialog.Dialog
		// Construct and Show on the Fyne thread. Building a PopUp off-thread
		// leaves a 1-object overlay; dialogLayout then panics on obj[2].
		fyne.DoAndWait(func() {
			d = dialog.NewInformation(title, message, parent.Window)
			d.SetOnClosed(func() { w.complete(btnOk, nil) })
			d.Show()
		})
		_ = w.wait()
		// Ok's Hide can race the wait return; ensure the popup is gone.
		fyne.DoAndWait(func() {
			d.Hide()
		})
	})
}

func CreateRegisterInformationDialogType() {
	t := core.NewRegisteringType("InformationDialog",
		core.ReferenceTypeZeroFn,
		classOnlyCheckCoerce("InformationDialog"),
	)
	registerInformationDialogMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
}

// --- ConfirmationDialog ---

func registerConfirmationDialogMembers(t *core.Type) {
	// ( parent:Window title:s message:s -- btn:i )
	t.RegisterNativeClassMethod("show", func(f *core.RunFrame, _ *core.TypeObjPair) {
		message := string(f.RP.DataStack.Pop().(core.String))
		title := string(f.RP.DataStack.Pop().(core.String))
		parent := dialogParent(f.RP.DataStack.Pop())
		w := beginModalWait()
		var d *dialog.ConfirmDialog
		// Same UI-thread construct+Show as InformationDialog. Completion is only
		// the NewConfirm bool callback (Yes → btnOk, No/Hide → btnCancel).
		fyne.DoAndWait(func() {
			d = dialog.NewConfirm(title, message, func(yes bool) {
				w.complete(confirmCloserIndex(yes), nil)
			}, parent.Window)
			d.Show()
		})
		res := w.wait()
		fyne.DoAndWait(func() {
			d.Hide()
		})
		f.RP.DataStack.Push(core.Int(res.CloserIndex))
	})
}

func CreateRegisterConfirmationDialogType() {
	t := core.NewRegisteringType("ConfirmationDialog",
		core.ReferenceTypeZeroFn,
		classOnlyCheckCoerce("ConfirmationDialog"),
	)
	registerConfirmationDialogMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
}

// --- ColorDialog ---

func registerColorDialogMembers(t *core.Type) {
	// ( parent:Window title:s message:s -- color:Color true | false )
	t.RegisterNativeClassMethod("show", func(f *core.RunFrame, _ *core.TypeObjPair) {
		message := string(f.RP.DataStack.Pop().(core.String))
		title := string(f.RP.DataStack.Pop().(core.String))
		parent := dialogParent(f.RP.DataStack.Pop())
		w := beginModalWait()
		var d *dialog.ColorPickerDialog
		// ColorPicker: color callback ≠ dialog.callback. Ok completes in the color
		// callback; Cancel/X via SetOnClosed (once ignores the Hide after Ok).
		// selectColor then rebuilds a ModalPopUp in updateUI — DoAndWait(Hide) after
		// wait can wedge the EventLoop while the UI thread is still in that rebuild
		// (dialog already gone; Quit/prompt stop working). Cleanup is fire-and-forget.
		fyne.DoAndWait(func() {
			d = dialog.NewColorPicker(title, message, func(c color.Color) {
				w.complete(btnOk, colorFromFyne(c))
			}, parent.Window)
			d.Advanced = true
			d.SetOnClosed(func() {
				w.complete(btnCancel, nil)
			})
			d.Show()
			d.Resize(fyne.NewSize(420, 520))
		})
		res := w.wait()
		fyne.Do(func() {
			if d != nil {
				d.Hide()
			}
			if c := parent.Window.Canvas(); c != nil {
				for _, o := range append([]fyne.CanvasObject(nil), c.Overlays().List()...) {
					c.Overlays().Remove(o)
				}
			}
		})
		if res.CloserIndex == btnOk {
			f.RP.DataStack.Push(res.Data.(core.Color))
			f.RP.DataStack.Push(core.Bool(true))
		} else {
			f.RP.DataStack.Push(core.Bool(false))
		}
	})
}

func CreateRegisterColorDialogType() {
	t := core.NewRegisteringType("ColorDialog",
		core.ReferenceTypeZeroFn,
		classOnlyCheckCoerce("ColorDialog"),
	)
	registerColorDialogMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
}

// --- FileDialog ---

func resizeFileDialog(d dialog.Dialog) {
	min := d.MinSize()
	d.Resize(fyne.NewSize(min.Width*1.5, min.Height*2))
}

func registerFileDialogMembers(t *core.Type) {
	// ( parent:Window -- path:s )  "" = cancel
	t.RegisterNativeClassMethod("showOpenFile", func(f *core.RunFrame, _ *core.TypeObjPair) {
		parent := dialogParent(f.RP.DataStack.Pop())
		w := beginModalWait()
		var d dialog.Dialog
		fyne.DoAndWait(func() {
			fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
				path := closeReaderPath(reader, err)
				if path == "" {
					w.complete(btnCancel, core.String(""))
				} else {
					w.complete(btnOk, core.String(path))
				}
			}, parent.Window)
			fd.SetView(dialog.ListView)
			d = fd
			d.Show()
			resizeFileDialog(d)
		})
		res := w.wait()
		fyne.DoAndWait(func() {
			d.Hide()
		})
		f.RP.DataStack.Push(res.Data.(core.String))
	})
	// ( parent:Window -- path:s w:TextWriter true | false )
	// Confirm creates or truncates the file (Fyne storage.Writer); caller owns close.
	t.RegisterNativeClassMethod("showSaveText", func(f *core.RunFrame, _ *core.TypeObjPair) {
		parent := dialogParent(f.RP.DataStack.Pop())
		w := beginModalWait()
		var d dialog.Dialog
		fyne.DoAndWait(func() {
			fd := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
				completeSaveWriter(w, writer, err, true)
			}, parent.Window)
			fd.SetView(dialog.ListView)
			d = fd
			d.Show()
			resizeFileDialog(d)
		})
		res := w.wait()
		fyne.DoAndWait(func() {
			d.Hide()
		})
		pushSaveWriterResult(f, res)
	})
	// ( parent:Window -- path:s w:ByteWriter true | false )
	// Confirm creates or truncates the file (Fyne storage.Writer); caller owns close.
	t.RegisterNativeClassMethod("showSaveBytes", func(f *core.RunFrame, _ *core.TypeObjPair) {
		parent := dialogParent(f.RP.DataStack.Pop())
		w := beginModalWait()
		var d dialog.Dialog
		fyne.DoAndWait(func() {
			fd := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
				completeSaveWriter(w, writer, err, false)
			}, parent.Window)
			fd.SetView(dialog.ListView)
			d = fd
			d.Show()
			resizeFileDialog(d)
		})
		res := w.wait()
		fyne.DoAndWait(func() {
			d.Hide()
		})
		pushSaveWriterResult(f, res)
	})
	// ( parent:Window -- path:s )  "" = cancel
	t.RegisterNativeClassMethod("showOpenFolder", func(f *core.RunFrame, _ *core.TypeObjPair) {
		parent := dialogParent(f.RP.DataStack.Pop())
		w := beginModalWait()
		var d dialog.Dialog
		fyne.DoAndWait(func() {
			fd := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
				path := folderPath(uri, err)
				if path == "" {
					w.complete(btnCancel, core.String(""))
				} else {
					w.complete(btnOk, core.String(path))
				}
			}, parent.Window)
			fd.SetView(dialog.ListView)
			d = fd
			d.Show()
			resizeFileDialog(d)
		})
		res := w.wait()
		fyne.DoAndWait(func() {
			d.Hide()
		})
		f.RP.DataStack.Push(res.Data.(core.String))
	})
}

func CreateRegisterFileDialogType() {
	t := core.NewRegisteringType("FileDialog",
		core.ReferenceTypeZeroFn,
		classOnlyCheckCoerce("FileDialog"),
	)
	registerFileDialogMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
}

func init() {
	CreateRegisterInformationDialogType()
	CreateRegisterConfirmationDialogType()
	CreateRegisterColorDialogType()
	CreateRegisterFileDialogType()
}
