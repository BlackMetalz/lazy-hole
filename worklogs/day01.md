# Day 1: Feb 2, 2026 to Feb 3, 2026

### Feb 2, 2026
Hmm, another shit going down. I'm gonna Agile method to build this project. Because it helps me break tasks into small chunks and focus on one thing at a time.

I think this project i will not use flat structure, i would like to use package?. 
```
go mod init github.com/blackmetalz/lazy-hole
go mod tidy
```

First, create main.go and types.go, ofc with some hello world to main.go. And yeah, to prevent classic: `go run main.go`, we need to run `go run .` because we not only using `main.go` and other go files, ex: `types.go`

First output:
```bash
go run .
Hello, World!
{Name:mysql-node-1 IP:10.0.0.5 User:kienlt SSH_Port:22 SSH_Key:~/.ssh/id_rsa}
```

But that is data created manually, let's use yaml to read it from file @sample/hosts.yaml

oke, implement LoadConfig function in config.go, it is simple to understand, but there is 1 thing i would notice again. It is fucking pointer `config := &Config{}`. Simply you just need to understand that yaml.Unmarshal required pointer to write data into it!

Hmm, not really that hard, i was too panicked. LOL. It is simple. I made some fully comment in config.go. Go read it if you curious!

Also here is Go playground example!
```go
package main

import (
	"encoding/json"
	"fmt"
)

type Config struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

func main() {
	data := []byte(`{"name":"lazy-hole","port":8080}`)

	// Case 1: Pass non-pointer → error
	var config1 Config
	err := json.Unmarshal(data, config1) // non-pointer!
	fmt.Println("Case 1 (non-pointer):")
	fmt.Println("  Error:", err)
	fmt.Println("  Config:", config1)

	fmt.Println()

	// Case 2: Pass pointer → works
	var config2 Config
	err = json.Unmarshal(data, &config2)
	fmt.Println("Case 2 (pointer):")
	fmt.Println("  Error:", err)
	fmt.Println("  Config:", config2)
}
```

Expected output:
```
prog.go:18:23: call of Unmarshal passes non-pointer as second argument

Go vet failed.

Case 1 (non-pointer):
  Error: json: Unmarshal(non-pointer main.Config)
  Config: { 0}

Case 2 (pointer):
  Error: <nil>
  Config: {lazy-hole 8080}
```

### Feb 3, 2026

I think i could have article for pointer in go alone, because it is really confusing for me. (Not really after 6 days since confusing xD)

But that will be for weekend, not today. Just save 2 little example in here for later usage

- https://go.dev/play/p/OeDBjdwDunD
- https://go.dev/play/p/1RqhrhNpbTu

So continue from Feb 2,2026.

```go
config := &Config{} 
```

1. `Config{}` : create new struct value (zero initialized)
2. `&` --> get address of struct value (`Config`)
3. `config` is pointer that pointed to that struct

When pass `config` into `Unmarshal(data, config)`
1. `Unmarshal` recieved address of `config`
2. It deference `(*config)` to get access and modify original data
3. It return original data that fill from YAML

So for Go idiom remember: when you need to `modify` input data --> use Pointer. Only read --> use value (copy)

I guess that is 30% of pointer in Go, but it is enough for me to move on.
Updated Feb 26, 2026: Holy fucking shit, at that time I'm so fucking bad LOLLL.

Next: validate IP and set default SSH_port

set default ssh port: just loop through config.Hosts and check if SSH_Port is 0, if yes, set it to 22. Why check for 0? because it is default value of int

For validate IP: use `net.ParseIP(host.IP)`

And yeah, another interesting thing is `fmt.Errorf`

First time I thought why Gemini recommend return 2 values
```go
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
```

I thought it was just for match 2 value return of func `func LoadConfig(path string) (*Config, error)` but no, it is not. It is for "Wrapping"
Let me copy an example:

Without Wrap:
```go
return nil, err
// Error: "open /invalid/path: no such file or directory"
// Where is the fucking error? where the fuck function call this shit?
```

With Wrap:
```go
return nil, fmt.Errorf("failed to read config file: %w", err)
// Error: "failed to read config file: open /invalid/path: no such file or directory"
// → Holy fucking shit: LoadConfig fail after read config file
```

Conclustion for `fmt.Errorf` and `%w` (wrap - keep chain, only use when caller want to check what the fuck is original error.): it basically wrap error together, add new context but still keep original error inside, so later we can get it by `errors.Unwrap` or check by `errors.Is` / `errors.As`. 

I will try to remember about return nil and fmt Errorf wrap. I think Gemini recommend me from ChgK8sCtx pet project. I get it now. Easy to debug when project goes big.

Update Feb 26, 2026: tried. Here is some example for it (thanks to Gemini generated xD)
```go
package main

import (
	"errors"
	"fmt"
	"os"
)

func readConfig(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		// Wrap error: thêm context nhưng giữ error gốc
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	return data, nil
}

func loadApp(path string) error {
	_, err := readConfig(path)
	if err != nil {
		// Wrap thêm 1 layer nữa
		return fmt.Errorf("failed to load app: %w", err)
	}
	return nil
}

func main() {
	err := loadApp("/not/exist/config.yaml")
	if err == nil {
		return
	}

	// 1. Print full error chain — thấy toàn bộ context
	fmt.Println("Full error:")
	fmt.Println(" ", err)

	fmt.Println()

	// 2. Unwrap từng layer
	fmt.Println("Unwrap 1 layer:")
	fmt.Println(" ", errors.Unwrap(err))

	fmt.Println()

	fmt.Println("Unwrap 2 layers (original error):")
	fmt.Println(" ", errors.Unwrap(errors.Unwrap(err)))

	fmt.Println()

	// 3. errors.Is — check error gốc xuyên qua mọi layer
	if errors.Is(err, os.ErrNotExist) {
		fmt.Println("errors.Is(err, os.ErrNotExist) = true")
		fmt.Println("  → Biết file không tồn tại, dù đã wrap 2 lần!")
	}

	fmt.Println()

	// 4. Bonus: nếu dùng %v thay vì %w thì sao?
	original := fmt.Errorf("original error")
	wrappedW := fmt.Errorf("with %%w: %w", original)
	wrappedV := fmt.Errorf("with %%v: %v", original)

	fmt.Println("%%w vs %%v:")
	fmt.Printf("  errors.Is(wrappedW, original) = %v\n", errors.Is(wrappedW, original))
	fmt.Printf("  errors.Is(wrappedV, original) = %v\n", errors.Is(wrappedV, original))
	fmt.Println("  → %v chỉ convert thành string, mất link tới error gốc!")
}
```

ok. That's all for Story 1.1

==================

