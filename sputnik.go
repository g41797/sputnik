package sputnik

import (
	"fmt"
	"time"
)

// Configuration
type ServerConfiguration any

// Configuration Factory
type ConfFactory func(confName string, result any) error

type Sputnik struct {
	// Configuration factory
	cnfFact ConfFactory

	// Block descriptor of used finisher
	fbd BlockDescriptor

	// Application blocks
	// Order in the list defines order of creation and initialization
	appBlocks []BlockDescriptor

	// Block Factories of the process
	blkFacts BlockFactories

	// Server connector plug-in
	cnt ServerConnector

	// Timeout for connect/reconnect/check connection
	to time.Duration

	// Descriptor of used connector block
	cnd BlockDescriptor
}

type SputnikOption func(sp *Sputnik)

func WithConfFactory(cf ConfFactory) SputnikOption {
	return func(sp *Sputnik) {
		sp.cnfFact = cf
	}
}

func WithFinisher(fbd BlockDescriptor) SputnikOption {
	return func(sp *Sputnik) {
		sp.fbd = fbd
	}
}

func WithAppBlocks(appBlocks []BlockDescriptor) SputnikOption {
	return func(sp *Sputnik) {
		sp.appBlocks = appBlocks
	}
}

func WithBlockFactories(blkFacts BlockFactories) SputnikOption {
	return func(sp *Sputnik) {
		sp.blkFacts = blkFacts
	}
}

func WithConnector(cnt ServerConnector, to time.Duration) SputnikOption {
	return func(sp *Sputnik) {
		sp.cnt = cnt
		sp.to = to
	}
}

func (sp *Sputnik) isValid() bool {
	return sp.cnfFact != nil && sp.appBlocks != nil
}

func NewSputnik(opts ...SputnikOption) (*Sputnik, error) {
	sp := new(Sputnik)

	// Pre-sets
	WithFinisher(FinisherDescriptor())(sp)
	WithBlockFactories(DefaultFactories())(sp)

	sp.cnd = BlockDescriptor{DefaultConnectorName, DefaultConnectorResponsibility}

	for _, opt := range opts {
		opt(sp)
	}

	ok := sp.isValid()
	if !ok {
		return nil, fmt.Errorf("not enough options for sputnik creation")
	}

	return sp, nil
}

// sputnik launcher
type Launch func() error

// sputnik shooter
type ShootDown func()

// Creates and initializes all blocks.
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
func (sputnik Sputnik) Prepare() (lfn Launch, st ShootDown, err error) {

	inr := new(initiator)

	inr.sputnik = sputnik

	err = inr.init(nil)

	if err != nil {
		return nil, nil, err
	}

	return inr.runInternal, inr.abort, nil
}

func (sputnik *Sputnik) createActiveBlocks() (activeBlocks, error) {

	dscrs := make([]BlockDescriptor, 0)
	dscrs = append(dscrs, sputnik.fbd)

	if sputnik.cnt != nil {
		dscrs = append(dscrs, ConnectorDescriptor())
	}

	dscrs = append(dscrs, sputnik.appBlocks...)

	abls := make(activeBlocks, 0)

	for _, bd := range dscrs {
		abl, err := sputnik.createByDescr(bd)
		if err != nil {
			return nil, err
		}
		abls = append(abls, abl)
	}

	return abls, nil
}

func (sputnik *Sputnik) createByDescr(bd BlockDescriptor) (*activeBlock, error) {
	b, err := sputnik.blkFacts.createByDescr(&bd)

	if err != nil {
		return nil, err
	}

	if !b.isValid(sputnik.cnt != nil) {
		return nil, fmt.Errorf("invalid callbacks in block: name =  %s resp = %s", bd.Name, bd.Responsibility)
	}

	abl := newActiveBlock(bd, b)

	return &abl, nil
}
