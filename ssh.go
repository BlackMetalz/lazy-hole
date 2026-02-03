package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/ssh"
)

// Default key path to try in order
// Support only id_rsa and id_ed25519 right now because I only understand these two xD
var defaultKeyPaths = []string{
	".ssh/id_rsa",
	".ssh/id_ed25519",
}

func connectSSH(host Host) (*ssh.Client, error) {
	var keyPath string
	homeDir := os.Getenv("HOME")

	// if ssh_key specified in config, use it
	if host.SSH_Key != "" {
		keyPath = host.SSH_Key
		// Need expand ~ to homedir
		// Because this shit wont work: os.ReadFile("~/.ssh/id_ed25519")
		if len(keyPath) >= 2 && keyPath[:2] == "~/" {
			keyPath = filepath.Join(homeDir, keyPath[2:])
		}

	} else {
		// try default keys
		for _, p := range defaultKeyPaths {
			fullPath := filepath.Join(homeDir, p)
			if _, err := os.Stat(fullPath); err == nil {
				keyPath = fullPath
				break
			}
		}
	}

	if keyPath == "" {
		return nil, fmt.Errorf("No SSH key found when make connectSSH")
	}

	// Read private key
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	// Parse private key
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	// Create SSH config
	sshConfig := &ssh.ClientConfig{
		User: host.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	// Connect
	address := fmt.Sprintf("%s:%d", host.IP, host.SSH_Port)
	client, err := ssh.Dial("tcp", address, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", host.Name, err)
	}

	return client, nil

}
