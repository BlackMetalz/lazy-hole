# Day 12: Mar 01, 2026

# Trying to finish some backlog with help of AI xD

### Make host list order is persistent, follow order in config!

But we need to understand problem first. That we are using goroutines in function `testAllHosts()`, whichever ssh connection finished first, it will be append to channel. After that, we will have a result from channel in order that ssh connection finished, not in order of config file!

```go
for i := 0; i < len(hosts); i++ {
    status := <-results // whoever finishes first gets appended first
    statuses = append(statuses, status)
}
```

So we only need to update func `testAllHosts()`. Gemini give me solution:

`pre-allocate` a slice with `len(hosts)` and write each result to its original index instead of appending from channel. Sound good to me!