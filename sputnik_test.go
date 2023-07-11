package sputnik_test

import (
	"testing"
	"time"

	"github.com/g41797/kissngoqueue"
	"github.com/g41797/sputnik"
)

// Configuration factory:
func dumbConf() any { return nil }

// Satellite has 3 blocks:
var blkList []sputnik.BlockDescriptor = []sputnik.BlockDescriptor{
	{"dumb", "1"},
	{"dumb", "2"},
	{"dumb", "3"},
}

// Satellite block
type dumbBlock struct {
	// Block controller
	bc sputnik.BlockController
	// Main queue of test
	q *kissngoqueue.Queue[sputnik.Msg]
	// Used for synchronization
	// between finish and run
	// This pattern may be used in real application
	stop chan struct{}
	done chan struct{}
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

// dumbBlock support all callbacks of Block:
//
// Init
func (dmb *dumbBlock) init(cnf any) error {
	dmb.stop = make(chan struct{}, 1)
	return nil
}

// Run:
func (dmb *dumbBlock) run(bc sputnik.BlockController) {

	// Save controller for further communication
	// with blocks
	dmb.bc = bc

	dmb.done = make(chan struct{})
	defer close(dmb.done)

	// select isn't required for one channel
	// in real application you can add "listening"
	// on another channels here e.g. timeouts or
	// redirected OnMsg|OnConnect|etc
	select {
	case <-dmb.stop:
		return
	}

	return
}

// Finish:
func (dmb *dumbBlock) finish(init bool) {
	close(dmb.stop) // Cancel Run

	if init {
		return
	}

	select {
	case <-dmb.done: // Wait finish of Run
		return
	}
	return
}

// OnServerConnect:
func (dmb *dumbBlock) serverConnected(connection any) {
	//Inform test about event
	m := make(sputnik.Msg)
	m["__name"] = "serverConnected"
	dmb.send(m)
	return
}

// OnServerDisconnect:
func (dmb *dumbBlock) serverDisConnected() {
	//Inform test about event
	m := make(sputnik.Msg)
	m["__name"] = "serverDisConnected"
	dmb.send(m)
	return
}

// OnMsg:
func (dmb *dumbBlock) eventReceived(msg sputnik.Msg) {
	//Inform test about event
	dmb.send(msg)
	return
}

func (dmb *dumbBlock) send(msg sputnik.Msg) {
	if dmb.q != nil {
		//Send message to test
		dmb.q.PutMT(msg)
	}
	return
}

func dumbSputnik(tb *testBlocks) sputnik.Sputnik {
	sp, _ := sputnik.NewSputnik(
		sputnik.WithConfFactory(dumbConf),
		sputnik.WithAppBlocks(blkList),
		sputnik.WithBlockFactories(tb.factories()),
	)
	return *sp
}

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
}

func NewTestBlocks() *testBlocks {
	tb := new(testBlocks)
	tb.q = kissngoqueue.NewQueue[sputnik.Msg]()
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

	factList := []struct {
		name string
		fact sputnik.BlockFactory
	}{
		{"dumb", tb.dbFact},
		{"finisher", finfct},
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

func TestPrepare(t *testing.T) {

	tb := NewTestBlocks()

	dsp := dumbSputnik(tb)

	_, kill, err := dsp.Prepare()

	if err != nil {
		t.Errorf("Prepare error %v", err)
	}

	tb.attachQueue()

	kill()

	return
}

func TestFinisher(t *testing.T) {

	tb := NewTestBlocks()

	dsp := dumbSputnik(tb)

	launch, kill, err := dsp.Prepare()

	if err != nil {
		t.Errorf("Prepare error %v", err)
	}

	tb.attachQueue()
	tb.launch = launch
	tb.kill = kill

	tb.run()

	time.Sleep(1 * time.Second)

	// Simulate SIGINT
	ok := tb.sendTo("finisher", make(sputnik.Msg))
	if !ok {
		t.Errorf("send to finisher failed")
	}

	tb.kill()

	<-tb.done

	return
}

func TestRun(t *testing.T) {

	tb := NewTestBlocks()

	dsp := dumbSputnik(tb)

	launch, kill, err := dsp.Prepare()

	if err != nil {
		t.Errorf("Prepare error %v", err)
	}

	tb.attachQueue()
	tb.launch = launch
	tb.kill = kill

	tb.run()

	time.Sleep(1 * time.Second)

	// Simulate ServerConnect
	tb.mainCntrl().ServerConnected(nil)
	if !tb.expect(3, "serverConnected") {
		t.Errorf("Wrong processing of serverconnected")
	}

	// Simulate ServerDisconnect
	tb.mainCntrl().ServerDisconnected()
	if !tb.expect(3, "serverDisConnected") {
		t.Errorf("Wrong processing of serverDisConnected")
	}

	tb.kill()

	<-tb.done

	return
}
