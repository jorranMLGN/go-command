package ui

import (
	"fmt"
	"image"
	"strings"
	"sync"
	"time"

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
	list.TextStyle = ui.NewStyle(ui.ColorWhite)
	list.WrapText = true

	return &History{
		list:    list,
		entries: make([]CommandHistoryEntry, 0),
	}
}

// AddEntry adds a new entry to the command history
func (h *History) AddEntry(entry CommandHistoryEntry) {
	// Prepend the entry to the history list so newest items appear first
	h.entries = append([]CommandHistoryEntry{entry}, h.entries...)
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
    
    // Calculate available width based on the view's rectangle
    availableWidth := max(h.rect.Max.X - h.rect.Min.X - 4, 20)
    
    for _, entry := range h.entries {
        // Create a more readable format with clear visual separation
        timeStr := formatTimestamp(entry.Timestamp, availableWidth)
        
        // Add each entry with horizontal dividers and consistent formatting
        items = append(items, fmt.Sprintf("--- %s ---\nCMD: %s\nDIR: %s\n",
            timeStr,
            truncateString(entry.Command, availableWidth),
            truncateDir(entry.WorkingDir, availableWidth),
        ))
    }
    h.list.Rows = items
}

// formatTimestamp formats the timestamp into a more readable format
// with different formats based on available width
func formatTimestamp(timestamp string, availableWidth int) string {
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return timestamp
	}
	
	// Use different time formats based on available width
 if availableWidth >= 40 {
		// Medium format: 15:04:05
		return t.Format("15:04:05")
	} else {
		// Short format: 15:04
		return t.Format("15:04")
	}
}

// truncateString truncates a string if it's longer than maxLen
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// truncateDir truncates a directory path from the beginning,
// keeping the last 3 directory levels or as much as fits in maxLen
func truncateDir(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}

	// Split the path into components
	parts := strings.Split(path, "/")
	
	// Start with just the last component
	result := parts[len(parts)-1]
	
	// Try to add more components from right to left
	for i := len(parts) - 2; i >= 0; i-- {
		// Check if we can add one more component
		if len(parts[i])+len(result)+1 <= maxLen-3 { // +1 for the slash, -3 for "..."
			result = parts[i] + "/" + result
		} else {
			// Can't add more, prepend ellipsis and stop
			return ".../" + result
		}
		
		// Keep only up to 3 directory levels
		if len(parts)-i >= 3 {
			if i > 0 { // If there are still components left
				return ".../" + result
			}
			break
		}
	}
	
	return result
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
