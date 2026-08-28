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

const assetsCollection = "assets"

// NewAssetCmd builds `castctl asset ...`.
func NewAssetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "asset",
		Short:   "Manage Live Stream assets (VOD images/videos for slates)",
		Aliases: []string{"assets"},
	}
	cmd.AddCommand(
		assetListCmd(),
		assetGetCmd(),
		assetCreateCmd(),
		assetDeleteCmd(),
	)
	return cmd
}

func assetColumns() []output.Column[*livestreampb.Asset] {
	return []output.Column[*livestreampb.Asset]{
		{Header: "id", Value: func(a *livestreampb.Asset) string { return parent.LastSegment(a.GetName()) }},
		{Header: "state", Value: func(a *livestreampb.Asset) string { return a.GetState().String() }},
		{Header: "created", Value: func(a *livestreampb.Asset) string { return ts(a.GetCreateTime()) }},
	}
}

func assetListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List assets",
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := cli.Current.RequireProjectLocation(); err != nil {
				return err
			}
			return withLiveStream(func(ctx context.Context, cl *livestream.Client) error {
				it := cl.ListAssets(ctx, &livestreampb.ListAssetsRequest{
					Parent: parent.Location(cli.Current.Project, cli.Current.Location),
				})
				var items []*livestreampb.Asset
				for {
					a, err := it.Next()
					if err == iterator.Done {
						break
					}
					if err != nil {
						return err
					}
					items = append(items, a)
				}
				if cli.Current.JSON {
					return output.JSONProtoList(items)
				}
				if len(items) == 0 {
					return output.Empty("assets", false)
				}
				return output.Table(items, assetColumns())
			})
		},
	}
}

func assetGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get an asset",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := cli.Current.RequireProjectLocation(); err != nil {
				return err
			}
			return withLiveStream(func(ctx context.Context, cl *livestream.Client) error {
				a, err := cl.GetAsset(ctx, &livestreampb.GetAssetRequest{
					Name: parent.Resource(cli.Current.Project, cli.Current.Location, assetsCollection, args[0]),
				})
				if err != nil {
					return err
				}
				return output.JSONProto(a)
			})
		},
	}
}

func assetCreateCmd() *cobra.Command {
	var file string
	c := &cobra.Command{
		Use:   "create <id>",
		Short: "Create an asset from -f spec.json",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := cli.Current.RequireProjectLocation(); err != nil {
				return err
			}
			asset := &livestreampb.Asset{}
			if err := cli.UnmarshalSpec(file, asset); err != nil {
				return err
			}
			return withLiveStream(func(ctx context.Context, cl *livestream.Client) error {
				op, err := cl.CreateAsset(ctx, &livestreampb.CreateAssetRequest{
					Parent:  parent.Location(cli.Current.Project, cli.Current.Location),
					AssetId: args[0],
					Asset:   asset,
				})
				if err != nil {
					return err
				}
				res, err := lro.Wait(ctx, op, "asset create")
				if err != nil || res == nil {
					return err
				}
				return output.JSONProto(res)
			})
		},
	}
	c.Flags().StringVarP(&file, "file", "f", "", "path to asset spec JSON (- for stdin)")
	_ = c.MarkFlagRequired("file")
	return c
}

func assetDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an asset",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := cli.Current.RequireProjectLocation(); err != nil {
				return err
			}
			return withLiveStream(func(ctx context.Context, cl *livestream.Client) error {
				op, err := cl.DeleteAsset(ctx, &livestreampb.DeleteAssetRequest{
					Name: parent.Resource(cli.Current.Project, cli.Current.Location, assetsCollection, args[0]),
				})
				if err != nil {
					return err
				}
				if err := lro.WaitEmpty(ctx, op, "asset delete"); err != nil {
					return err
				}
				cli.Infof("Deleted asset %s", args[0])
				return nil
			})
		},
	}
}
