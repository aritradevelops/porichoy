package cmd

import "github.com/spf13/cobra"

type AdminInfo struct {
	email           string
	name            string
	password        string
	confirmPassword string
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Porichoy",
	Long:  `Initialize Porichoy`,
	Run: func(cmd *cobra.Command, args []string) {
		adminInfo := AdminInfo{}

	},
}
