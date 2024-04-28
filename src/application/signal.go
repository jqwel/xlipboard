package application

import (
	"os"
	"os/signal"
	"syscall"
)

func init() {
	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-signalChannel
		switch sig {
		case os.Interrupt:
			onExit()
			os.Exit(0)
		case syscall.SIGTERM:
			// handle SIGTERM
			onExit()
			os.Exit(0)
		}
	}()
}
