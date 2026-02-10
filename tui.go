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
		case EffectPartition:
			effectStr += fmt.Sprintf(" (Partition:%s)", e.Target)
		}
	}

	return fmt.Sprintf("%-15s %s%s", status.Host.Name, statusIcon, effectStr)

}

// Show Action Menu, display menu actions for host selected
// When user press enter in 1 host, this fucking menu will pop up
func (t *TUI) showActionMenu(status HostStatus) {
	// Create new list for menu
	actionList := tview.NewList()
	actionList.SetTitle(" Actions for " + status.Host.Name + " ").SetBorder(true)

	// if host without sudo, display warning and return
	// Without sudo we can not do anything!
	if !status.Sudo {
		// Modal equal to popup with msg
		modal := tview.NewModal().SetText("This host has NO SUDO access!\nCan not do anything!").AddButtons([]string{"OK"}).SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			// When user click ok, return to host list
			t.app.SetRoot(t.hostList, true)
		})

		t.app.SetRoot(modal, true)
		return
	}

	// Add action options
	actionList.AddItem("Blackhole", "Drop traffic to IP/CIDR", 'b', func() {
		t.showInputForm(status, "blackhole")
	})

	actionList.AddItem("Latency", "Add network delay", 'l', func() {
		t.showInputForm(status, "latency")
	})

	actionList.AddItem("Packet Loss", "Drop random packets", 'p', func() {
		t.showInputForm(status, "packetloss")
	})
	actionList.AddItem("IPtables Partition", "Block source IP", 'i', func() {
		t.showInputForm(status, "partition")
	})

	// Restore single effect
	actionList.AddItem("Restore Single", "Remove one rule", 's', func() {
		t.showRestoreMenu(status)
	})

	// Restore all rules
	actionList.AddItem("Restore All", "Remove all rules", 'r', func() {
		// call restoreHost directly
		err := restoreHost(status.Client, status.Host.Name)
		if err != nil {
			t.showMessage("Error: " + err.Error())
		} else {
			t.showMessage("Restored " + status.Host.Name)
		}
	})

	actionList.AddItem("Back", "Return to host list", 0, func() {
		t.refreshHostList() // refresh data first
		t.app.SetRoot(t.hostList, true)
	})

	// Keyboard handler for menu
	actionList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			// REFRESH THE FUCKING DATA!
			t.refreshHostList()
			// Return to host list
			t.app.SetRoot(t.hostList, true)
		}
		return event
	})

	// Display action menu instead of host list
	t.app.SetRoot(actionList, true)
}

// Show message that display message in pop up
func (t *TUI) showMessage(msg string) {
	modal := tview.NewModal().SetText(msg).AddButtons([]string{"OK"}).SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		// STILL NEED TO REFRESH FIRST, EVERY FUCKING TIME!!!
		// When click OK, refresh, damn it!
		t.refreshHostList()
		t.app.SetRoot(t.hostList, true)
	})

	t.app.SetRoot(modal, true)
}

// Show input form, place holder
// Display form that take params by action type
func (t *TUI) showInputForm(status HostStatus, actionType string) {
	// t.showMessage("Input form for " + actionType + " - Comming in story 4.4!")

	// tview.Newform => form has input field, like <form> in HTML
	form := tview.NewForm()
	form.SetBorder(true)

	switch actionType {
	case "blackhole":
		form.SetTitle(" Blackhole - " + status.Host.Name + " ")
		// AddInputField(label, value, width, validateFunc, doneFunc)
		form.AddInputField("Target IP/CIDR", "", 30, nil, nil)
		form.AddButton("Apply", func() {
			// GetFormItem(0) = Get the first field
			// .(*tview.InputField) = type assertion?
			target := form.GetFormItem(0).(*tview.InputField).GetText()
			err := addBlackHole(status.Client, status.Host.Name, target)
			if err != nil {
				t.showMessage("Error: " + err.Error())
			} else {
				t.showMessage("Blackhole added: " + target)
			}
		})
	case "latency":
		form.SetTitle(" Latency - " + status.Host.Name + " ")
		form.AddInputField("Interface:", "eth0", 20, nil, nil)
		form.AddInputField("Delay (e.g. 100ms): ", "", 20, nil, nil)
		form.AddButton("Apply", func() {
			iface := form.GetFormItem(0).(*tview.InputField).GetText()
			delay := form.GetFormItem(1).(*tview.InputField).GetText()
			err := addLatency(status.Client, status.Host.Name, iface, delay)
			if err != nil {
				t.showMessage("Error: " + err.Error())
			} else {
				t.showMessage("Latency added: " + delay)
			}
		})
	case "packetloss":
		form.SetTitle(" Packet Loss - " + status.Host.Name + " ")
		form.AddInputField("Interface:", "eth0", 20, nil, nil)
		form.AddInputField("Loss % (e.g. 10%): ", "", 20, nil, nil)
		form.AddButton("Apply", func() {
			iface := form.GetFormItem(0).(*tview.InputField).GetText()
			percent := form.GetFormItem(1).(*tview.InputField).GetText()
			err := addPacketLoss(status.Client, status.Host.Name, iface, percent)
			if err != nil {
				t.showMessage("Error: " + err.Error())
			} else {
				t.showMessage("Packet Loss added: " + percent)
			}
		})

	case "partition":
		form.SetTitle(" Partition - " + status.Host.Name + " ")
		form.AddInputField("Source IP to block:", "", 30, nil, nil)
		form.AddButton("Apply", func() {
			sourceIP := form.GetFormItem(0).(*tview.InputField).GetText()
			err := addPartition(status.Client, status.Host.Name, sourceIP)
			if err != nil {
				t.showMessage("Error: " + err.Error())
			} else {
				t.showMessage("Partition added: " + sourceIP)
			}
		})
	}

	// Add Cancel button - return back to action menu
	form.AddButton("Cancel", func() {
		t.showActionMenu(status)
	})

	// ESC press - return back to action menu
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			t.refreshHostList() // REFRESH, DAMN IT!
			t.showActionMenu(status)
		}
		return event
	})

	t.app.SetRoot(form, true)

}

// refreshHostList, update all host list
// Need to call this func everytime when we back to host list to have latest data
// Fix issue rules added but host show nothing
func (t *TUI) refreshHostList() {
	/*
		t.hostList.Clear() // remove old list

		// DEBUG: show effect count in title
		allEffects := effectTracker.GetAll()
		t.hostList.SetTitle(fmt.Sprintf(" Hosts (tracked: %d hosts) ", len(allEffects)))

		// build new list with latest data from tracker!
		for i, status := range t.statuses {
			label := t.formatHostLabel(status)

			t.hostList.AddItem(label, status.Host.IP, rune('1'+i), func() {
				t.showActionMenu(status)
			})
		}
	*/

	// Display again with host list
	// t.app.SetRoot(t.hostList, true)
	// Caller will handle this, prevent fucking conflict!
	/*
		So this is complete fucked, we need to create fucking complete new list
	*/
	t.hostList = tview.NewList()
	t.hostList.SetTitle(" Hosts ").SetBorder(true)

	allEffects := effectTracker.GetAll()
	t.hostList.SetTitle(fmt.Sprintf(" Hosts (tracked: %d) ", len(allEffects)))

	// Add hosts to list
	for i, status := range t.statuses {
		label := t.formatHostLabel(status)
		var shortKey rune
		if i < 9 {
			shortKey = rune('1' + i)
		} else {
			shortKey = 0 // No shortkey for more than 9 hosts
		}
		t.hostList.AddItem(label, status.Host.IP, shortKey, func() {
			t.showActionMenu(status)
		})
	}

	// Re-add keyboard handler because new list!
	t.hostList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape || event.Rune() == 'q' {
			t.app.Stop()
		}
		return event
	})
}

// showRestoreMenu, func support for single restore action for specific host.
func (t *TUI) showRestoreMenu(status HostStatus) {
	effects := effectTracker.Get(status.Host.Name)

	// If no effects found
	if len(effects) == 0 {
		t.showMessage("No active effects on " + status.Host.Name)
		return
	}

	restoreList := tview.NewList()
	restoreList.SetTitle(" Restore - " + status.Host.Name + "").SetBorder(true)

	// Add each effect into list
	for i, e := range effects {
		label := fmt.Sprintf("%s: %s %s", e.Type, e.Target, e.Value)

		var shortcut rune
		if i < 9 {
			shortcut = rune('1' + i)
		} else {
			shortcut = 0
		}

		restoreList.AddItem(label, "", shortcut, func() {
			err := removeSingleEffect(status.Client, status.Host.Name, e)
			if err != nil {
				t.showMessage("Error: " + err.Error())
			} else {
				t.showMessage("Removed: " + e.Type + " " + e.Target)
			}
		})
	}

	// Back Button
	restoreList.AddItem("Back", "", 0, func() {
		t.showActionMenu(status)
	})

	// Capture ESC to go back
	restoreList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			t.showActionMenu(status)
		}
		return event
	})

	// Set root.
	t.app.SetRoot(restoreList, true)
}
