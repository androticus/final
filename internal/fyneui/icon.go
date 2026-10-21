package fyneui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"github.com/androticus/final/internal/core"
)

// Icon is a class-only factory type. Each class property returns a Fyne
// fyne.Resource (opaque on the stack) for factories such as Button.newWithIcon.
// The Fyne Icon widget is not surfaced yet.

func iconClassConst(t *core.Type, name string, resFn func() fyne.Resource) {
	t.RegisterClassProperty(name, "Any",
		func(_ any) any { return resFn() },
		nil,
	)
}

func themeIcon(name fyne.ThemeIconName) func() fyne.Resource {
	return func() fyne.Resource { return theme.Current().Icon(name) }
}

func registerIconMembers(t *core.Type) {
	// Theme icons in https://docs.fyne.io/explore/icons/ list order (IconName without prefix).
	iconClassConst(t, "Account", theme.AccountIcon)
	iconClassConst(t, "ArrowDropDown", theme.MenuDropDownIcon)
	iconClassConst(t, "ArrowDropUp", theme.MenuDropUpIcon)
	iconClassConst(t, "BrokenImage", theme.BrokenImageIcon)
	iconClassConst(t, "Calendar", theme.CalendarIcon)
	iconClassConst(t, "Cancel", theme.CancelIcon)
	iconClassConst(t, "CheckButton", theme.CheckButtonIcon)
	iconClassConst(t, "CheckButtonChecked", theme.CheckButtonCheckedIcon)
	iconClassConst(t, "CheckButtonFill", theme.CheckButtonFillIcon)
	iconClassConst(t, "CheckButtonPartial", themeIcon(theme.IconNameCheckButtonPartial))
	iconClassConst(t, "ColorAchromatic", theme.ColorAchromaticIcon)
	iconClassConst(t, "ColorChromatic", theme.ColorChromaticIcon)
	iconClassConst(t, "ColorPalette", theme.ColorPaletteIcon)
	iconClassConst(t, "Computer", theme.ComputerIcon)
	iconClassConst(t, "Confirm", theme.ConfirmIcon)
	iconClassConst(t, "ContentAdd", theme.ContentAddIcon)
	iconClassConst(t, "ContentClear", theme.ContentClearIcon)
	iconClassConst(t, "ContentCopy", theme.ContentCopyIcon)
	iconClassConst(t, "ContentCut", theme.ContentCutIcon)
	iconClassConst(t, "ContentPaste", theme.ContentPasteIcon)
	iconClassConst(t, "ContentRedo", theme.ContentRedoIcon)
	iconClassConst(t, "ContentRemove", theme.ContentRemoveIcon)
	iconClassConst(t, "ContentUndo", theme.ContentUndoIcon)
	iconClassConst(t, "Delete", theme.DeleteIcon)
	iconClassConst(t, "Desktop", theme.DesktopIcon)
	iconClassConst(t, "Document", theme.DocumentIcon)
	iconClassConst(t, "DocumentCreate", theme.DocumentCreateIcon)
	iconClassConst(t, "DocumentPrint", theme.DocumentPrintIcon)
	iconClassConst(t, "DocumentSave", theme.DocumentSaveIcon)
	iconClassConst(t, "Download", theme.DownloadIcon)
	iconClassConst(t, "DragCornerIndicator", themeIcon(theme.IconNameDragCornerIndicator))
	iconClassConst(t, "Error", theme.ErrorIcon)
	iconClassConst(t, "File", theme.FileIcon)
	iconClassConst(t, "FileApplication", theme.FileApplicationIcon)
	iconClassConst(t, "FileAudio", theme.FileAudioIcon)
	iconClassConst(t, "FileImage", theme.FileImageIcon)
	iconClassConst(t, "FileText", theme.FileTextIcon)
	iconClassConst(t, "FileVideo", theme.FileVideoIcon)
	iconClassConst(t, "Folder", theme.FolderIcon)
	iconClassConst(t, "FolderNew", theme.FolderNewIcon)
	iconClassConst(t, "FolderOpen", theme.FolderOpenIcon)
	iconClassConst(t, "Grid", theme.GridIcon)
	iconClassConst(t, "Help", theme.HelpIcon)
	iconClassConst(t, "History", theme.HistoryIcon)
	iconClassConst(t, "Home", theme.HomeIcon)
	iconClassConst(t, "Info", theme.InfoIcon)
	iconClassConst(t, "List", theme.ListIcon)
	iconClassConst(t, "Login", theme.LoginIcon)
	iconClassConst(t, "Logout", theme.LogoutIcon)
	iconClassConst(t, "MailAttachment", theme.MailAttachmentIcon)
	iconClassConst(t, "MailCompose", theme.MailComposeIcon)
	iconClassConst(t, "MailForward", theme.MailForwardIcon)
	iconClassConst(t, "MailReply", theme.MailReplyIcon)
	iconClassConst(t, "MailReplyAll", theme.MailReplyAllIcon)
	iconClassConst(t, "MailSend", theme.MailSendIcon)
	iconClassConst(t, "MediaFastForward", theme.MediaFastForwardIcon)
	iconClassConst(t, "MediaFastRewind", theme.MediaFastRewindIcon)
	iconClassConst(t, "MediaMusic", theme.MediaMusicIcon)
	iconClassConst(t, "MediaPause", theme.MediaPauseIcon)
	iconClassConst(t, "MediaPhoto", theme.MediaPhotoIcon)
	iconClassConst(t, "MediaPlay", theme.MediaPlayIcon)
	iconClassConst(t, "MediaRecord", theme.MediaRecordIcon)
	iconClassConst(t, "MediaReplay", theme.MediaReplayIcon)
	iconClassConst(t, "MediaSkipNext", theme.MediaSkipNextIcon)
	iconClassConst(t, "MediaSkipPrevious", theme.MediaSkipPreviousIcon)
	iconClassConst(t, "MediaStop", theme.MediaStopIcon)
	iconClassConst(t, "MediaVideo", theme.MediaVideoIcon)
	iconClassConst(t, "Menu", theme.MenuIcon)
	iconClassConst(t, "MenuExpand", theme.MenuExpandIcon)
	iconClassConst(t, "MoreHorizontal", theme.MoreHorizontalIcon)
	iconClassConst(t, "MoreVertical", theme.MoreVerticalIcon)
	iconClassConst(t, "MoveDown", theme.MoveDownIcon)
	iconClassConst(t, "MoveUp", theme.MoveUpIcon)
	iconClassConst(t, "NavigateBack", theme.NavigateBackIcon)
	iconClassConst(t, "NavigateNext", theme.NavigateNextIcon)
	iconClassConst(t, "Question", theme.QuestionIcon)
	iconClassConst(t, "RadioButton", theme.RadioButtonIcon)
	iconClassConst(t, "RadioButtonChecked", theme.RadioButtonCheckedIcon)
	iconClassConst(t, "RadioButtonFill", theme.RadioButtonFillIcon)
	iconClassConst(t, "Search", theme.SearchIcon)
	iconClassConst(t, "SearchReplace", theme.SearchReplaceIcon)
	iconClassConst(t, "Settings", theme.SettingsIcon)
	iconClassConst(t, "Storage", theme.StorageIcon)
	iconClassConst(t, "Upload", theme.UploadIcon)
	iconClassConst(t, "ViewFullScreen", theme.ViewFullScreenIcon)
	iconClassConst(t, "ViewRefresh", theme.ViewRefreshIcon)
	iconClassConst(t, "ViewRestore", theme.ViewRestoreIcon)
	iconClassConst(t, "ViewZoomFit", theme.ZoomFitIcon)
	iconClassConst(t, "ViewZoomIn", theme.ZoomInIcon)
	iconClassConst(t, "ViewZoomOut", theme.ZoomOutIcon)
	iconClassConst(t, "Visibility", theme.VisibilityIcon)
	iconClassConst(t, "VisibilityOff", theme.VisibilityOffIcon)
	iconClassConst(t, "VolumeDown", theme.VolumeDownIcon)
	iconClassConst(t, "VolumeMute", theme.VolumeMuteIcon)
	iconClassConst(t, "VolumeUp", theme.VolumeUpIcon)
	iconClassConst(t, "Warning", theme.WarningIcon)
	iconClassConst(t, "WindowClose", theme.WindowCloseIcon)
	iconClassConst(t, "WindowMaximize", theme.WindowMaximizeIcon)
	iconClassConst(t, "WindowMinimize", theme.WindowMinimizeIcon)
}

func CreateRegisterIconType() {
	t := core.NewRegisteringType("Icon",
		core.ReferenceTypeZeroFn,
		classOnlyCheckCoerce("Icon"),
	)
	registerIconMembers(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
}

func init() {
	CreateRegisterIconType()
}
