package sputnik

var _ BlockController = &controller{}

type controller struct {
	bd  BlockDescriptor
	bl  Block
	abs activeBlocks
	mpr *msgProcessor
}

func attachController(abl *activeBlock, abs activeBlocks) {
	cn := new(controller)
	cn.bd = abl.bd
	cn.bl = abl.bl
	cn.mpr = newMsgProcessor(cn.bl.OnMsg)
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
	if cn.bl.OnMsg == nil {
		return false
	}

	return cn.mpr.submit(msg)
}

func (cn *controller) ServerConnected(sc any, lp any) bool {
	if cn.bl.OnConnect == nil {
		return false
	}

	go cn.bl.OnConnect(sc, lp)

	return true
}

func (cn *controller) ServerDisconnected(sc any) bool {
	if cn.bl.OnDisconnect == nil {
		return false
	}

	go cn.bl.OnDisconnect(sc)

	return true
}

func (cn *controller) Finish() {

	icn := cn.abs[0].bc

	fm := make(Msg)
	fm["__name"] = finishedMsg
	fm["__resp"] = cn.bd.Responsibility

	go func(fn Finish, bc BlockController, m Msg) {
		fn(false)
		icn.Send(fm)
	}(cn.bl.Finish, icn, fm)
}
