package sputnik

import (
	"fmt"
	"time"
)

// Configuration Factory
type ConfFactory func() any

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
	cnt Connector

	// Timeout for connect/reconnect/check connection
	to time.Duration

	// Block descriptor of used connector
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

func WithConnector(cnt Connector, to time.Duration) SputnikOption {
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
func (spk Sputnik) Prepare() (lfn Launch, st ShootDown, err error) {

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
	dscrs = append(dscrs, spk.fbd)

	if spk.cnt != nil {
		dscrs = append(dscrs, ConnectorDescriptor())
	}

	dscrs = append(dscrs, spk.appBlocks...)

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
	b, err := spk.blkFacts.createByDescr(&bd)

	if err != nil {
		return nil, err
	}

	if !b.isValid(spk.cnt != nil) {
		return nil, fmt.Errorf("invalid callbacks in block: name =  %s resp = %s", bd.Name, bd.Responsibility)
	}

	abl := newActiveBlock(bd, b)

	return &abl, nil
}
