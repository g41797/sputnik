package sputnik

type sputnik struct {
	spp SpacePort
	abs activeBlocks
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
	return
}

func (sp *sputnik) serverDisConnected(connection any) {
	return
}

func (sp *sputnik) eventReceived(ev Event) {
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

	if err == nil {
		sp.abs = make(activeBlocks, 0)
		sp.abs = append(sp.abs, sp.activeinitiator())
		sp.abs = append(sp.abs, ibs...)

		// TODO add block controllers !!!

		return nil
	}

	for i := len(ibs) - 1; i >= 0; i-- {
		ibs[i].finish()
	}

	return err
}

func (sp *sputnik) runInternal() (err error) {

	for {

		break
	}

	return
}

func (sp *sputnik) finishInternal() {
	return
}

func (sp *sputnik) factory() Block {
	return Block{
		Init:   sp.init,
		Run:    sp.run,
		Finish: sp.finish,

		OnConnect:    sp.serverConnected,
		OnDisconnect: sp.serverDisConnected,
		OnEvent:      sp.eventReceived,
	}
}

func (sp *sputnik) activeinitiator() activeBlock {
	return newActiveBlock(
		BlockDescriptor{InitiatorResponsibility, InitiatorResponsibility}, sp.factory())
}
