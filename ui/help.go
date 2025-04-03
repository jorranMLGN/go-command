package ui

import (
	"github.com/gizak/termui/v3/widgets"
)

// NewHelpView creates a help view widget
func NewHelpView() *widgets.Paragraph {
	help := widgets.NewParagraph()
	help.Title = "Help"
	help.Text = `
Command Presets Manager - Help

Navigation:
  ↑/↓ or j/k: Navigate up and down in lists
  Tab: Move to next field in forms
  Shift+Tab: Move to previous field in forms

Main Screen:
  Enter: Execute selected command preset
  a: Add a new command preset
  d: Delete the selected preset
  i: Import presets from a file
  e: Export presets to a file
  h: Show this help screen
  q or Ctrl+C: Quit application

Form Screen:
  Enter: Submit form or move to next field
  Esc: Cancel and return to previous screen

About:
  This application allows you to manage and execute command presets
  with specific working directories across different operating systems.
`
	return help
}