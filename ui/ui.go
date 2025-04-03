// Update to View type
const (
	ListView     View = "list"
	FormView     View = "form"
	HelpView     View = "help"
	ConfirmView  View = "confirm"
	FileDialogView View = "file_dialog"  // New view type
)

// Update to App struct
type App struct {
	store        *data.Store
	currentView  View
	presetList   *widgets.List
	form         *Form
	help         *widgets.Paragraph
	confirm      *widgets.Paragraph
	statusBar    *widgets.Paragraph
	fileDialog   *FileDialog  // New field for file dialog
	selectedPresetIndex int
	pendingAction func() error
}

// Update the getCurrentView function to include the file dialog
func (a *App) getCurrentView() []ui.Drawable {
	switch a.currentView {
	case ListView:
		return []ui.Drawable{a.presetList, a.statusBar}
	case FormView:
		return []ui.Drawable{a.form, a.statusBar}
	case HelpView:
		return []ui.Drawable{a.help, a.statusBar}
	case ConfirmView:
		return []ui.Drawable{a.presetList, a.confirm, a.statusBar}
	case FileDialogView:
		if a.fileDialog != nil {
			return append(a.fileDialog.Draw(), a.statusBar)
		}
		return []ui.Drawable{a.statusBar}
	default:
		return []ui.Drawable{a.presetList, a.statusBar}
	}
}

// Add a handler for file dialog events
func (a *App) handleFileDialogEvent(e ui.Event) {
	if a.fileDialog == nil {
		a.currentView = ListView
		return
	}

	continueDialog, err := a.fileDialog.HandleEvent(e)
	if err != nil {
		a.statusBar.Text = fmt.Sprintf("Error: %v", err)
	}

	if !continueDialog {
		a.currentView = ListView
		a.fileDialog = nil
	}
}

// Update the handleEvent function to include file dialog handling
func (a *App) handleEvent(e ui.Event) {
	switch a.currentView {
	case ListView:
		a.handleListViewEvent(e)
	case FormView:
		a.handleFormViewEvent(e)
	case HelpView:
		a.handleHelpViewEvent(e)
	case ConfirmView:
		a.handleConfirmViewEvent(e)
	case FileDialogView:
		a.handleFileDialogEvent(e)
	}
}