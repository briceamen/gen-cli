package runtimecli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"generative-cli/generated/commands"
)

// NewRootCommand builds the root command for the generated CLI.
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:          "scalingo-gen",
		Short:        "Generated Scalingo CLI",
		Long:         "A generated CLI built from go-scalingo methods using the manifest as source of truth.",
		SilenceUsage: true,
	}

	commands.RegisterAll(rootCmd)

	return rootCmd
}

// Execute runs the generated CLI.
func Execute() {
	if err := NewRootCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
