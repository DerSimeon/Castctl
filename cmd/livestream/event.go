package livestream

import (
	"context"

	livestream "cloud.google.com/go/video/livestream/apiv1"
	"cloud.google.com/go/video/livestream/apiv1/livestreampb"
	"github.com/simeon/castctl/internal/cli"
	"github.com/simeon/castctl/internal/output"
	"github.com/simeon/castctl/internal/parent"
	"github.com/spf13/cobra"
	"google.golang.org/api/iterator"
)

const eventsCollection = "events"

// NewEventCmd builds `castctl event ...`. Events live under a channel, so every
// subcommand requires --channel.
func NewEventCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "event",
		Short:   "Manage channel events (ad breaks, slates, splices)",
		Aliases: []string{"events"},
	}
	cmd.AddCommand(
		eventListCmd(),
		eventGetCmd(),
		eventCreateCmd(),
		eventDeleteCmd(),
	)
	return cmd
}

func channelName(channel string) string {
	return parent.Resource(cli.Current.Project, cli.Current.Location, channelsCollection, channel)
}

func eventColumns() []output.Column[*livestreampb.Event] {
	return []output.Column[*livestreampb.Event]{
		{Header: "id", Value: func(e *livestreampb.Event) string { return parent.LastSegment(e.GetName()) }},
		{Header: "state", Value: func(e *livestreampb.Event) string { return e.GetState().String() }},
		{Header: "created", Value: func(e *livestreampb.Event) string { return ts(e.GetCreateTime()) }},
	}
}

func eventListCmd() *cobra.Command {
	var channel string
	c := &cobra.Command{
		Use:   "list",
		Short: "List events on a channel",
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := cli.Current.RequireProjectLocation(); err != nil {
				return err
			}
			return withLiveStream(func(ctx context.Context, cl *livestream.Client) error {
				it := cl.ListEvents(ctx, &livestreampb.ListEventsRequest{Parent: channelName(channel)})
				var items []*livestreampb.Event
				for {
					e, err := it.Next()
					if err == iterator.Done {
						break
					}
					if err != nil {
						return err
					}
					items = append(items, e)
				}
				if cli.Current.JSON {
					return output.JSONProtoList(items)
				}
				if len(items) == 0 {
					return output.Empty("events", false)
				}
				return output.Table(items, eventColumns())
			})
		},
	}
	requireChannel(c, &channel)
	return c
}

func eventGetCmd() *cobra.Command {
	var channel string
	c := &cobra.Command{
		Use:   "get <id>",
		Short: "Get an event",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := cli.Current.RequireProjectLocation(); err != nil {
				return err
			}
			return withLiveStream(func(ctx context.Context, cl *livestream.Client) error {
				e, err := cl.GetEvent(ctx, &livestreampb.GetEventRequest{
					Name: parent.Child(channelName(channel), eventsCollection, args[0]),
				})
				if err != nil {
					return err
				}
				return output.JSONProto(e)
			})
		},
	}
	requireChannel(c, &channel)
	return c
}

func eventCreateCmd() *cobra.Command {
	var channel, file string
	c := &cobra.Command{
		Use:   "create <id>",
		Short: "Create an event from -f spec.json",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := cli.Current.RequireProjectLocation(); err != nil {
				return err
			}
			ev := &livestreampb.Event{}
			if err := cli.UnmarshalSpec(file, ev); err != nil {
				return err
			}
			return withLiveStream(func(ctx context.Context, cl *livestream.Client) error {
				// CreateEvent is synchronous (no LRO).
				res, err := cl.CreateEvent(ctx, &livestreampb.CreateEventRequest{
					Parent:  channelName(channel),
					EventId: args[0],
					Event:   ev,
				})
				if err != nil {
					return err
				}
				return output.JSONProto(res)
			})
		},
	}
	requireChannel(c, &channel)
	c.Flags().StringVarP(&file, "file", "f", "", "path to event spec JSON (- for stdin)")
	_ = c.MarkFlagRequired("file")
	return c
}

func eventDeleteCmd() *cobra.Command {
	var channel string
	c := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an event",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := cli.Current.RequireProjectLocation(); err != nil {
				return err
			}
			return withLiveStream(func(ctx context.Context, cl *livestream.Client) error {
				// DeleteEvent is synchronous.
				if err := cl.DeleteEvent(ctx, &livestreampb.DeleteEventRequest{
					Name: parent.Child(channelName(channel), eventsCollection, args[0]),
				}); err != nil {
					return err
				}
				cli.Infof("Deleted event %s", args[0])
				return nil
			})
		},
	}
	requireChannel(c, &channel)
	return c
}

func requireChannel(c *cobra.Command, target *string) {
	c.Flags().StringVar(target, "channel", "", "channel ID the event belongs to")
	_ = c.MarkFlagRequired("channel")
}
