package finaltui

import "github.com/androticus/final/finalembed"

// UiCmd selects which host UI modal to raise. Go enum; not a Final type.
type UiCmd int

const (
	UiCmdPrompt UiCmd = iota + 1
	UiCmdField
)

// Btn is Ok/Cancel. Exposed on Dialog as class readonly Int constants.
type Btn int

const (
	BtnCancel Btn = 0
	BtnOk     Btn = 1
)

// UiStatus is the transport-level response code (not the user's button choice).
type UiStatus int

const (
	UiStatusOK UiStatus = iota
	UiStatusError
)

// DOField is the shared string enum for DynamicObject keys (Final natives + cmd/final).
type DOField string

const (
	DOFieldTitle         DOField = "title"
	DOFieldBodyText      DOField = "bodyText"
	DOFieldDefaultButton DOField = "defaultButton"
	DOFieldChoice        DOField = "choice"
	DOFieldFields        DOField = "fields"
)

// MaxFieldDialogFields is the practical limit for one modal.
const MaxFieldDialogFields = 8

// UiRequest is Final/Tui → host UI (Go chan; not a Final Pipe).
type UiRequest struct {
	Cmd    UiCmd
	Params *finalembed.DynamicObject
}

// UiResponse is host UI → Final/Tui.
type UiResponse struct {
	Status UiStatus
	Err    string
	Result *finalembed.DynamicObject
}

func (f DOField) String() string { return string(f) }
