package checkpoint

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jpconstantineau/dupectl/pkg/logger"
)

// SetupSignalHandler creates a context that is cancelled on SIGINT/SIGTERM
// Returns context and a cleanup function that saves checkpoint
func SetupSignalHandler(onShutdown func()) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Info("Received signal %v, initiating graceful shutdown...", sig)

		// Cancel context to stop operations
		cancel()

		// Save checkpoint with timeout
		done := make(chan struct{})
		go func() {
			if onShutdown != nil {
				onShutdown()
			}
			close(done)
		}()

		// Wait up to 5 seconds for checkpoint save
		select {
		case <-done:
			logger.Info("Checkpoint saved, exiting")
		case <-time.After(5 * time.Second):
			logger.Warn("Checkpoint save timeout, forcing exit")
		}

		os.Exit(130) // Standard exit code for SIGINT
	}()

	return ctx, cancel
}
