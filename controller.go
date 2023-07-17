package sputnik

var _ BlockController = &controller{}

type controller struct {
	descriptor BlockDescriptor
	block      *Block
	actBlks    activeBlocks
	mpr        *msgProcessor
}

func attachController(resp string, actBlks activeBlocks) {
	abl, _ := actBlks.getABl(resp)
	cn := new(controller)
	cn.descriptor = abl.descriptor
	cn.block = abl.block
	cn.mpr = newMsgProcessor(cn.block.onMsg)
	cn.actBlks = actBlks
	abl.controller = cn
}

func (cn *controller) Controller(resp string) (bc BlockController, exists bool) {

	abl, exists := cn.actBlks.getABl(resp)

	if !exists {
		return nil, false
	}

	return abl.controller, true
}

func (cn *controller) Descriptor() BlockDescriptor {
	return cn.descriptor
}

func (cn *controller) Send(msg Msg) bool {
	if msg == nil {
		return false
	}

	if cn.block.onMsg == nil {
		return false
	}
	sok := cn.mpr.submit(msg)

	return sok
}

func (cn *controller) ServerConnected(sc ServerConnection) bool {
	if cn.block.onConnect == nil {
		return false
	}

	go cn.block.onConnect(sc)

	return true
}

func (cn *controller) ServerDisconnected() bool {
	if cn.block.onDisconnect == nil {
		return false
	}

	go cn.block.onDisconnect()

	return true
}

func (cn *controller) Finish() {

	icn := cn.actBlks[0].controller
	resp := cn.descriptor.Responsibility

	// This message will be processed by initiator:
	fm := make(Msg)
	fm["__name"] = finishedMsg
	fm["__resp"] = resp

	// Dedicate goroutine for finish of the block
	go func(fn Finish, bc BlockController, m Msg, pr *msgProcessor) {
		fn(false)
		if pr != nil {
			pr.cancel()
		}
		icn.Send(fm) // Send message to initiator about finished block
	}(cn.block.finish, icn, fm, cn.mpr)

	return
}
