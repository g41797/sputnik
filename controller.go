package sputnik

var _ BlockController = &controller{}

type controller struct {
	bd  BlockDescriptor
	bl  *Block
	abs activeBlocks
	mpr *msgProcessor
}

func attachController(resp string, abs activeBlocks) {
	abl, _ := abs.getABl(resp)
	cn := new(controller)
	cn.bd = abl.bd
	cn.bl = abl.bl
	cn.mpr = newMsgProcessor(cn.bl.onMsg)
	cn.abs = abs
	abl.bc = cn
}

func (cn *controller) Controller(resp string) (bc BlockController, exists bool) {

	abl, exists := cn.abs.getABl(resp)

	if !exists {
		return nil, false
	}

	return abl.bc, true
}

func (cn *controller) Descriptor() BlockDescriptor {
	return cn.bd
}

func (cn *controller) Send(msg Msg) bool {
	if msg == nil {
		return false
	}

	if cn.bl.onMsg == nil {
		return false
	}
	sok := cn.mpr.submit(msg)

	return sok
}

func (cn *controller) ServerConnected(sc any, lp any) bool {
	if cn.bl.onConnect == nil {
		return false
	}

	go cn.bl.onConnect(sc, lp)

	return true
}

func (cn *controller) ServerDisconnected() bool {
	if cn.bl.onDisconnect == nil {
		return false
	}

	go cn.bl.onDisconnect()

	return true
}

func (cn *controller) Finish() {

	icn := cn.abs[0].bc
	resp := cn.bd.Responsibility

	// This message will be processed by initiator:
	fm := make(Msg)
	fm["__name"] = finishedMsg
	fm["__resp"] = resp

	// Dedicate goroutine for finish of the block
	go func(fn Finish, bc BlockController, m Msg) {
		fn(false)
		icn.Send(fm) // Send message to initiator about finished block
	}(cn.bl.finish, icn, fm)

	return
}
