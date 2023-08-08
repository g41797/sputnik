package sputnik

import (
	"github.com/g41797/kissngoqueue"
)

const EchoBlockName = "echo"

// Echo block is used for debugging
// It replaces not implemented block
// Just assign required responsibility.
// During creation it also get Queue[Msg]
// - created by test
// - stored in factory
// Every retrieved message will be put in this queue.
// for further getting by test
type echo struct {
	q    *kissngoqueue.Queue[Msg]
	done chan struct{}
}

func (bl *echo) init(_ ConfFactory) error {
	bl.done = make(chan struct{})
	return nil
}

func (bl *echo) finish(init bool) {
	close(bl.done)
	return
}

func (bl *echo) run(_ BlockCommunicator) {
	defer bl.q.CancelMT()
	<-bl.done
	return
}

func (bl *echo) onMsg(msg Msg) {
	if bl.q != nil {
		bl.q.PutMT(msg)
	}
	return
}

type echoFactory struct {
	q *kissngoqueue.Queue[Msg]
}

func (ef *echoFactory) createEchoBlock() *Block {
	echo := new(echo)
	echo.q = ef.q
	block := NewBlock(
		WithInit(echo.init),
		WithRun(echo.run),
		WithFinish(echo.finish),
		WithOnMsg(echo.onMsg))
	return block

}

func EchoBlockFactory(q *kissngoqueue.Queue[Msg]) BlockFactory {
	ef := new(echoFactory)
	ef.q = q
	return ef.createEchoBlock
}
