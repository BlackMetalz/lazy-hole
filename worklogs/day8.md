# Day 8: Feb 10, 2026

# Epic 4 - Terminal UI

## Story 4.2: Action menu for selected host
Goal: When user select host and press `enter`, show manu actions!

Closure issue, that happened in day3 but I don't understand, I will try again in day 8.

We had, example: `hosts = [hostA, hostB, hostC]`

Problem
```go
for i, status := range t.statuses {
    t.hostList.AddItem(..., func() {
        t.showActionMenu(status)   // ← false here!
    })
}
```

In loop, Go doesn't create new variable `status` each iteration, it just reuse the same variable and change value inside that.

Each func (closure) that you created, it will remember address of variable, hostC everytime --> wrong

So solution here is copy value of `status` to new variable `s`, then pass it to func.

But this is fixed in go 1.22++, not worth for remember but good to know Go was had a such as bug!

So output of this story:
![alt text](../images/03.png)