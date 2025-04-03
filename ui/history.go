package ui

import (
	"image"
	"sync"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

// CommandHistoryEntry represents a single command execution history item
type CommandHistoryEntry struct {
	Timestamp  string
	Command    string
	WorkingDir string
	Status     string
}

// History represents the command history view
type History struct {
	list       *widgets.List
	rect       image.Rectangle
	entries    []CommandHistoryEntry
	sync.Mutex // Add mutex for thread-safety
}

// Lock implements the Drawable interface
func (h *History) Lock() {
	h.Mutex.Lock()
}

// Unlock implements the Drawable interface
func (h *History) Unlock() {
	h.Mutex.Unlock()
}

// NewHistoryView creates a new history view
func NewHistoryView() *History {
	list := widgets.NewList()
	list.Title = "Command History"
	list.TextStyle = ui.NewStyle(ui.ColorYellow)
	list.WrapText = true
	list.SelectedRowStyle = ui.NewStyle(ui.ColorBlack, ui.ColorYellow)

	return &History{
		list:    list,
		entries: make([]CommandHistoryEntry, 0),
	}
}

// AddEntry adds a new entry to the command history
func (h *History) AddEntry(entry CommandHistoryEntry) {
	h.entries = append(h.entries, entry)
	h.updateList()
}

// SetEntries sets the history entries
func (h *History) SetEntries(entries []CommandHistoryEntry) {
	h.entries = entries
	h.updateList()
}

// updateList updates the list widget with current entries
func (h *History) updateList() {
	items := make([]string, 0, len(h.entries))
	for _, entry := range h.entries {
		items = append(items,
			entry.Timestamp+" | "+entry.Command+" | "+entry.WorkingDir+" | "+entry.Status)
	}
	h.list.Rows = items
}

// SetRect implements the Drawable interface
func (h *History) SetRect(x1, y1, x2, y2 int) {
	h.rect = image.Rect(x1, y1, x2, y2)
	h.list.SetRect(x1, y1, x2, y2)
}

// GetRect implements the Drawable interface
func (h *History) GetRect() image.Rectangle {
	return h.rect
}

// Draw implements the Drawable interface
func (h *History) Draw(buf *ui.Buffer) {
	h.list.Draw(buf)
}

// HandleEvent handles UI events
func (h *History) HandleEvent(e ui.Event) {
	switch e.ID {
	case "j", "<Down>":
		h.list.ScrollDown()
	case "k", "<Up>":
		h.list.ScrollUp()
	}
}
