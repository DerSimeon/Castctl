package cmd

import (
	"fmt"
	"os"

	"github.com/simeon/castctl/internal/config"
	"github.com/simeon/castctl/internal/output"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Read and write persistent defaults (~/.castctl/config.yaml)",
	}

	set := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a config value (e.g. project, location)",
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := config.Set(args[0], args[1]); err != nil {
				return err
			}
			fmt.Printf("Set %s = %s\n", args[0], args[1])
			return nil
		},
	}

	get := &cobra.Command{
		Use:   "get <key>",
		Short: "Get a stored config value",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			v, err := config.Get(args[0])
			if err != nil {
				return err
			}
			fmt.Println(v)
			return nil
		},
	}

	view := &cobra.Command{
		Use:   "view",
		Short: "Show all stored config plus the file path",
		RunE: func(_ *cobra.Command, _ []string) error {
			all, err := config.All()
			if err != nil {
				return err
			}
			if flagJSON {
				return output.JSONValue(all)
			}
			path, _ := config.FilePath()
			fmt.Printf("Config file: %s\n", path)
			if _, statErr := os.Stat(path); statErr != nil {
				fmt.Println("(not created yet)")
				return nil
			}
			for k, v := range all {
				fmt.Printf("%s: %v\n", k, v)
			}
			return nil
		},
	}

	cmd.AddCommand(set, get, view)
	return cmd
}
