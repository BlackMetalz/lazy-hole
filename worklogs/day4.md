# Day 4: Feb 5, 2026

### Story 3.5: Remove latency 

Command: `sudo tc qdisc del dev eth0 root`

Handle when delete rule not exists
```bash
root@kienlt-jump:~# tc qdisc del dev eth0 root 2> abc.txt
root@kienlt-jump:~# cat abc.txt
Error: Cannot delete qdisc with handle of zero.
```

So you clearly can see we direct error to `abc.txt`. That is why we use to check strings contains in `result.Stderr`

### Story 3.6: Add packet loss
Command: `sudo tc qdisc add dev eth0 root netem loss 10%`

Need to handle duplicate also
```bash
root@kienlt-jump:~# tc qdisc add dev eth0 root netem loss 5%
root@kienlt-jump:~# tc qdisc add dev eth0 root netem loss 5%
Error: Exclusivity flag on, cannot modify.
```

Hmm, rename `removeLatency` to `removeTCRules`

### Story 3.7: Block traffic

Command: `sudo iptables -A INPUT -s <IP> -j DROP`

And yeah, common cheatsheet for iptables here bro:
```bash
Iptables is Linux's classic firewall tool for managing packet filtering and NAT rules. Here's a cheatsheet of common commands for quick reference.

## Basic Commands
- List all rules: `iptables -L -v -n`
- List with line numbers: `iptables -L -v -n --line-numbers`
- Flush all rules: `iptables -F`
- Set default policy (e.g., DROP input): `iptables -P INPUT DROP`
- Save rules: `iptables-save > /etc/iptables.rules`
- Restore rules: `iptables-restore < /etc/iptables.rules` [gist.github](https://gist.github.com/davydany/0ad377f6de3c70056d2bd0f1549e1017)

## Common Rules
- Allow SSH (port 22): `iptables -A INPUT -p tcp --dport 22 -j ACCEPT`
- Allow HTTP/HTTPS: `iptables -A INPUT -p tcp --dport 80 -j ACCEPT` and `--dport 443`
- Allow established connections: `iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT`
- Block all other input: `iptables -A INPUT -j DROP`
- Allow loopback: `iptables -A INPUT -i lo -j ACCEPT` [andreafortuna](https://andreafortuna.org/2019/05/08/iptables-a-simple-cheatsheet/)

## Delete Rules
- Delete by line number: `iptables -D INPUT 5`
- Delete specific rule: `iptables -D INPUT -p tcp --dport 22 -j ACCEPT`
- Delete chain: `iptables -X customchain` [andreafortuna](https://andreafortuna.org/2019/05/08/iptables-a-simple-cheatsheet/)

## NAT Examples
- Masquerade for outbound: `iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE`
- Port forward: `iptables -t nat -A PREROUTING -p tcp --dport 8080 -j REDIRECT --to-port 80` [digitalocean](https://www.digitalocean.com/community/tutorials/iptables-essentials-common-firewall-rules-and-commands)

## Chain Management
| Chain   | Purpose                  |
|---------|--------------------------|
| INPUT   | Packets to this host    |
| OUTPUT  | Packets from this host  |
| FORWARD | Routed packets          |
| PREROUTING | Incoming before routing |
| POSTROUTING | Outgoing after routing |  [bashsenpai](https://bashsenpai.com/resources/cheatsheets/iptables)

Run as root or with sudo. For persistence, use tools like iptables-persistent on Debian-based systems.
```


## Epic 5 instead of Epic 4
- Easy to finish first
- No need rush for MVP

### Story 5.1 Track active effects

HOLY FUCKING SHIT. I COMPLETE UNDERSTAND NOTHING!!!!!

only function i uderstand is delete() which is build in
```
// The delete built-in function deletes the element with the specified key
// (m[key]) from the map. If m is nil or there is no such element, delete
```

```go
package main

import "fmt"

func main() {
	_map := map[string]int{
		"banana":    1,
		"apple":     2,
		"holy_fuck": 3,
	}

	fmt.Printf("Before Delete: %v\n", _map)

	// Remove holy_fuck key from _map/ Format delete(map,key)
	delete(_map, "holy_fuck")

	fmt.Printf("After Delete: %v\n", _map)
}
```

Holy fucking shit, Gemini thought I was middle/senior in Go? I'm just intern >.>
Give up? no, just split them into small slice to continue

So struct is data container which we used to store data. Gemini trying to teach me about method receiver (struct with methods). New lesson: Go allow attach func INTO struct. It is called Method. Hmmm

So my understanding: struct + function = method.

// Normal function
```go
func getName(h Host) string {
    return h.Name
}
```

// Method - attach into struct Host
```go
func (h Host) GetName() string {
    return h.Name
}
```

Syntax: `func (h Host)` means this fucking function belong to `Host` struct
Usage:
```go
host := Host{Name: "host-name-1", IP: "1.1.1.1"}

// Normal func
name := GetName(host)

// Method
name := host.GetName() // call as property of host
```

Complete example: https://go.dev/play/p/Zun1edKXRiD
```go
package main

import "fmt"

type Host struct {
	Name string
	IP   string
}

// Normal Func
func getName(h Host) string {
	return h.Name
}

// Method
func (h Host) getName() string {
	return h.Name
}

func main() {
	newHost := Host{Name: "host-name-1", IP: "1.1.1.1"}

	fmt.Println(newHost)

	// Call from normal func
	fmt.Println(getName(newHost))

	// Call from method
	fmt.Println(newHost.getName())
}
```

Hmmm, it seems related to `interfaces` in Go, but it told me that I don't need to understand `interfaces` right now. I remember keyword `interfaces` because in K8S Operator, `interfaces` is one of required before write any operator.

in tracker.go which i copied, the reason to use methods is much more simple. It is to make code more readable and maintainable.
```go
tracker := NewEffectTracker()
tracker.Add("host1", effect) // Call method
tracker.Remove("host1", effect) // Call method
tracker.Get("host1")        // Call method
```

We can do like that, instead of
```go
effects := make(map[string][]ActiveEffect) // Make new map with key is string and value is slice of ActiveEffect

AddEffect(effects, "host1", effect)
RemoveEffect(effects, "host1", effect)
GetEffect(effects, "host1")
```

Hmm. I see make map, I haven't familiar with that syntax, but not anymore after this example
```go
package main

import "fmt"

type ActiveEffect struct {
	Type   string 
	Target string 
	Value  string 
}

func main() {
    // Create map with string key and value is slice of ActiveEffect
	chaosMap := map[string][]ActiveEffect{
		"web-01": {
			{
				Type:   "latency",
				Target: "10.0.2.45",
				Value:  "120ms",
			},
			{
				Type:   "packetloss",
				Target: "10.0.2.45",
				Value:  "8%",
			},
		},

		"api-03": {
			{
				Type:   "blackhole",
				Target: "172.16.10.0/24",
				Value:  "",
			},
		},
	}

	for hostname, effects := range chaosMap {
		fmt.Printf("Host: %s\n", hostname)
		for i, eff := range effects {
			fmt.Printf("  #%d: %s → %s = %s\n", i+1, eff.Type, eff.Target, eff.Value)
		}
		fmt.Println()
	}
}
```

ok, that's for simple method. Now let's talk about pointer method
```go
func (t *EffectTracker) Add(...) {
```

From what i still remember, this will modify original EffectTracker struct. And without pointer, it will only modify value of "copy" struct, so after function end, original struct changed nothing.

So next is `sync.Mutex`, Gemini said this is simple?
Issue: when 2 goroutines access map together --> crash. Or we can be understand that this shit prevent race condition when multiple goroutines access map together.
Solution: Mutext = Lock. Only 1 goroutine can access map at a time. 
```go
t.mu.Lock() // Lock this shit
// Do some stuff here with map
t.mu.Unlock() // Unlock this shit
```

And `defer t.mu.Unlock()` auto unlock when function return, to not forget unlock.

Constructor pattern: Holy fucking shiet, I don't expect doing design pattern in this KISS project. But Gemini said it is good practice to use constructor pattern.
```go
func NewEffectTracker() *EffectTracker {
	return &EffectTracker{
		effects: make(map[string][]ActiveEffect),
	}
}
```

Go doesn't has `class` or `constructor`. Pattern is `convention` to create struct that already initialized. Why it needed?
```go
tracker := EffectTracker{} // effects = nil. Map is not initialized
tracker.Add("host1", ...) // Panic, nill map
```

So `NewEffectTracker()` is just a function to create EffectTracker struct that already initialized map. 

And next Remove From slice with tricky syntax
```go
t.effects[hostname] = append(effects[:i], effects[i+1:]...)
```

Let's breakdown:
- `effects[:i]` : slice from start to index i (not include i)
- `effects[i+1:]` : slice from i+1 to the end
- `append(...)` : append slice to slice, remove element at index i

Ex:
```
Before: [A, B, C, D, E]  (remove C at index 2)
[:2]  = [A, B]
[3:]  = [D, E]
After = [A, B, D, E]
```