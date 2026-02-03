package main

type Host struct {
	Name     string `yaml:"name"`
	IP       string `yaml:"ip"`
	User     string `yaml:"ssh_user"`
	SSH_Port int    `yaml:"ssh_port"`
	SSH_Key  string `yaml:"ssh_key"`
}

// wrapper struct for config file
type Config struct {
	Hosts []Host `yaml:"hosts"`
}
