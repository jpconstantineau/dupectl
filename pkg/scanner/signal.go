package scanner

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// SetupSignalHandler creates a context that cancels on SIGINT or SIGTERM
func SetupSignalHandler() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		fmt.Printf("\n\nReceived signal %v, stopping scan and saving checkpoint...\n", sig)
		cancel()
	}()

	return ctx
}
