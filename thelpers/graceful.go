package thelpers

import (
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
)

type Graceful struct {
	InterruptSignal chan struct{}
	*errgroup.Group
}

func NewGraceful(interruptSignal chan struct{}) *Graceful {
	return &Graceful{
		InterruptSignal: interruptSignal,
		Group:           &errgroup.Group{},
	}
}

func (g *Graceful) Go(f func() error) {
	g.Group.Go(func() error {
		<-g.InterruptSignal
		return f()
	})
}

func (g *Graceful) GoNoErr(f func()) {
	g.Group.Go(func() error {
		<-g.InterruptSignal
		f()
		return nil
	})
}

func StopSignal() chan struct{} {
	stop := make(chan struct{})

	go func() {
		stopSig := make(chan os.Signal, 1)
		signal.Notify(stopSig, os.Interrupt, syscall.SIGTERM)

		<-stopSig
		close(stop)
	}()

	return stop
}
