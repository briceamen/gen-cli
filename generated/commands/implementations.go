package commands

import (
	"context"
	"fmt"

	scalingo "github.com/Scalingo/go-scalingo/v8"
	"github.com/spf13/cobra"

	"generative-cli/config"
	"generative-cli/render"
)

func init() {
	// Override the generated stubs with working implementations
	appsListCmd.RunE = appsListImpl
	appsShowCmd.RunE = appsShowImpl
	regionsListCmd.RunE = regionsListImpl
	stacksListCmd.RunE = stacksListImpl
}

// newClient creates a new Scalingo client using the config package
func newClient(ctx context.Context) (*scalingo.Client, error) {
	token, err := config.C.LoadAuth()
	if err != nil {
		return nil, err
	}

	return scalingo.New(ctx, scalingo.ClientConfig{
		APIToken: token,
		Region:   config.C.GetRegion(),
	})
}

func appsListImpl(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	client, err := newClient(ctx)
	if err != nil {
		fmt.Println(render.RenderError(err))
		return err
	}

	result, err := client.AppsList(ctx)
	if err != nil {
		fmt.Println(render.RenderError(err))
		return err
	}

	output, _ := cmd.Flags().GetString("output")
	format := render.FormatTable
	if output == "json" {
		format = render.FormatJSON
	}

	rendered, err := render.RenderResult(result, format)
	if err != nil {
		fmt.Println(render.RenderError(err))
		return err
	}
	fmt.Println(rendered)
	return nil
}

func appsShowImpl(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	client, err := newClient(ctx)
	if err != nil {
		fmt.Println(render.RenderError(err))
		return err
	}

	appName, _ := cmd.Flags().GetString("app-name")
	if appName == "" {
		return fmt.Errorf("--app-name is required")
	}

	result, err := client.AppsShow(ctx, appName)
	if err != nil {
		fmt.Println(render.RenderError(err))
		return err
	}

	output, _ := cmd.Flags().GetString("output")
	format := render.FormatDetail
	if output == "json" {
		format = render.FormatJSON
	}

	rendered, err := render.RenderResult(result, format)
	if err != nil {
		fmt.Println(render.RenderError(err))
		return err
	}
	fmt.Println(rendered)
	return nil
}

func regionsListImpl(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	token, err := config.C.LoadAuth()
	if err != nil {
		fmt.Println(render.RenderError(err))
		return err
	}

	// Regions endpoint doesn't require a region
	client, err := scalingo.New(ctx, scalingo.ClientConfig{
		APIToken: token,
	})
	if err != nil {
		fmt.Println(render.RenderError(err))
		return err
	}

	result, err := client.RegionsList(ctx)
	if err != nil {
		fmt.Println(render.RenderError(err))
		return err
	}

	output, _ := cmd.Flags().GetString("output")
	format := render.FormatTable
	if output == "json" {
		format = render.FormatJSON
	}

	rendered, err := render.RenderResult(result, format)
	if err != nil {
		fmt.Println(render.RenderError(err))
		return err
	}
	fmt.Println(rendered)
	return nil
}

func stacksListImpl(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	client, err := newClient(ctx)
	if err != nil {
		fmt.Println(render.RenderError(err))
		return err
	}

	result, err := client.StacksList(ctx)
	if err != nil {
		fmt.Println(render.RenderError(err))
		return err
	}

	output, _ := cmd.Flags().GetString("output")
	format := render.FormatTable
	if output == "json" {
		format = render.FormatJSON
	}

	rendered, err := render.RenderResult(result, format)
	if err != nil {
		fmt.Println(render.RenderError(err))
		return err
	}
	fmt.Println(rendered)
	return nil
}
