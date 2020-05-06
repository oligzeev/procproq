package util

import (
	"context"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

func StartSignalReceiver(groupCtx context.Context, done context.CancelFunc) error {
	log.Trace("SignalReceiver: starting")

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-signalChannel:
		log.Tracef("SignalReceiver: done with %s", sig)
		done()
	case <-groupCtx.Done():
		log.Trace("SignalReceiver: exit")
		return groupCtx.Err()
	}
	return nil
}
