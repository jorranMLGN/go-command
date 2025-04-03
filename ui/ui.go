package ui

import (
	"fmt"
	"os/exec"
	"runtime"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/jorranMLGN/go-command/data"
)

// View represents the different UI views
type View string

const (
	ListView      View = "list"
	FormView      View = "form"
	HelpView      View = "help"
	ConfirmView   View = "confirm"
	FileDialogView View = "file_dialog"  // New view type
)

// App represents the application UI
type App struct {
	store              *data.Store
	currentView        View
	presetList         *widgets.List
	form               *Form
	help               *widgets.Paragraph
	confirm            *widgets.Paragraph
	statusBar          *widgets.Paragraph
	fileDialog         *FileDialog  // New field for file dialog
	selectedPresetIndex int
	pendingAction      func() error
}

// NewApp creates a new UI application
func NewApp(store *data.Store) *App {
	return &App{
		store:       store,
		currentView: ListView,
		selectedPresetIndex: -1,
	}
}

// Run starts the UI
func (a *App) Run() error {
	if err := ui.Init(); err != nil {
		return fmt.Errorf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	a.setupUI()
	a.updateListItems()

	uiEvents := ui.PollEvents()
	for {
		e := <-uiEvents
		switch e.ID {
		case "q", "<C-c>":
			return nil
		default:
			a.handleEvent(e)
		}
		ui.Render(a.getCurrentView()...)
	}
}

// setupUI initializes all UI components
func (a *App) setupUI() {
	// List view
	a.presetList = widgets.NewList()
	a.presetList.Title = "Command Presets"
	a.presetList.TextStyle = ui.NewStyle(ui.ColorYellow)
	a.presetList.WrapText = true
	a.presetList.SelectedRowStyle = ui.NewStyle(ui.ColorBlack, ui.ColorYellow)

	// Form view
	a.form = NewForm()

	// Help view
	a.help = widgets.NewParagraph()
	a.help.Title = "Help"
	a.help.Text = `
    List View:
      ↑/↓: Navigate presets
      Enter: Execute selected command
      a: Add new preset
      d: Delete selected preset
      i: Import presets from file
      e: Export presets to file
      h: Show this help
      q: Quit

    Form View:
      Tab: Navigate form fields
      Enter: Submit form
      Esc: Cancel and return to list

    General:
      Ctrl+C: Quit application
	`

	// Confirm dialog
	a.confirm = widgets.NewParagraph()
	a.confirm.Title = "Confirm"

	// Status bar
	a.statusBar = widgets.NewParagraph()
	a.statusBar.Border = false

	a.resize()
}

// resize updates UI component sizes based on terminal dimensions
func (a *App) resize() {
	termWidth, termHeight := ui.TerminalDimensions()

	// Size list view
	a.presetList.SetRect(0, 0, termWidth, termHeight-2)

	// Size form view
	a.form.SetRect(5, 5, termWidth-5, termHeight-5)

	// Size help view
	a.help.SetRect(5, 5, termWidth-5, termHeight-5)

	// Size confirm dialog
	a.confirm.SetRect(10, 10, termWidth-10, 15)

	// Size status bar
	a.statusBar.SetRect(0, termHeight-2, termWidth, termHeight)
}

// updateListItems updates the list with current presets
func (a *App) updateListItems() {
	items := make([]string, 0, len(a.store.PresetList.Presets))
	for _, preset := range a.store.PresetList.Presets {
		items = append(items, fmt.Sprintf("%s | %s | %s", preset.Name, preset.Command, preset.WorkingDir))
	}
	a.presetList.Rows = items
}

// getCurrentView returns the UI elements for the current view
func (a *App) getCurrentView() []ui.Drawable {
	switch a.currentView {
	case ListView:
		return []ui.Drawable{a.presetList, a.statusBar}
	case FormView:
		// Form implements Drawable now
		return []ui.Drawable{a.form, a.statusBar}
	case HelpView:
		return []ui.Drawable{a.help, a.statusBar}
	case ConfirmView:
		return []ui.Drawable{a.presetList, a.confirm, a.statusBar}
	case FileDialogView:
		if a.fileDialog != nil {
			// FileDialog implements Drawable now
			return []ui.Drawable{a.fileDialog, a.statusBar}
		}
		return []ui.Drawable{a.statusBar}
	default:
		return []ui.Drawable{a.presetList, a.statusBar}
	}
}

// handleEvent processes UI events
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

// handleListViewEvent handles events in the list view
func (a *App) handleListViewEvent(e ui.Event) {
	switch e.ID {
	case "j", "<Down>":
		a.presetList.ScrollDown()
	case "k", "<Up>":
		a.presetList.ScrollUp()
	case "<Enter>":
		a.executeSelectedCommand()
	case "a":
		a.form.Reset()
		a.currentView = FormView
		a.statusBar.Text = "Add new command preset. Press Esc to cancel."
	case "d":
		if len(a.store.PresetList.Presets) > 0 && a.presetList.SelectedRow >= 0 {
			a.selectedPresetIndex = a.presetList.SelectedRow
			a.confirm.Text = "Are you sure you want to delete this preset? (y/n)"
			a.pendingAction = a.deleteSelectedPreset
			a.currentView = ConfirmView
		}
	case "i":
		a.importPresets()
	case "e":
		a.exportPresets()
	case "h":
		a.currentView = HelpView
	}
}

// handleFormViewEvent handles events in the form view
func (a *App) handleFormViewEvent(e ui.Event) {
	switch e.ID {
	case "<Escape>":
		a.currentView = ListView
		a.statusBar.Text = "Form canceled."
	case "<Tab>":
		a.form.NextField()
	case "<Backtab>":
		a.form.PrevField()
	case "<Enter>":
		if a.form.ActiveField == 3 { // Submit button
			preset := data.Preset{
				Name:       a.form.NameField.Text,
				Command:    a.form.CommandField.Text,
				WorkingDir: a.form.WorkdirField.Text,
			}
			a.store.AddPreset(preset)
			if err := a.store.Save(); err != nil {
				a.statusBar.Text = fmt.Sprintf("Error saving preset: %v", err)
			} else {
				a.statusBar.Text = "Preset added successfully."
				a.updateListItems()
				a.currentView = ListView
			}
		} else {
			a.form.NextField()
		}
	default:
		a.form.HandleInput(e)
	}
}

// handleHelpViewEvent handles events in the help view
func (a *App) handleHelpViewEvent(e ui.Event) {
	switch e.ID {
	case "<Escape>", "<Enter>", "h":
		a.currentView = ListView
	}
}

// handleConfirmViewEvent handles events in confirmation dialogs
func (a *App) handleConfirmViewEvent(e ui.Event) {
	switch e.ID {
	case "y", "Y":
		if a.pendingAction != nil {
			if err := a.pendingAction(); err != nil {
				a.statusBar.Text = fmt.Sprintf("Error: %v", err)
			}
			a.pendingAction = nil
		}
		a.currentView = ListView
	case "n", "N", "<Escape>":
		a.pendingAction = nil
		a.currentView = ListView
		a.statusBar.Text = "Action canceled."
	}
}

// handleFileDialogEvent handles events in the file dialog
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

// executeSelectedCommand runs the selected command in a new terminal window
func (a *App) executeSelectedCommand() {
	if a.presetList.SelectedRow < 0 || a.presetList.SelectedRow >= len(a.store.PresetList.Presets) {
		a.statusBar.Text = "No preset selected."
		return
	}

	preset := a.store.PresetList.Presets[a.presetList.SelectedRow]
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		// Use common terminals on Linux
		cmd = exec.Command("xterm", "-e", fmt.Sprintf("cd %s && %s; bash", preset.WorkingDir, preset.Command))
		// Alternative: terminal -e "cd dir && command"
	case "darwin":
		// macOS terminal
		applescript := fmt.Sprintf(`tell app "Terminal" to do script "cd %s && %s"`, preset.WorkingDir, preset.Command)
		cmd = exec.Command("osascript", "-e", applescript)
	case "windows":
		// Windows command prompt
		cmd = exec.Command("cmd", "/C", "start", "cmd", "/K", fmt.Sprintf("cd /d %s && %s", preset.WorkingDir, preset.Command))
	default:
		a.statusBar.Text = "Unsupported operating system."
		return
	}

	if err := cmd.Start(); err != nil {
		a.statusBar.Text = fmt.Sprintf("Error executing command: %v", err)
	} else {
		a.statusBar.Text = "Command started in new terminal."
	}
}

// deleteSelectedPreset removes the selected preset
func (a *App) deleteSelectedPreset() error {
	if err := a.store.RemovePreset(a.selectedPresetIndex); err != nil {
		return err
	}

	if err := a.store.Save(); err != nil {
		return err
	}

	a.updateListItems()
	a.statusBar.Text = "Preset deleted successfully."
	return nil
}

// importPresets imports presets from a file using a file dialog
func (a *App) importPresets() {
	// Create file dialog
	dialog := NewFileDialog(false, func(filePath string) error {
		err := a.store.ImportPresets(filePath)
		if err != nil {
			return err
		}

		// Save the combined presets
		if err := a.store.Save(); err != nil {
			return err
		}

		// Update the UI
		a.updateListItems()
		a.statusBar.Text = "Presets imported successfully."
		return nil
	})

	// Show file dialog
	a.fileDialog = dialog
	a.currentView = FileDialogView
	a.statusBar.Text = "Select a JSON file to import presets."
}

// exportPresets exports presets to a file using a file dialog
func (a *App) exportPresets() {
	// Create file dialog
	dialog := NewFileDialog(true, func(filePath string) error {
		err := a.store.ExportPresets(filePath)
		if err != nil {
			return err
		}

		a.statusBar.Text = "Presets exported successfully to " + filePath
		return nil
	})

	// Show file dialog
	a.fileDialog = dialog
	a.currentView = FileDialogView
	a.statusBar.Text = "Choose a location to save presets."
}