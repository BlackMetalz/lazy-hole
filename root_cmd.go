package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "0.1.0"

var rootCmd = &cobra.Command{
	Use:     "lazy-hole",
	Short:   "Network chaos engineering tool",
	Long:    `lazy-hole - A CLI/TUI tool to simulate network failures for testing distributed systems.`,
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Start TUI here (future)
		fmt.Println("lazy-hole v" + version)
		fmt.Println("TUI mode coming soon...")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}
