package ui

import (
	"image"
	"os"
	"path/filepath"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

// FileDialog represents a file selection/input dialog
type FileDialog struct {
	Title          *widgets.Paragraph
	PathInput      *widgets.Paragraph
	CurrentPath    *widgets.Paragraph
	FileList       *widgets.List
	InfoText       *widgets.Paragraph
	IsOpen         bool
	IsExport       bool
	Callback       func(string) error
	x1, y1, x2, y2 int // Coordinates for the dialog
}

// Draw implements termui.Drawable interface
func (d *FileDialog) Draw(buf *ui.Buffer) {
	// Draw all dialog components
	d.Title.Draw(buf)
	d.PathInput.Draw(buf)
	d.CurrentPath.Draw(buf)
	d.FileList.Draw(buf)
	d.InfoText.Draw(buf)
}

// NewFileDialog creates a new file dialog
func NewFileDialog(isExport bool, callback func(string) error) *FileDialog {
	title := widgets.NewParagraph()
	if isExport {
		title.Title = "Export Presets"
	} else {
		title.Title = "Import Presets"
	}

	pathInput := widgets.NewParagraph()
	pathInput.Title = "File Path"
	pathInput.Text = ""
	pathInput.BorderStyle = ui.NewStyle(ui.ColorYellow)

	currentPath := widgets.NewParagraph()
	currentPath.Title = "Current Directory"
	homedir, _ := os.UserHomeDir()
	currentPath.Text = homedir

	fileList := widgets.NewList()
	fileList.Title = "Files"
	fileList.SelectedRowStyle = ui.NewStyle(ui.ColorBlack, ui.ColorYellow)

	infoText := widgets.NewParagraph()
	if isExport {
		infoText.Text = "Enter export file path or press Enter to browse. Press Esc to cancel."
	} else {
		infoText.Text = "Enter import file path or press Enter to browse. Press Esc to cancel."
	}
	infoText.Border = false

	dialog := &FileDialog{
		Title:       title,
		PathInput:   pathInput,
		CurrentPath: currentPath,
		FileList:    fileList,
		InfoText:    infoText,
		IsOpen:      true,
		IsExport:    isExport,
		Callback:    callback,
	}

	// Initial file listing
	err := dialog.updateFileList(homedir)
	if err != nil {
		return nil
	}

	return dialog
}

// SetRect sets the dialog dimensions
func (d *FileDialog) SetRect(x1, y1, x2, y2 int) {
	// Store the coordinates
	d.x1, d.y1, d.x2, d.y2 = x1, y1, x2, y2

	height := y2 - y1

	// Rest of the method remains the same
	titleHeight := 3
	inputHeight := 3
	pathHeight := 3
	infoHeight := 3

	// Title at the top
	d.Title.SetRect(x1, y1, x2, y1+titleHeight)

	// Info text at the bottom
	d.InfoText.SetRect(x1, y2-infoHeight, x2, y2)

	// Path input near the top
	d.PathInput.SetRect(x1+1, y1+titleHeight, x2-1, y1+titleHeight+inputHeight)

	// Current path display
	d.CurrentPath.SetRect(x1+1, y1+titleHeight+inputHeight, x2-1, y1+titleHeight+inputHeight+pathHeight)

	// File list takes the remaining space
	listHeight := height - (titleHeight + inputHeight + pathHeight + infoHeight)
	d.FileList.SetRect(x1+1, y1+titleHeight+inputHeight+pathHeight, x2-1, y1+titleHeight+inputHeight+pathHeight+listHeight)
}

// DrawableComponents returns the UI elements to render
func (d *FileDialog) DrawableComponents() []ui.Drawable {
	return []ui.Drawable{
		d.Title,
		d.PathInput,
		d.CurrentPath,
		d.FileList,
		d.InfoText,
	}
}

// updateFileList updates the file listing for the given directory
func (d *FileDialog) updateFileList(dirPath string) error {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	// Update current path
	d.CurrentPath.Text = dirPath

	// Add parent directory option
	fileItems := []string{".. (Parent Directory)"}

	// Add directories first
	for _, file := range files {
		if file.IsDir() {
			fileItems = append(fileItems, "📁 "+file.Name())
		}
	}

	// Add files
	for _, file := range files {
		if !file.IsDir() {
			// For import, only show json files
			if !d.IsExport && filepath.Ext(file.Name()) != ".json" {
				continue
			}
			fileItems = append(fileItems, "📄 "+file.Name())
		}
	}

	d.FileList.Rows = fileItems
	return nil
}

// HandleEvent processes dialog events
func (d *FileDialog) HandleEvent(e ui.Event) (bool, error) {
	switch e.ID {
	case "<Escape>":
		return false, nil // Cancel dialog

	case "<Enter>":
		// If in file list mode and an item is selected
		if d.FileList.SelectedRow >= 0 && d.FileList.SelectedRow < len(d.FileList.Rows) {
			selectedItem := d.FileList.Rows[d.FileList.SelectedRow]

			// Handle parent directory
			if selectedItem == ".. (Parent Directory)" {
				parent := filepath.Dir(d.CurrentPath.Text)
				return true, d.updateFileList(parent)
			}

			// Extract actual name without prefix
			var name string
			if selectedItem[:2] == "📁 " {
				name = selectedItem[2:]
				// Navigate to directory
				newPath := filepath.Join(d.CurrentPath.Text, name)
				return true, d.updateFileList(newPath)
			} else if selectedItem[:2] == "📄 " {
				name = selectedItem[2:]
				// Select file
				filePath := filepath.Join(d.CurrentPath.Text, name)
				d.PathInput.Text = filePath

				// If exporting, we need to ensure it has .json extension
				if d.IsExport && filepath.Ext(filePath) != ".json" {
					filePath += ".json"
					d.PathInput.Text = filePath
				}

				// Submit the path if it's valid
				if d.Callback != nil {
					if err := d.Callback(d.PathInput.Text); err != nil {
						d.InfoText.Text = "Error: " + err.Error()
						return true, nil
					}
					return false, nil // Close dialog on success
				}
			}
		} else if d.PathInput.Text != "" {
			// Direct path entry submission
			if d.Callback != nil {
				path := d.PathInput.Text

				// If exporting, ensure it has .json extension
				if d.IsExport && filepath.Ext(path) != ".json" {
					path += ".json"
				}

				if err := d.Callback(path); err != nil {
					d.InfoText.Text = "Error: " + err.Error()
					return true, nil
				}
				return false, nil // Close dialog on success
			}
		}

	case "j", "<Down>":
		d.FileList.ScrollDown()

	case "k", "<Up>":
		d.FileList.ScrollUp()

	case "<Backspace>":
		if len(d.PathInput.Text) > 0 {
			d.PathInput.Text = d.PathInput.Text[:len(d.PathInput.Text)-1]
		}

	default:
		// Add character to the path input if it's a regular key
		if len(e.ID) == 1 {
			d.PathInput.Text += e.ID
		} else if e.ID == "<Space>" {
			d.PathInput.Text += " "
		} else if e.ID == "/" || e.ID == "\\" {
			d.PathInput.Text += string(filepath.Separator)
		}
	}

	return true, nil // Continue dialog
}

func (d *FileDialog) GetRect() image.Rectangle {
	return image.Rect(d.x1, d.y1, d.x2, d.y2)
}

func (d *FileDialog) Lock() {}

func (d *FileDialog) Unlock() {}
