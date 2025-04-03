package ui

import (
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"image"
)

// Form represents a form for adding/editing presets
type Form struct {
	Title          *widgets.Paragraph
	NameLabel      *widgets.Paragraph
	NameField      *widgets.Paragraph
	CommandLabel   *widgets.Paragraph
	CommandField   *widgets.Paragraph
	WorkdirLabel   *widgets.Paragraph
	WorkdirField   *widgets.Paragraph
	SubmitButton   *widgets.Paragraph
	ActiveField    int
	Fields         []*widgets.Paragraph
	x1, y1, x2, y2 int // Added fields to store coordinates
}

// NewForm creates a new form
func NewForm() *Form {
	title := widgets.NewParagraph()
	title.Title = "Add Command Preset"
	title.Border = false

	nameLabel := widgets.NewParagraph()
	nameLabel.Text = "Name:"
	nameLabel.Border = false

	nameField := widgets.NewParagraph()
	nameField.Text = ""
	nameField.Border = true

	cmdLabel := widgets.NewParagraph()
	cmdLabel.Text = "Command:"
	cmdLabel.Border = false

	cmdField := widgets.NewParagraph()
	cmdField.Text = ""
	cmdField.Border = true

	dirLabel := widgets.NewParagraph()
	dirLabel.Text = "Working Directory:"
	dirLabel.Border = false

	dirField := widgets.NewParagraph()
	dirField.Text = ""
	dirField.Border = true

	submitBtn := widgets.NewParagraph()
	submitBtn.Text = "[ Save ]"
	submitBtn.Border = true

	fields := []*widgets.Paragraph{nameField, cmdField, dirField, submitBtn}

	return &Form{
		Title:        title,
		NameLabel:    nameLabel,
		NameField:    nameField,
		CommandLabel: cmdLabel,
		CommandField: cmdField,
		WorkdirLabel: dirLabel,
		WorkdirField: dirField,
		SubmitButton: submitBtn,
		ActiveField:  0,
		Fields:       fields,
	}
}

// Draw implements termui.Drawable interface
func (f *Form) Draw(buf *ui.Buffer) {
	// Draw all form components
	f.Title.Draw(buf)
	f.NameLabel.Draw(buf)
	f.NameField.Draw(buf)
	f.CommandLabel.Draw(buf)
	f.CommandField.Draw(buf)
	f.WorkdirLabel.Draw(buf)
	f.WorkdirField.Draw(buf)
	f.SubmitButton.Draw(buf)
}

// SetRect sets the form dimensions
func (f *Form) SetRect(x1, y1, x2, y2 int) {
	// Store the coordinates
	f.x1, f.y1, f.x2, f.y2 = x1, y1, x2, y2

	width := x2 - x1
	padding := 1
	fieldHeight := 3
	labelHeight := 1

	fieldWidth := width - 2*padding

	// Use these variables in calculations
	currentY := y1 + padding

	// Title
	f.Title.SetRect(x1+padding, currentY, x2-padding, currentY+labelHeight)
	currentY += labelHeight + 1

	// Name field - use fieldWidth
	f.NameLabel.SetRect(x1+padding, currentY, x1+padding+fieldWidth, currentY+labelHeight)
	currentY += labelHeight
	f.NameField.SetRect(x1+padding, currentY, x1+padding+fieldWidth, currentY+fieldHeight)
	currentY += fieldHeight + 1

	// Command field - use fieldWidth
	f.CommandLabel.SetRect(x1+padding, currentY, x1+padding+fieldWidth, currentY+labelHeight)
	currentY += labelHeight
	f.CommandField.SetRect(x1+padding, currentY, x1+padding+fieldWidth, currentY+fieldHeight)
	currentY += fieldHeight + 1

	// Working directory field - use fieldWidth
	f.WorkdirLabel.SetRect(x1+padding, currentY, x1+padding+fieldWidth, currentY+labelHeight)
	currentY += labelHeight
	f.WorkdirField.SetRect(x1+padding, currentY, x1+padding+fieldWidth, currentY+fieldHeight)
	currentY += fieldHeight + 2

	// Submit button - ensure it doesn't exceed remaining height
	buttonWidth := 20
	buttonX := x1 + (width-buttonWidth)/2
	remainingHeight := y2 - currentY - padding
	if remainingHeight >= fieldHeight {
		f.SubmitButton.SetRect(buttonX, currentY, buttonX+buttonWidth, currentY+min(fieldHeight, remainingHeight))
	} else {
		f.SubmitButton.SetRect(buttonX, currentY, buttonX+buttonWidth, currentY+1)
	}

	f.updateActiveField()
}

// updateActiveField updates the visual style of the active field
func (f *Form) updateActiveField() {
	for i, field := range f.Fields {
		if i == f.ActiveField {
			field.BorderStyle = ui.NewStyle(ui.ColorYellow)
		} else {
			field.BorderStyle = ui.NewStyle(ui.ColorWhite)
		}
	}
}

// NextField moves to the next field in the form
func (f *Form) NextField() {
	f.ActiveField = (f.ActiveField + 1) % len(f.Fields)
	f.updateActiveField()
}

// PrevField moves to the previous field in the form
func (f *Form) PrevField() {
	f.ActiveField = (f.ActiveField - 1 + len(f.Fields)) % len(f.Fields)
	f.updateActiveField()
}

// Reset clears all form fields
func (f *Form) Reset() {
	f.NameField.Text = ""
	f.CommandField.Text = ""
	f.WorkdirField.Text = ""
	f.ActiveField = 0
	f.updateActiveField()
}

// HandleInput processes keyboard input for the active field
func (f *Form) HandleInput(e ui.Event) {
	// Only process character inputs, not special keys
	if len(e.ID) == 1 {
		field := f.Fields[f.ActiveField]
		field.Text += e.ID
	} else if e.ID == "<Space>" {
		field := f.Fields[f.ActiveField]
		field.Text += " "
	} else if e.ID == "<Backspace>" && f.ActiveField != 3 { // Not button
		field := f.Fields[f.ActiveField]
		if len(field.Text) > 0 {
			field.Text = field.Text[:len(field.Text)-1]
		}
	}
}

func (f *Form) GetRect() image.Rectangle {
	return image.Rect(f.x1, f.y1, f.x2, f.y2)
}

func (f *Form) Lock() {}

func (f *Form) Unlock() {}
