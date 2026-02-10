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

## Story 4.4: Parameter input for actions
Need to do this before story 4.3!

Goal: Apply rules from TUI for real.

New concept:
- `tview.NewForm()` : Form with input field, like HTML form
- `AddInputField()`: Add field for text input
- `AddButton()`: Add button, like HTML button
- `GetFormItem(0)`: Get field by index
- `.*tview.InputField`: Type assertion, convert interface to struct??? No idea about this. Maybe convert interface to struct

![alt text](../images/04.png)

Result, hmmm? 
![alt text](../images/05.png)

Let's find out..
in root_cmd.go, we have define for that close
```go

					// Close connection here
					// status.Client.Close() // Temp comment for Story 4.4
				} else {
					fmt.Printf("%s: Sudo access NOT OK!\n", status.Host.User)
				}
			} else {
				fmt.Printf("%s: Failed. Issue %v\n", status.Host.Name, status.Error)
			}
```

After comment it
![alt text](../images/06.png.png).

And show other form we just added

Latency:
![alt text](../images/07.png)

Packet Loss:
![alt text](../images/08.png)

Test restore all
Before:
```bash
root@kienlt-jump:~# ip route|grep black
blackhole 99.99.99.99
```

![alt text](../images/09.png)

After:
```bash
root@kienlt-jump:~# ip route|grep black
root@kienlt-jump:~#
```

Exit and clean all effect Rule works well. They are tracked in memory.

