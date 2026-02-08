# Day 6: Feb 8, 2026

## Story 5.3: Remove all effect for all hosts for restore
Logic is pretty simple and easy to understand:

- Get effects from tracker (global tracker)
- loop through hostname that has effects
- find client for each hostname from hostStatuses
- Call func restoreHost() for each host
- Count and report number of host restored

## Story 5.4: Auto-restore on exit
Goal: when user exit application, auto restore all effects (remove all affects applied on hosts)

Logic is simple also:
- Listen to signals (ctrl+C/kill): `signal.Notify()`
- `<-c` - wait/block until received signal