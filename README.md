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
- Current state: story 1.1
```bash
go mod tidy
go run .
```