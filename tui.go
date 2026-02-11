package main

import (
	"fmt"
	"strings"

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

	// Layout = hostList + footer
	layout *tview.Flex

	// Data from ssh test connection
	statuses []HostStatus

	// Flag for lock refresh
	refreshing bool

	// For filter Text
	filterText string
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
			// Check active effects before fucking quit!
			allEffects := effectTracker.GetAll()
			if len(allEffects) > 0 {
				// Count total
				total := 0
				for _, effects := range allEffects {
					total += len(effects)
				}

				msg := fmt.Sprintf("%d rules still active on %d hosts.\nQuit anyway?", total, len(allEffects))
				t.showConfirmDialog(msg, func() {
					t.app.Stop() // If user Yes -> quit
				})
			} else {
				t.app.Stop() // No rule, just quit!
			}
		}

		// Story 5.3 - View protected IPs
		if event.Rune() == 'p' {
			t.showProtectedIPs()
		}

		// Story 6.1 - Refresh host
		if event.Rune() == 'r' {
			t.refreshHostStatus()
		}

		// Story 6.2 - Help overlay
		if event.Rune() == '?' {
			t.showHelp()
		}

		// Stort 6.3 - Filter host
		if event.Rune() == '/' {
			t.showFilterDialog()
		}

		return event
	})

	// Build layout = hostList + footer
	t.buildLayout()

	// SetRoot = which widget will display
	// EnableMouse = allow mouse interaction
	// Run() = when event loop, block until stop()
	return t.app.SetRoot(t.layout, true).EnableMouse(true).Run()
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
		case EffectPortBlock:
			effectStr += fmt.Sprintf(" (PortBlock:%s:%s)", e.Target, e.Value)
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

	// Popup msg if host is not connected!
	if !status.Connected {
		// Modal equal to popup with msg
		modal := tview.NewModal().SetText("This host is not connected!!\nCan't do anything!").AddButtons([]string{"OK"}).SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			// When user click ok, return to host list
			t.app.SetRoot(t.layout, true)
		})

		t.app.SetRoot(modal, true)
		return
	} else if !status.Sudo {
		// if host without sudo, display warning and return
		// Without sudo we can not do anything!
		// Modal equal to popup with msg
		modal := tview.NewModal().SetText("This host has NO SUDO access!\nCan't do anything!").AddButtons([]string{"OK"}).SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			// When user click ok, return to host list
			t.app.SetRoot(t.layout, true)
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

	actionList.AddItem("Port Block", "Block specific port from IP", 'd', func() {
		t.showInputForm(status, "portblock")
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
		t.app.SetRoot(t.layout, true)
	})

	// Keyboard handler for menu
	actionList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			// REFRESH THE FUCKING DATA!
			t.refreshHostList()
			// Return to host list
			t.app.SetRoot(t.layout, true)
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
		t.app.SetRoot(t.layout, true)
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

			// Story 5.1, prevent self-lock
			if status.SSH_SourceIP == target {
				t.showConfirmDialog("WARNING: You are trying to block your own IP!\nAre you sure?", func() {
					// if User don't care. Allow it!
					err := addBlackHole(status.Client, status.Host.Name, target)
					if err != nil {
						t.showMessage("Error: " + err.Error())
					} else {
						t.showMessage("Blackhole added: " + target)
					}
				})
				return
			}

			err := addBlackHole(status.Client, status.Host.Name, target)
			if err != nil {
				t.showMessage("Error: " + err.Error())
			} else {
				t.showMessage("Blackhole added: " + target)
			}
		})
	case "latency":
		// Call listInterfaces first before display form
		interfaces, err := listInterfaces(status.Client)
		if err != nil {
			t.showMessage("Error listing interfaces: " + err.Error())
			return
		}

		if len(interfaces) == 1 {
			// 1 interface → show in title, no field needed!
			form.SetTitle(fmt.Sprintf(" Latency - %s (%s) ", status.Host.Name, interfaces[0]))
			form.AddInputField("Delay (e.g. 100ms): ", "", 20, nil, nil)
			form.AddButton("Apply", func() {
				iface := interfaces[0]
				delay := form.GetFormItem(0).(*tview.InputField).GetText()
				err := addLatency(status.Client, status.Host.Name, iface, delay)
				if err != nil {
					t.showMessage("Error: " + err.Error())
				} else {
					t.showMessage("Latency added: " + delay + " on " + iface)
				}
			})
		} else {
			// Multiple interfaces → dropdown
			form.SetTitle(" Latency - " + status.Host.Name + " ")
			form.AddDropDown("Interface:", interfaces, 0, nil)
			form.AddInputField("Delay (e.g. 100ms): ", "", 20, nil, nil)
			form.AddButton("Apply", func() {
				_, iface := form.GetFormItem(0).(*tview.DropDown).GetCurrentOption()
				delay := form.GetFormItem(1).(*tview.InputField).GetText()
				err := addLatency(status.Client, status.Host.Name, iface, delay)
				if err != nil {
					t.showMessage("Error: " + err.Error())
				} else {
					t.showMessage("Latency added: " + delay + " on " + iface)
				}
			})
		}
	case "packetloss":
		// Call listInterfaces first
		interfaces, err := listInterfaces(status.Client)
		if err != nil {
			t.showMessage("Error listing interfaces: " + err.Error())
			return
		}

		if len(interfaces) == 1 {
			// 1 interface → show in title, no field needed!
			form.SetTitle(fmt.Sprintf(" Packet Loss - %s (%s) ", status.Host.Name, interfaces[0]))
			form.AddInputField("Loss % (e.g. 10): ", "", 20, nil, nil)
			form.AddButton("Apply", func() {
				iface := interfaces[0]
				lossPercent := form.GetFormItem(0).(*tview.InputField).GetText()
				err := addPacketLoss(status.Client, status.Host.Name, iface, lossPercent)
				if err != nil {
					t.showMessage("Error: " + err.Error())
				} else {
					t.showMessage("Packet loss added: " + lossPercent + "% on " + iface)
				}
			})
		} else {
			// Multiple interfaces → dropdown
			form.SetTitle(" Packet Loss - " + status.Host.Name + " ")
			form.AddDropDown("Interface:", interfaces, 0, nil)
			form.AddInputField("Loss % (e.g. 10): ", "", 20, nil, nil)
			form.AddButton("Apply", func() {
				_, iface := form.GetFormItem(0).(*tview.DropDown).GetCurrentOption()
				lossPercent := form.GetFormItem(1).(*tview.InputField).GetText()
				err := addPacketLoss(status.Client, status.Host.Name, iface, lossPercent)
				if err != nil {
					t.showMessage("Error: " + err.Error())
				} else {
					t.showMessage("Packet loss added: " + lossPercent + "% on " + iface)
				}
			})
		}

	case "partition":
		form.SetTitle(" Partition - " + status.Host.Name + " ")
		form.AddInputField("Source IP to block:", "", 30, nil, nil)
		form.AddButton("Apply", func() {
			sourceIP := form.GetFormItem(0).(*tview.InputField).GetText()

			// Prevent self-lock
			if status.SSH_SourceIP == sourceIP {
				t.showConfirmDialog("WARNING: You are trying to block your own IP!", func() {
					// if User don't care. Allow it!
					err := addPartition(status.Client, status.Host.Name, sourceIP)
					if err != nil {
						t.showMessage("Error: " + err.Error())
					} else {
						t.showMessage("Partition added: " + sourceIP)
					}
				})
				return
			}

			err := addPartition(status.Client, status.Host.Name, sourceIP)
			if err != nil {
				t.showMessage("Error: " + err.Error())
			} else {
				t.showMessage("Partition added: " + sourceIP)
			}
		})

	case "portblock":
		form.SetTitle(" Port Block - " + status.Host.Name + " ")
		form.AddInputField("Source IP:", "", 30, nil, nil)
		form.AddInputField("Port (e.g. 3306):", "", 10, nil, nil)
		form.AddButton("Apply", func() {
			sourceIP := form.GetFormItem(0).(*tview.InputField).GetText()
			port := form.GetFormItem(1).(*tview.InputField).GetText()

			// Prevent self-lock
			if status.SSH_SourceIP == sourceIP {
				t.showConfirmDialog("WARNING: You are trying to block your own IP!", func() {
					err := addPortBlock(status.Client, status.Host.Name, sourceIP, port)
					if err != nil {
						t.showMessage("Error: " + err.Error())
					} else {
						t.showMessage("Port blocked: " + sourceIP + ":" + port)
					}
				})
				return
			}

			err := addPortBlock(status.Client, status.Host.Name, sourceIP, port)
			if err != nil {
				t.showMessage("Error: " + err.Error())
			} else {
				t.showMessage("Port blocked: " + sourceIP + ":" + port)
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

	// Add hosts to list (with filter)
	displayIdx := 0
	for _, status := range t.statuses {
		// Story 6.3 - Filter by name
		if t.filterText != "" && !strings.Contains(
			strings.ToLower(status.Host.Name),
			strings.ToLower(t.filterText),
		) {
			continue // Skip host not matching filter
		}

		label := t.formatHostLabel(status)
		var shortKey rune
		if displayIdx < 9 {
			shortKey = rune('1' + displayIdx)
		} else {
			shortKey = 0 // No shortkey for more than 9 hosts
		}
		t.hostList.AddItem(label, status.Host.IP, shortKey, func() {
			t.showActionMenu(status)
		})
		displayIdx++
	}

	// Re-add keyboard handler because new list!
	t.hostList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape || event.Rune() == 'q' {
			// CÙNG LOGIC như Run()!
			allEffects := effectTracker.GetAll()
			if len(allEffects) > 0 {
				total := 0
				for _, effects := range allEffects {
					total += len(effects)
				}
				msg := fmt.Sprintf("%d rules still active on %d hosts.\nQuit anyway?", total, len(allEffects))
				t.showConfirmDialog(msg, func() {
					t.app.Stop()
				})
			} else {
				t.app.Stop()
			}
		}

		// Story 5.3 - View protected IPs
		if event.Rune() == 'p' {
			t.showProtectedIPs()
		}

		if event.Rune() == 'r' {
			t.refreshHostStatus()
		}

		// Story 6.2 - Help overlay
		if event.Rune() == '?' {
			t.showHelp()
		}

		// Stort 6.3 - Filter host
		if event.Rune() == '/' {
			t.showFilterDialog()
		}

		return event
	})

	// Rebuild layout with new list
	t.buildLayout()
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

// Refresh host func
func (t *TUI) refreshHostStatus() {

	// Prevent spam refresh!
	if t.refreshing {
		t.showMessage("Already refreshing, hold on!")
		return // hold on!
	}

	t.refreshing = true // Lock

	// Show refreshing... msg
	t.hostList.SetTitle(" Refreshing... ")
	t.app.ForceDraw() // this does force render immediately

	// Re0test ssh to all hosts
	// Get hosts from current statuses
	var hosts []Host
	for _, s := range t.statuses {
		// Close old connections
		if s.Client != nil {
			s.Client.Close()
		}

		hosts = append(hosts, s.Host)
	}

	// Re-connect
	t.statuses = testAllHosts(hosts)

	// Refresh display
	t.refreshHostList()

	// Count connected hosts
	connected := 0
	for _, s := range t.statuses {
		if s.Connected {
			connected++
		}
	}

	t.refreshing = false // Unlock before showMessage
	msg := fmt.Sprintf("Refreshed! %d/%d hosts connected", connected, len(t.statuses))
	t.showMessage(msg)
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

	// MIDDLE = commands in 2 columns
	headerMid := tview.NewTextView().
		SetDynamicColors(true).
		SetText(
			"[aqua](r)[-] Refresh    [aqua](ESC)[-] Back\n" +
				"[aqua](p)[-] Protected  [aqua](Enter)[-] Select\n" +
				"[aqua](?)[-] Help       [aqua](q)[-] Quit\n" +
				"[aqua](/)[-] Filter",
		)

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
	t.layout = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 5, 0, false).   // Header 5 lines (logo height)
		AddItem(t.hostList, 0, 1, true) // List with focus
}

// showHelp displays all keyboard shortcuts
// with better format!
func (t *TUI) showHelp() {
	helpText := "Keyboard Shortcuts:\n\n" +
		"Arrow Keys - Navigate hosts\n" +
		"Enter - Select host\n" +
		"Esc - Go back / Quit\n" +
		"r - Refresh host status\n" +
		"p - View protected IPs\n" +
		"/ - Filter hosts\n" +
		"? - This help\n" +
		"q - Quit\n\n" +
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
