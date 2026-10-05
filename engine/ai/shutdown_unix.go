//go:build !windows

package ai

import (
	"os"
	"os/signal"
	"syscall"
)

func shutdownSignals() []os.Signal {
	return []os.Signal{
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGABRT,
	}
}

func terminateAfterShutdown(sig os.Signal) {
	signal.Reset(sig)

	if signal, ok := sig.(syscall.Signal); ok {
		if err := syscall.Kill(syscall.Getpid(), signal); err == nil {
			return
		}
	}

	os.Exit(1)
}
