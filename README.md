# Overview

### What is Lazy Hole?
A CLI/TUI tool to simulate network failures for testing distributed systems (like MySQL Galera cluster). Instead of remembering complex `tc` and `ip route` commands, you use an interactive interface.

### Why build this?
- Testing: How does your app behave when network fails?
- Learning: SSH in Go, TUI frameworks, goroutines, channels
- Productivity: I hate googling `tc qdisc` syntax every time xD

### Architecture
Run on jump host → SSH to target hosts → Execute network commands remotely

# Installation

### Build from source
```bash
go mod tidy
go build -o lazy-hole .
```

### Install from release

#### Ubuntu/Linux (amd64)
```bash
curl -sL https://github.com/BlackMetalz/lazy-hole/releases/latest/download/lazy-hole-linux-amd64 -o /tmp/lazy-hole
chmod +x /tmp/lazy-hole
sudo mv /tmp/lazy-hole /usr/local/bin/lazy-hole
```

#### macOS (Apple Silicon)
```bash
curl -sL https://github.com/BlackMetalz/lazy-hole/releases/latest/download/lazy-hole-darwin-arm64 -o /tmp/lazy-hole
curl -sL https://github.com/BlackMetalz/lazy-hole/releases/latest/download/lazy-hole-darwin-arm64 -o /tmp/lazy-hole
chmod +x /tmp/lazy-hole
sudo mv /tmp/lazy-hole /usr/local/bin/lazy-hole
```
