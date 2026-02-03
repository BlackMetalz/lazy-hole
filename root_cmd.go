package main

import (
	"fmt"
	"os"

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
