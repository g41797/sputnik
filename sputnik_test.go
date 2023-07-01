package sputnik_test

import (
	"testing"

	"github.com/g41797/sputnik"
)

var factList []string = []string{"dumb", "finisher"}

var blkList []sputnik.BlockDescriptor = []sputnik.BlockDescriptor{
	{"dumb", "1"},
	{"dumb", "2"},
	{"dumb", "3"},
}

func factories() sputnik.BlockFactories {
	res := make(sputnik.BlockFactories)
	for _, name := range factList {
		sputnik.RegisterBlockFactoryInner(name, dbFact, res)
	}
	return res
}

func dumbConf() any { return nil }

func dumbSpacePort() sputnik.SpacePort {
	return sputnik.SpacePort{
		CnfFct:    dumbConf,
		AppBlocks: blkList,
		Finisher:  sputnik.BlockDescriptor{"finisher", "finisher"},
		BlkFact:   factories(),
	}
}

func TestPrepare(t *testing.T) {

	dsp := dumbSpacePort()

	_, _, err := sputnik.Prepare(dsp)

	if err != nil {
		t.Errorf("Prepare error %v", err)
	}

	return
}

type dumbBlock struct {
	bc   sputnik.BlockController
	done chan struct{}
}

func (db *dumbBlock) init(cnf any) error {
	db.done = make(chan struct{}, 1)
	return nil
}

func (db *dumbBlock) finish() {
	close(db.done)
	return
}

func (db *dumbBlock) run(bc sputnik.BlockController) {
	select {
	case <-db.done:
		return
	}
}

func (db *dumbBlock) eventReceived(msg sputnik.Msg) {
	msg = nil
}

func dbFact() sputnik.Block {
	db := new(dumbBlock)
	return sputnik.Block{
		Init:   db.init,
		Finish: db.finish,
		Run:    db.run,
		OnMsg:  db.eventReceived,
	}
}
