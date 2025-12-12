package generatorcli

import (
	"fmt"

	"github.com/spf13/cobra"

	"generative-cli/generator"
)

var outputPath string

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate CLI commands from manifest",
	Long:  `Generate Cobra commands based on entries in manifest.toml (manifest is treated as read-only).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		manifest, err := generator.LoadManifest("manifest.toml")
		if err != nil {
			return fmt.Errorf("failed to load manifest: %w", err)
		}

		// Resolve SDK path and parse for full method signatures (including all return types)
		resolvedSDKPath, err := resolveSDKPath(sdkPath)
		if err != nil {
			return fmt.Errorf("failed to resolve SDK path: %w", err)
		}

		fmt.Printf("Parsing SDK at: %s\n", resolvedSDKPath)

		// Parse SDK to get full method signatures with all return types
		services, structs, err := generator.ParseSDKWithStructs(resolvedSDKPath)
		if err != nil {
			return fmt.Errorf("failed to parse SDK: %w", err)
		}

		fmt.Printf("Parsed %d services and %d structs from SDK\n", len(services), len(structs))

		// Detect method chaining patterns (e.g., logsURL param -> LogsURL method)
		fmt.Println("Detecting method chaining patterns...")
		services = generator.DetectMethodChaining(services)

		// Build a map of methods to generate based on manifest
		methodsToGen := manifest.MethodsToGenerateSet()

		// Filter parsed services to only include methods marked for generation
		// Also include hidden methods that are needed for chaining
		methods := make(map[string][]generator.Method)
		for _, svc := range services {
			for _, method := range svc.Methods {
				key := svc.Name + "." + method.Name
				if methodsToGen[key] {
					methods[svc.Name] = append(methods[svc.Name], method)
				}
			}
		}

		if countMethods(methods) == 0 {
			fmt.Println("No methods marked for generation in manifest")
			return nil
		}

		fmt.Printf("Generating commands for %d methods across %d services\n", countMethods(methods), len(methods))

		// Generate code
		if err := generator.GenerateCommands(methods, structs, outputPath); err != nil {
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
