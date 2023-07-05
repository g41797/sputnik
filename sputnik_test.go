package sputnik_test

import (
	"testing"

	"github.com/g41797/kissngoqueue"
	"github.com/g41797/sputnik"
)

var factList []string = []string{"dumb", "finisher"}

var blkList []sputnik.BlockDescriptor = []sputnik.BlockDescriptor{
	{"dumb", "1"},
	{"dumb", "2"},
	{"dumb", "3"},
}

type testBlocks struct {
	dbl    []*dumbBlock
	q      *kissngoqueue.Queue[sputnik.Msg]
	launch sputnik.Launch
	kill   sputnik.ShootDown
}

func NewTestBlocks() *testBlocks {
	tb := new(testBlocks)
	tb.q = kissngoqueue.NewQueue[sputnik.Msg]()
	return tb
}

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

func (tb *testBlocks) sendTo(resp string, msg sputnik.Msg) bool {
	bc, exists := tb.dbl[0].bc.Controller((resp))

	if !exists {
		return false
	}

	return bc.Send(msg)
}

func (tb *testBlocks) factories() sputnik.BlockFactories {
	res := make(sputnik.BlockFactories)
	for _, name := range factList {
		sputnik.RegisterBlockFactoryInner(name, tb.dbFact, res)
	}
	return res
}

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

type dumbBlock struct {
	bc   sputnik.BlockController
	q    *kissngoqueue.Queue[sputnik.Msg]
	stop chan struct{}
	done chan struct{}
}

func (db *dumbBlock) init(cnf any) error {
	db.stop = make(chan struct{}, 1)
	return nil
}

func (db *dumbBlock) run(bc sputnik.BlockController) {

	db.done = make(chan struct{})
	defer close(db.done)

	select {
	case <-db.stop:
		return
	}
}

func (db *dumbBlock) finish(init bool) {
	close(db.stop)

	if init {
		return
	}

	select {
	case <-db.done:
		return
	}
}

func (db *dumbBlock) serverConnected(connection any, logger any) {
	m := make(sputnik.Msg)
	m["__name"] = "serverConnected"
	db.send(m)
}

func (db *dumbBlock) serverDisConnected(connection any) {
	m := make(sputnik.Msg)
	m["__name"] = "serverDisConnected"
	db.send(m)
}

func (db *dumbBlock) eventReceived(msg sputnik.Msg) {
	db.send(msg)
}

func (db *dumbBlock) send(msg sputnik.Msg) {
	if db.q != nil {
		db.q.PutMT(msg)
	}
}

func dumbConf() any { return nil }

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

	tb.kill()

	return
}
