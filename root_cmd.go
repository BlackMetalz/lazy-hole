package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time" // For calculate time

	"github.com/spf13/cobra"
)

var (
	version    = "v0.1.0"
	configPath string
)

var rootCmd = &cobra.Command{
	Use:     "lazy-hole",
	Short:   "Network chaos engineering tool",
	Long:    `lazy-hole - A CLI/TUI tool to simulate network failures for testing distributed systems.`,
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {

		// Load Config
		config, err := LoadConfig(configPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		// TODO: Start TUI here (future)
		// fmt.Println("TUI mode coming soon...")
		// Already done in begin of EPIC 4 xD

		fmt.Println("lazy-hole " + version)
		fmt.Printf("Loaded %d hosts from %s\n", len(config.Hosts), configPath)

		// Test SSH Connection to each host. (Sequential)
		// for _, host := range config.Hosts {
		// 	client, err := connectSSH(host)
		// 	if err != nil {
		// 		fmt.Printf("Failed to connect to %s via port %d: %v\n", host.Name, host.SSH_Port, err)
		// 		continue
		// 	}
		// 	fmt.Printf("Successfully connected to %s\n", host.Name)
		// 	defer client.Close()
		// }

		// Test SSH Connection to each host (Parallel)
		fmt.Println("\nTesting SSH connections... >.>")
		startTime := time.Now() // start counting time
		statuses := testAllHosts(config.Hosts)

		// Setup cleanup handler after we have connections
		setupCleanUp(statuses)

		timeElapsed := time.Since(startTime) // end counting time

		for _, status := range statuses {
			// Check if connected is true
			if status.Connected {
				fmt.Printf("%s: Connected!\n", status.Host.Name)
				if status.Sudo {

					// Test sudo access
					// fmt.Printf("%s: Sudo access OK!\n", status.Host.User)

					// Test hostnamectl command
					/*
						result, err := runCommand(status.Client, "hostnamectl")
						if err != nil {
							fmt.Printf("  %s: Failed to run command: %v\n", status.Host.Name, err)
						} else {
							fmt.Printf("  %s: %s\n", status.Host.Name, result.Stdout)
						}
					*/

					// Test blackhole
					/*
						err = addBlackHole(status.Client, "9.9.9.9")
						if err != nil {
							fmt.Printf("Blackhole error: %v\n", err)
						} else {
							fmt.Printf("Blackhole added successfully for %s\n", status.Host.Name)
						}
					*/

					// Test list interface
					/*
						interfaces, err := listInterfaces(status.Client)
						if err != nil {
							fmt.Printf("Error while list interfaces: %v\n", err)
						} else {
							fmt.Printf("Interfaces: %v\n", interfaces)
						}
					*/

					// Test add latency
					/*
						interfaces, _ := listInterfaces(status.Client)
						if len(interfaces) > 0 {
							// Example we have eth0 and lo interface only
							// So eth0 interface will be used
							err := addLatency(status.Client, status.Host.Name, interfaces[0], "100ms")
							if err != nil {
								fmt.Printf("latency increase error: %v\n", err)
							} else {
								fmt.Printf("Added 100ms latency to %s\n", interfaces[0])
							}
						}
					*/

					// Close connection here
					// status.Client.Close() // Temp comment for Story 4.4
				} else {
					fmt.Printf("%s: Sudo access NOT OK!\n", status.Host.User)
				}
			} else {
				fmt.Printf("%s: Failed. Issue %v\n", status.Host.Name, status.Error)
			}
		}

		fmt.Printf("\nTotal time elapsed for testing all hosts: %s\n", timeElapsed)

		// Start TUI when run
		tui := NewTUI(statuses)
		if err := tui.Run(); err != nil {
			fmt.Printf("TUI Error: %v\n", err)
		}

		// fmt.Println("DEBUG: TUI exited!")
		// fmt.Printf("DEBUG: effects count = %d\n", len(effectTracker.GetAll()))

		// Cleanup after TUI exit (ESC/q)
		// Use latest statuses from TUI (may have been refreshed!)
		currentStatuses := tui.GetStatuses()

		// Count actual effects (not just hosts in map)
		totalEffects := 0
		for _, effects := range effectTracker.GetAll() {
			totalEffects += len(effects)
		}
		if totalEffects > 0 {
			fmt.Println("\nCleaning up effects...")
			restoreAll(currentStatuses)
		}

		// Close all SSH connections after exit!
		for _, status := range currentStatuses {
			if status.Client != nil {
				status.Client.Close()
			}
		}
	},
}

// Add flag for config path
func init() {
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "sample/hosts.yaml", "Path to config file")
}

// setupCleanUp will setup clean up function for root command
func setupCleanUp(hostStatuses []HostStatus) {
	// Create channel to receive signal
	c := make(chan os.Signal, 1) // buffer size

	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c // Wait for Ctrl+C or kill signal
		fmt.Println("\nReceived interrupt signal, cleaning up!")

		// Count actual effects (not just hosts in map)
		totalEffects := 0
		for _, effects := range effectTracker.GetAll() {
			totalEffects += len(effects)
		}
		if totalEffects > 0 {
			restoreAll(hostStatuses)
		}

		// Close all SSH connections
		for _, status := range hostStatuses {
			if status.Client != nil {
				status.Client.Close()
			}
		}

		fmt.Println("Bye!")
		os.Exit(0)
	}()
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}
