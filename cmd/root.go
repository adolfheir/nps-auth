package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "nps-auth",
}

func init() {
	rootCmd.AddCommand(ServerCmd)
	rootCmd.AddCommand(ClientCmd)
}

// Execute : apply commands
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
