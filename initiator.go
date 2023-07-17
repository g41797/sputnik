package sputnik

import (
	"sync"

	"github.com/g41797/kissngoqueue"
)

type initiator struct {
	lock          sync.Mutex
	sputnik       Sputnik
	actBlks       activeBlocks
	q             *kissngoqueue.Queue[Msg]
	runStarted    bool
	abortStarted  bool
	finishStarted bool
	finishedBlks  int
	done          chan struct{}
}

// Factory of initiator:
func (inr *initiator) factory() *Block {
	return NewBlock(
		WithInit(inr.init),
		WithRun(inr.run),
		WithFinish(inr.finish),
		WithOnConnect(inr.serverConnected),
		WithOnDisConnect(inr.serverDisConnected),
		WithOnMsg(inr.msgReceived),
	)
}

func (inr *initiator) init(_ ServerConfiguration) error {

	appBlks, err := inr.sputnik.createActiveBlocks()
	if err != nil {
		return err
	}

	ibs := make(activeBlocks, 0)

	for _, abl := range appBlks {
		err = abl.init(inr.sputnik.cnfFact())
		if err != nil {
			break
		}
		ibs = append(ibs, abl)
	}

	if err != nil {
		for i := len(ibs) - 1; i >= 0; i-- {
			ibs[i].finish()
		}

		return err
	}

	inr.actBlks = make(activeBlocks, 0)
	inr.actBlks = append(inr.actBlks, inr.activeinitiator())
	inr.actBlks = append(inr.actBlks, ibs...)

	inr.addControllers()

	inr.q = kissngoqueue.NewQueue[Msg]()

	inr.setupConnector()

	return nil
}

func (inr *initiator) setupConnector() {
	connector := inr.sputnik.cnt
	to := inr.sputnik.to

	if connector == nil {
		return
	}

	setupMsg := make(Msg)

	setupMsg["__connector"] = connector
	setupMsg["__timeout"] = to

	cbl, _ := inr.actBlks.getABl(DefaultConnectorResponsibility)

	cbl.controller.Send(setupMsg)

	return
}

func (inr *initiator) run(_ BlockController) {

	inr.done = make(chan struct{})
	defer close(inr.done)

	if !inr.activate() {
		return
	}

	// Main loop
	for {
		nm, ok := inr.q.Get()
		if !ok { //all blocks finished
			break
		}
		if nm == nil {
			continue
		}
		inr.processMsg(nm)
	}

	return
}

func (inr *initiator) activate() bool {
	inr.lock.Lock()
	defer inr.lock.Unlock()

	if inr.abortStarted {
		return false
	}

	// Start active blacks on own goroutines
	for _, abl := range inr.actBlks[1:] {
		go func(fr Run, bc BlockController) {
			fr(bc)
		}(abl.block.run, abl.controller)
	}

	inr.runStarted = true
	return true
}

const (
	finishMsg   = "__finish"
	finishedMsg = "__finished"
)

func (inr *initiator) finish(init bool) {
	if init {
		return
	}

	m := make(Msg)
	m["__name"] = finishMsg
	inr.q.PutMT(m)
	return
}

func (inr *initiator) msgReceived(msg Msg) {
	inr.q.PutMT(msg)
	return
}

func (inr *initiator) serverConnected(connection ServerConnection) {
	for _, abl := range inr.actBlks[1:] {
		abl.controller.ServerConnected(connection)
	}
	return
}

func (inr *initiator) serverDisConnected() {
	for _, abl := range inr.actBlks[1:] {
		abl.controller.ServerDisconnected()
	}
	return
}

func (inr *initiator) runInternal() (err error) {

	inr.run(nil)

	return nil
}

// TODO Add timeout for abort/ShootDown
func (inr *initiator) abort() {

	if inr.finishBeforeLaunch() {
		return
	}

	inr.finish(false)

	select {
	case <-inr.done:
		return
	}
}

func (inr *initiator) finishBeforeLaunch() bool {
	inr.lock.Lock()
	defer inr.lock.Unlock()

	if inr.runStarted {
		return false
	}

	for i := len(inr.actBlks) - 1; i > 0; i-- {
		inr.actBlks[i].finish()
	}

	inr.abortStarted = true
	return true
}

func (inr *initiator) activeinitiator() *activeBlock {
	ibl := newActiveBlock(
		BlockDescriptor{InitiatorResponsibility, InitiatorResponsibility}, inr.factory())
	return &ibl
}

func (inr *initiator) addControllers() {
	for _, abl := range inr.actBlks {
		attachController(abl.descriptor.Responsibility, inr.actBlks)
	}
	return
}

func (inr *initiator) processMsg(m Msg) {

	mn, exists := m["__name"]

	if !exists {
		return
	}

	name, ok := mn.(string)

	if !ok {
		return
	}

	switch name {
	case finishMsg:
		inr.processFinish()
	case finishedMsg:
		inr.processFinished()
	}

	return
}

func (inr *initiator) processFinish() {
	if inr.finishStarted {
		return
	}
	inr.finishStarted = true

	for i := len(inr.actBlks) - 1; i > 0; i-- {
		inr.actBlks[i].controller.Finish()
	}

	inr.finishedBlks = 0
}

func (inr *initiator) processFinished() {
	inr.finishedBlks++
	if inr.finishedBlks == len(inr.actBlks)-1 {
		inr.q.CancelMT() // stop main loop
	}
	return
}
