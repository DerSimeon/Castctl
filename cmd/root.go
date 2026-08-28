package cmd

import (
	"fmt"
	"os"

	"github.com/simeon/castctl/cmd/livestream"
	"github.com/simeon/castctl/cmd/transcoder"
	"github.com/simeon/castctl/internal/cli"
	"github.com/simeon/castctl/internal/config"
	"github.com/spf13/cobra"
)

var (
	flagProject  string
	flagLocation string
	flagJSON     bool
	flagAsync    bool
)

// Execute is the CLI entry point.
func Execute(version string) {
	root := newRootCmd(version)
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func newRootCmd(version string) *cobra.Command {
	root := &cobra.Command{
		Use:           "castctl",
		Short:         "Manage GCP Live Stream and Transcoder APIs",
		Long:          "castctl is a CLI for Google Cloud's Live Stream API and Transcoder API.\nAuthentication uses Application Default Credentials (ADC).",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(c *cobra.Command, _ []string) error {
			s, err := config.Resolve(flagProject, flagLocation, flagJSON, flagAsync)
			if err != nil {
				return err
			}
			cli.Current = s
			return nil
		},
	}

	pf := root.PersistentFlags()
	pf.StringVar(&flagProject, "project", "", "GCP project ID (env GOOGLE_CLOUD_PROJECT, or config)")
	pf.StringVar(&flagLocation, "location", "", "GCP region, e.g. us-central1 (env CASTCTL_LOCATION, or config)")
	pf.BoolVar(&flagJSON, "json", false, "output JSON instead of a table")
	pf.BoolVar(&flagAsync, "async", false, "return immediately on long-running operations instead of waiting")

	root.AddCommand(
		newConfigCmd(),
		newAuthCmd(),
		livestream.NewInputCmd(),
		livestream.NewChannelCmd(),
		livestream.NewEventCmd(),
		livestream.NewClipCmd(),
		livestream.NewAssetCmd(),
		transcoder.NewJobCmd(),
		transcoder.NewJobTemplateCmd(),
	)
	return root
}
