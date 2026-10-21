package fyneui

import (
	"fyne.io/fyne/v2"
	"github.com/androticus/final/internal/core"
)

// Container is the Host Final type for bare *fyne.Container values returned by
// layout factories (Box, Grid, Border). Fyne's driver type-switches on
// *fyne.Container, so wrappers must not enter the object tree.

func CreateRegisterContainerType() {
	t := core.NewRegisteringType("Container",
		core.MakeHostTypeZero[*fyne.Container](),
		core.MakeHostTypeCheckCoerce[*fyne.Container]("Container"),
	)
	applyCanvasMixin[*fyne.Container](t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	core.RegisterHostType[*fyne.Container](t)
}

func init() {
	CreateRegisterContainerType()
}
