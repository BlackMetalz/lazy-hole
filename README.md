# Overview

## What is Lazy Hole?
A CLI/TUI tool to simulate network failures for testing distributed systems (like MySQL Galera cluster). Instead of remembering complex `tc qdisc`, `iptables` and `ip route` commands, you use an interactive interface.

## Motivation to build this trash?
- Testing: How does your app behave when network fails?
- Learning: SSH in Go, TUI frameworks, goroutines, channels and more....
- Productivity: I hate googling or typing `tc qdisc`, `iptables`, `ip route` commands every time xD
- Finally, I want a tool for daily usage that give a happy time when working xD

## Architecture
Run on jump host → SSH to target hosts → Execute network commands remotely

## Features
- **Blackhole routing** — Drop all traffic to specific IP/CIDR (ip route blackhole). Simulates DNS/routing failure, network down!
- **Latency injection** — Add delay to network interfaces (tc qdisc)
- **Packet loss** — Simulate unreliable network (tc qdisc)
- ~~**Network partition** — Block traffic from specific source IPs (iptables). Simulates firewall misconfiguration!~~ (Removed)
- **Port blocking** — Block specific port from source IP (iptables)
- **Auto-restore** — Cleanup all effects on exit (Ctrl+C safe)
- **K9s-style TUI** — Interactive terminal UI with keyboard shortcuts (Motivation: I love k9s)
- **Host filtering** — Filter hosts in TUI

## Requirements
- Target hosts need `tc` + `ip route` + `iptables` (commonly installed in Linux!)
- SSH key-based authentication
- Sudo access in target hosts

# Installation

## Build from source
```bash
go mod tidy
go build -o lazy-hole .
```

## Install from release

### Ubuntu/Linux (amd64)
```bash
curl -sL https://github.com/BlackMetalz/lazy-hole/releases/latest/download/lazy-hole-linux-amd64 -o /tmp/lazy-hole
chmod +x /tmp/lazy-hole
sudo mv /tmp/lazy-hole /usr/local/bin/lazy-hole
lazy-hole -v
```

### macOS (Apple Silicon)
```bash
curl -sL https://github.com/BlackMetalz/lazy-hole/releases/latest/download/lazy-hole-darwin-arm64 -o /tmp/lazy-hole
chmod +x /tmp/lazy-hole
sudo mv /tmp/lazy-hole /usr/local/bin/lazy-hole
lazy-hole -v
```

## Usage

### 1. Create config file
```yaml
hosts:
  - name: mysql-node-1
    ip: 10.0.0.5
    ssh_user: kienlt
    ssh_key: ~/.ssh/id_rsa
  - name: mysql-node-2
    ip: 10.0.0.6
    ssh_user: kienlt
```

### 2. Run
```bash
lazy-hole -c /path/to/hosts.yaml
```

## Keyboard Shortcuts

### Host View (default)

| Key | Action |
|-----|--------|
| `Enter` | Open action menu for selected host |
| `r` | Refresh host status |
| `/` | Filter hosts by name |
| `p` | Show protected IPs (SSH source) |
| `u` | Undo last action |
| `h` | View action history |
| `g` | Switch to group view |
| `?` | Show help |
| `Esc`/`q` | Quit (auto-restore effects) |

### Group View

| Key | Action |
|-----|--------|
| `Enter` | Open action menu for selected group |
| `l`/`Esc` | Switch back to host view |
| `u` | Undo last action |
| `h` | View action history |
| `/` | Filter hosts |
| `?` | Show help |
| `q` | Quit |
