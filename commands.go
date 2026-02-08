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
func addBlackHole(client *ssh.Client, hostname, target string) error {
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
	} else {
		// add track the effect bro!
		effectTracker.Add(hostname, ActiveEffect{
			Type:   EffectBlackHole,
			Target: target,
			Value:  "", // Blackhole doesn't require value.
		})
	}

	return nil
}

// Remove blackhole route
func removeBlackHole(client *ssh.Client, hostname, target string) error {
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
	} else {
		// remove track the effect bro!
		effectTracker.Remove(hostname, ActiveEffect{
			Type:   EffectBlackHole,
			Target: target,
			Value:  "",
		})
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
func addLatency(client *ssh.Client, hostname, iface, delay string) error {
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
	} else {
		// add track latency bro
		effectTracker.Add(hostname, ActiveEffect{
			Type:   EffectLatency,
			Target: iface,
			Value:  delay,
		})
	}

	return nil
}

// This can be understand as removeLatency from addLatency
func removeTCRules(client *ssh.Client, hostname, iface string) error {
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
	} else {
		effectTracker.Remove(hostname, ActiveEffect{
			Type:   EffectLatency,
			Target: iface,
			Value:  "", // remember no need value i guess?
		})
	}

	return nil
}

func addPacketLoss(client *ssh.Client, hostname, iface, percent string) error {
	// validate percent (0-100)
	if !strings.HasSuffix(percent, "%") {
		return fmt.Errorf("invalid format: %s (use e.g., '10%%')", percent)
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
	} else {
		effectTracker.Add(hostname, ActiveEffect{
			Type:   EffectPacketLoss,
			Target: iface,
			Value:  percent,
		})
	}

	return nil
}

// Block incomming traffic from source IP (network partition)
func addPartition(client *ssh.Client, hostname, sourceIP string) error {
	if net.ParseIP(sourceIP) == nil {
		return fmt.Errorf("invalid IP: %s", sourceIP)
	}

	// Check if rule exists first
	checkCmd := fmt.Sprintf("sudo iptables -C INPUT -s %s -j DROP", sourceIP)
	checkResult, err := runCommand(client, checkCmd)
	if checkResult.ExitCode == 0 {
		return fmt.Errorf("partition rule already exists for %s", sourceIP)
	}

	cmd := fmt.Sprintf("sudo iptables -A INPUT -s %s -j DROP", sourceIP)
	result, err := runCommand(client, cmd)
	if err != nil {
		return fmt.Errorf("failed to add partition: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("Command failed: %s", result.Stderr)
	} else {
		effectTracker.Add(hostname, ActiveEffect{
			Type:   EffectPartition,
			Target: sourceIP,
			Value:  "",
		})
	}

	return nil
}

// Remove partition
func removePartition(client *ssh.Client, hostname, sourceIP string) error {
	if net.ParseIP(sourceIP) == nil {
		return fmt.Errorf("invalid IP: %s", sourceIP)
	}

	// -D instead of -A for delete rule!
	cmd := fmt.Sprintf("sudo iptables -D INPUT -s %s -j DROP", sourceIP)
	result, err := runCommand(client, cmd)
	if err != nil {
		return fmt.Errorf("failed to remove partition: %w", err)
	}

	if result.ExitCode != 0 {
		if strings.Contains(result.Stderr, "Bad rule (does a matching rule exist in that chain?)") {
			return fmt.Errorf("no partition rule to remove for %s", sourceIP)
		}
		return fmt.Errorf("command failed: %s", result.Stderr)
	} else {
		effectTracker.Remove(hostname, ActiveEffect{
			Type:   EffectPartition,
			Target: sourceIP,
			Value:  "",
		})
	}

	return nil
}

// Restore single host - remove all effects passed into that fucking host
func restoreHost(client *ssh.Client, hostname string) error {

	effects := effectTracker.Get(hostname)

	if len(effects) == 0 {
		return fmt.Errorf("No fucking effects to restore for %s", hostname)
	}

	for _, effect := range effects {
		var err error
		switch effect.Type {
		case EffectBlackHole:
			err = removeBlackHole(client, hostname, effect.Target)
		case EffectLatency, EffectPacketLoss:
			err = removeTCRules(client, hostname, effect.Target)
		case EffectPartition:
			err = removePartition(client, hostname, effect.Target)
		}

		if err != nil {
			// Log error but still continue xD
			fmt.Printf("Warning: failed to remove %s on %s: %v\n", effect.Type, hostname, err)
		}
	}

	// Clear all tracked effects for this host in memory
	effectTracker.Clear(hostname)
	return nil
}

// Restore all hosts - remove all effects from all hosts
func restoreAll(hostStatuses []HostStatus) error {
	// Get all affect from global tracker
	allEffects := effectTracker.GetAll()

	if len(allEffects) == 0 {
		return fmt.Errorf("No effects to restore, bro!")
	}

	restored := 0
	for hostname := range allEffects {
		// find the client for this host
		var client *ssh.Client
		for _, status := range hostStatuses {
			if status.Host.Name == hostname && status.Connected {
				client = status.Client
				break
			}
		}

		if client == nil {
			fmt.Printf("Warning: can not restore %s (not connected)\n", hostname)
			continue
		}

		err := restoreHost(client, hostname)
		if err != nil {
			fmt.Printf("Warning: failed to restore %s: %v\n", hostname, err)
		} else {
			// increase counter
			restored++
		}
	}

	fmt.Printf("Restored %d hosts \n", restored)
	return nil
}
