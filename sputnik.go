package sputnik

type sputnik struct {
	spp SpacePort
	abs activeBlocks
}

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

func (sp *sputnik) init(cnf any) error {
	return nil
}

func (sp *sputnik) run(ibc BlockController) {
	return
}

func (sp *sputnik) finish() {
	return
}

func (sp *sputnik) serverConnected(connection any, logger any) {
	for _, abl := range sp.abs[1:] {
		if abl.bc != nil {
			abl.bc.ServerConnected(connection, logger)
			continue
		}
	}
}

func (sp *sputnik) serverDisConnected(connection any) {

	for _, abl := range sp.abs[1:] {
		if abl.bc != nil {
			abl.bc.ServerDisconnected(connection)
			continue
		}
	}
}

func (sp *sputnik) msgReceived(msg Msg) {
	return
}

func (sp *sputnik) initInternal() error {

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

	return nil
}

func (sp *sputnik) runInternal() (err error) {

	for _, abl := range sp.abs {
		go func(fr Run, bc BlockController) {
			fr(bc)
		}(abl.bl.Run, abl.bc)
	}

	return nil
}

func (sp *sputnik) finishInternal() {
	return
}

func (sp *sputnik) abort() {
	return
}

func (sp *sputnik) activeinitiator() *activeBlock {
	ibl := newActiveBlock(
		BlockDescriptor{InitiatorResponsibility, InitiatorResponsibility}, sp.factory())
	return &ibl
}

func (sp *sputnik) addControllers() {
	for _, abl := range sp.abs {
		cn := new(controller)
		cn.bd = abl.bd
		cn.bl = abl.bl
		cn.abs = sp.abs
		abl.bc = cn
	}
}
