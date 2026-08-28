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
)

const clipsCollection = "clips"

// NewClipCmd builds `castctl clip ...`. Clips are cut from a channel's stream,
// so every subcommand requires --channel.
func NewClipCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "clip",
		Short:   "Manage Live Stream clips (VOD cut from a channel)",
		Aliases: []string{"clips"},
	}
	cmd.AddCommand(
		clipListCmd(),
		clipGetCmd(),
		clipCreateCmd(),
		clipDeleteCmd(),
	)
	return cmd
}

func clipColumns() []output.Column[*livestreampb.Clip] {
	return []output.Column[*livestreampb.Clip]{
		{Header: "id", Value: func(c *livestreampb.Clip) string { return parent.LastSegment(c.GetName()) }},
		{Header: "state", Value: func(c *livestreampb.Clip) string { return c.GetState().String() }},
		{Header: "created", Value: func(c *livestreampb.Clip) string { return ts(c.GetCreateTime()) }},
	}
}

func clipListCmd() *cobra.Command {
	var channel string
	c := &cobra.Command{
		Use:   "list",
		Short: "List clips on a channel",
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := cli.Current.RequireProjectLocation(); err != nil {
				return err
			}
			return withLiveStream(func(ctx context.Context, cl *livestream.Client) error {
				it := cl.ListClips(ctx, &livestreampb.ListClipsRequest{Parent: channelName(channel)})
				var items []*livestreampb.Clip
				for {
					cp, err := it.Next()
					if err == iterator.Done {
						break
					}
					if err != nil {
						return err
					}
					items = append(items, cp)
				}
				if cli.Current.JSON {
					return output.JSONProtoList(items)
				}
				if len(items) == 0 {
					return output.Empty("clips", false)
				}
				return output.Table(items, clipColumns())
			})
		},
	}
	requireChannel(c, &channel)
	return c
}

func clipGetCmd() *cobra.Command {
	var channel string
	c := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a clip",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := cli.Current.RequireProjectLocation(); err != nil {
				return err
			}
			return withLiveStream(func(ctx context.Context, cl *livestream.Client) error {
				cp, err := cl.GetClip(ctx, &livestreampb.GetClipRequest{
					Name: parent.Child(channelName(channel), clipsCollection, args[0]),
				})
				if err != nil {
					return err
				}
				return output.JSONProto(cp)
			})
		},
	}
	requireChannel(c, &channel)
	return c
}

func clipCreateCmd() *cobra.Command {
	var channel, file string
	c := &cobra.Command{
		Use:   "create <id>",
		Short: "Create a clip from -f spec.json",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := cli.Current.RequireProjectLocation(); err != nil {
				return err
			}
			clip := &livestreampb.Clip{}
			if err := cli.UnmarshalSpec(file, clip); err != nil {
				return err
			}
			return withLiveStream(func(ctx context.Context, cl *livestream.Client) error {
				op, err := cl.CreateClip(ctx, &livestreampb.CreateClipRequest{
					Parent: channelName(channel),
					ClipId: args[0],
					Clip:   clip,
				})
				if err != nil {
					return err
				}
				res, err := lro.Wait(ctx, op, "clip create")
				if err != nil || res == nil {
					return err
				}
				return output.JSONProto(res)
			})
		},
	}
	requireChannel(c, &channel)
	c.Flags().StringVarP(&file, "file", "f", "", "path to clip spec JSON (- for stdin)")
	_ = c.MarkFlagRequired("file")
	return c
}

func clipDeleteCmd() *cobra.Command {
	var channel string
	c := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a clip",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := cli.Current.RequireProjectLocation(); err != nil {
				return err
			}
			return withLiveStream(func(ctx context.Context, cl *livestream.Client) error {
				op, err := cl.DeleteClip(ctx, &livestreampb.DeleteClipRequest{
					Name: parent.Child(channelName(channel), clipsCollection, args[0]),
				})
				if err != nil {
					return err
				}
				if err := lro.WaitEmpty(ctx, op, "clip delete"); err != nil {
					return err
				}
				cli.Infof("Deleted clip %s", args[0])
				return nil
			})
		},
	}
	requireChannel(c, &channel)
	return c
}
