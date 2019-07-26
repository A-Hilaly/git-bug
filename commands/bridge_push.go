package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/MichaelMure/git-bug/bridge"
	"github.com/MichaelMure/git-bug/bridge/core"
	"github.com/MichaelMure/git-bug/cache"
	"github.com/MichaelMure/git-bug/util/interrupt"
)

func runBridgePush(cmd *cobra.Command, args []string) error {
	backend, err := cache.NewRepoCache(repo)
	if err != nil {
		return err
	}
	defer backend.Close()
	interrupt.RegisterCleaner(backend.Close)

	var b *core.Bridge

	if len(args) == 0 {
		b, err = bridge.DefaultBridge(backend)
	} else {
		b, err = bridge.LoadBridge(backend, args[0])
	}

	if err != nil {
		return err
	}

	parentCtx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	interrupt.RegisterCleaner(func() error {
		cancel()
		return nil
	})

	// TODO: by default export only new events
	out, err := b.ExportAll(ctx, time.Time{})
	if err != nil {
		return err
	}

	for result := range out {
		if result.Err != nil {
			fmt.Println(result.Err, result.Reason)
		} else {
			fmt.Printf("%s: %s\n", result.String(), result.ID)
		}
	}

	return nil
}

var bridgePushCmd = &cobra.Command{
	Use:     "push [<name>]",
	Short:   "Push updates.",
	PreRunE: loadRepo,
	RunE:    runBridgePush,
	Args:    cobra.MaximumNArgs(1),
}

func init() {
	bridgeCmd.AddCommand(bridgePushCmd)
}
