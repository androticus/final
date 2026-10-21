package finaltui

import (
	"strings"

	"github.com/androticus/final/finalembed"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const pagePrompt = "prompt"

func (ui *appUI) buildPromptPage() tview.Primitive {
	ui.promptBody = tview.NewTextView().
		SetWrap(true).
		SetWordWrap(true).
		SetTextAlign(tview.AlignCenter)

	ui.promptOk = tview.NewButton("Ok").
		SetSelectedFunc(func() { ui.finishPrompt(BtnOk) })
	ui.promptCancel = tview.NewButton("Cancel").
		SetSelectedFunc(func() { ui.finishPrompt(BtnCancel) })

	btnRow := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(ui.promptOk, 10, 0, true).
		AddItem(nil, 2, 0, false).
		AddItem(ui.promptCancel, 10, 0, false).
		AddItem(nil, 0, 1, false)

	// Vertical center: spacers around a fixed-height body (max 3 lines).
	bodyCol := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(ui.promptBody, 3, 0, false).
		AddItem(nil, 0, 1, false)

	ui.promptForm = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(bodyCol, 0, 1, false).
		AddItem(btnRow, 3, 0, true)
	ui.promptForm.SetBorder(true).SetTitle(" Prompt ")
	ui.promptForm.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Key() {
		case tcell.KeyEscape:
			ui.finishPrompt(BtnCancel)
			return nil
		case tcell.KeyTab:
			ui.cyclePromptButton(true)
			return nil
		case tcell.KeyBacktab:
			ui.cyclePromptButton(false)
			return nil
		case tcell.KeyEnter:
			ui.activateFocusedPromptButton()
			return nil
		}
		return ev
	})

	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(ui.promptForm, 10, 0, true).
			AddItem(nil, 0, 1, false), 60, 0, true).
		AddItem(nil, 0, 1, false)
}

func (ui *appUI) startUIReqDrain() {
	Tui_.ArmUI()
	go func() {
		for req := range Tui_.UIRequests() {
			r := req
			ui.app.QueueUpdateDraw(func() {
				ui.dispatchUIReq(r)
			})
		}
	}()
}

func (ui *appUI) showPromptReq(req *UiRequest) {
	title, body, defBtn := readPromptParams(req.Params)
	ui.promptForm.SetTitle(" " + title + " ")
	ui.promptBody.SetText(trimPromptBody(body))
	ui.promptFocusOk = defBtn != BtnCancel
	ui.pages.SwitchToPage(pagePrompt)
	ui.focusPromptButton()
}

func readPromptParams(params *finalembed.DynamicObject) (title, body string, defBtn Btn) {
	title = "Prompt"
	defBtn = BtnOk
	if params == nil {
		return title, body, defBtn
	}
	if v := params.Get(DOFieldTitle.String()); v != nil {
		title = string(v.(finalembed.String))
	}
	if v := params.Get(DOFieldBodyText.String()); v != nil {
		body = string(v.(finalembed.String))
	}
	if v := params.Get(DOFieldDefaultButton.String()); v != nil {
		if c, ok := v.(finalembed.Int); ok {
			defBtn = Btn(c)
		}
	}
	return title, body, defBtn
}

func trimPromptBody(body string) string {
	lines := strings.Split(body, "\n")
	if len(lines) > 3 {
		lines = lines[:3]
	}
	return strings.Join(lines, "\n")
}

func (ui *appUI) focusPromptButton() {
	if ui.promptFocusOk {
		ui.app.SetFocus(ui.promptOk)
	} else {
		ui.app.SetFocus(ui.promptCancel)
	}
}

func (ui *appUI) cyclePromptButton(forward bool) {
	if forward {
		ui.promptFocusOk = !ui.promptFocusOk
	} else {
		ui.promptFocusOk = !ui.promptFocusOk
	}
	ui.focusPromptButton()
}

func (ui *appUI) activateFocusedPromptButton() {
	if ui.promptFocusOk {
		ui.finishPrompt(BtnOk)
	} else {
		ui.finishPrompt(BtnCancel)
	}
}

func (ui *appUI) finishPrompt(btn Btn) {
	name, _ := ui.pages.GetFrontPage()
	if name != pagePrompt {
		return
	}
	ui.pages.SwitchToPage(pageMain)
	ui.app.SetFocus(ui.prompt)
	Tui_.Complete(CompleteChoice(btn))
}

func (ui *appUI) cancelDialogIfOpen() {
	name, _ := ui.pages.GetFrontPage()
	switch name {
	case pagePrompt:
		ui.finishPrompt(BtnCancel)
	case pageFields:
		ui.finishFields(BtnCancel, nil)
	default:
		Tui_.Complete(CompleteChoice(BtnCancel))
	}
}
