/*
Copyright © 2025
*/
package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/aritradeveops/porichoy/porichoyctl/cmd/app"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	appName         = "porichoyctl"
	defaultRootHost = "http://localhost:8080"
	delim           = "."
	envPrefix       = "PORICHOY_"
)

var (
	cfgFile string
	Config  *koanf.Koanf
)

var rootCmd = &cobra.Command{
	Use:   appName,
	Short: "Configure and Manage Porichoy Deployments",
	Long:  `Porichoyctl is a CLI tool for managing and configuring Porichoy deployments.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	Config = koanf.New(delim)

	// Register config loader
	cobra.OnInitialize(initConfig)

	// Persistent flags (available to all commands)
	rootCmd.PersistentFlags().StringVar(
		&cfgFile,
		"config",
		"",
		"config file (default is $HOME/."+appName+".yaml)",
	)

	rootCmd.PersistentFlags().String(
		"host",
		defaultRootHost,
		"Porichoy API host",
	)

	// Commands
	rootCmd.AddCommand(configureCmd)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(app.NewCmd())
}

func initConfig() {
	loadConfigFlag()
	loadConfigFile()
	loadEnv()
	loadFlags(rootCmd)
}

func configFilePath() string {
	if cfgFile != "" {
		return cfgFile
	}

	home, err := homedir.Dir()
	cobra.CheckErr(err)

	return filepath.Join(home, "."+appName+".yaml")
}

func loadConfigFile() {
	path := configFilePath()

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return
	}

	err := Config.Load(file.Provider(path), yaml.Parser())
	cobra.CheckErr(err)
}

func loadEnv() {
	err := Config.Load(
		env.Provider(
			envPrefix,
			delim,
			func(s string) string {
				// PORICHOY_HOST -> host
				key := strings.TrimPrefix(s, envPrefix)
				return strings.ToLower(key)
			},
		),
		nil,
	)
	cobra.CheckErr(err)
}

func loadFlags(cmd *cobra.Command) {
	err := Config.Load(
		posflag.Provider(cmd.Flags(), delim, Config),
		nil,
	)
	cobra.CheckErr(err)
}

func writeConfig() error {
	path := configFilePath()

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	final, err := Config.Marshal(yaml.Parser())
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return os.WriteFile(path, final, 0600)
}
func loadConfigFlag() {
	fs := pflag.NewFlagSet("config", pflag.ContinueOnError)
	fs.StringVar(&cfgFile, "config", "", "")

	// Ignore errors to avoid failing on unknown flags
	_ = fs.Parse(os.Args[1:])
}
