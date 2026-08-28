// Package lro provides a generic waiter for the long-running operations
// returned by the Live Stream API (channels, inputs, assets).
package lro

import (
	"context"

	gax "github.com/googleapis/gax-go/v2"
	"github.com/simeon/castctl/internal/cli"
)

// Operation is the shared surface of the generated *Operation wrappers.
type Operation[T any] interface {
	Wait(context.Context, ...gax.CallOption) (*T, error)
	Name() string
}

// Wait blocks on op unless --async is set. In async mode it prints the
// operation name to stderr and returns (nil, nil).
func Wait[T any](ctx context.Context, op Operation[T], kind string) (*T, error) {
	if cli.Current.Async {
		cli.Infof("%s operation started (async): %s", kind, op.Name())
		return nil, nil
	}
	cli.Infof("Waiting for %s to complete...", kind)
	return op.Wait(ctx)
}

// EmptyOperation is the surface of generated *Operation wrappers whose result
// is google.protobuf.Empty (Wait returns only an error) — e.g. deletes.
type EmptyOperation interface {
	Wait(context.Context, ...gax.CallOption) error
	Name() string
}

// WaitEmpty blocks on an empty-result operation unless --async is set.
func WaitEmpty(ctx context.Context, op EmptyOperation, kind string) error {
	if cli.Current.Async {
		cli.Infof("%s operation started (async): %s", kind, op.Name())
		return nil
	}
	cli.Infof("Waiting for %s to complete...", kind)
	return op.Wait(ctx)
}
