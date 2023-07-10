package sputnik

type Sputnik struct {
	// Configuration factory
	CnfFct func() any

	// Block descriptor of used finisher
	Finisher BlockDescriptor

	// Application blocks
	// Order in the list defines order of creation and initialization
	AppBlocks []BlockDescriptor

	// Block Factories of the process
	BlkFact BlockFactories
}

func NewSputnik(cnfn func() any, appBlocks []BlockDescriptor) Sputnik {
	return Sputnik{
		CnfFct:    cnfn,
		AppBlocks: appBlocks,
		Finisher:  FinisherDescriptor(),
		BlkFact:   DefaultFactories(),
	}
}

// sputnik launcher
type Launch func() error

// sputnik shooter
type ShootDown func()

// Prepare sputnik for launch
//
// If creation and initialization of any block failed:
//
//   - Finish is called on all already initialized blocks
//
//   - Order of finish - reversal of initialization
//
//     = Returned error describes reason of the failure
//
// Otherwise returned 2 functions for sputnik management:
//
//   - lfn - Launch of the sputnik , exit from this function will be
//     after signal for shutdown of the process  or after call of
//     second returned function (see below)
//
//   - st - ShootDown of sputnik - abort flight
func Prepare(spk Sputnik) (lfn Launch, st ShootDown, err error) {

	inr := new(initiator)

	inr.spk = spk

	err = inr.init(nil)

	if err != nil {
		return nil, nil, err
	}

	return inr.runInternal, inr.abort, nil
}

func (spk *Sputnik) createActiveBlocks() (activeBlocks, error) {

	dscrs := make([]BlockDescriptor, 0)
	dscrs = append(dscrs, spk.Finisher)
	dscrs = append(dscrs, spk.AppBlocks...)

	abls := make(activeBlocks, 0)

	for _, bd := range dscrs {
		abl, err := spk.createByDescr(bd)
		if err != nil {
			return nil, err
		}
		abls = append(abls, abl)
	}

	return abls, nil
}

func (spk *Sputnik) createByDescr(bd BlockDescriptor) (*activeBlock, error) {
	b, err := spk.BlkFact.createByDescr(&bd)

	if err != nil {
		return nil, err
	}

	abl := newActiveBlock(bd, b)

	return &abl, nil
}
