package cmd

import (
	"fmt"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/spf13/cobra"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Manage porichoyctl configuration",
}
var configureSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		Config.Set(args[0], args[1])
		fmt.Printf("here")
		if err := writeConfig(); err != nil {
			return err
		}

		cmd.Printf("✔ %s set to %s\n", args[0], args[1])
		return nil
	},
}

var configureGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if !Config.Exists(args[0]) {
			cmd.Println("not set")
			return
		}

		cmd.Println(Config.String(args[0]))
	},
}

var configureViewCmd = &cobra.Command{
	Use:   "view",
	Short: "View full configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		out, err := Config.Marshal(yaml.Parser())
		if err != nil {
			return err
		}

		cmd.Println(string(out))
		return nil
	},
}

func init() {
	configureCmd.AddCommand(configureSetCmd)
	configureCmd.AddCommand(configureGetCmd)
	configureCmd.AddCommand(configureViewCmd)
}
