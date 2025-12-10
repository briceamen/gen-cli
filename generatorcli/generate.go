package generatorcli

import (
	"fmt"

	"github.com/spf13/cobra"

	"generative-cli/generator"
)

var (
	sdkPath    string
	outputPath string
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate CLI commands from manifest",
	Long:  `Generate Cobra commands based on entries in manifest.toml (manifest is treated as read-only).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		manifest, err := generator.LoadManifest("manifest.toml")
		if err != nil {
			return fmt.Errorf("failed to load manifest: %w", err)
		}

		methods := manifest.MethodsToGenerate()
		if len(methods) == 0 {
			fmt.Println("No methods marked for generation in manifest")
			return nil
		}

		fmt.Printf("Generating commands for %d methods across %d services\n", countMethods(methods), len(methods))

		// Generate code
		if err := generator.GenerateCommands(methods, outputPath); err != nil {
			return fmt.Errorf("failed to generate commands: %w", err)
		}

		// Generate spec
		if err := generator.GenerateSpec(methods, outputPath); err != nil {
			return fmt.Errorf("failed to generate spec: %w", err)
		}

		fmt.Println("Generation complete!")
		return nil
	},
}

func init() {
	generateCmd.Flags().StringVarP(&outputPath, "output", "o", "generated/commands", "Output path for generated commands")
}

func countMethods(services map[string][]generator.Method) int {
	count := 0
	for _, methods := range services {
		count += len(methods)
	}
	return count
}
