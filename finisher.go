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
	RegisterBlockFactory(DefaultFinisherName, FinisherBlockFactory)
}

func FinisherBlockFactory() Block {
	finisher := new(finisher)
	block := Block{
		Init:   finisher.init,
		Run:    finisher.run,
		Finish: finisher.finish,

		OnMsg: finisher.debug,
	}
	return block
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
	signal.Notify(bl.term, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	defer signal.Reset(syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer close(bl.term)

	select {
	case <-bl.done:
		return

	case <-bl.term:
		ibc, _ := bl.bc.Controller(InitiatorResponsibility)
		ibc.Finish()
	}

	return
}

func (bl *finisher) finish(init bool) {
	if init {
		return
	}
	close(bl.done)
	return
}

// Any received message interpreted as SIGQUIT
// Used for testing.
func (bl *finisher) debug(msg Msg) {
	bl.term <- syscall.SIGQUIT
}
