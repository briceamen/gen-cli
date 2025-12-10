package generatorcli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"generative-cli/generator"
)

var defaultSDKPath = filepath.Join("vendor", "github.com", "Scalingo", "go-scalingo", "v8")

var updateManifestCmd = &cobra.Command{
	Use:   "update-manifest",
	Short: "Scan the SDK and add missing methods to manifest.toml",
	Long:  "Parse the go-scalingo SDK and append any missing methods to manifest.toml without generating code.",
	RunE: func(cmd *cobra.Command, args []string) error {
		resolvedSDKPath, err := resolveSDKPath(sdkPath)
		if err != nil {
			return err
		}

		fmt.Printf("Parsing SDK at: %s\n", resolvedSDKPath)

		services, err := generator.ParseSDK(resolvedSDKPath)
		if err != nil {
			return fmt.Errorf("failed to parse SDK: %w", err)
		}

		fmt.Printf("Found %d services\n", len(services))

		manifest, err := generator.LoadManifest("manifest.toml")
		if err != nil {
			fmt.Println("No existing manifest found, creating new one")
			manifest = generator.NewManifest()
		}

		newMethods := generator.DiffServices(services, manifest)
		if len(newMethods) == 0 {
			fmt.Println("Manifest already up to date")
			return nil
		}

		fmt.Printf("Adding %d new methods to manifest\n", countMethods(newMethods))
		manifest.AddServices(services)
		manifest.EnsureParamNames()

		if err := manifest.Save("manifest.toml"); err != nil {
			return fmt.Errorf("failed to save manifest: %w", err)
		}

		fmt.Println("Manifest updated!")
		return nil
	},
}

func init() {
	updateManifestCmd.Flags().StringVarP(&sdkPath, "sdk-path", "s", "", "Path to go-scalingo SDK (defaults to vendored copy)")
}

func resolveSDKPath(flagValue string) (string, error) {
	if flagValue != "" {
		if _, err := os.Stat(flagValue); err != nil {
			if os.IsNotExist(err) {
				return "", fmt.Errorf("sdk path %s does not exist", flagValue)
			}
			return "", fmt.Errorf("failed to read sdk path %s: %w", flagValue, err)
		}
		return flagValue, nil
	}

	if _, err := os.Stat(defaultSDKPath); err == nil {
		return defaultSDKPath, nil
	} else if os.IsNotExist(err) {
		return "", fmt.Errorf("vendored sdk not found at %s; provide --sdk-path", defaultSDKPath)
	} else {
		return "", fmt.Errorf("failed to read vendored sdk path %s: %w", defaultSDKPath, err)
	}
}
