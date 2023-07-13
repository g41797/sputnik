package sputnik_test

import (
	"time"

	"github.com/g41797/kissngoqueue"
	"github.com/g41797/sputnik"
)

// Test helper:
type testBlocks struct {
	// All blocks
	dbl []*dumbBlock
	// Test queue
	q *kissngoqueue.Queue[sputnik.Msg]
	// Launcher
	launch sputnik.Launch
	// ShootDown
	kill sputnik.ShootDown
	// Signalling channel
	done chan struct{}
	// Connector
	conntr dummyConnector
	to     time.Duration
}

func NewTestBlocks() *testBlocks {
	tb := new(testBlocks)
	tb.q = kissngoqueue.NewQueue[sputnik.Msg]()
	tb.conntr = dummyConnector{}
	tb.to = time.Millisecond * 100
	return tb
}

// Expectation:
// - get n messages from blocks
// - with "__name" == <name>
func (tb *testBlocks) expect(n int, name string) bool {
	for i := 0; i < n; i++ {
		msg, ok := tb.q.Get()
		if !ok {
			return false
		}

		mn, exists := msg["__name"]

		if !exists {
			return false
		}

		mname, ok := mn.(string)

		if !ok {
			return false
		}

		if mname != name {
			return false
		}

	}

	return true
}

// Send msg to block using it's responsibility
// Use this pattern in real application for
// negotiation between blocks
func (tb *testBlocks) sendTo(resp string, msg sputnik.Msg) bool {
	cn := tb.dbl[0].bc
	bc, exists := cn.Controller(resp)

	if !exists {
		return false
	}
	sok := bc.Send(msg)
	return sok
}

func (tb *testBlocks) mainCntrl() sputnik.BlockController {
	mcn, _ := tb.dbl[0].bc.Controller(sputnik.InitiatorResponsibility)
	return mcn
}

// Run Launcher on dedicated goroutine
// Test controls execution via sputnik API
// Results received using queue
func (tb *testBlocks) run() {
	tb.done = make(chan struct{})

	go func(l sputnik.Launch, done chan struct{}) {
		defer close(done)
		l()
	}(tb.launch, tb.done)

	return
}

// Registration of factories for test environment
// For this case init() isn't used
// use this pattern for the case when you don't need
// dynamic registration: all blocks (and factories) are
// known in advance.
func (tb *testBlocks) factories() sputnik.BlockFactories {
	res := make(sputnik.BlockFactories)

	finfct, _ := sputnik.Factory(sputnik.DefaultFinisherName)
	confct, _ := sputnik.Factory(sputnik.DefaultConnectorName)

	factList := []struct {
		name string
		fact sputnik.BlockFactory
	}{
		{"dumb", tb.dbFact},
		{sputnik.DefaultFinisherName, finfct},
		{sputnik.DefaultConnectorName, confct},
	}

	for _, fd := range factList {
		sputnik.RegisterBlockFactoryInner(fd.name, fd.fact, res)
	}
	return res
}

func (tb *testBlocks) attachQueue() {
	for i, _ := range tb.dbl {
		tb.dbl[i].q = tb.q
	}
	return
}

// Block factory:
func (tb *testBlocks) dbFact() *sputnik.Block {
	dmb := new(dumbBlock)
	tb.dbl = append(tb.dbl, dmb)
	return sputnik.NewBlock(
		sputnik.WithInit(dmb.init),
		sputnik.WithRun(dmb.run),
		sputnik.WithFinish(dmb.finish),

		sputnik.WithOnMsg(dmb.eventReceived),
		sputnik.WithOnConnect(dmb.serverConnected),
		sputnik.WithOnDisConnect(dmb.serverDisConnected),
	)
}

// Satellite has 3 app. blocks:
var blkList []sputnik.BlockDescriptor = []sputnik.BlockDescriptor{
	{"dumb", "1"},
	{"dumb", "2"},
	{"dumb", "3"},
}

// Configuration factory:
func dumbConf() any { return nil }

func dumbSputnik(tb *testBlocks) sputnik.Sputnik {
	sp, _ := sputnik.NewSputnik(
		sputnik.WithConfFactory(dumbConf),
		sputnik.WithAppBlocks(blkList),
		sputnik.WithBlockFactories(tb.factories()),
		sputnik.WithConnector(&tb.conntr, tb.to),
	)
	return *sp
}
