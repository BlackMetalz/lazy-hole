package main

import (
	"fmt"
	"net"
	"strings"

	"golang.org/x/crypto/ssh"
)

// Validate IP or CIDR format

func isValidIPOrCIDR(input string) bool {
	// try as IP
	if net.ParseIP(input) != nil {
		return true
	}

	// try as CIDR
	_, _, err := net.ParseCIDR(input)
	return err == nil
}

// Add blackhole route
func addBlackHole(client *ssh.Client, target string) error {
	// Validate Input
	if !isValidIPOrCIDR(target) {
		return fmt.Errorf("Invalid IP or CIDR: %s", target)
	}

	// Command to add blackhole route
	cmd := fmt.Sprintf("sudo ip route add blackhole %s", target)

	result, err := runCommand(client, cmd)
	if err != nil {
		return fmt.Errorf("Failed to add blackhole route: %w", err)
	}

	// Check for route already exists
	// RTNETLINK answers: File exists
	if result.ExitCode != 0 {
		if strings.Contains(result.Stderr, "File exists") {
			return fmt.Errorf("Route already exists: %s", target)
		}

		// return other error
		return fmt.Errorf("Command failed: %s", result.Stderr)
	}

	return nil
}

// Remove blackhole route
func removeBlackHole(client *ssh.Client, target string) error {
	// Validate Input
	if !isValidIPOrCIDR(target) {
		return fmt.Errorf("Invalid IP or CIDR: %s", target)
	}

	// Command to remove blackhole route
	cmd := fmt.Sprintf("sudo ip route del blackhole %s", target)

	result, err := runCommand(client, cmd)
	if err != nil {
		return fmt.Errorf("Failed to remove blackhole route: %w", err)
	}

	// Check for route not exists
	// RTNETLINK answers: No such process
	if result.ExitCode != 0 {
		if strings.Contains(result.Stderr, "No such process") {
			return fmt.Errorf("Route not exists: %s", target)
		}

		// return other error
		return fmt.Errorf("Command failed: %s", result.Stderr)
	}

	return nil
}
