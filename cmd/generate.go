package cmd

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
	Short: "Generate CLI commands from SDK",
	Long:  `Parse the go-scalingo SDK and generate Cobra commands for new endpoints.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Parsing SDK at: %s\n", sdkPath)

		// Parse SDK interfaces
		services, err := generator.ParseSDK(sdkPath)
		if err != nil {
			return fmt.Errorf("failed to parse SDK: %w", err)
		}

		fmt.Printf("Found %d services\n", len(services))

		// Load existing manifest
		manifest, err := generator.LoadManifest("manifest.toml")
		if err != nil {
			fmt.Println("No existing manifest found, creating new one")
			manifest = generator.NewManifest()
		}

		// Find new methods
		newMethods := generator.DiffServices(services, manifest)
		if len(newMethods) == 0 {
			fmt.Println("No new methods found")
			return nil
		}

		fmt.Printf("Found %d new methods to generate\n", countMethods(newMethods))

		// Generate code
		if err := generator.GenerateCommands(newMethods, outputPath); err != nil {
			return fmt.Errorf("failed to generate commands: %w", err)
		}

		// Generate spec
		if err := generator.GenerateSpec(newMethods, outputPath); err != nil {
			return fmt.Errorf("failed to generate spec: %w", err)
		}

		// Update manifest
		manifest.AddServices(services)
		if err := manifest.Save("manifest.toml"); err != nil {
			return fmt.Errorf("failed to save manifest: %w", err)
		}

		fmt.Println("Generation complete!")
		return nil
	},
}

func init() {
	generateCmd.Flags().StringVarP(&sdkPath, "sdk-path", "s", "", "Path to go-scalingo SDK (required)")
	generateCmd.Flags().StringVarP(&outputPath, "output", "o", "generated/commands", "Output path for generated commands")
	_ = generateCmd.MarkFlagRequired("sdk-path")
}

func countMethods(services map[string][]generator.Method) int {
	count := 0
	for _, methods := range services {
		count += len(methods)
	}
	return count
}
