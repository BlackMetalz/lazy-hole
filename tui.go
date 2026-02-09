package main

import (
	"fmt"

	// tcell is a lib low-level that handle keyboard/mouse events
	// tview ise tcell
	// tview is TUI lib (terminal UI)
	// it is fucking like HTML/CSS for terminal, beautiful!
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// TUI holds the application state
type TUI struct {
	// Every tview app need a application
	app *tview.Application

	// Widget display hosts list
	hostList *tview.List

	// Data from ssh test connection
	statuses []HostStatus
}

// NewTUI creates a new TUI instance with data from SSH test connection
func NewTUI(statuses []HostStatus) *TUI {
	// return TUI struct
	return &TUI{
		app:      tview.NewApplication(), // new app container
		statuses: statuses,               // save data for later display
	}
}

// start the TUI
// Reflection: can be understand as app.Run()
func (t *TUI) Run() error {

	// List widget, with empty list
	t.hostList = tview.NewList()

	// Set fucking title that display in border
	t.hostList.SetTitle(" Hosts ").SetBorder(true)

	// Add hosts to list
	for i, status := range t.statuses {
		// Format label from helper func
		label := t.formatHostLabel(status)

		// Add item to list, func of tview, take 4 params
		// 1: label
		// 2: short description
		// 3: shortkey: 1,2,3....
		// 4: selectedFunc: func call when Enter - nil ==> nothing
		t.hostList.AddItem(label, status.Host.IP, rune('1'+i), nil)
	}

	// Keyboard handler - capture keyboard events
	t.hostList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// If user click escape or 'q' ==> exit app
		if event.Key() == tcell.KeyEscape || event.Rune() == 'q' {
			t.app.Stop()
		}

		return event
	})

	// SetRoot = which widget will display
	// EnableMouse = allow mouse interaction
	// Run() = when event loop, block until stop()
	return t.app.SetRoot(t.hostList, true).EnableMouse(true).Run()
}

// formatHostLabel create string display for 1 host
// Example output: "mysql-node-1" Healthy [2 rules]
// Helper func
func (t *TUI) formatHostLabel(status HostStatus) string {
	var statusIcon string

	// color tags: [red], [green], [yellow], [white]
	// [white] reset color into default

	if !status.Connected {
		// No connection
		statusIcon = "[red] FAILED[white]"
	} else if !status.Sudo {
		// No Sudo
		statusIcon = "[yellow] NO SUDO[white]"
	} else {
		statusIcon = "[green] HEALTHY[white]"
	}

	// count effect active on this running host
	effects := effectTracker.Get(status.Host.Name)
	effectCount := ""
	if len(effects) > 0 {
		effectCount = fmt.Sprintf("[%d rule]", len(effects))
	}

	// %-15s = format string, padding 15 chars, left align
	return fmt.Sprintf("%-15s %s%s", status.Host.Name, statusIcon, effectCount)
}
