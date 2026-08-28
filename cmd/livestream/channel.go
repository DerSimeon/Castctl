package livestream

import (
	"context"

	livestream "cloud.google.com/go/video/livestream/apiv1"
	"cloud.google.com/go/video/livestream/apiv1/livestreampb"
	"github.com/simeon/castctl/internal/cli"
	"github.com/simeon/castctl/internal/lro"
	"github.com/simeon/castctl/internal/output"
	"github.com/simeon/castctl/internal/parent"
	"github.com/spf13/cobra"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

const channelsCollection = "channels"

// NewChannelCmd builds `castctl channel ...`.
func NewChannelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "channel",
		Short:   "Manage Live Stream channels",
		Aliases: []string{"channels"},
	}
	cmd.AddCommand(
		channelListCmd(),
		channelGetCmd(),
		channelCreateCmd(),
		channelUpdateCmd(),
		channelDeleteCmd(),
		channelStartCmd(),
		channelStopCmd(),
	)
	return cmd
}

func channelColumns() []output.Column[*livestreampb.Channel] {
	return []output.Column[*livestreampb.Channel]{
		{Header: "id", Value: func(c *livestreampb.Channel) string { return parent.LastSegment(c.GetName()) }},
		{Header: "state", Value: func(c *livestreampb.Channel) string { return c.GetStreamingState().String() }},
		{Header: "created", Value: func(c *livestreampb.Channel) string { return ts(c.GetCreateTime()) }},
	}
}

func channelListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List channels",
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := cli.Current.RequireProjectLocation(); err != nil {
				return err
			}
			return withLiveStream(func(ctx context.Context, c *livestream.Client) error {
				it := c.ListChannels(ctx, &livestreampb.ListChannelsRequest{
					Parent: parent.Location(cli.Current.Project, cli.Current.Location),
				})
				var items []*livestreampb.Channel
				for {
					ch, err := it.Next()
					if err == iterator.Done {
						break
					}
					if err != nil {
						return err
					}
					items = append(items, ch)
				}
				if cli.Current.JSON {
					return output.JSONProtoList(items)
				}
				if len(items) == 0 {
					return output.Empty("channels", false)
				}
				return output.Table(items, channelColumns())
			})
		},
	}
}

func channelGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a channel",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := cli.Current.RequireProjectLocation(); err != nil {
				return err
			}
			return withLiveStream(func(ctx context.Context, c *livestream.Client) error {
				ch, err := c.GetChannel(ctx, &livestreampb.GetChannelRequest{
					Name: parent.Resource(cli.Current.Project, cli.Current.Location, channelsCollection, args[0]),
				})
				if err != nil {
					return err
				}
				return output.JSONProto(ch)
			})
		},
	}
}

func channelCreateCmd() *cobra.Command {
	var file string
	c := &cobra.Command{
		Use:   "create <id>",
		Short: "Create a channel from -f spec.json",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := cli.Current.RequireProjectLocation(); err != nil {
				return err
			}
			ch := &livestreampb.Channel{}
			if err := cli.UnmarshalSpec(file, ch); err != nil {
				return err
			}
			return withLiveStream(func(ctx context.Context, cl *livestream.Client) error {
				op, err := cl.CreateChannel(ctx, &livestreampb.CreateChannelRequest{
					Parent:    parent.Location(cli.Current.Project, cli.Current.Location),
					ChannelId: args[0],
					Channel:   ch,
				})
				if err != nil {
					return err
				}
				res, err := lro.Wait(ctx, op, "channel create")
				if err != nil || res == nil {
					return err
				}
				return output.JSONProto(res)
			})
		},
	}
	c.Flags().StringVarP(&file, "file", "f", "", "path to channel spec JSON (- for stdin)")
	_ = c.MarkFlagRequired("file")
	return c
}

func channelUpdateCmd() *cobra.Command {
	var file, mask string
	c := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a channel from -f spec.json",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := cli.Current.RequireProjectLocation(); err != nil {
				return err
			}
			ch := &livestreampb.Channel{}
			if err := cli.UnmarshalSpec(file, ch); err != nil {
				return err
			}
			ch.Name = parent.Resource(cli.Current.Project, cli.Current.Location, channelsCollection, args[0])
			req := &livestreampb.UpdateChannelRequest{Channel: ch}
			if mask != "" {
				req.UpdateMask = &fieldmaskpb.FieldMask{Paths: splitCSV(mask)}
			}
			return withLiveStream(func(ctx context.Context, cl *livestream.Client) error {
				op, err := cl.UpdateChannel(ctx, req)
				if err != nil {
					return err
				}
				res, err := lro.Wait(ctx, op, "channel update")
				if err != nil || res == nil {
					return err
				}
				return output.JSONProto(res)
			})
		},
	}
	c.Flags().StringVarP(&file, "file", "f", "", "path to channel spec JSON (- for stdin)")
	c.Flags().StringVar(&mask, "update-mask", "", "comma-separated field paths to update")
	_ = c.MarkFlagRequired("file")
	return c
}

func channelDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a channel",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := cli.Current.RequireProjectLocation(); err != nil {
				return err
			}
			return withLiveStream(func(ctx context.Context, cl *livestream.Client) error {
				op, err := cl.DeleteChannel(ctx, &livestreampb.DeleteChannelRequest{
					Name: parent.Resource(cli.Current.Project, cli.Current.Location, channelsCollection, args[0]),
				})
				if err != nil {
					return err
				}
				if err := lro.WaitEmpty(ctx, op, "channel delete"); err != nil {
					return err
				}
				cli.Infof("Deleted channel %s", args[0])
				return nil
			})
		},
	}
}

func channelStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start <id>",
		Short: "Start streaming a channel",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := cli.Current.RequireProjectLocation(); err != nil {
				return err
			}
			return withLiveStream(func(ctx context.Context, cl *livestream.Client) error {
				op, err := cl.StartChannel(ctx, &livestreampb.StartChannelRequest{
					Name: parent.Resource(cli.Current.Project, cli.Current.Location, channelsCollection, args[0]),
				})
				if err != nil {
					return err
				}
				if _, err := lro.Wait(ctx, op, "channel start"); err != nil {
					return err
				}
				cli.Infof("Started channel %s", args[0])
				return nil
			})
		},
	}
}

func channelStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop <id>",
		Short: "Stop streaming a channel",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := cli.Current.RequireProjectLocation(); err != nil {
				return err
			}
			return withLiveStream(func(ctx context.Context, cl *livestream.Client) error {
				op, err := cl.StopChannel(ctx, &livestreampb.StopChannelRequest{
					Name: parent.Resource(cli.Current.Project, cli.Current.Location, channelsCollection, args[0]),
				})
				if err != nil {
					return err
				}
				if _, err := lro.Wait(ctx, op, "channel stop"); err != nil {
					return err
				}
				cli.Infof("Stopped channel %s", args[0])
				return nil
			})
		},
	}
}
