package generatorcli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// sdkPath is shared between generate and update-manifest commands
var sdkPath string

// NewRootCommand builds the root command for the generator CLI.
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "generative-cli",
		Short: "CLI generator for Scalingo SDK",
		Long:  `A CLI tool that introspects the go-scalingo SDK and generates Cobra commands for all available endpoints.`,
	}

	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(updateManifestCmd)

	return rootCmd
}

// Execute runs the generator CLI.
func Execute() {
	if err := NewRootCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
