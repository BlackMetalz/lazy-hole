# lazy-hole v1 (MVP) - Expected Output

## What You'll Have When MVP is Done

### 🖥️ CLI Tool
Run `lazy-hole -c hosts.yaml` to:
- Load host configuration from YAML
- SSH to all hosts in parallel
- Check sudo access on each host
- Display host status (HEALTHY / NO SUDO / FAILED)

### 🌐 Network Commands
Apply these effects on any remote host:

| Command | What It Does |
|---------|--------------|
| **Blackhole** | Drop all traffic to specific IP/CIDR |
| **Latency** | Add network delay (e.g., 100ms) |
| **Packet Loss** | Drop random % of packets (e.g., 10%) |
| **Partition** | Block incoming traffic from specific IP |

### 📺 TUI (Terminal UI)
Interactive interface with:
- Host list with status indicators (green/yellow/red)
- Arrow key navigation
- Action menu per host
- Real-time display of active rules
- Parameter input for each action

### 🛡️ Safety Features
- Detect & protect SSH source IP
- Auto-timeout for rules
- "Restore All" to remove all effects
- Warning when quitting with active rules

---

## Example Usage Flow

```
1. Run: lazy-hole -c hosts.yaml
2. See host list with status
3. Select a host → Action menu
4. Choose: [L] Latency
5. Enter: 100ms
6. See: "mysql-node-1: DELAY 100ms on eth0"
7. Press [R] Restore All
8. Rules removed, back to normal
```

---

## Files You'll Create

```
lazy-hole/
├── main.go           # Entry point
├── root_cmd.go       # Cobra CLI
├── config.go         # YAML config loader
├── types.go          # Structs
├── ssh.go            # SSH connection
├── commands.go       # Network commands
├── tracker.go        # Effect tracking
├── tui.go            # Terminal UI (future)
└── hosts.yaml        # Your host config
```

---

## Story Points Summary

| Epic | Description | Points |
|------|-------------|--------|
| 1 | Project Setup & Configuration | 4 |
| 2 | SSH Connection | 10 |
| 3 | Core Commands (MVP) | 18 |
| 4 | TUI (Interactive Mode) | 14 |
| 5 | Safety Features | 7 |
| 6 | Polish & UX | 5 |
| **Total** | | **58 points** |
