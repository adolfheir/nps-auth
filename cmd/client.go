package cmd

import (
	"nps-auth/internal/client"

	"github.com/spf13/cobra"
)

var ClientCmd = &cobra.Command{
	Use: "client",
	RunE: func(cmd *cobra.Command, args []string) error {

		client.Init()

		return nil
	},
}
