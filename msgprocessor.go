package sputnik

import (
	"sync"

	"github.com/g41797/kissngoqueue"
)

// Helper of controller. All messages send to block
// are processed using queue on the same goroutine.
type msgProcessor struct {
	fnc  OnMsg
	q    *kissngoqueue.Queue[Msg]
	once sync.Once
}

func newMsgProcessor(fnc OnMsg) *msgProcessor {
	pr := msgProcessor{
		fnc: fnc,
		q:   kissngoqueue.NewQueue[Msg](),
	}
	return &pr
}

func (pr *msgProcessor) submit(msg Msg) (cancelled bool) {
	pr.once.Do(func() { go pr.process() })
	return !pr.q.PutMT(msg)
}

func (pr *msgProcessor) cancel() {
	pr.q.CancelMT()
}

func (pr *msgProcessor) process() {
	msg, ok := pr.q.Get()
	if !ok {
		return
	}

	pr.fnc(msg)

}
