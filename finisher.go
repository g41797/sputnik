package sputnik

import (
	"os"
	"os/signal"
	"syscall"
)

func FinisherDescriptor() BlockDescriptor {
	return BlockDescriptor{DefaultFinisherName, DefaultFinisherResponsibility}
}

func init() {
	RegisterBlockFactory(DefaultFinisherName, finisherBlockFactory)
}

func finisherBlockFactory() *Block {
	finisher := new(finisher)
	block := NewBlock(
		WithInit(finisher.init),
		WithRun(finisher.run),
		WithFinish(finisher.finish),
		WithOnMsg(finisher.debug))
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
