# Day 2: Feb 4, 2026

### Story 1.2

Add cobra CLI for lazy-hole with --help and --version, it is copy-able from ChgK8sCtx project xD

### Story 1.3

I noticed there is `Fprintln(os.Stderr, err)` in `rootCmd.go`
I start to ask why, and did little research about it, what is different between it and `fmt.Errorf()` in day1.md

So quick report:

`fmt.Errorf()`:
- Purpose: create error to return
- Return: `error` object
- Usage: Inside function

`fmt.Fprintln(os.Stderr, err)`:
- Purpose: print error to stderr
- Return: written bytes
- Usage: In main function to print error. With this define, we can redirect error to file or pipe to another command: `./app > output.txt 2> error.txt` like this. ==> go LFCS course, IO redirection part bro!


Result for story 1.3 after we add new flag, it is just show new flag in command help. LOL

Before
```bash
kienlt@Luongs-MacBook-Pro lazy-hole % go run . --help
lazy-hole - A CLI/TUI tool to simulate network failures for testing distributed systems.

Usage:
  lazy-hole [flags]

Flags:
  -h, --help      help for lazy-hole
  -v, --version   version for lazy-hole
```

After
```bash
kienlt@Luongs-MacBook-Pro lazy-hole % go run . --help
lazy-hole - A CLI/TUI tool to simulate network failures for testing distributed systems.

Usage:
  lazy-hole [flags]

Flags:
  -c, --config string   Path to config file (default "sample/hosts.yaml")
  -h, --help            help for lazy-hole
  -v, --version         version for lazy-hole
  ```
```

And with default config:
```bash
go run .
invalid IP address: 10.0.0.999
exit status 1
```

Some other cases
```bash
# ====== NON EXISTS =======
go run . -c /nonexistent.yaml
failed to read config file: open /nonexistent.yaml: no such file or directory
exit status 1
# ========= Short flag ======
go run . -c sample/hosts.yaml
invalid IP address: 10.0.0.999
exit status 1
# ========= Full flag ======
go run . --config sample/hosts.yaml
invalid IP address: 10.0.0.999
exit status 1
```