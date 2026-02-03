package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")

	// newHost := Host{
	// 	Name:     "mysql-node-1",
	// 	IP:       "10.0.0.5",
	// 	User:     "kienlt",
	// 	SSH_Port: 22,
	// 	SSH_Key:  "~/.ssh/id_rsa",
	// }

	// fmt.Printf("%+v\n", newHost)

	config, err := LoadConfig("sample/hosts.yaml")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", config)
}
