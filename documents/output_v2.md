# lazy-hole v2 - Expected Output (Future Enhancements)

> ⚠️ **Prerequisite:** Complete MVP (v1) first!

## What v2 Adds

### 🌐 Additional Network Commands

| Command | What It Does |
|---------|--------------|
| **Bandwidth Limit** | Limit speed (e.g., 10mbit) |
| **Corruption** | Flip random bits in packets |
| **Reordering** | Deliver packets out of order |
| **Duplication** | Send same packet twice |
| **Interface Down** | Disable entire NIC |
| **REJECT vs DROP** | Choose error behavior |

### 🎯 Preset Scenarios (One-Click)

| Scenario | What It Does |
|----------|--------------|
| **Split-Brain** | Two hosts can't see each other (DB testing) |
| **Network Partition** | Isolate groups of hosts |
| **WAN Simulation** | Add 50ms + 0.1% loss (cross-DC) |
| **Custom Scenarios** | Save/load your own presets |

### 📝 Action History & Audit

- Log all actions with timestamp
- Undo last action
- Export current state to YAML
- Import saved state

### 🖥️ Advanced Host Management

- Host groups (by role: databases, web, etc.)
- Add/edit/remove hosts in TUI
- Background health monitoring

### ⚡ Multi-Action Support

- Combine rules: delay + loss + corruption
- Apply to multiple hosts at once
- Detailed rule summary view

### ⏰ Scheduling

- Schedule rules for specific time
- Periodic chaos mode (random chaos every X hours)

---

## Priority Recommendation

**Do First (High Value):**
1. Split-brain scenario - Your original use case!
2. Action history - Important for production use
3. Apply to multiple hosts - Big time saver

**Do Second (Nice to Have):**
4. Bandwidth limiting
5. Slow network preset
6. Undo feature

**Do Later (Advanced):**
7. Scheduling
8. Host management
9. Reorder/Duplicate

---

## Story Points Summary

| Epic | Description | Points |
|------|-------------|--------|
| V2-1 | Additional Network Commands | 14 |
| V2-2 | Preset Scenarios | 10 |
| V2-3 | Action History & Audit | 8 |
| V2-4 | Advanced Host Management | 10 |
| V2-5 | Multi-Action Support | 8 |
| V2-6 | Scheduling | 6 |
| **Total** | | **56 points** |
