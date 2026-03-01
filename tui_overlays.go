package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// showHelp displays all keyboard shortcuts
// with better format!
func (t *TUI) showHelp() {
	helpText := "Keyboard Shortcuts:\n\n" +
		"Arrow Keys - Navigate hosts\n" +
		"Enter - Select host\n" + // 2
		"Esc - Go back / Quit\n" + // 3
		"r - Refresh host status\n" + // 4
		"p - View protected IPs\n" + // 5
		"/ - Filter hosts\n" + // 6
		"? - This help\n" + // 7
		"u - Undo last action\n" + // 8
		"h - History\n" + // 9
		"q - Quit\n\n" + // 1
		"Press OK to close."

	modal := tview.NewModal().
		SetText(helpText).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			t.app.SetRoot(t.layout, true)
		})

	t.app.SetRoot(modal, true)
}

// showFilterDialog displays popup for user to type filter text
func (t *TUI) showFilterDialog() {
	inputField := tview.NewInputField().
		SetLabel("Filter host: ").
		SetText(t.filterText). // Show current filter
		SetFieldWidth(30)

	inputField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			// Apply filter
			t.filterText = inputField.GetText()
			t.refreshHostList()
			t.app.SetRoot(t.layout, true)
		} else if key == tcell.KeyEscape {
			// Cancel, keep old filter
			t.app.SetRoot(t.layout, true)
		}
	})

	// Center the input in a modal-like layout
	// FlexRow (Spacer, InputRow, Spacer)
	inputRow := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(nil, 0, 1, false).
		AddItem(inputField, 40, 0, true).
		AddItem(nil, 0, 1, false)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(inputRow, 1, 0, true).
		AddItem(nil, 0, 1, false)

	t.app.SetRoot(flex, true)
}

// showHistory displays action history log in a scrollable view
func (t *TUI) showHistory() {
	entries, err := actionLogger.ReadHistory()
	if err != nil {
		t.showMessage("Error reading history: " + err.Error())
		return
	}

	if len(entries) == 0 {
		t.showMessage("No history yet!")
		return
	}

	// Build text from entries
	text := ""
	for _, e := range entries {
		text += fmt.Sprintf("[yellow]%s[-] | [green]%s[-] | %s | %s | %s\n",
			e.Timestamp, e.Hostname, e.Action, e.Params, e.Result)
	}

	// Create scrollable text view
	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetText(text).
		SetScrollable(true)

	textView.SetTitle(" Action History (ESC to close) ").SetBorder(true)

	// Scroll to bottom (latest entries)
	textView.ScrollToEnd()

	textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			t.app.SetRoot(t.layout, true)
		}
		return event
	})

	t.app.SetRoot(textView, true)
}

// showUndoConfirm shows what will be undone and asks for confirmation
func (t *TUI) showUndoConfirm() {
	action := undoStack.Peek()
	if action == nil {
		t.showMessage("Nothing to undo!")
		return
	}

	// Build description of what will be undone
	desc := fmt.Sprintf("Undo: %s %s", action.Effect.Type, action.Effect.Target)
	if action.Effect.Value != "" {
		desc += " (" + action.Effect.Value + ")"
	}
	desc += " on " + action.Hostname
	if action.BatchID != "" {
		desc += " (batch/group action)"
	}
	desc += "?"

	t.showConfirmDialog(desc, func() {
		// PopBatch: if batch → pop all with same BatchID, else pop 1
		batch := undoStack.PopBatch()
		if len(batch) == 0 {
			t.showMessage("Nothing to undo!")
			return
		}

		succeeded := 0
		failed := 0
		for _, undone := range batch {
			err := removeSingleEffect(undone.Client, undone.Hostname, undone.Effect)
			if err != nil {
				failed++
			} else {
				actionLogger.Log(undone.Hostname, "UNDO", undone.Effect.Type+" "+undone.Effect.Target, "SUCCESS")
				succeeded++
			}
		}

		// Refresh current view so stale data is gone!
		if t.viewMode == "groups" {
			t.buildGroupList()
			t.buildLayout()
		} else {
			t.refreshHostList()
		}

		msg := fmt.Sprintf("Undone %d action(s)", succeeded)
		if failed > 0 {
			msg += fmt.Sprintf(", %d failed", failed)
		}
		t.showMessage(msg)
	})
}

// show protected ips
func (t *TUI) showProtectedIPs() {
	list := tview.NewList()
	list.SetTitle(" Protected IPs ").SetBorder(true)

	for i, status := range t.statuses {
		if status.SSH_SourceIP != "" {
			label := fmt.Sprintf("%-15s SSH source: %s", status.Host.Name, status.SSH_SourceIP)

			var shortcut rune
			if i < 9 {
				shortcut = rune('1' + i)
			} else {
				shortcut = 0
			}

			list.AddItem(label, "", shortcut, nil)
		}
	}

	// If not protected IP
	// Count list
	if list.GetItemCount() == 0 {
		list.AddItem("No protected IPs detected??", "", 0, nil)
	}

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			t.app.SetRoot(t.layout, true)
		}
		return event
	})

	t.app.SetRoot(list, true)
}
