package sputnik

import (
	"fmt"
)

// BlockFactory should be provided for every block in the process
type BlockFactory func() *Block

type BlockFactories map[string]BlockFactory

// BlockFactory registered in the process via RegisterBlockFactory
// Please pay attention that panic called for any error during registration.
//
// Use init() for registration of BlockFactory
//
//	 func init() {
//			sputnik.RegisterBlockFactory("syslogPublisher", slpbFactory)
//		}
func RegisterBlockFactory(name string, bf BlockFactory) {
	if name == "" {
		panic("RegisterBlockFactory: empty block name")
	}
	if bf == nil {
		panic(fmt.Errorf("RegisterBlockFactory: nil block factory for %s", name))
	}

	if blockFactories == nil {
		blockFactories = make(map[string]BlockFactory)
	}

	err := RegisterBlockFactoryInner(name, bf, blockFactories)

	if err != nil {
		panic(err)
	}
}

func DefaultFactories() BlockFactories {
	if blockFactories == nil {
		blockFactories = make(map[string]BlockFactory)
	}
	return blockFactories
}

func RegisterBlockFactoryInner(name string, bf BlockFactory, facts BlockFactories) error {

	if _, ok := facts[name]; ok {
		return fmt.Errorf("RegisterBlockFactory: %s already registered", name)
	}
	facts[name] = bf
	return nil
}

func Factory(name string) (BlockFactory, error) {
	bfs := DefaultFactories()

	fct, exists := bfs[name]

	if !exists {
		return nil, fmt.Errorf("factory for %s does not exist", name)
	}
	return fct, nil
}

func (bfs BlockFactories) createByName(name string) (blk *Block, exist bool) {
	fct, ok := bfs[name]

	if !ok {
		return nil, false
	}

	bl := fct()

	return bl, true
}

func (bfs BlockFactories) createByDescr(bd *BlockDescriptor) (blk *Block, err error) {
	blk, ok := bfs.createByName(bd.Name)

	if !ok {
		return nil, fmt.Errorf("Creation of block [name: %s resp: %s] failed", bd.Name, bd.Responsibility)
	}

	return blk, nil
}

var blockFactories = make(map[string]BlockFactory)
