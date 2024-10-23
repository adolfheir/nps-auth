package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var VersionCmd = &cobra.Command{
	Use:     "version",
	Short:   "Get the version of nps-auth",
	Example: "nps-auth version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf(`nps-auth version: %v`, "v0.0.1")
	},
}
