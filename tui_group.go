package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// groupEntry holds ordered group data for stable index mapping
type groupEntry struct {
	name    string
	members []string
}

// buildGroupList, create list from group
func (t *TUI) buildGroupList() {
	t.groupList = tview.NewList()
	t.groupList.SetTitle(" Groups ").SetBorder(true)

	// Collect groups from statuses into a map first
	groupMap := make(map[string][]string)
	for _, s := range t.statuses {
		if s.Host.Group != "" {
			groupMap[s.Host.Group] = append(groupMap[s.Host.Group], s.Host.Name)
		}
	}

	// Convert to ordered slice so index → group is stable
	var groupEntries []groupEntry
	for gName, gMembers := range groupMap {
		groupEntries = append(groupEntries, groupEntry{name: gName, members: gMembers})
	}

	// Add each group to list view
	for idx, g := range groupEntries {
		// Build effect details for each member (same style as host view)
		effectStr := ""
		for _, memberName := range g.members {
			effects := effectTracker.Get(memberName)
			for _, e := range effects {
				switch e.Type {
				case EffectBlackHole:
					effectStr += fmt.Sprintf(" (BlackHole:%s→%s)", memberName, e.Target)
				case EffectLatency:
					effectStr += fmt.Sprintf(" (Latency:%s %s→%s)", e.Value, e.Target, memberName)
				case EffectPacketLoss:
					effectStr += fmt.Sprintf(" (PacketLoss:%s%% %s→%s)", e.Value, e.Target, memberName)
				case EffectPortBlock:
					effectStr += fmt.Sprintf(" (PortBlock:%s:%s→%s)", e.Target, e.Value, memberName)
				}
			}
		}
		// Build a label to display in group views
		labels := fmt.Sprintf("%s (%d hosts)%s", g.name, len(g.members), effectStr)
		var shortcut rune
		// If there is more than 9 group, > 9th++, they will not able to receive shortcut. Haha
		if idx < 9 {
			shortcut = rune('1' + idx)
		} else {
			shortcut = 0
		}

		// Capture loop vars for closures
		gName := g.name
		gMembers := g.members
		t.groupList.AddItem(labels, strings.Join(gMembers, ", "), shortcut, func() {
			t.showGroupActionMenu(gName, gMembers)
		})
	}

	// ESC = back to hosts. Specific handler
	// Ughhh, I hate this, need to refactor this?? Tired of typing/copying SetInputCapture...
	// Group-specific + common shortkeys (no middleware in tview, must duplicate)
	t.groupList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Group-specific keys
		if event.Key() == tcell.KeyEscape || event.Rune() == 'l' {
			t.viewMode = "hosts"
			t.refreshHostList()
			t.app.SetRoot(t.layout, true)
		}
		// Common keys (same as setupHostListKeys)
		if event.Rune() == 'q' {
			t.app.Stop()
		}
		if event.Rune() == 'h' {
			t.showHistory()
		}
		if event.Rune() == 'u' {
			t.showUndoConfirm()
		}
		if event.Rune() == '?' {
			t.showHelp()
		}
		if event.Rune() == '/' {
			t.showFilterDialog()
		}
		// k = Kill all rules for the highlighted group
		if event.Rune() == 'k' {
			idx := t.groupList.GetCurrentItem()
			if idx >= 0 && idx < len(groupEntries) {
				g := groupEntries[idx]
				t.killGroupRules(g.name, g.members)
			}
		}
		return event
	})
}

// Switch to group view
func (t *TUI) switchToGroupView() {
	t.viewMode = "groups"
	t.buildGroupList()
	// Rebuild layout with group list
	t.buildLayout()
	t.app.SetRoot(t.layout, true)
}

// killGroupRules removes all active rules for every host in a group.
// Shows a confirm dialog first, then restores each connected member.
func (t *TUI) killGroupRules(groupName string, members []string) {
	// Count total active effects across the group
	totalEffects := 0
	for _, memberName := range members {
		totalEffects += len(effectTracker.Get(memberName))
	}

	if totalEffects == 0 {
		t.showMessage(fmt.Sprintf("Group [%s]: no active rules to remove!", groupName))
		return
	}

	msg := fmt.Sprintf("Remove ALL %d rules from group [%s] (%d hosts)?", totalEffects, groupName, len(members))
	t.showConfirmDialog(msg, func() {
		success := 0
		skipped := 0
		var errors []string

		for _, memberName := range members {
			// Find client for this member
			var status *HostStatus
			for i := range t.statuses {
				if t.statuses[i].Host.Name == memberName {
					status = &t.statuses[i]
					break
				}
			}
			if status == nil || !status.Connected || !status.Sudo {
				skipped++
				continue
			}
			err := restoreHost(status.Client, memberName)
			if err != nil {
				errors = append(errors, memberName+": "+err.Error())
			} else {
				success++
			}
		}

		result := fmt.Sprintf("Group [%s] Kill Rules: %d/%d hosts restored", groupName, success, len(members))
		if skipped > 0 {
			result += fmt.Sprintf("\n%d skipped (disconnected/no sudo)", skipped)
		}
		if len(errors) > 0 {
			result += "\n\nErrors:\n" + strings.Join(errors, "\n")
		}

		// Rebuild group list so effects disappear immediately
		t.buildGroupList()
		t.buildLayout()
		t.showMessage(result)
	})
}

// func showGroupActionMenu, shơ action menu fỏ group
func (t *TUI) showGroupActionMenu(groupName string, members []string) {
	actionList := tview.NewList()
	actionList.SetTitle(" Actions for group: " + groupName + "").SetBorder(true)

	actionList.AddItem("Blackhole", "Drop traffic to IP/CIDR (manual input)", 'b', func() {
		t.showGroupInputForm(groupName, members, "blackhole")
	})
	actionList.AddItem("Blackhole by Group", "Block all IPs of another group", 'B', func() {
		t.showGroupBlackholeByGroup(groupName, members)
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
		// Add q to quit in group view!
		if event.Rune() == 'q' {
			t.app.Stop()
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
	// Snapshot undo stack size before applying, so we can tag new entries with batch ID
	stackBefore := undoStack.Len()

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

		if actionType != "blackhole" && err != nil {
			errors = append(errors, memberName+": "+err.Error())
		} else if actionType != "blackhole" {
			success++
		}
	}

	// Tag all new undo entries with same batch ID for batch undo!
	if undoStack.Len() > stackBefore {
		batchID := fmt.Sprintf("group-%s-%d", groupName, time.Now().UnixNano())
		undoStack.TagBatch(stackBefore, batchID)
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

	// Refresh group view so active rules show immediately!
	t.buildGroupList()
	t.buildLayout()

	t.showMessage(msg)
}

// showGroupBlackholeByGroup shows a list of OTHER groups to blackhole
// Collects all IPs from selected target group, then applies blackhole on all hosts in current group
func (t *TUI) showGroupBlackholeByGroup(groupName string, members []string) {
	// Collect all groups and their IPs (exclude current group)
	// map[groupName] => []IP
	groupIPs := make(map[string][]string)
	for _, s := range t.statuses {
		if s.Host.Group != "" && s.Host.Group != groupName {
			groupIPs[s.Host.Group] = append(groupIPs[s.Host.Group], s.Host.IP)
		}
	}

	if len(groupIPs) == 0 {
		t.showMessage("No other groups found!")
		return
	}

	// Build list of target groups
	list := tview.NewList()
	list.SetTitle(" Blackhole by Group - from: " + groupName + " ").SetBorder(true)

	idx := 0
	for targetGroup, ips := range groupIPs {
		label := fmt.Sprintf("%s (%d IPs: %s)", targetGroup, len(ips), strings.Join(ips, ", "))
		var shortcut rune
		if idx < 9 {
			shortcut = rune('1' + idx)
		} else {
			shortcut = 0
		}

		// Capture for closure
		targetIPs := strings.Join(ips, ",")
		targetName := targetGroup

		list.AddItem(label, "", shortcut, func() {
			// Confirm before applying
			msg := fmt.Sprintf("Blackhole all IPs of group <%s> on all hosts in group <%s>?\nIPs: %s", targetName, groupName, targetIPs)
			t.showConfirmDialog(msg, func() {
				t.applyGroupAction(groupName, members, "blackhole", targetIPs, "")
			})
		})
		idx++
	}

	list.AddItem("Back", "", 0, func() {
		t.showGroupActionMenu(groupName, members)
	})

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			t.showGroupActionMenu(groupName, members)
		}
		return event
	})

	t.app.SetRoot(list, true)
}
