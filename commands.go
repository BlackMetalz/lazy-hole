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

// List network interfaces (exclude loopback "lo")
func listInterfaces(client *ssh.Client) ([]string, error) {
	// Command to list interfaces
	result, err := runCommand(client, "ls /sys/class/net/")
	if err != nil {
		return nil, fmt.Errorf("Failed to list interfaces: %w", err)
	}

	if result.ExitCode != 0 {
		return nil, fmt.Errorf("Command failed: %s", result.Stderr)
	}

	// parse output - split by whitespace
	raw_interfaces := strings.Fields(result.Stdout)

	// filter out "lo"
	var interfaces []string
	for _, iface := range raw_interfaces {
		if iface != "lo" {
			interfaces = append(interfaces, iface)
		}
	}

	return interfaces, nil
}

// Add latency to interface with tc command
func addLatency(client *ssh.Client, iface, delay string) error {
	// Validate delay format (e.g., "100ms", "50ms")
	if !strings.HasSuffix(delay, "ms") {
		return fmt.Errorf("invalid delay format: %s (use e.g., '100ms')", delay)
	}

	// Create command
	cmd := fmt.Sprintf("sudo tc qdisc add dev %s root netem delay %s", iface, delay)

	result, err := runCommand(client, cmd)
	if err != nil {
		return fmt.Errorf("Failed to add latency: %w", err)
	}

	// handle Error: Exclusivity flag on, cannot modify.
	if result.ExitCode != 0 {
		if strings.Contains(result.Stderr, "Exclusivity flag on, cannot modify.") {
			// try change instead of add
			cmd = fmt.Sprintf("sudo tc qdisc change dev %s root netem delay %s", iface, delay)

			result, err = runCommand(client, cmd)
			if err != nil {
				return fmt.Errorf("Failed to add latency: %w", err)
			}

			if result.ExitCode != 0 {
				return fmt.Errorf("Change tc qdisc failed: %s", result.Stderr)
			}
		} else {
			return fmt.Errorf("Command failed: %s", result.Stderr)
		}
	}

	return nil
}

func removeTCRules(client *ssh.Client, iface string) error {
	cmd := fmt.Sprintf("sudo tc qdisc del dev %s root", iface)

	result, err := runCommand(client, cmd)
	if err != nil {
		return fmt.Errorf("Failed to remove latency: %w", err)
	}

	if result.ExitCode != 0 {
		if strings.Contains(result.Stderr, "Cannot delete qdisc with handle of zero.") {
			return fmt.Errorf("no tc rules on %s", iface)
		}

		return fmt.Errorf("command failed: %s", result.Stderr)
	}

	return nil
}

func addPacketLoss(client *ssh.Client, iface, percent string) error {
	// validate percent (0-100)
	if !strings.HasSuffix(percent, "%") {
		return fmt.Errorf("invalid format: %s (use e.g., '10%')", percent)
	}

	cmd := fmt.Sprintf("sudo tc qdisc add dev %s root netem loss %s", iface, percent)

	result, err := runCommand(client, cmd)
	if err != nil {
		return fmt.Errorf("failed to add packet loss: %w", err)
	}

	if result.ExitCode != 0 {
		if strings.Contains(result.Stderr, "Exclusivity flag on, cannot modify") {
			cmd = fmt.Sprintf("sudo tc qdisc change dev %s root netem loss %s", iface, percent)
			result, err = runCommand(client, cmd)
			if err != nil || result.ExitCode != 0 {
				return fmt.Errorf("Change failed: %s", result.Stderr)
			}
		} else {
			return fmt.Errorf("Command failed: %s", result.Stderr)
		}
	}

	return nil
}
