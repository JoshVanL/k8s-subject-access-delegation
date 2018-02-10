package signals

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

var onlyOneSignalHandler = make(chan struct{})

func RunSignalHandler(log *logrus.Entry) (stop <-chan struct{}) {
	close(onlyOneSignalHandler) // panics when called twice

	stopCh := make(chan struct{})
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	go func(log *logrus.Entry) {
		<-ch
		log.Warn("Controller received interrupt. Shutting down...")
		close(stopCh)
		<-ch
		log.Warn("Force Closed.")
		os.Exit(1)
	}(log)

	return stopCh
}
