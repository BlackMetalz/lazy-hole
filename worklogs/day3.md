# Day 3: Feb 4, 2026

## Story 2.2 Goroutines, concurrency.

Interesting section. I believe i heard a lot about goroutines or concurrency in Go, but i never make an effort to make it work. So what is Goroutine?
- Lightweight thread, managed by Go runtime.
- Create goroutine by `go` before function call
- Why lightweight? ~2KB stack size, vs 8 MB(default?, output of `ulimit -s`) for thread in Linux.
Ex:
```go
go doSomething()
```

Hmm, I need to understand Channel as well. So what is Channel?
- pipe to communicate between goroutines
- send and receive data safely
- make channel syntax: `ch := make(chan)`
That is theory, no idea how it works!
```go
ch := make(chan int) // unbeffered
ch := make(chan int,10) // buffered (capacity 10)
```

So pattern for this shit: SSH to all hosts in parallel
- Spawn N goroutines (N = number of host in host config!)
- Each goroutine connect SSH, send result into channel
- Main goroutine collect N result from channel!

This shit is literal hard! Without help of AI i don't think i can make it works under 30 minutes!

But there is something need to notice or write down.
Why unbuffered channel can be deadlock?
Copied this shit from Gemini xD
```go
ch := make(chan int) // init unbuffered channel

// Goroutine 1 want to send
ch <- 1 // Block // Wait receiver

// Goroutine 2 want to send
ch <- 2 // Block // Wait receiver
```
with Buffered
```go
ch := make(chan int, 2) // buffer size 2

ch <- 1 // OK, send buffer slot 1
ch <- 2 // ok, send buffer slot 2
```

Still not enough to understand, lets dig deeper!
So let's go back to theory. channel is pipe between goroutines and Main(receiver)
```# By gemini
                    ┌─────────────────┐
Goroutine 1 ───────►│                 │
Goroutine 2 ───────►│   CHANNEL       │───────► Main (receiver)
Goroutine 3 ───────►│   (pipe)        │
                    └─────────────────┘
```

Wait receiver? Wait who?. Receiver = Main goroutine (we often knows as `<-results`)

```
Main goroutine          Goroutine 1         Goroutine 2     Goroutine 3
     |                      |                   |                   |
     |--spawn-------------->|                   |                   |
     |--spawn---------------------------------->|                   |
     |--spawn-------------------------------------------------->|
     |                      |                   |                   |
     |                   connect...          connect...         connect...
     |                      |                   |                   |
     |<--result-------------|                   |                   |
     |<--result---------------------------------|                   |
     |<--result---------------------------------------------------------|
     |
  collect all
```


// Sender (goroutines)
results <- status // I want put fucking status into pipe!

// Receiver (main)
status := <-results // I want take fucking status from pipe!


unbuffered = direct handshake
buffered = indirect handshake

So with `unbuffered` channel, send and receive must happen at the same time! (like handshake)

Goroutine: "I have fucking status, who want to receive it?" --> Waiting

Main: "Ok, i want to receive fucking status" --> Waiting

When both ready --> Transfer happens
Both waiting for each other --> Deadlock!
hmm, reflection with real example rock!

Buffered = indirect handshake.

Goroutine: "I have fucking status" --> Put into pipe (buffer) (if pipe not full) --> No need to wait for fucking receiver!

Main: Can be late, take `status` from pipe (buffer) (if pipe not empty) --> No need to wait for fucking sender!

Buffered channel doesn't make receiver receive faster. It just make sender not wait for receiver!

Both buffered and unbuffered have same speed for receive, the different is sender(goroutine), not receiver(main)

```go
status := <-results  // Receive from channel
```

This still block of channel empty(have no data)
```
buffered (empty): status := <-results // Block, waiting for data
unbuffered (empty): status := <-results // Block, waiting for data
```


Hmm, conclusion, my understanding about goroutines and channel in Vietnamese
```
Vậy chốt lại theo ý hiểu của mình, chúng ta đã biết có bao nhiêu host, thì tạo 1 cái channel có buffer, nghĩa là tạo 1 channel có số buffer bằng với tổng số host, len(hosts). 
Rồi từ đó cho chạy vào goroutines, mỗi go routines được spawn ra thì sẽ được đẩy vào channel , rồi từ đó thằng main receiver  []HostStatus (vì khi đã lấy đủ data thì tất cả sẽ được collect về thằng này để biết được là các host kia có ssh ok ko, status ra sao), còn như bro nói thì 1 host check mất 10s, 30 host thì nó sẽ spawn ra đủ 30 cái goroutine và chúng ta quất 1 cái channel có buffer size là 30 thì cũng chỉ mất 10s cho 30 host.
```

AI Review:
- 30 hosts → 30 goroutines spawn CÙNG LÚC
- Channel buffer size 30 → tất cả có thể send mà KHÔNG CẦN CHỜ
- Công việc parallel → ~10s cho 30 hosts (thay vì 300s sequential)

Still not finish for story 2.2 xD

So order is randm in goroutines, because which goroutine is finish first, it will print first.
```bash
go run . -c sample/live.yaml
lazy-hole v0.1.0
Loaded 10 hosts from sample/live.yaml

Testing SSH connections... >.>
mysql-node-4: Connected!
mysql-node-1: Connected!
mysql-node-10: Connected!
mysql-node-3: Connected!
mysql-node-9: Connected!
mysql-node-8: Connected!
mysql-node-2: Connected!
mysql-node-6: Connected!
mysql-node-7: Connected!
mysql-node-5: Connected!

⏱️ Total time elapsed for testing all hosts: 173.456792ms
```

So why it is not follow by order?
```go
for i := 0; i < len(hosts); i++ {
    status := <-results  // get result if AVAILABLE first.
}
```

Channel is FIFO but goroutines is not finish in order of spawn.

There is still one thing about closure bug in go. But i'm not able to understand at this time, you know too much shit on Goroutines and channel, that already made my brain blow up! 
Copied ...

Bug version: host = last value of loop
```go
for _, host := range hosts {
    go func() {
        connectSSH(host)  // what is host?
    }()
}
```

progress
```
Loop iteration 1: host = A → goroutine 1 scheduled (haven't run)
Loop iteration 2: host = B → goroutine 2 scheduled (haven't run)
Loop iteration 3: host = C → goroutine 3 scheduled (haven't run)
Loop done: host = C

Goroutine 1 runs: connectSSH(host) → host = C (false!)
Goroutine 2 runs: connectSSH(host) → host = C (false!)
Goroutine 3 runs: connectSSH(host) → host = C (true somehow!)
```

Fix version:
```go
for _, host := range hosts {
    go func(h Host) {
        connectSSH(h)  // h là COPY của host tại thời điểm gọi
    }(host)  // pass host value ngay lập tức
}
```

progress why: when run loop, it will create new variable h for each iteration, so each goroutine will have its own copy of host, and won't be affected by other goroutines

```
Loop iteration 1: host = A → goroutine 1 gets h = A (copied!)
Loop iteration 2: host = B → goroutine 2 gets h = B (copied!)
Loop iteration 3: host = C → goroutine 3 gets h = C (copied!)
```

TLDR: pass param: copy value at spawn moment --> each goroutine have correct data

## Story 2.3, check sudo access (1 point)

Nothing much to explain, check if user has sudo. If not, print error and exit.
Goal: After ssh connect, check host has sudo access or not.

Result:
```bash
go run . -c sample/live.yaml
lazy-hole v0.1.0
Loaded 10 hosts from sample/live.yaml

Testing SSH connections... >.>
mysql-node-8: Connected!
  root: Sudo access OK!
mysql-node-1: Connected!
  ubuntu: Sudo access OK!
mysql-node-3: Connected!
  kienlt: Sudo access NOT OK!
mysql-node-2: Connected!
  kienlt: Sudo access NOT OK!
mysql-node-9: Connected!
  root: Sudo access OK!
mysql-node-7: Connected!
  root: Sudo access OK!
mysql-node-4: Connected!
  root: Sudo access OK!
mysql-node-5: Connected!
  root: Sudo access OK!
mysql-node-10: Connected!
  root: Sudo access OK!
mysql-node-6: Connected!
  root: Sudo access OK!

⏱️ Total time elapsed for testing all hosts: 802.7655ms
```

And yeah, I almost forgot about `defer` in `func checkSudo`
```go
func checkSudo(client *ssh.Client) bool {
	session, err := client.NewSession() // 1. Create Session
	if err != nil {
		return false
	}

	// Close session after check done
	defer session.Close() // 2. Register defer (not run yet)

	// sudo -n = non-interactive, fails if password required
	err = session.Run("sudo -n true") // 3. Run command
	return err == nil // true if sudo works! 
    // func return. After that, defer run
}
```

So defer is not run when function return, it run after function return to make sure we have closed session!


## Story 2.4: Run command on remote host (2 points)
Goal: function to run any command in remote host, capture output and return it.

New thing learn:
- `bytes.Buffer` : capture output from remote host into memory
- `session.Stdout/Stderr`: redirect output of command
- Type assertion: `err.(*ssh.ExitError)` - check error type to get exit code

Funny when i tried to use closed connection to run command xD
```
  mysql-node-1: Failed to run command: failed to create session: read tcp 1.1.3.4:12345->host-ip-here:22: use of closed network connection
```

Little explain for `bytes.buffer`
```
Remote Host                   Your Program (Go)
    │                              │
    │  SSH Session                 │
    │◄────────────────────────────►│
    │                              │
    │  Run "hostname"              │
    │─────────────────────────────►│
    │                              │
    │  Output: "mysql-node-1"      │
    │◄─────────────────────────────│
    │                              │
    │           ┌─────────────┐    │
    │           │ stdout      │    │  ← bytes.Buffer contain output
    │           │ "mysql..."  │    │
    │           └─────────────┘    │
```

So another for `err.(*ssh.ExitError)`.

When `session.Run()` fail, it will return `error`
```
type error interface {
	Error() string
}
```
As you can see, there is no fucking exit code here!

So we depend on SSH package solution. SSH package defined `*ssh.ExitError` - which contain exit code of remote command.
```go
type ExitError struct {
    // ... fields
}

func (e *ExitError) ExitStatus() int  // ← Exit code!
```

type assertion - check and convert
```go
if exitErr, ok := err.(*ssh.ExitError); ok {
    // ok = true: err REALLY is *ssh.ExitError
    // exitErr = converted value, can call ExitStatus()
    exitCode = exitErr.ExitStatus()
}
```

example:
```go
err = session.Run("exit 42")  // Command exit with code 42

// err with type is error, but REALLY is *ssh.ExitError
if exitErr, ok := err.(*ssh.ExitError); ok {
    fmt.Println(exitErr.ExitStatus())  // 42
}
```

TL;DR: Type assertion: check this fucking error to see it is special or not, if not get info from it which is exit code.

I understand those shit about 20%, but good to go LOL


## Epic 3. Core command (MVP)

## Story 3.1: Add blackhole route

Goal: add blackhole route to remote host.

Progress: everything is pretty basic, nothing advanced like Epic 2!
Validate IR / CIDR, because `ip route add blackhole <ip>` will fail if not valid IP or CIDR.

After validate, used function `runCommand(client, "command-here")` to run command on remote host.

Handled blackroute exists by checking output of command, if contains `File exists` then return error. 

Example:
```bash
kienlt@Luongs-MacBook-Pro lazy-hole % go run . -c sample/single.yaml
lazy-hole v0.1.0
Loaded 1 hosts from sample/single.yaml

Testing SSH connections... >.>
mysql-node-1: Connected!
Blackhole added successfully for mysql-node-1

⏱️ Total time elapsed for testing all hosts: 237.129375ms
kienlt@Luongs-MacBook-Pro lazy-hole % go run . -c sample/single.yaml
lazy-hole v0.1.0
Loaded 1 hosts from sample/single.yaml

Testing SSH connections... >.>
mysql-node-1: Connected!
Blackhole error: Route already exists: 9.9.9.9

⏱️ Total time elapsed for testing all hosts: 507.752375ms
```

### Story 3.2: Remove blackhole route

"Anh đã mơ về, ngôi nhà và những đứa trẻ."

I will choose option read-able, that mean i will create 2 separate function to manage add/remove blackhole
