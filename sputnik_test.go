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
	// This pattern meu be used in real application
	stop chan struct{}
	done chan struct{}
}

// Block factory:
func (tb *testBlocks) dbFact() sputnik.Block {
	db := new(dumbBlock)
	tb.dbl = append(tb.dbl, db)
	return sputnik.Block{
		Init:   db.init,
		Run:    db.run,
		Finish: db.finish,

		OnMsg:        db.eventReceived,
		OnConnect:    db.serverConnected,
		OnDisconnect: db.serverDisConnected,
	}
}

// dumbBlock support all callbacks of Block:
//
// Init
func (db *dumbBlock) init(cnf any) error {
	db.stop = make(chan struct{}, 1)
	return nil
}

// Run:
func (db *dumbBlock) run(bc sputnik.BlockController) {

	db.done = make(chan struct{})
	defer close(db.done)

	select {
	case <-db.stop:
		return
	}

	return
}

// Finish:
func (db *dumbBlock) finish(init bool) {
	close(db.stop) // Cancel Run

	if init {
		return
	}

	select {
	case <-db.done: // Wait finish of Run
		return
	}
	return
}

// OnServerConnect:
func (db *dumbBlock) serverConnected(connection any, logger any) {
	//Inform test about event
	m := make(sputnik.Msg)
	m["__name"] = "serverConnected"
	db.send(m)
	return
}

// OnServerDisconnect:
func (db *dumbBlock) serverDisConnected(connection any) {
	//Inform test about event
	m := make(sputnik.Msg)
	m["__name"] = "serverDisConnected"
	db.send(m)
	return
}

// OnMsg:
func (db *dumbBlock) eventReceived(msg sputnik.Msg) {
	//Inform test about event
	db.send(msg)
	return
}

func (db *dumbBlock) send(msg sputnik.Msg) {
	if db.q != nil {
		//Send message to test
		db.q.PutMT(msg)
	}
	return
}

// Spaceport of sputnik:
func dumbSpacePort(tb *testBlocks) sputnik.SpacePort {
	return sputnik.SpacePort{
		CnfFct:    dumbConf,
		AppBlocks: blkList,
		Finisher:  sputnik.BlockDescriptor{"finisher", "finisher"},
		BlkFact:   tb.factories(),
	}
}

func TestPrepare(t *testing.T) {

	tb := new(testBlocks)

	dsp := dumbSpacePort(tb)

	_, kill, err := sputnik.Prepare(dsp)

	if err != nil {
		t.Errorf("Prepare error %v", err)
	}

	kill()

	return
}

func TestFinisher(t *testing.T) {

	tb := new(testBlocks)

	dsp := dumbSpacePort(tb)

	launch, kill, err := sputnik.Prepare(dsp)

	if err != nil {
		t.Errorf("Prepare error %v", err)
	}

	tb.launch = launch
	tb.kill = kill

	tb.run()

	time.Sleep(1 * time.Second)

	tb.kill()

	<-tb.done

	return
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
	bc, exists := tb.dbl[0].bc.Controller((resp))

	if !exists {
		return false
	}

	return bc.Send(msg)
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

	var factList []string = []string{"dumb", "finisher"}

	for _, name := range factList {
		sputnik.RegisterBlockFactoryInner(name, tb.dbFact, res)
	}
	return res
}
