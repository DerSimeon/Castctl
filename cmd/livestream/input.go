package livestream

import (
	"context"

	livestream "cloud.google.com/go/video/livestream/apiv1"
	"cloud.google.com/go/video/livestream/apiv1/livestreampb"
	"github.com/simeon/castctl/internal/cli"
	"github.com/simeon/castctl/internal/client"
	"github.com/simeon/castctl/internal/lro"
	"github.com/simeon/castctl/internal/output"
	"github.com/simeon/castctl/internal/parent"
	"github.com/spf13/cobra"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

const inputsCollection = "inputs"

// NewInputCmd builds `castctl input ...`.
func NewInputCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "input",
		Short:   "Manage Live Stream inputs (RTMP/SRT endpoints)",
		Aliases: []string{"inputs"},
	}
	cmd.AddCommand(
		inputListCmd(),
		inputGetCmd(),
		inputCreateCmd(),
		inputUpdateCmd(),
		inputDeleteCmd(),
	)
	return cmd
}

func withLiveStream(fn func(context.Context, *livestream.Client) error) error {
	ctx := context.Background()
	c, err := client.LiveStream(ctx)
	if err != nil {
		return err
	}
	defer c.Close()
	return fn(ctx, c)
}

func inputColumns() []output.Column[*livestreampb.Input] {
	return []output.Column[*livestreampb.Input]{
		{Header: "id", Value: func(i *livestreampb.Input) string { return parent.LastSegment(i.GetName()) }},
		{Header: "type", Value: func(i *livestreampb.Input) string { return i.GetType().String() }},
		{Header: "tier", Value: func(i *livestreampb.Input) string { return i.GetTier().String() }},
		{Header: "uri", Value: func(i *livestreampb.Input) string { return i.GetUri() }},
	}
}

func inputListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List inputs",
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := cli.Current.RequireProjectLocation(); err != nil {
				return err
			}
			return withLiveStream(func(ctx context.Context, c *livestream.Client) error {
				it := c.ListInputs(ctx, &livestreampb.ListInputsRequest{
					Parent: parent.Location(cli.Current.Project, cli.Current.Location),
				})
				var items []*livestreampb.Input
				for {
					in, err := it.Next()
					if err == iterator.Done {
						break
					}
					if err != nil {
						return err
					}
					items = append(items, in)
				}
				if cli.Current.JSON {
					return output.JSONProtoList(items)
				}
				if len(items) == 0 {
					return output.Empty("inputs", false)
				}
				return output.Table(items, inputColumns())
			})
		},
	}
}

func inputGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get an input",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := cli.Current.RequireProjectLocation(); err != nil {
				return err
			}
			return withLiveStream(func(ctx context.Context, c *livestream.Client) error {
				in, err := c.GetInput(ctx, &livestreampb.GetInputRequest{
					Name: parent.Resource(cli.Current.Project, cli.Current.Location, inputsCollection, args[0]),
				})
				if err != nil {
					return err
				}
				return output.JSONProto(in)
			})
		},
	}
}

func inputCreateCmd() *cobra.Command {
	var file, inputType string
	c := &cobra.Command{
		Use:   "create <id>",
		Short: "Create an input from -f spec.json (or --type for a minimal input)",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := cli.Current.RequireProjectLocation(); err != nil {
				return err
			}
			input := &livestreampb.Input{}
			if file != "" {
				if err := cli.UnmarshalSpec(file, input); err != nil {
					return err
				}
			} else if inputType != "" {
				t, ok := livestreampb.Input_Type_value[inputType]
				if !ok {
					return errBadEnum("type", inputType, livestreampb.Input_Type_name)
				}
				input.Type = livestreampb.Input_Type(t)
			} else {
				return errNeedSpecOr("--type")
			}
			return withLiveStream(func(ctx context.Context, cl *livestream.Client) error {
				op, err := cl.CreateInput(ctx, &livestreampb.CreateInputRequest{
					Parent:  parent.Location(cli.Current.Project, cli.Current.Location),
					InputId: args[0],
					Input:   input,
				})
				if err != nil {
					return err
				}
				res, err := lro.Wait(ctx, op, "input create")
				if err != nil || res == nil {
					return err
				}
				return output.JSONProto(res)
			})
		},
	}
	c.Flags().StringVarP(&file, "file", "f", "", "path to input spec JSON (- for stdin)")
	c.Flags().StringVar(&inputType, "type", "", "input type: RTMP_PUSH or SRT_PUSH (when no -f)")
	return c
}

func inputUpdateCmd() *cobra.Command {
	var file, mask string
	c := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an input from -f spec.json",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := cli.Current.RequireProjectLocation(); err != nil {
				return err
			}
			input := &livestreampb.Input{}
			if err := cli.UnmarshalSpec(file, input); err != nil {
				return err
			}
			input.Name = parent.Resource(cli.Current.Project, cli.Current.Location, inputsCollection, args[0])
			req := &livestreampb.UpdateInputRequest{Input: input}
			if mask != "" {
				req.UpdateMask = &fieldmaskpb.FieldMask{Paths: splitCSV(mask)}
			}
			return withLiveStream(func(ctx context.Context, cl *livestream.Client) error {
				op, err := cl.UpdateInput(ctx, req)
				if err != nil {
					return err
				}
				res, err := lro.Wait(ctx, op, "input update")
				if err != nil || res == nil {
					return err
				}
				return output.JSONProto(res)
			})
		},
	}
	c.Flags().StringVarP(&file, "file", "f", "", "path to input spec JSON (- for stdin)")
	c.Flags().StringVar(&mask, "update-mask", "", "comma-separated field paths to update")
	return c
}

func inputDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an input",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := cli.Current.RequireProjectLocation(); err != nil {
				return err
			}
			return withLiveStream(func(ctx context.Context, cl *livestream.Client) error {
				op, err := cl.DeleteInput(ctx, &livestreampb.DeleteInputRequest{
					Name: parent.Resource(cli.Current.Project, cli.Current.Location, inputsCollection, args[0]),
				})
				if err != nil {
					return err
				}
				if err := lro.WaitEmpty(ctx, op, "input delete"); err != nil {
					return err
				}
				cli.Infof("Deleted input %s", args[0])
				return nil
			})
		},
	}
}
