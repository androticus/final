package fyneui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/androticus/final/internal/core"
)

func NewRichTextWithText(text string) *widget.RichText {
	return widget.NewRichTextWithText(text)
}

func NewRichTextFromMarkdown(content string) *widget.RichText {
	return widget.NewRichTextFromMarkdown(content)
}

func registerRichTextMembers(t *core.Type) {
	// ( text:String -- )
	t.RegisterNativeMethod("setText", func(f *core.RunFrame, ths *core.TypeObjPair) {
		text := string(f.RP.DataStack.Pop().(core.String))
		rt := ths.Obj.(*widget.RichText)
		fyne.Do(func() {
			rt.Segments = []widget.RichTextSegment{
				&widget.TextSegment{Style: widget.RichTextStyleInline, Text: text},
			}
			rt.Refresh()
		})
	})

	// ( content:String -- )
	t.RegisterNativeMethod("parseMarkdown", func(f *core.RunFrame, ths *core.TypeObjPair) {
		content := string(f.RP.DataStack.Pop().(core.String))
		rt := ths.Obj.(*widget.RichText)
		fyne.Do(func() {
			rt.ParseMarkdown(content)
		})
	})

	// ( text:String -- rt:RichText )
	t.RegisterNativeClassMethod("newWithText", func(f *core.RunFrame, _ *core.TypeObjPair) {
		text := string(f.RP.DataStack.Pop().(core.String))
		f.RP.DataStack.Push(NewRichTextWithText(text))
	})

	// ( content:String -- rt:RichText )
	t.RegisterNativeClassMethod("newFromMarkdown", func(f *core.RunFrame, _ *core.TypeObjPair) {
		content := string(f.RP.DataStack.Pop().(core.String))
		f.RP.DataStack.Push(NewRichTextFromMarkdown(content))
	})
}

func CreateRegisterRichTextType() {
	t := core.NewRegisteringType("RichText",
		core.MakeHostTypeZero[*widget.RichText](),
		core.MakeHostTypeCheckCoerce[*widget.RichText]("RichText"),
	)
	registerRichTextMembers(t)
	applyCanvasMixin[*widget.RichText](t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	core.RegisterHostType[*widget.RichText](t)
}

func init() {
	CreateRegisterRichTextType()
}
