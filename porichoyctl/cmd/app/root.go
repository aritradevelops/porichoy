package app

import (
	"github.com/spf13/cobra"
)

var appCmd = &cobra.Command{
	Use:   "app",
	Short: "App management",
	Long:  `App management`,
}

func NewCmd() *cobra.Command {
	appCmd.AddCommand(appAddCmd)
	return appCmd
}
