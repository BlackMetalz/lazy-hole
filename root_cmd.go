package main

import (
	"fmt"
	"os"
	"time" // For calculate time

	"github.com/spf13/cobra"
)

var (
	version    = "0.1.0"
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
		fmt.Println("lazy-hole v" + version)
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
					interfaces, err := listInterfaces(status.Client)
					if err != nil {
						fmt.Printf("Error while list interfaces: %v\n", err)
					} else {
						fmt.Printf("Interfaces: %v\n", interfaces)
					}

					// Close connection here
					status.Client.Close()
				} else {
					fmt.Printf("%s: Sudo access NOT OK!\n", status.Host.User)
				}
			} else {
				fmt.Printf("%s: Failed. Issue %v\n", status.Host.Name, status.Error)
			}
		}

		fmt.Printf("\n⏱️ Total time elapsed for testing all hosts: %s\n", timeElapsed)

	},
}

// Add flag for config path
func init() {
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "sample/hosts.yaml", "Path to config file")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}
