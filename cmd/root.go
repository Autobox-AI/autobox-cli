package cmd

import (
	"fmt"
	"os"

	"github.com/Autobox-AI/autobox-cli/internal/config"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	verbose bool
	noColor bool
	output  string
)

var rootCmd = &cobra.Command{
	Use:   "autobox",
	Short: "Autobox CLI - Manage AI simulation containers",
	Long: `Autobox CLI is a command-line tool for managing Autobox AI simulations.
	
It provides functionality to launch, monitor, and manage simulation containers
running the Autobox Engine, with support for metrics collection and status tracking.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if noColor {
			color.NoColor = true
		}
		if err := config.Init(); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing config: %v\n", err)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.autobox/autobox.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "table", "output format (table|json|yaml)")

	addCommands()
}

func addCommands() {
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(metricsCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(terminateCmd)
	rootCmd.AddCommand(versionCmd)
}