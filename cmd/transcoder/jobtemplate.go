package transcoder

import (
	"context"

	transcoder "cloud.google.com/go/video/transcoder/apiv1"
	"cloud.google.com/go/video/transcoder/apiv1/transcoderpb"
	"github.com/simeon/castctl/internal/cli"
	"github.com/simeon/castctl/internal/output"
	"github.com/simeon/castctl/internal/parent"
	"github.com/spf13/cobra"
	"google.golang.org/api/iterator"
)

const jobTemplatesCollection = "jobTemplates"

// NewJobTemplateCmd builds `castctl job-template ...`.
func NewJobTemplateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "job-template",
		Short:   "Manage Transcoder job templates",
		Aliases: []string{"job-templates", "jobtemplate"},
	}
	cmd.AddCommand(
		jobTemplateListCmd(),
		jobTemplateGetCmd(),
		jobTemplateCreateCmd(),
		jobTemplateDeleteCmd(),
	)
	return cmd
}

func jobTemplateColumns() []output.Column[*transcoderpb.JobTemplate] {
	return []output.Column[*transcoderpb.JobTemplate]{
		{Header: "id", Value: func(t *transcoderpb.JobTemplate) string { return parent.LastSegment(t.GetName()) }},
	}
}

func jobTemplateListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List job templates",
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := cli.Current.RequireProjectLocation(); err != nil {
				return err
			}
			return withTranscoder(func(ctx context.Context, c *transcoder.Client) error {
				it := c.ListJobTemplates(ctx, &transcoderpb.ListJobTemplatesRequest{
					Parent: parent.Location(cli.Current.Project, cli.Current.Location),
				})
				var items []*transcoderpb.JobTemplate
				for {
					t, err := it.Next()
					if err == iterator.Done {
						break
					}
					if err != nil {
						return err
					}
					items = append(items, t)
				}
				if cli.Current.JSON {
					return output.JSONProtoList(items)
				}
				if len(items) == 0 {
					return output.Empty("job templates", false)
				}
				return output.Table(items, jobTemplateColumns())
			})
		},
	}
}

func jobTemplateGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a job template",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := cli.Current.RequireProjectLocation(); err != nil {
				return err
			}
			return withTranscoder(func(ctx context.Context, c *transcoder.Client) error {
				t, err := c.GetJobTemplate(ctx, &transcoderpb.GetJobTemplateRequest{
					Name: parent.Resource(cli.Current.Project, cli.Current.Location, jobTemplatesCollection, args[0]),
				})
				if err != nil {
					return err
				}
				return output.JSONProto(t)
			})
		},
	}
}

func jobTemplateCreateCmd() *cobra.Command {
	var file string
	c := &cobra.Command{
		Use:   "create <id>",
		Short: "Create a job template from -f spec.json",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := cli.Current.RequireProjectLocation(); err != nil {
				return err
			}
			tmpl := &transcoderpb.JobTemplate{}
			if err := cli.UnmarshalSpec(file, tmpl); err != nil {
				return err
			}
			return withTranscoder(func(ctx context.Context, c *transcoder.Client) error {
				res, err := c.CreateJobTemplate(ctx, &transcoderpb.CreateJobTemplateRequest{
					Parent:        parent.Location(cli.Current.Project, cli.Current.Location),
					JobTemplateId: args[0],
					JobTemplate:   tmpl,
				})
				if err != nil {
					return err
				}
				return output.JSONProto(res)
			})
		},
	}
	c.Flags().StringVarP(&file, "file", "f", "", "path to job-template spec JSON (- for stdin)")
	_ = c.MarkFlagRequired("file")
	return c
}

func jobTemplateDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a job template",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := cli.Current.RequireProjectLocation(); err != nil {
				return err
			}
			return withTranscoder(func(ctx context.Context, c *transcoder.Client) error {
				if err := c.DeleteJobTemplate(ctx, &transcoderpb.DeleteJobTemplateRequest{
					Name: parent.Resource(cli.Current.Project, cli.Current.Location, jobTemplatesCollection, args[0]),
				}); err != nil {
					return err
				}
				cli.Infof("Deleted job template %s", args[0])
				return nil
			})
		},
	}
}
