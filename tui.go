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
	actionList.AddItem("Blackhole", "Drop traffic to IP/CIDR (Layer 3)", 'b', func() {
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

	actionList.AddItem("Port Block", "Block specific port from IP (Layer 4)", 'd', func() {
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
		// Feb 12, 2026. I added feature to support multiple ip for blackhole
		// Support comma-separated IPs, e.g. 192.168.3.11,192.168.13.11
		form.AddInputField("Target IP/CIDR (comma-separated)", "", 50, nil, nil)
		form.AddButton("Apply", func() {
			input := form.GetFormItem(0).(*tview.InputField).GetText()

			// Split by comma and trim whitespace
			parts := strings.Split(input, ",")
			var targets []string
			for _, p := range parts {
				trimmed := strings.TrimSpace(p)
				// if trimmed is not empty, append to target lists
				if trimmed != "" {
					targets = append(targets, trimmed)
				}
			}

			// Count
			if len(targets) == 0 {
				t.showMessage("Error: no valid targets provided")
				return
			}

			// Check self-lock for any of the targets
			hasSelfLock := false
			for _, target := range targets {
				if status.SSH_SourceIP == target {
					hasSelfLock = true
					break
				}
			}

			// Func that actually applies all blackholes and shows summary
			// Apply for each target, if getting error, append to errors slice!
			applyAll := func() {
				success := 0
				var errors []string
				for _, target := range targets {
					err := addBlackHole(status.Client, status.Host.Name, target)
					if err != nil {
						errors = append(errors, target+": "+err.Error())
					} else {
						success++
					}
				}

				// Build summary message
				msg := fmt.Sprintf("Blackhole: %d/%d added", success, len(targets))
				// If there are errors, append them to the message
				if len(errors) > 0 {
					msg += "\n\nFailed:\n" + strings.Join(errors, "\n")
				}
				t.showMessage(msg)
			}

			// Check self lock, prevent blackhole yourself LOL
			if hasSelfLock {
				t.showConfirmDialog("WARNING: One of the targets is your SSH source IP!\nAre you sure?", applyAll)
			} else {
				applyAll()
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
		// Feb 12, 2026. I added feature to support multiple ip for network partition
		// Support comma-separated IPs, e.g. 192.168.3.11,192.168.13.11
		form.AddInputField("Source IP to block (comma-separated):", "", 50, nil, nil)
		form.AddButton("Apply", func() {
			input := form.GetFormItem(0).(*tview.InputField).GetText()

			// Split by comma and trim whitespace
			parts := strings.Split(input, ",")
			var targets []string
			for _, p := range parts {
				// if trimmed is not empty, append to target lists
				trimmed := strings.TrimSpace(p)
				if trimmed != "" {
					targets = append(targets, trimmed)
				}
			}

			// Count
			if len(targets) == 0 {
				t.showMessage("Error: no valid targets provided")
				return
			}

			// Check self-lock for any of the targets
			hasSelfLock := false
			for _, target := range targets {
				if status.SSH_SourceIP == target {
					hasSelfLock = true
					break
				}
			}

			// Func that actually applies all blackholes and shows summary
			// Apply for each target, if getting error, append to errors slice!
			applyAll := func() {
				success := 0
				var errors []string
				for _, target := range targets {
					err := addPartition(status.Client, status.Host.Name, target)
					if err != nil {
						errors = append(errors, target+": "+err.Error())
					} else {
						success++
					}
				}

				// Build summary message
				msg := fmt.Sprintf("Partition: %d/%d added", success, len(targets))
				if len(errors) > 0 {
					msg += "\n\nFailed:\n" + strings.Join(errors, "\n")
				}
				t.showMessage(msg)
			}

			// Check self lock, prevent partition yourself LOL
			if hasSelfLock {
				// Show confirm dialog take 2 param, if yes applyAll!
				t.showConfirmDialog("WARNING: One of the targets is your SSH source IP!\nAre you sure?", applyAll)
			} else {
				applyAll()
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
	// Re-apply keyboard shortcuts for new list
	t.setupHostListKeys()

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

// setupHostListKeys sets all keyboard shortcuts for the host list.
// Single source of truth! Called by both Run() and refreshHostList()
func (t *TUI) setupHostListKeys() {
	t.hostList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape || event.Rune() == 'q' {
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

// buildGroupList, create list from group
func (t *TUI) buildGroupList() {
	t.groupList = tview.NewList()
	t.groupList.SetTitle(" Groups ").SetBorder(true)

	// Collect groups from statuses
	// map[groupName] => []hostName
	groups := make(map[string][]string)
	for _, s := range t.statuses {
		// Check group field is not empty
		if s.Host.Group != "" {
			// append host name into group
			groups[s.Host.Group] = append(groups[s.Host.Group], s.Host.Name)
		}
	}

	// Add each group to list view
	idx := 0 // Init index
	for groupName, members := range groups {
		// Build a label to display in group views
		labels := fmt.Sprintf("%s (%d hosts)", groupName, len(members))
		var shortcut rune
		// If there is more than 9 group, > 9th++, they will not able to receive shortcut. Haha
		if idx < 9 {
			shortcut = rune('1' + idx)
		} else {
			shortcut = 0
		}

		t.groupList.AddItem(labels, strings.Join(members, ", "), shortcut, func() {
			t.showGroupActionMenu(groupName, members)
		})

		idx++
	}

	// ESC = back to hosts. Specific handler
	// Ughhh, I hate this, need to refactor this?? Tired of typing/copying SetInputCapture...
	t.groupList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape || event.Rune() == 'l' {
			t.viewMode = "hosts"
			t.refreshHostList()
			t.app.SetRoot(t.layout, true)
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
	desc += " on " + action.Hostname + "?"

	t.showConfirmDialog(desc, func() {
		// Pop from stack and execute undo
		undone := undoStack.Pop()
		if undone == nil {
			t.showMessage("Nothing to undo!")
			return
		}

		err := removeSingleEffect(undone.Client, undone.Hostname, undone.Effect)
		if err != nil {
			t.showMessage("Undo failed: " + err.Error())
		} else {
			actionLogger.Log(undone.Hostname, "UNDO", undone.Effect.Type+" "+undone.Effect.Target, "SUCCESS")
			t.showMessage("Undone: " + undone.Effect.Type + " " + undone.Effect.Target + " on " + undone.Hostname)
		}
	})
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

// Switch to group view
func (t *TUI) switchToGroupView() {
	t.viewMode = "groups"
	t.buildGroupList()
	// Rebuild layout with group list
	t.buildLayout()
	t.app.SetRoot(t.layout, true)
}

// func showGroupActionMenu, shơ action menu fỏ group
func (t *TUI) showGroupActionMenu(groupName string, members []string) {
	actionList := tview.NewList()
	actionList.SetTitle(" Actions for group: " + groupName + "").SetBorder(true)

	actionList.AddItem("Blackhole", "Drop traffic to IP/CIDR", 'b', func() {
		t.showGroupInputForm(groupName, members, "blackhole")
	})
	actionList.AddItem("Latency", "Add network delay", 'l', func() {
		t.showGroupInputForm(groupName, members, "latency")
	})
	// actionList.AddItem("Partition", "Block source IP", 'i', func() {
	// 	t.showGroupInputForm(groupName, members, "partition")
	// })
	actionList.AddItem("Port Block", "Block specific port from IP", 'd', func() {
		t.showGroupInputForm(groupName, members, "portblock")
	})
	actionList.AddItem("Back", "Return to group list", 0, func() {
		t.switchToGroupView()
	})
	actionList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			t.switchToGroupView()
		}
		return event
	})
	t.app.SetRoot(actionList, true)

}

// showGroupInputForm shows input form for group action
func (t *TUI) showGroupInputForm(groupName string, members []string, actionType string) {
	form := tview.NewForm()
	form.SetBorder(true)
	form.SetTitle(fmt.Sprintf(" %s - Group: %s (%d hosts) ", actionType, groupName, len(members)))

	switch actionType {
	case "blackhole":
		form.AddInputField("Target IP/CIDR:", "", 50, nil, nil)
		form.AddButton("Apply to All", func() {
			target := form.GetFormItem(0).(*tview.InputField).GetText()
			t.applyGroupAction(groupName, members, actionType, target, "")
		})
	case "latency":
		form.AddInputField("Interface:", "", 20, nil, nil)
		form.AddInputField("Delay (e.g. 100ms):", "", 20, nil, nil)
		form.AddButton("Apply to All", func() {
			iface := form.GetFormItem(0).(*tview.InputField).GetText()
			delay := form.GetFormItem(1).(*tview.InputField).GetText()
			t.applyGroupAction(groupName, members, actionType, iface, delay)
		})
	// case "partition":
	// 	form.AddInputField("Source IP to block:", "", 50, nil, nil)
	// 	form.AddButton("Apply to All", func() {
	// 		sourceIP := form.GetFormItem(0).(*tview.InputField).GetText()
	// 		t.applyGroupAction(groupName, members, actionType, sourceIP, "")
	// 	})
	case "portblock":
		form.AddInputField("Source IP:", "", 30, nil, nil)
		form.AddInputField("Port:", "", 10, nil, nil)
		form.AddButton("Apply to All", func() {
			sourceIP := form.GetFormItem(0).(*tview.InputField).GetText()
			port := form.GetFormItem(1).(*tview.InputField).GetText()
			t.applyGroupAction(groupName, members, actionType, sourceIP, port)
		})
	}

	form.AddButton("Cancel", func() {
		t.showGroupActionMenu(groupName, members)
	})

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			t.showGroupActionMenu(groupName, members)
		}
		return event
	})
	t.app.SetRoot(form, true)
}

// applyGroupAction applies action to all hosts in a group
// Flow: loop hosts → skip disconnected → apply action → collect results → show summary popup.
func (t *TUI) applyGroupAction(groupName string, members []string, actionType, target, value string) {
	// Init var
	success := 0
	skipped := 0
	total := 0
	var errors []string

	for _, memberName := range members {
		// Find HostStatus by name
		var status *HostStatus
		for i := range t.statuses {
			if t.statuses[i].Host.Name == memberName {
				status = &t.statuses[i]
				break
			}
		}

		// skip if not found or not connected!
		if status == nil || !status.Connected || !status.Sudo {
			skipped++
			continue
		}

		// Apply action based on type
		var err error
		switch actionType {
		case "blackhole":
			// err = addBlackHole(status.Client, status.Host.Name, target)
			targets := strings.Split(target, ",")
			for _, _target := range targets {
				trimmed := strings.TrimSpace(_target)
				if trimmed == "" {
					continue
				}
				total++
				var ehhh error
				ehhh = addBlackHole(status.Client, status.Host.Name, trimmed)
				if ehhh != nil {
					errors = append(errors, memberName+":"+trimmed+": "+ehhh.Error())
				} else {
					success++
				}
			}
		case "latency":
			total++
			err = addLatency(status.Client, status.Host.Name, target, value)
		// case "partition":
		// 	err = addPartition(status.Client, status.Host.Name, target)
		case "portblock":
			total++
			err = addPortBlock(status.Client, status.Host.Name, target, value)
		}

		if err != nil {
			errors = append(errors, memberName+": "+err.Error())
		} else {
			success++
		}
	}

	// Build summary!
	// Count by target now! Example 3 ip apply in group that has 3 hosts --> 9!
	msg := fmt.Sprintf("Group [%s] %s: %d/%d succeeded", groupName, actionType, success, total)
	if skipped > 0 {
		msg += fmt.Sprintf("\n%d skipped (disconnected/no sudo)", skipped)
	}
	if len(errors) > 0 {
		msg += "\n\nFailed:\n" + strings.Join(errors, "\n")
	}
	t.showMessage(msg)
}
