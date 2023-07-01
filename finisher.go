package sputnik

import (
	"os"
	"os/signal"
	"syscall"
)

const DefaultFinisherName = "finisher"
const DefaultFinisherResponsibility = "finisher"

func FinisherDescriptor() BlockDescriptor {
	return BlockDescriptor{DefaultFinisherName, DefaultFinisherResponsibility}
}

func init() {
	RegisterBlockFactory(DefaultFinisherName,
		func() Block {
			finisher := new(finisher)
			block := Block{
				Init:   finisher.init,
				Run:    finisher.run,
				Finish: finisher.finish,
			}
			return block
		})
}

type finisher struct {
	done chan struct{}
	term chan os.Signal

	bc BlockController
}

func (bl *finisher) init(conf any) error {
	return nil
}

func (bl *finisher) run(self BlockController) {
	bl.bc = self

	bl.done = make(chan struct{})

	bl.term = make(chan os.Signal, 3)
	defer close(bl.term)

	signal.Notify(bl.term, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case <-bl.done:
		return

	case <-bl.term:
		if bl.bc != nil {
			ibc, ok := bl.bc.Controller(InitiatorResponsibility)
			if ok {
				ibc.Finish()
			}
		}
	}
}

func (bl *finisher) finish() {
	close(bl.done)
	return
}

// TODO: check context usage + signal.NotifyContext ...
