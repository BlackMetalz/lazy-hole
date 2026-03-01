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

	// Layout = hostList/groupList + footer
	layout *tview.Flex

	// Data from ssh test connection
	statuses []HostStatus

	// Flag for lock refresh
	refreshing bool

	// For filter Text
	filterText string

	// widget for groups view, same list hosts list
	groupList *tview.List

	// view mode, allow switch between groups/hosts view
	viewMode string
}

// NewTUI creates a new TUI instance with data from SSH test connection
func NewTUI(statuses []HostStatus) *TUI {
	// return TUI struct
	return &TUI{
		app:      tview.NewApplication(), // new app container
		statuses: statuses,               // save data for later display
		viewMode: "hosts",                // Default view is host view!
	}
}

// GetStatuses returns current statuses (may be updated after refresh)
func (t *TUI) GetStatuses() []HostStatus {
	return t.statuses
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
		// 4: update selectedFunc
		var shortKey rune
		if i < 9 {
			shortKey = rune('1' + i)
		} else {
			shortKey = 0 // No shortkey for more than 9 hosts
		}
		t.hostList.AddItem(label, status.Host.IP, shortKey, func() {
			t.showActionMenu(status) // Use s, not status
		})
	}

	// Setup keyboard shortcuts for host list (shared with refreshHostList)
	t.setupHostListKeys()

	// Build layout = hostList + footer
	t.buildLayout()

	// SetRoot = which widget will display
	// EnableMouse = allow mouse interaction
	// Run() = when event loop, block until stop()
	return t.app.SetRoot(t.layout, true).Run()
}

// formatHostLabel create string display for 1 host
// Example output: "mysql-node-1" Healthy [2 rules]
// Helper func
func (t *TUI) formatHostLabel(status HostStatus) string {
	var statusIcon string

	// color tags: [red], [green], [yellow], [white]
	// [white] reset color into default

	if !status.Connected {
		statusIcon = "[red]FAILED[-]"
	} else if !status.Sudo {
		statusIcon = "[yellow]NO SUDO[-]"
	} else {
		statusIcon = "[green]HEALTHY[-]"
	}

	// // count effect active on this running host
	// effects := effectTracker.Get(status.Host.Name)
	// fmt.Println("Effect count: ", len(effects))

	// effectCount := ""
	// if len(effects) > 0 {
	// 	effectCount = fmt.Sprintf("(%d rules)", len(effects))
	// }

	// // %-15s = format string, padding 15 chars, left align
	// return fmt.Sprintf("%-15s %s%s", status.Host.Name, statusIcon, effectCount)

	// Get effects for this host
	effects := effectTracker.Get(status.Host.Name)

	/* stop count in story 4.5
	effectCount := ""
	if len(effects) > 0 {
		effectCount = fmt.Sprintf("(%d rules)", len(effects))
	}
	*/

	// Story 4.5, display detail each effect:
	effectStr := "" // Init
	for _, e := range effects {
		switch e.Type {
		case EffectBlackHole:
			effectStr += fmt.Sprintf(" (BlackHole:%s)", e.Target)
		case EffectLatency:
			effectStr += fmt.Sprintf(" (Latency:%s %s)", e.Value, e.Target)
		case EffectPacketLoss:
			effectStr += fmt.Sprintf(" (PacketLoss:%s%% %s)", e.Value, e.Target)
		// case EffectPartition:
		// 	effectStr += fmt.Sprintf(" (Partition:%s)", e.Target)
		case EffectPortBlock:
			effectStr += fmt.Sprintf(" (PortBlock:%s:%s)", e.Target, e.Value)
		}
	}

	return fmt.Sprintf("%-15s %s%s", status.Host.Name, statusIcon, effectStr)

}

// Show message that display message in pop up
func (t *TUI) showMessage(msg string) {
	modal := tview.NewModal().SetText(msg).AddButtons([]string{"OK"}).SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		// STILL NEED TO REFRESH FIRST, EVERY FUCKING TIME!!!
		// When click OK, refresh, damn it!
		t.refreshHostList()
		t.app.SetRoot(t.layout, true)
	})

	t.app.SetRoot(modal, true)
}

// showConfirm Dialog func
func (t *TUI) showConfirmDialog(msg string, onConfirm func()) {
	modal := tview.NewModal().SetText(msg).AddButtons([]string{"Yes", "No"}).SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Yes" {
			onConfirm()
		} else {
			t.refreshHostList()
			t.app.SetRoot(t.layout, true)
		}
	})

	t.app.SetRoot(modal, true)
}

// setupHostListKeys sets all keyboard shortcuts for the host list.
// Single source of truth! Called by both Run() and refreshHostList()
func (t *TUI) setupHostListKeys() {
	t.hostList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape || event.Rune() == 'q' {
			allEffects := effectTracker.GetAll()
			total := 0
			for _, effects := range allEffects {
				total += len(effects)
			}
			// Don't show if we have 0 rule active!
			if total > 0 {
				msg := fmt.Sprintf("%d rules still active on %d hosts.\nQuit anyway?", total, len(allEffects))
				t.showConfirmDialog(msg, func() {
					t.app.Stop()
				})
			} else {
				t.app.Stop()
			}
		}

		if event.Rune() == 'p' {
			t.showProtectedIPs()
		}
		if event.Rune() == 'r' {
			t.refreshHostStatus()
		}
		if event.Rune() == '?' {
			t.showHelp()
		}
		if event.Rune() == '/' {
			t.showFilterDialog()
		}
		if event.Rune() == 'h' {
			t.showHistory()
		}
		if event.Rune() == 'u' {
			t.showUndoConfirm()
		}
		if event.Rune() == 'g' {
			t.switchToGroupView()
		}

		return event
	})
}

// buildLayout creates Flex layout = K9s-style header + hostList
func (t *TUI) buildLayout() {
	// LEFT = version info (dynamic from root_cmd.go)
	leftText := "[yellow]lazy-hole[-] " + version
	if t.filterText != "" {
		leftText += "\n[red]filtered keyword: " + t.filterText + "[-]"
	}
	headerLeft := tview.NewTextView().
		SetDynamicColors(true).
		SetText(leftText)

	// MIDDLE = commands in 2 columns using Grid
	// (-1,-1) = 2 equal columns!
	headerMid := tview.NewGrid().SetColumns(-1, -1)

	// Need call separated because it will return `*Box`, not `*Grid` so we can not add 2 columns using
	// headerMid.AddItem below
	// headerMid.SetBorder(true)

	// We will have max 5 command for each columns!

	leftMidCol := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetText(
			"[aqua](r)[-]" + " Refresh\n" +
				"[aqua](p)[-]" + " Protected\n" +
				"[aqua](?)[-]" + " Help\n" +
				"[aqua](/)[-]" + " Filter\n" +
				"[aqua](h)[-]" + " History applied\n",
		)

	rightMidCol := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetText(
			"[aqua](ESC)[-]" + " Back\n" +
				"[aqua](q)[-]" + " Quit\n" +
				"[aqua](u)[-]" + " Undo last rule\n" +
				"[aqua](g)[-]" + " Switch group view\n" +
				"[aqua](l)[-]" + " Switch host view\n",
		)

	headerMid.
		AddItem(leftMidCol, 0, 0, 1, 1, 0, 0, false).
		AddItem(rightMidCol, 0, 1, 1, 1, 0, 0, false)

	// headerMid := tview.NewTextView().
	// 	SetDynamicColors(true).
	// 	SetText(
	// 		"[aqua](r)[-] Refresh    [aqua](ESC)[-] Back\n" +
	// 			"[aqua](p)[-] Protected  [aqua](Enter)[-] Select\n" +
	// 			"[aqua](?)[-] Help       [aqua](q)[-] Quit\n" +
	// 			"[aqua](/)[-] Filter	 [aqua](u)[-] Undo last rule\n" +
	// 			"[aqua](h)[-] History applied\n",
	// 	)

	// RIGHT = ASCII art logo (block chars)
	headerRight := tview.NewTextView().
		SetDynamicColors(true).
		SetText("[yellow]" + logo + "[-]")

	// Header = horizontal flex (left + mid + right)
	header := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(headerLeft, 20, 0, false). // Fix 20 chars
		AddItem(headerMid, 0, 1, false).   // Flexible middle
		AddItem(headerRight, 25, 0, false) // Fix 22 chars for logo

	// Main layout = vertical flex
	// Remember setBorder eaten 2 lines, if we set border for header, we will need add 7 lines to header. Not 5!
	var activeList *tview.List
	if t.viewMode == "groups" {
		activeList = t.groupList
	} else {
		activeList = t.hostList
	}

	t.layout = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 5, 0, false).   // Header 5 lines (logo height). Each header will have max 5 commands!
		AddItem(activeList, 0, 1, true) // List with focus
}
