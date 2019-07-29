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

func runBridgePull(cmd *cobra.Command, args []string) error {
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
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	done := make(chan struct{})

	interrupt.RegisterCleaner(func() error {
		// send signal to stop the importer
		cancel()

		// block until importer gracefully shutdown
		<-done
		close(done)
		return nil
	})

	// TODO: by default import only new events
	events, err := b.ImportAll(ctx, time.Time{})
	if err != nil {
		return err
	}

	for result := range events {
		if result.Err != nil {
			fmt.Println(result.Err, result.Reason)
		} else {
			fmt.Printf("%s: %s\n", result.String(), result.ID)
		}
	}

	// send done signal
	done <- struct{}{}

	return nil
}

var bridgePullCmd = &cobra.Command{
	Use:     "pull [<name>]",
	Short:   "Pull updates.",
	PreRunE: loadRepo,
	RunE:    runBridgePull,
	Args:    cobra.MaximumNArgs(1),
}

func init() {
	bridgeCmd.AddCommand(bridgePullCmd)
}
