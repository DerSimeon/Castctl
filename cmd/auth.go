package cmd

import (
	"context"
	"fmt"

	"github.com/simeon/castctl/internal/cli"
	"github.com/simeon/castctl/internal/client"
	"github.com/simeon/castctl/internal/output"
	"github.com/spf13/cobra"
)

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Inspect Application Default Credentials",
	}

	status := &cobra.Command{
		Use:   "status",
		Short: "Verify ADC is usable and show the resolved identity and project",
		RunE: func(c *cobra.Command, _ []string) error {
			ctx := context.Background()
			email, err := client.ADCEmail(ctx)
			if err != nil {
				return err
			}
			principal := email
			if principal == "" {
				principal = "(user credentials — email not exposed by ADC)"
			}
			result := map[string]string{
				"credentials": "ok",
				"principal":   principal,
				"project":     cli.Current.Project,
				"location":    cli.Current.Location,
			}
			if cli.Current.JSON {
				return output.JSONValue(result)
			}
			fmt.Println("Application Default Credentials: OK")
			fmt.Printf("Principal: %s\n", principal)
			fmt.Printf("Project:   %s\n", orNotSet(cli.Current.Project))
			fmt.Printf("Location:  %s\n", orNotSet(cli.Current.Location))
			return nil
		},
	}

	cmd.AddCommand(status)
	return cmd
}

func orNotSet(s string) string {
	if s == "" {
		return "(not set)"
	}
	return s
}
