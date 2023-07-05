package sputnik

import (
	"sync"

	"github.com/g41797/kissngoqueue"
)

type sputnik struct {
	lock          sync.Mutex
	spp           SpacePort
	abs           activeBlocks
	q             *kissngoqueue.Queue[Msg]
	runStarted    bool
	abortStarted  bool
	finishStarted bool
	finishedBlks  int
	done          chan struct{}
}

// sputnik has functionality of "initiator" block:

// Factory of initiator:
func (sp *sputnik) factory() Block {
	return Block{
		Init:   sp.init,
		Run:    sp.run,
		Finish: sp.finish,

		OnConnect:    sp.serverConnected,
		OnDisconnect: sp.serverDisConnected,
		OnMsg:        sp.msgReceived,
	}
}

// Callback of "initiator" block
func (sp *sputnik) init(_ any) error {

	appBlks, err := sp.spp.createActiveBlocks()
	if err != nil {
		return err
	}

	ibs := make(activeBlocks, 0)

	for _, abl := range appBlks {
		err = abl.init(sp.spp.CnfFct())
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

	sp.abs = make(activeBlocks, 0)
	sp.abs = append(sp.abs, sp.activeinitiator())
	sp.abs = append(sp.abs, ibs...)

	sp.addControllers()

	sp.q = kissngoqueue.NewQueue[Msg]()

	return nil
}

func (sp *sputnik) run(_ BlockController) {

	if !sp.activate() {
		return
	}

	// Main loop
	for {
		nm, ok := sp.q.Get()
		if !ok { //all blocks finished
			break
		}
		if nm == nil {
			continue
		}
		sp.processMsg(nm)
	}

	return
}

func (sp *sputnik) activate() bool {
	sp.lock.Lock()
	defer sp.lock.Unlock()

	if sp.abortStarted {
		return false
	}

	sp.done = make(chan struct{})
	defer close(sp.done)

	// Start active blacks on own goroutines
	for _, abl := range sp.abs[1:] {
		go func(fr Run, bc BlockController) {
			fr(bc)
		}(abl.bl.Run, abl.bc)
	}

	sp.runStarted = true
	return true
}

const (
	finishMsg   = "__finish"
	finishedMsg = "__finished"
)

func (sp *sputnik) finish(init bool) {
	if init {
		return
	}

	m := make(Msg)
	m["__name"] = finishMsg
	sp.q.PutMT(m)
}

func (sp *sputnik) msgReceived(msg Msg) {
	sp.q.PutMT(msg)
}

func (sp *sputnik) serverConnected(connection any, logger any) {
	for _, abl := range sp.abs[1:] {
		abl.bc.ServerConnected(connection, logger)
	}
}

func (sp *sputnik) serverDisConnected(connection any) {
	for _, abl := range sp.abs[1:] {
		abl.bc.ServerDisconnected(connection)
	}
}

func (sp *sputnik) runInternal() (err error) {

	sp.run(nil)

	return nil
}

// TODO Add timeout for abort/ShootDown
func (sp *sputnik) abort() {

	if sp.finishBeforeLaunch() {
		return
	}

	sp.finish(false)

	select {
	case <-sp.done:
		return
	}
}

func (sp *sputnik) finishBeforeLaunch() bool {
	sp.lock.Lock()
	defer sp.lock.Unlock()

	if sp.runStarted {
		return false
	}

	for i := len(sp.abs) - 1; i >= 0; i-- {
		sp.abs[i].finish()
	}

	sp.abortStarted = true
	return true
}

func (sp *sputnik) activeinitiator() *activeBlock {
	ibl := newActiveBlock(
		BlockDescriptor{InitiatorResponsibility, InitiatorResponsibility}, sp.factory())
	return &ibl
}

func (sp *sputnik) addControllers() {
	for _, abl := range sp.abs {
		attachController(abl, sp.abs)
	}
}

func (sp *sputnik) processMsg(m Msg) {

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
		sp.processFinish()
	case finishedMsg:
		sp.processFinished()
	}
}

func (sp *sputnik) processFinish() {
	if sp.finishStarted {
		return
	}
	sp.finishStarted = true

	for i := len(sp.abs) - 1; i >= 0; i-- {
		sp.abs[i].bc.Finish()
	}

	sp.finishedBlks = 0
}

func (sp *sputnik) processFinished() {
	sp.finishedBlks++
	if sp.finishedBlks == len(sp.abs)-1 {
		sp.q.CancelMT() // stop main loop
	}
}
