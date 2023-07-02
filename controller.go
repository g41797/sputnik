package sputnik

var _ BlockController = &controller{}

type controller struct {
	bd  BlockDescriptor
	bl  Block
	abs activeBlocks
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

	//TODO replace with kissngoqueue
	go cn.bl.OnMsg(msg)

	return true
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
	go cn.bl.Finish()
}
