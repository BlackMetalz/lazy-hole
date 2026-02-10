package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// Test all hosts connection in PARALLEL
func testAllHosts(hosts []Host) []HostStatus {
	// Yes, I understand this shit.
	// Input is slice of Host, that loads all host from config
	// Output is slide of Host Status, that contains connection result for each host

	// create channel to collect results
	// good shit, count total host to make buffered channel xD
	results := make(chan HostStatus, len(hosts))

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

	// Collect all results
	var statuses []HostStatus
	for i := 0; i < len(hosts); i++ {
		status := <-results // Receive from channel
		statuses = append(statuses, status)
		// status.Client.Close() // Close connection after test done!
		// Temp comment this out to avoid closing connection before using it
	}

	return statuses

}

// Check sudo access
func checkSudo(client *ssh.Client) bool {
	session, err := client.NewSession()
	if err != nil {
		return false
	}

	// Close session after check done
	defer session.Close()

	// sudo -n = non-interactive, fails if password required
	err = session.Run("sudo -n true")
	return err == nil // true if sudo works!
}

// Run command on remote host
func runCommand(client *ssh.Client, cmd string) (CommandResult, error) {
	session, err := client.NewSession() // Init new session
	if err != nil {
		return CommandResult{}, fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close() // Close session after run done

	// capture stdout and stderr
	// Init 2 empty buffer in memory
	var stdout, stderr bytes.Buffer

	// Tell the fucking session that send output to my fucking buffer

	// stdout of remote --> buffer stdout
	session.Stdout = &stdout // Redirect stdout to buffer

	// stderr of remote --> buffer stderr
	session.Stderr = &stderr // Redirect stderr to buffer

	// Run command
	err = session.Run(cmd) // Output of remote command will be written to stdout and stderr buffer

	// Get exit code from error
	exitCode := 0 // Default exit code
	if err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			exitCode = exitErr.ExitStatus()
			err = nil // run command, just non-zero exit code, not error
		}
	}

	return CommandResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}, err
}

func detectSSHSourceIP(client *ssh.Client) (string, error) {
	// Get self ip!
	result, err := runCommand(client, "echo $SSH_CLIENT | awk '{print $1}'")

	if err != nil {
		return "", err
	}
	return strings.TrimSpace(result.Stdout), nil
}
