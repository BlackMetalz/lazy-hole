package main

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

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
	actionList.AddItem("Blackhole", "Drop traffic to IP/CIDR (manual input)", 'b', func() {
		t.showInputForm(status, "blackhole")
	})

	actionList.AddItem("Blackhole by Group", "Block all IPs of a group", 'B', func() {
		t.showHostBlackholeByGroup(status)
	})

	actionList.AddItem("Latency", "Add network delay", 'l', func() {
		t.showInputForm(status, "latency")
	})

	actionList.AddItem("Packet Loss", "Drop random packets", 'p', func() {
		t.showInputForm(status, "packetloss")
	})
	// actionList.AddItem("IPtables Partition", "Block source IP", 'i', func() {
	// 	t.showInputForm(status, "partition")
	// })

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

		/*
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
		*/

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

// showHostBlackholeByGroup shows list of groups to blackhole on a SINGLE host
func (t *TUI) showHostBlackholeByGroup(status HostStatus) {
	// Collect all groups and their IPs
	// map[groupName] => []IP
	groupIPs := make(map[string][]string)
	for _, s := range t.statuses {
		if s.Host.Group != "" {
			groupIPs[s.Host.Group] = append(groupIPs[s.Host.Group], s.Host.IP)
		}
	}

	if len(groupIPs) == 0 {
		t.showMessage("No groups found!")
		return
	}

	list := tview.NewList()
	list.SetTitle(" Blackhole by Group - " + status.Host.Name + " ").SetBorder(true)

	idx := 0
	for targetGroup, ips := range groupIPs {
		label := fmt.Sprintf("%s (%d IPs: %s)", targetGroup, len(ips), strings.Join(ips, ", "))
		var shortcut rune
		if idx < 9 {
			shortcut = rune('1' + idx)
		} else {
			shortcut = 0
		}

		targetIPs := strings.Join(ips, ",")
		targetName := targetGroup

		list.AddItem(label, "", shortcut, func() {
			msg := fmt.Sprintf("Blackhole all IPs of group <%s> on <%s>?\nIPs: %s", targetName, status.Host.Name, targetIPs)
			t.showConfirmDialog(msg, func() {
				// Apply blackhole for each IP
				targets := strings.Split(targetIPs, ",")
				success := 0
				var errors []string
				for _, target := range targets {
					trimmed := strings.TrimSpace(target)
					if trimmed == "" {
						continue
					}
					err := addBlackHole(status.Client, status.Host.Name, trimmed)
					if err != nil {
						errors = append(errors, trimmed+": "+err.Error())
					} else {
						success++
					}
				}
				msg := fmt.Sprintf("Blackhole on %s: %d/%d added", status.Host.Name, success, len(targets))
				if len(errors) > 0 {
					msg += "\n\nFailed:\n" + strings.Join(errors, "\n")
				}
				t.showMessage(msg)
			})
		})
		idx++
	}

	list.AddItem("Back", "", 0, func() {
		t.showActionMenu(status)
	})

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			t.showActionMenu(status)
		}
		return event
	})

	t.app.SetRoot(list, true)
}
