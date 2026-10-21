package fyneui

import (
	"github.com/androticus/final/internal/core"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

const fyneAppID = "ai.androticus.final-gui.preferences"

type App struct {
	fyne.App
}

func NewApp() *App {
	app.SetMetadata(fyne.AppMetadata{
		ID: fyneAppID,
		Migrations: map[string]bool{
			"fyneDo": true,
		},
	})
	return &App{App: app.NewWithID(fyneAppID)}
}

func (a *App) FinalType() *core.Type {
	return appType
}

func registerAppMembers(t *core.Type) {
	t.RegisterClassProperty("app", "App",
		func(_ any) any { return TheApp },
		nil,
	)
}

func CreateRegisterAppType() {
	t := core.NewRegisteringType("App",
		core.ReferenceTypeZeroFn,
		core.MakePointerTypeCheckCoerce[App]("App"),
	)
	registerAppMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	appType = t
}

var (
	TheApp  *App
	appType *core.Type
	_       core.Typed = (*App)(nil)
)

func init() {
	core.EnableBootstrapTranche3()
	TheApp = NewApp()
	CreateRegisterAppType()
}
