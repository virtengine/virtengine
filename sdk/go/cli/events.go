package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"

	"github.com/cosmos/cosmos-sdk/client"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"

	"github.com/virtengine/virtengine/sdk/go/util/events"
	"github.com/virtengine/virtengine/sdk/go/util/pubsub"
)

// EventsCmd prints out events in real time
func EventsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "events",
		Short: "Prints out virtengine events in real time",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunForeverWithContext(cmd.Context(), func(ctx context.Context) error {
				return getEvents(ctx, cmd, args)
			})
		},
	}

	cmd.Flags().String(cflags.FlagNode, "tcp://localhost:26657", "The node address")
	if err := viper.BindPFlag(cflags.FlagNode, cmd.Flags().Lookup(cflags.FlagNode)); err != nil {
		panic(fmt.Sprintf("failed to bind flag %s: %v", cflags.FlagNode, err))
	}

	return cmd
}

func getEvents(ctx context.Context, cmd *cobra.Command, _ []string) error {
	cctx := client.GetClientContextFromCmd(cmd)

	node, err := cctx.GetNode()
	if err != nil {
		return err
	}

	bus := pubsub.NewBus()
	defer bus.Close()

	group, ctx := errgroup.WithContext(ctx)

	subscriber, err := bus.Subscribe()
	if err != nil {
		return err
	}

	evtSvc, err := events.NewEvents(ctx, node, "virtengine-cli", bus)
	if err != nil {
		return err
	}

	group.Go(func() error {
		<-ctx.Done()
		evtSvc.Shutdown()

		return nil
	})

	group.Go(func() error {
		for {
			select {
			case <-subscriber.Done():
				return nil
			case ev := <-subscriber.Events():
				if err := PrintJSON(cctx, ev); err != nil {
					return err
				}
			}
		}
	})

	err = group.Wait()
	if err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	return nil
}
