package main

import "golang.org/x/crypto/ssh"

type Host struct {
	Name     string `yaml:"name"`
	IP       string `yaml:"ip"`
	User     string `yaml:"ssh_user"`
	SSH_Port int    `yaml:"ssh_port"`
	SSH_Key  string `yaml:"ssh_key"`
	Group    string `yaml:"group,omitempty"`
}

// wrapper struct for config file
type Config struct {
	Hosts []Host `yaml:"hosts"`
}

// Host Status store connection result for each host
type HostStatus struct {
	Host         Host
	Connected    bool
	Error        error
	Client       *ssh.Client
	Sudo         bool   // sudo access check
	SSH_SourceIP string // SSH source IP
}

// CommandResult, stores output of remote command
type CommandResult struct {
	Stdout   string
	Stderr   string
	ExitCode int // Exit code of remote command
}
