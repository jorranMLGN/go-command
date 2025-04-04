package ui

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/jorranMLGN/go-command/data"
)

// View represents the different UI views
type View string

const (
	ListView       View = "list"
	FormView       View = "form"
	HelpView       View = "help"
	ConfirmView    View = "confirm"
	FileDialogView View = "file_dialog"
	HistoryView    View = "history" // Correct declaration
)

// App represents the application UI
type App struct {
	store               *data.Store
	currentView         View
	presetList          *widgets.List
	form                *Form
	help                *widgets.Paragraph
	confirm             *widgets.Paragraph
	statusBar           *widgets.Paragraph
	fileDialog          *FileDialog
	history             *History
	commandHistory      []CommandHistoryEntry
	selectedPresetIndex int
	pendingAction       func() error
}

// NewApp creates a new UI application
func NewApp(store *data.Store) *App {
	return &App{
		store:               store,
		currentView:         ListView,
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
	ui.Render(a.getCurrentView()...)

	uiEvents := ui.PollEvents()
	for {
		e := <-uiEvents
		switch e.ID {
		case "q", "<C-c>":
			return nil
		case "<Resize>":
			a.handleResize()
		default:
			a.handleEvent(e)
		}
		ui.Render(a.getCurrentView()...)
	}
}

// handleResize responds to terminal window resize events
func (a *App) handleResize() {
	a.resize()

	// If file dialog is active, resize it
	if a.currentView == FileDialogView && a.fileDialog != nil {
		termWidth, termHeight := ui.TerminalDimensions()
		a.fileDialog.SetRect(5, 5, termWidth-5, termHeight-7)
	}

	// Form needs to update its component positions
	if a.currentView == FormView {
		a.form.UpdateLayout()
	}

	a.updateListItems()
	a.updateStatusBarTips()
}

// setupUI initializes all UI components
func (a *App) setupUI() {
	// List view
	a.presetList = widgets.NewList()
	a.presetList.Title = "Command Presets"
	a.presetList.TextStyle = ui.NewStyle(ui.ColorWhite)
	a.presetList.WrapText = true
	a.presetList.SelectedRowStyle = ui.NewStyle(ui.ColorWhite, ui.ColorMagenta, ui.ModifierBold)
	a.presetList.SelectedRow = -1 // Explicitly set to -1 initially

	// Form view
	a.form = NewForm()

	// Help view
	a.help = widgets.NewParagraph()
	a.help.Title = "Help"
	a.help.Text = `
    List View:
      ↑/↓: Navigate presets
   Space: Toggle selected preset
      ->: Execute selected preset
      Enter: Execute Toggled Presets
      a: Add new preset
      d: Delete selected preset
      i: Import presets from file
      e: Export presets to file
      v: View command history
      h: Show this help
      q: Quit

    Form View:
      Tab: Navigate form fields
      Enter: Submit form
      Esc: Cancel and return to list
      
    History View:
      ↑/↓: Navigate history
      Esc: Return to list view

    General:
      Ctrl+C: Quit application
`

	// Confirm dialog
	a.confirm = widgets.NewParagraph()
	a.confirm.Title = "Confirm"

	// Status bar
	a.statusBar = widgets.NewParagraph()
	a.statusBar.Border = true
	a.statusBar.Title = "Status | Made by JorranMLGN"
	a.statusBar.TextStyle = ui.NewStyle(ui.ColorWhite, ui.ColorClear, ui.ModifierBold)
	a.statusBar.Text = "Welcome! Press 'a' to add a preset, 'h' for help"

	// command history
	a.history = NewHistoryView()
	a.commandHistory = make([]CommandHistoryEntry, 0)

	a.resize()
	a.updateStatusBarTips()
}

// updateStatusBarTips updates the status bar with contextual keybinding hints
func (a *App) updateStatusBarTips() {
	switch a.currentView {
	case ListView:
		if len(a.store.PresetList.Presets) == 0 {
			a.statusBar.Text = "No presets | a: Add | i: Import | h: Help | q: Quit"
		} else {
			a.statusBar.Text = "↑/↓: Navigate | Space: Toggle | Enter: Run selected | → : Execute | a: Add | d: Delete | h: Help | i: Import | e: Export | v: View history | q: Quit"
		}
	case FormView:
		a.statusBar.Text = "Tab: Next field | Shift+Tab: Previous | Enter: Submit | Esc: Cancel"
	case HelpView:
		a.statusBar.Text = "Press Esc or Enter to return to list view"
	case ConfirmView:
		a.statusBar.Text = "y: Confirm | n: Cancel | Esc: Cancel"
	case FileDialogView:
		a.statusBar.Text = "↑/↓: Navigate | Enter: Select | Esc: Cancel"
	case HistoryView:
		a.statusBar.Text = "Command History | ↑/↓: Navigate | Esc: Return to list"
	}
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
	if len(a.store.PresetList.Presets) == 0 {
		// Make status bar larger when no presets are available
		a.statusBar.SetRect(0, termHeight-7, termWidth, termHeight)
	} else {
		a.statusBar.SetRect(0, termHeight-5, termWidth, termHeight)
	}
	a.history.SetRect(0, 0, termWidth, termHeight-5)

}

// updateListItems updates the list of presets in the UI
func (a *App) updateListItems() {
	items := make([]string, 0, len(a.store.PresetList.Presets))
	for _, preset := range a.store.PresetList.Presets {
		var name string
		enabledMark := " [ ]"

		if preset.Enabled {
			enabledMark = " [✓]"
			name = strings.ToUpper(preset.Name) // Make toggled items bold via uppercase
		} else {
			name = preset.Name
		}

		items = append(items, fmt.Sprintf("%-3s %-20.20s | %-40.40s | %-20.50s",
			enabledMark, name, preset.Command, preset.WorkingDir))
	}

	a.presetList.Rows = items

	// Set SelectedRow based on whether there are items
	if len(items) > 0 {
		if a.presetList.SelectedRow < 0 || a.presetList.SelectedRow >= len(items) {
			a.presetList.SelectedRow = 0
		}
	} else {
		a.presetList.SelectedRow = -1
	}
}

// executePreset runs the command for the given preset index
func (a *App) runEnabledPresets() {
	enabledPresets := 0
	for _, preset := range a.store.PresetList.Presets {
		if preset.Enabled {
			enabledPresets++
		}
	}

	if enabledPresets == 0 {
		a.statusBar.Text = "No presets selected. Toggle presets with 'Space' key."
		return
	}

	a.statusBar.Text = fmt.Sprintf("Executing %d selected commands...", enabledPresets)

	// Run all enabled presets
	executedCount := 0
	for i, preset := range a.store.PresetList.Presets {
		if preset.Enabled {
			if err := a.executePreset(i); err != nil {
				a.statusBar.Text = fmt.Sprintf("Error executing preset '%s': %v",
					preset.Name, err)
				return
			}
			executedCount++
		}
	}

	a.statusBar.Text = fmt.Sprintf("Successfully executed %d commands.", executedCount)
}

// executePreset runs the command for the given preset index
func (a *App) executePreset(index int) error {
	if index < 0 || index >= len(a.store.PresetList.Presets) {
		return fmt.Errorf("invalid preset index")
	}

	preset := a.store.PresetList.Presets[index]
	var cmd *exec.Cmd

	// Add command to history
	a.history.AddEntry(
		CommandHistoryEntry{
			Timestamp:  time.Now().Format(time.RFC3339),
			Command:    preset.Command,
			WorkingDir: preset.WorkingDir,
			Status:     "Started",
		},
	)

	switch runtime.GOOS {
	case "linux":
		terminalCmd, err := getLinuxTerminalCommand(preset.WorkingDir, preset.Command)
		if err != nil {
			return err
		}
		cmd = terminalCmd
	case "darwin":
		applescript := fmt.Sprintf(`tell app "Terminal" to do script "cd %s && %s"`,
			preset.WorkingDir, preset.Command)
		cmd = exec.Command("osascript", "-e", applescript)
	case "windows":
		cmd = exec.Command("cmd", "/C", "start", "cmd", "/K",
			fmt.Sprintf("cd /d %s && %s", preset.WorkingDir, preset.Command))
	default:
		return fmt.Errorf("unsupported operating system")
	}

	return cmd.Start()
}

// toggleSelectedPreset toggles the enabled state of the selected preset
func (a *App) toggleSelectedPreset() {
	if len(a.store.PresetList.Presets) == 0 || a.presetList.SelectedRow < 0 {
		return
	}

	idx := a.presetList.SelectedRow
	a.store.PresetList.Presets[idx].Enabled = !a.store.PresetList.Presets[idx].Enabled

	a.updateListItems()

	if err := a.store.Save(); err != nil {
		a.statusBar.Text = fmt.Sprintf("Error saving preset state: %v", err)
	}
}

// getCurrentView returns the UI elements for the current view
func (a *App) getCurrentView() []ui.Drawable {
	switch a.currentView {
	case ListView:
		// Check if presets list is empty or selected row is -1
		if len(a.store.PresetList.Presets) == 0 || a.presetList.SelectedRow == -1 {
			a.statusBar.Text = "No command presets available\n\n" +
				"Available commands:\n" +
				"• Press 'a' to add a new preset\n" +
				"• Press 'i' to import presets from file\n" +
				"• Press 'h' to view all commands and help\n" +
				"• Press 'q' to quit"
			return []ui.Drawable{a.statusBar}
		}
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
	case HistoryView:
		return []ui.Drawable{a.history, a.statusBar}
	default:
		return []ui.Drawable{a.presetList, a.statusBar}
	}
}

// Add a new handler for history view events
func (a *App) handleHistoryViewEvent(e ui.Event) {
	switch e.ID {
	case "<Escape>", "<Enter>":
		a.currentView = ListView
	default:
		a.history.HandleEvent(e)
	}
}

// handleEvent processes UI events
func (a *App) handleEvent(e ui.Event) {
	prevView := a.currentView

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
	case HistoryView:
		a.handleHistoryViewEvent(e)
	}

	if prevView != a.currentView {
		a.updateStatusBarTips()
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
		a.runEnabledPresets()
	case "<Right>":
		a.executeSelectedCommand()
	case "<Space>":
		a.toggleSelectedPreset()
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
	case "v": // New shortcut for viewing history
		a.currentView = HistoryView
		a.updateStatusBarTips()
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

// getLinuxTerminalCommand returns the appropriate command for the available terminal
func getLinuxTerminalCommand(workingDir, command string) (*exec.Cmd, error) {
	// Escape the command properly for shell execution
	shellCommand := fmt.Sprintf("cd '%s' && %s; exec bash",
		strings.Replace(workingDir, "'", "'\\''", -1),
		command)

	// Try common terminals with correct arguments
	terminals := []struct {
		name string
		args []string
	}{
		// Most common terminal on Ubuntu/GNOME
		{"gnome-terminal", []string{"--", "bash", "-c", shellCommand}},
		// KDE terminal
		{"konsole", []string{"--noclose", "-e", "bash", "-c", shellCommand}},
		// XFCE terminal
		{"xfce4-terminal", []string{"--hold", "-e", "bash -c '" + shellCommand + "'"}},
		// Popular alternative terminal
		{"terminator", []string{"-e", "bash -c '" + shellCommand + "'"}},
		// Fallback terminal
		{"xterm", []string{"-hold", "-e", "bash -c '" + shellCommand + "'"}},
		// Other terminals
		{"tilix", []string{"-e", "bash -c '" + shellCommand + "'"}},
		{"mate-terminal", []string{"--disable-factory", "-e", "bash -c '" + shellCommand + "'"}},
	}

	for _, term := range terminals {
		path, err := exec.LookPath(term.name)
		if err == nil {
			return exec.Command(path, term.args...), nil
		}
	}

	return nil, fmt.Errorf("no suitable terminal emulator found")
}

// executeSelectedCommand runs the selected command in a new terminal window
func (a *App) executeSelectedCommand() {
	if len(a.store.PresetList.Presets) == 0 {
		a.statusBar.Text = "No presets available. Press 'a' to add a new preset."
		return
	}

	if a.presetList.SelectedRow < 0 || a.presetList.SelectedRow >= len(a.store.PresetList.Presets) {
		a.statusBar.Text = "No preset selected."
		return
	}

	preset := a.store.PresetList.Presets[a.presetList.SelectedRow]
	var cmd *exec.Cmd

	// Add command to history
	a.history.AddEntry(
		CommandHistoryEntry{
			Timestamp:  time.Now().Format(time.RFC3339),
			Command:    preset.Command,
			WorkingDir: preset.WorkingDir,
			Status:     "Started",
		},
	)

	switch runtime.GOOS {
	case "linux":
		// Find an available terminal emulator
		terminalCmd, err := getLinuxTerminalCommand(preset.WorkingDir, preset.Command)
		if err != nil {
			a.statusBar.Text = "Error: No suitable terminal emulator found. Please install one of: gnome-terminal, konsole, xfce4-terminal, terminator, xterm."
			return
		}
		cmd = terminalCmd
	case "darwin":
		// macOS terminal (unchanged)
		applescript := fmt.Sprintf(`tell app "Terminal" to do script "cd %s && %s"`, preset.WorkingDir, preset.Command)
		cmd = exec.Command("osascript", "-e", applescript)
	case "windows":
		// Windows command prompt (unchanged)
		cmd = exec.Command("cmd", "/C", "start", "cmd", "/K", fmt.Sprintf("cd /d %s && %s", preset.WorkingDir, preset.Command))
	default:
		a.statusBar.Text = "Unsupported operating system."
		return
	}

	// Start the command but don't wait for it
	if err := cmd.Start(); err != nil {
		a.statusBar.Text = fmt.Sprintf("Error executing command: %v", err)
		if runtime.GOOS == "linux" {
			a.statusBar.Text += "\nPlease install a terminal emulator (gnome-terminal, konsole, xfce4-terminal, terminator, xterm)."
		}
	} else {
		a.statusBar.Title = "Command started in new terminal."

		// after 5 seconds, update the status bar back to the tips

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
