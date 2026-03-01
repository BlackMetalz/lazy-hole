# Day 1:
Nothing much, I understand every concept in day 1. But I think i still need to write about wrap error!

# Day 2:
Many thing to write for this day.

Fprint vs Sprint
### Fprint
It print to writer (io.Writer) we passed
```go
fmt.Fprintln(os.Stderr, err)  // → write into stderr (fd 2)
fmt.Fprintln(os.Stdout, err)  // → write into stdout (fd 1)
fmt.Fprintln(file, err)       // → write into file
fmt.Fprintln(httpWriter, err) // → write into HTTP response
```

Like unix philosophy: Everything is a file descriptor. Fprint take destination as first argument.

### Sprint
It return string, simply return formatted string and you have to print it.
```go
s := fmt.Sprintf("hello %s", "world")  // s = "hello world" — return string
fmt.Println(s)                          // you still need to print it self.
```

### SSH
Copy and paste code for ssh to host and execute command.

# Day 3:
Real fun things.

### Concurrency with Goroutines and Channels

I will write down again with my understand
So I have func testAllHosts that run goroutines to ssh to multiple host together.

First, i created buffered channel with length is length of hosts
```go
results := make(chan HostStatus, len(hosts))
```

Second, i spawn goroutine for each host
```go
	// spawn goroutine for each host
	for _, host := range hosts {
		go func(h Host) {
			client, err := connectSSH(h)

			status := HostStatus{
				Host:      h,
				Connected: err == nil,
				Error:     err,
				Client:    client,
			}

			if status.Connected {
				// Check sudo if connected
				status.Sudo = checkSudo(client)

				// Prevent self-block
				sourceIP, _ := detectSSHSourceIP(status.Client)
				status.SSH_SourceIP = sourceIP
			}

			// This shit is import
			results <- status // send result into fucking channel
		}(host) // Fucking second important.
		// Pass host as argument to goroutine to avoid closure capture the last value of host?
	}
```

After everything, we take result from channel
```go
	// Collect all results
	var statuses []HostStatus
	for i := 0; i < len(hosts); i++ {
		status := <-results // Receive from channel
		statuses = append(statuses, status)
		// status.Client.Close() // Close connection after test done!
		// Temp comment this out to avoid closing connection before using it
	}

```

I need to rewrite because issue related to goroutines not copy value of host was fixed in Go1.22+

### Many more. Holy fucking shit for day 3!!!

# Day 4:

### sync.Mutex
Mutex = Mutal Exclusion : Only 1 person can enter at a time.

Main understand: prevent race condition when multiple goroutines access same variable. Especially in map. Go map not concurrent-safe by default, together write will be panic!

I like gemini explaination. HAHA
```
Mutex = Mutual Exclusion = "chỉ 1 thằng được vào tại 1 thời điểm"

Nó giống cái cửa toilet — khóa lại khi bro vào, thằng khác phải đợi bro ra mới vào được. Nếu không có khóa → 2 thằng vào cùng lúc → disaster 💀
```

### Constructor pattern
So this is for go specific, because Go doesn't have class, so no constructor like Python/Java (I don't remember i used construct in python also. LOL)

And what the fuck is the problem?. Zero value in Go not always "ready to use".
Go have theory "zero value should be useful"
https://dave.cheney.net/2013/01/19/what-is-the-zero-value-and-why-is-it-useful#:~:text=Mutex%20is%20declared.,following%20code%20will%20output%20false.

But not alway struct able to achieve that point.

```go
// Init new tracker
tracker := EffectTracker{} // Effects = nil --> map not init
tracker.Add("host1", ...) // Panic, write nil into map
```

That is why we have to make map before write, zero value of map is `nil`, read is oke (return zero value), but write will be panic!

So convention NewXxx()
```go
func NewEffectTracker() *EffectTracker {
    return &EffectTracker{
        effects: make(map[string][]ActiveEffect),
    }
}
```

This is convention, not required syntax. But it is good practice!

So when we need it, when not?

Need:
- Map --> need make() first
- Channel --> need make() first

I asked Gemini to generate code for me for this example:
```go
package main

import "fmt"

type EffectTracker struct {
	effects map[string][]string
}

// Constructor pattern
func NewEffectTracker() *EffectTracker {
	return &EffectTracker{
		effects: make(map[string][]string),
	}
}

func (t *EffectTracker) Add(hostname, effect string) {
	t.effects[hostname] = append(t.effects[hostname], effect)
}

func main() {
	// Case 1: WITHOUT constructor → PANIC
	fmt.Println("=== Case 1: Zero value (no constructor) ===")
	badTracker := EffectTracker{} // effects = nil
	fmt.Printf("effects == nil? %v\n", badTracker.effects == nil)

	// Reading nil map → ok, return zero value
	fmt.Printf("Read nil map: %v (no panic)\n", badTracker.effects["host1"])

	// Writing nil map → PANIC!
	fmt.Println("About to write to nil map...")
	// Uncomment line below to see panic:
	// badTracker.Add("host1", "blackhole")  // panic: assignment to entry in nil map

	fmt.Println()

	// Case 2: WITH constructor → works fine
	fmt.Println("=== Case 2: With NewEffectTracker() ===")
	goodTracker := NewEffectTracker() // effects = initialized map
	fmt.Printf("effects == nil? %v\n", goodTracker.effects == nil)

	goodTracker.Add("host1", "blackhole")
	goodTracker.Add("host1", "latency-100ms")
	goodTracker.Add("host2", "packet-loss-5%")

	for host, effects := range goodTracker.effects {
		fmt.Printf("%s: %v\n", host, effects)
	}
}
```

That's all i'm able to understand right now about constructor pattern. LOL

And I have no fucking idea about this trick, design pattern
```go
// Remove specific effect from host
func (t *EffectTracker) Remove(hostname string, effect ActiveEffect) {
	t.mu.Lock()
	defer t.mu.Unlock()

	effects := t.effects[hostname]
	for i, e := range effects {
		if e.Type == effect.Type && e.Target == effect.Target {
			t.effects[hostname] = append(effects[:i], effects[i+1:]...)
			return
		}
	}
}
```

Lets me search if any better solution for this! Complete hard to understand! Asked AI several time, it said this is the best already. Since performance will never be an issue even this is O(n) complexity, but n=10 is not really an issue to change. Seem legit to me!

### Slicing xD
I have no idea, I thought i'm able to understand it well from tutorials, but no...
Another real example for me to learn:
```go
package main

import "fmt"

func main() {
	s := []string{
		"A", "B", "C", "D", "E",
	}
	fmt.Println("Slice of string lenth: ", len(s))

	// Remove element number 2 (C) from this fucking slice of string
	i := 2
	fmt.Println(s[:i])
	fmt.Println(s[:i+1])

	// We have 2 fucking ways to remove
	// FIRST //
	var x []string // nil slice. Not able to append directly like: x = append(s[:i], s[i+1:])
	x = append(s[:i], s[i+1:]...)
	fmt.Println("First method: ", x)

	// SECOND //
	// Make it zero value
	x2 := make([]string, 0) // init zero value slice of string
	x2 = append(x2, s[:i]...)
	x2 = append(x2, s[i+1:]...)
	fmt.Println("Second method: ", x2)

}
```