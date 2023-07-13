package sputnik_test

import (
	"github.com/g41797/kissngoqueue"
	"github.com/g41797/sputnik"
)

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
