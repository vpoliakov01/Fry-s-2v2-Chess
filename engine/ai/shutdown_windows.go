//go:build windows

package ai

import "os"

func shutdownSignals() []os.Signal {
	return []os.Signal{os.Interrupt}
}

func terminateAfterShutdown(os.Signal) {
	os.Exit(1)
}
