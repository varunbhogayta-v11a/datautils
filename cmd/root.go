package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "datautil",
	Short: "A CLI tool for dataset processing",
	Long: `datautil is a powerful CLI tool for processing datasets.
It supports filtering, transforming, validating, and exporting data in various formats.

Supported formats: CSV, JSON, XML, Excel

Examples:
  datautil filter --input data.csv --where "age > 25"
  datautil transform --input data.csv --add "full_name=first+last"
  datautil validate --input data.csv --required name,email`,
	Version: "1.0.0",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")
	rootCmd.PersistentFlags().Bool("verbose", false, "verbose output")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("./configs")
		viper.AddConfigPath("$HOME/.datautil")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
