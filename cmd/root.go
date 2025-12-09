package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"generative-cli/generated/commands"
)

var rootCmd = &cobra.Command{
	Use:   "generative-cli",
	Short: "CLI generator for Scalingo SDK",
	Long:  `A CLI tool that introspects the go-scalingo SDK and generates Cobra commands for all available endpoints.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(generateCmd)

	// Register all generated commands
	commands.RegisterAll(rootCmd)
}
