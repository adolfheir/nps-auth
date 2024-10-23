package cmd

import (
	"nps-auth/internal/server"

	"github.com/spf13/cobra"
)

var ServerCmd = &cobra.Command{
	Use: "server",
	RunE: func(cmd *cobra.Command, args []string) error {

		server.Init()

		return nil
	},
}
